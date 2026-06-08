package socks5

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
	"multi-protocol-proxy/internal/auth"
	"multi-protocol-proxy/internal/bandwidth"
	"multi-protocol-proxy/internal/config"
	"multi-protocol-proxy/internal/ui"
)

const (
	Version5 = 0x05

	MethodNoAuth       = 0x00
	MethodUserPass     = 0x02
	MethodNoAcceptable = 0xFF

	UserPassVersion = 0x01

	CmdConnect = 0x01
	CmdBind    = 0x02
	CmdUDP     = 0x03

	AddrTypeIPv4   = 0x01
	AddrTypeDomain = 0x03
	AddrTypeIPv6   = 0x04

	ReplySucceeded          = 0x00
	ReplyGeneralFailure     = 0x01
	ReplyConnectionNotAllowed = 0x02
	ReplyNetworkUnreachable = 0x03
	ReplyHostUnreachable    = 0x04
	ReplyConnectionRefused  = 0x05
	ReplyTTLExpired         = 0x06
	ReplyCmdNotSupported    = 0x07
	ReplyAddrTypeNotSupported = 0x08
)

type Server struct {
	Config    *config.Config
	UserStore *auth.UserStore
	Bandwidth *bandwidth.Tracker

	ln       net.Listener
	wg       sync.WaitGroup
	shutdown chan struct{}
}

func NewServer(cfg *config.Config, userStore *auth.UserStore, bw *bandwidth.Tracker) *Server {
	return &Server{
		Config:    cfg,
		UserStore: userStore,
		Bandwidth: bw,
		shutdown:  make(chan struct{}),
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := s.Config.Env.SOCKS5Port
	if addr == "" {
		addr = ":1080"
	}

	var err error
	s.ln, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	ui.LogStatus("info", "SOCKS5 Proxy listening on "+addr)

	go s.watchShutdown(ctx)

	for {
		select {
		case <-s.shutdown:
			return nil
		default:
		}

		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return nil
			default:
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return err
			}
		}

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.handleConnection(ctx, c)
		}(conn)
	}
}

func (s *Server) watchShutdown(ctx context.Context) {
	<-ctx.Done()
	close(s.shutdown)
	if s.ln != nil {
		s.ln.Close()
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	startTime := time.Now()
	clientIP := conn.RemoteAddr().String()

	if !s.UserStore.CheckIPAllowed(clientIP) {
		MetricAuthFailures.WithLabelValues("ip_blocked").Inc()
		ui.LogStatus("warn", "SOCKS5 IP blocked: "+clientIP)
		return
	}

	MetricActiveConns.Inc()
	defer MetricActiveConns.Dec()

	conn.SetDeadline(time.Now().Add(30 * time.Second))

	var username string
	var err error
	username, err = s.handleMethodNegotiation(conn)
	if err != nil {
		ui.LogStatus("error", "SOCKS5 method negotiation failed: "+err.Error())
		return
	}

	isSuperAdmin := false
	user := s.UserStore.GetUser(username)
	if user != nil && user.Role == "super_admin" {
		if _, ok := s.UserStore.IsSuperAdminIP(clientIP); ok {
			isSuperAdmin = true
			ui.LogStatus("info", "SOCKS5 super_admin verified: "+username+" from "+clientIP)
		}
	}

	if !isSuperAdmin {
		if !s.UserStore.CheckRateLimit(username) {
			MetricRateLimited.WithLabelValues(username).Inc()
			ui.LogStatus("warn", "SOCKS5 rate limited: "+username)
			return
		}
	}

	if !isSuperAdmin && user != nil {
		if !s.UserStore.CheckExpiry(username) {
			ui.LogStatus("warn", "SOCKS5 account expired: "+username)
			return
		}

		if s.Bandwidth != nil && !s.Bandwidth.CheckAllowance(username, user.BandwidthLimitGB) {
			ui.LogStatus("warn", "SOCKS5 bandwidth exceeded: "+username)
			return
		}

		if s.Bandwidth != nil && !s.Bandwidth.CheckConnLimit(username, user.MaxConnections) {
			ui.LogStatus("warn", "SOCKS5 connection limit reached: "+username)
			return
		}
	}

	if s.Bandwidth != nil {
		s.Bandwidth.IncrementConns(username)
		defer s.Bandwidth.DecrementConns(username)
	}

	targetAddr, err := s.handleRequest(conn)
	if err != nil {
		ui.LogStatus("error", "SOCKS5 request failed: "+err.Error())
		return
	}

	targetConn, err := net.DialTimeout("tcp", targetAddr, 30*time.Second)
	if err != nil {
		s.sendReply(conn, ReplyHostUnreachable, nil)
		MetricErrors.WithLabelValues("dial_failed").Inc()
		return
	}
	defer targetConn.Close()

	localAddr := targetConn.LocalAddr().(*net.TCPAddr)
	s.sendReply(conn, ReplySucceeded, localAddr)

	MetricConnections.WithLabelValues(username).Inc()

	conn.SetDeadline(time.Time{})
	targetConn.SetDeadline(time.Time{})

	var relayClient, relayTarget net.Conn
	relayClient = conn
	relayTarget = targetConn
	if user != nil && user.BandwidthSpeedMbps > 0 {
		relayClient = bandwidth.NewThrottledConn(conn, user.BandwidthSpeedMbps).(*bandwidth.ThrottledConn)
		relayTarget = bandwidth.NewThrottledConn(targetConn, user.BandwidthSpeedMbps).(*bandwidth.ThrottledConn)
	}

	var upBytes, downBytes int64
	done := make(chan struct{}, 2)

	go func() {
		n, _ := io.Copy(relayTarget, relayClient)
		upBytes = n
		done <- struct{}{}
	}()

	go func() {
		n, _ := io.Copy(relayClient, relayTarget)
		downBytes = n
		done <- struct{}{}
	}()


	<-done

	duration := time.Since(startTime).Seconds()
	MetricBytes.WithLabelValues(username, "upstream").Add(float64(upBytes))
	MetricBytes.WithLabelValues(username, "downstream").Add(float64(downBytes))
	MetricDuration.Observe(duration)

	if s.Bandwidth != nil {
		s.Bandwidth.RecordBytes(username, upBytes, downBytes)
	}
}

func (s *Server) handleMethodNegotiation(conn net.Conn) (string, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", err
	}

	if buf[0] != Version5 {
		return "", errors.New("unsupported SOCKS version")
	}

	numMethods := int(buf[1])
	methods := make([]byte, numMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return "", err
	}

	hasUserPass := false
	for _, method := range methods {
		if method == MethodUserPass {
			hasUserPass = true
			break
		}
	}

	if !hasUserPass {
		conn.Write([]byte{Version5, MethodNoAcceptable})
		MetricAuthFailures.WithLabelValues("no_auth_method").Inc()
		return "", errors.New("no acceptable auth method")
	}

	
	conn.Write([]byte{Version5, MethodUserPass})


	return s.authenticateUser(conn)
}

func (s *Server) handleMethodNegotiationNoAuth(conn net.Conn) (string, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", err
	}

	if buf[0] != Version5 {
		return "", errors.New("unsupported SOCKS version")
	}

	numMethods := int(buf[1])
	methods := make([]byte, numMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return "", err
	}

	conn.Write([]byte{Version5, MethodNoAuth})
	return "", nil 

func (s *Server) authenticateUser(conn net.Conn) (string, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", err
	}

	if buf[0] != UserPassVersion {
		return "", errors.New("unsupported auth version")
	}

	
	usernameLen := int(buf[1])
	username := make([]byte, usernameLen)
	if _, err := io.ReadFull(conn, username); err != nil {
		return "", err
	}


	if _, err := io.ReadFull(conn, buf[:1]); err != nil {
		return "", err
	}

	passwordLen := int(buf[0])
	password := make([]byte, passwordLen)
	if _, err := io.ReadFull(conn, password); err != nil {
		return "", err
	}


	_, valid := s.UserStore.ValidateCredentials(string(username), string(password))
	if !valid {
		conn.Write([]byte{UserPassVersion, 0x01})
		MetricAuthFailures.WithLabelValues("invalid_credentials").Inc()
		ui.LogStatus("warn", "SOCKS5 auth failed for: "+string(username))
		return "", errors.New("authentication failed")
	}

	conn.Write([]byte{UserPassVersion, 0x00})
	return string(username), nil
}

func (s *Server) handleRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", err
	}

	if buf[0] != Version5 {
		return "", errors.New("unsupported version")
	}

	cmd := buf[1]
	addrType := buf[3]

	if cmd != CmdConnect {
		s.sendReply(conn, ReplyCmdNotSupported, nil)
		return "", errors.New("unsupported command")
	}

	var host string
	switch addrType {
	case AddrTypeIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()

	case AddrTypeDomain:
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return "", err
		}
		domainLen := int(buf[0])
		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", err
		}
		host = string(domain)

	case AddrTypeIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", err
		}
		host = net.IP(addr).String()

	default:
		s.sendReply(conn, ReplyAddrTypeNotSupported, nil)
		return "", errors.New("unsupported address type")
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBuf)

	return fmt.Sprintf("%s:%d", host, port), nil
}

func (s *Server) sendReply(conn net.Conn, reply byte, addr *net.TCPAddr) {
	resp := make([]byte, 10)
	resp[0] = Version5
	resp[1] = reply
	resp[2] = 0x00 
	resp[3] = AddrTypeIPv4

	if addr != nil {
		ip := addr.IP.To4()
		if ip != nil {
			copy(resp[4:8], ip)
		}
		binary.BigEndian.PutUint16(resp[8:10], uint16(addr.Port))
	}

	conn.Write(resp)
}

func (s *Server) Shutdown(ctx context.Context) error {
	close(s.shutdown)
	if s.ln != nil {
		s.ln.Close()
	}
	s.wg.Wait()
	return nil
}
