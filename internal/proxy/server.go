package proxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"multi-protocol-proxy/internal/config"
	"multi-protocol-proxy/internal/ui"
)

type Server struct {
	Config   *config.Config
	ln       net.Listener
	connSem  chan struct{}  
	wg       sync.WaitGroup 
	shutdown chan struct{} 
	mu   sync.RWMutex
	cert *tls.Certificate
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		Config:   cfg,
		connSem:  make(chan struct{}, cfg.MaxConns),
		shutdown: make(chan struct{}),
	}
}

func (s *Server) Reload() error {
	cert, err := tls.LoadX509KeyPair(s.Config.CertFile, s.Config.KeyFile)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.cert = &cert
	s.mu.Unlock()

	ui.LogStatus("success", "Certificates reloaded from disk")
	return nil
}

func (s *Server) getCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cert, nil
}

func (s *Server) Start(ctx context.Context) error {
	// 1. Initial certificate load
	if err := s.Reload(); err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		GetCertificate: s.getCertificate,
		MinVersion:     tls.VersionTLS12,
		NextProtos:     []string{"http/1.1"},
	}

	var err error
	s.ln, err = tls.Listen("tcp", s.Config.Listen, tlsConfig)
	if err != nil {
		return err
	}

	metricsAddr := s.Config.MetricsListen
	if strings.HasPrefix(metricsAddr, ":") {
		metricsAddr = "localhost" + metricsAddr
	}
	ui.LogStatus("info", "Metrics: http://"+metricsAddr+"/metrics")
	
	Stats.AllowedOrigin = s.Config.Env.AllowedOrigin
	
	ui.LogStatus("info", "Stats API: https://" + s.Config.Env.APIDomain + "/api/stats")

	go s.watchShutdown(ctx)

	for {
		select {
		case <-s.shutdown:
			return s.drainConnections()
		default:
		}

		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return s.drainConnections()
			default:
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return err
			}
		}

		select {
		case s.connSem <- struct{}{}:
			s.wg.Add(1)
			go func(c net.Conn) {
				defer s.wg.Done()
				defer func() { <-s.connSem }() 
				HandleConnection(ctx, c, s.Config)
			}(conn)
		default:
			MetricConnectionsRejected.Inc()
			ui.LogStatus("warn", "Connection rejected: at max capacity ("+itoa(s.Config.MaxConns)+")")
			conn.Close()
		}
	}
}

func (s *Server) watchShutdown(ctx context.Context) {
	<-ctx.Done()
	ui.LogStatus("warn", "Shutdown signal received...")
	close(s.shutdown)
	s.ln.Close()
}

func (s *Server) drainConnections() error {
	activeConns := GetActiveConns()
	if activeConns > 0 {
		ui.LogStatus("info", "Draining "+itoa(activeConns)+" active connections (30s timeout)...")
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		ui.LogStatus("success", "All connections drained. Goodbye.")
	case <-time.After(30 * time.Second):
		ui.LogStatus("warn", "Drain timeout reached. Forcing shutdown.")
	}

	return nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}


func PeekSNI(conn net.Conn) (string, []byte, error) {
	buf := make([]byte, 16384)
	n, err := conn.Read(buf)
	if err != nil {
		return "", nil, err
	}
	data := buf[:n]
	sni := extractSNI(data)
	return sni, data, nil
}

func extractSNI(data []byte) string {
	if len(data) < 5 {
		return ""
	}

	if data[0] != 0x16 {
		return ""
	}

	pos := 5

	if len(data) < pos+4 {
		return ""
	}

	if data[pos] != 0x01 {
		return ""
	}

	pos += 4

	if len(data) < pos+34 {
		return ""
	}
	pos += 34

	if len(data) < pos+1 {
		return ""
	}
	sessionIDLen := int(data[pos])
	pos += 1 + sessionIDLen

	if len(data) < pos+2 {
		return ""
	}
	cipherSuitesLen := int(data[pos])<<8 | int(data[pos+1])
	pos += 2 + cipherSuitesLen

	if len(data) < pos+1 {
		return ""
	}
	compressionLen := int(data[pos])
	pos += 1 + compressionLen

	if len(data) < pos+2 {
		return ""
	}
	extensionsLen := int(data[pos])<<8 | int(data[pos+1])
	pos += 2

	endPos := pos + extensionsLen
	if endPos > len(data) {
		endPos = len(data)
	}

	for pos+4 <= endPos {
		extType := int(data[pos])<<8 | int(data[pos+1])
		extLen := int(data[pos+2])<<8 | int(data[pos+3])
		pos += 4

		if extType == 0x0000 { 
			if pos+5 > endPos {
				return ""
			}
			if data[pos+2] != 0x00 {
				return ""
			}
			nameLen := int(data[pos+3])<<8 | int(data[pos+4])
			if pos+5+nameLen > endPos {
				return ""
			}
			return string(data[pos+5 : pos+5+nameLen])
		}
		pos += extLen
	}

	return ""
}

func HandleConnection(ctx context.Context, clientConn net.Conn, cfg *config.Config) {
	defer clientConn.Close()

	MetricActiveConns.Inc()
	defer MetricActiveConns.Dec()

	startTime := time.Now()
	timeout := time.Duration(cfg.TimeoutSec) * time.Second

	clientConn.SetDeadline(time.Now().Add(10 * time.Second))

	sni, initialData, err := PeekSNI(clientConn)
	if err != nil {
		MetricErrorsTotal.WithLabelValues("peek_failed").Inc()
		Stats.RecordError()
		ui.LogStatus("error", "Failed to peek SNI: "+err.Error())
		return
	}

	target, allowed := cfg.Hosts[strings.ToLower(sni)]
	if !allowed || sni == "" {
		if len(initialData) > 0 && initialData[0] != 0x16 {
			handleInternalAPI(clientConn, initialData)
			return
		}

		MetricErrorsTotal.WithLabelValues("unauthorized_sni").Inc()
		Stats.RecordError()
		ui.LogStatus("error", "Unauthorized SNI: "+sni)
		return
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	upConn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		MetricErrorsTotal.WithLabelValues("dial_failed").Inc()
		Stats.RecordError()
		ui.LogStatus("error", "Target unreachable: "+target+" - "+err.Error())
		return
	}
	defer upConn.Close()

	if len(initialData) > 0 {
		if _, err := upConn.Write(initialData); err != nil {
			MetricErrorsTotal.WithLabelValues("write_failed").Inc()
			return
		}
	}

	MetricRelayTotal.WithLabelValues(sni).Inc()
	Stats.RecordRelay()

	clientConn.SetDeadline(time.Time{})
	upConn.SetDeadline(time.Time{})

	done := make(chan struct{}, 2)
	var upBytes, downBytes int64

	copyData := func(dst, src net.Conn, bytes *int64) {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, 32*1024)
		for {
			src.SetDeadline(time.Now().Add(timeout))
			select {
			case <-ctx.Done():
				return
			default:
			}
			nr, er := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[:nr])
				if nw > 0 {
					*bytes += int64(nw)
				}
				if ew != nil {
					break
				}
			}
			if er != nil {
				break
			}
		}
	}

	go copyData(upConn, clientConn, &upBytes)
	go copyData(clientConn, upConn, &downBytes)

	select {
	case <-done:
	case <-ctx.Done():
	}

	duration := time.Since(startTime).Seconds()
	MetricConnectionDuration.Observe(duration)
	MetricBytesTotal.WithLabelValues(sni, "upstream").Add(float64(upBytes))
	MetricBytesTotal.WithLabelValues(sni, "downstream").Add(float64(downBytes))
	Stats.RecordBytes(upBytes + downBytes)

	ui.LogRelay(sni, clientConn.RemoteAddr().String(), upBytes, downBytes)
}

func handleInternalAPI(conn net.Conn, initialData []byte) {
	ui.LogStatus("info", "Handling API request from "+conn.RemoteAddr().String())
	
	reader := io.MultiReader(bytes.NewReader(initialData), conn)
	br := bufio.NewReader(reader)

	req, err := http.ReadRequest(br)
	if err != nil {
		if err != io.EOF {
			ui.LogStatus("error", "API ReadRequest error: "+err.Error())
		}
		return
	}

	w := &simpleResponseWriter{
		conn:   conn,
		header: make(http.Header),
	}

	switch req.URL.Path {
	case "/api/stats":
		StatsHandler(w, req)
	case "/api/history":
		HistoryHandler(w, req)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}

	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
}

type simpleResponseWriter struct {
	conn        net.Conn
	header      http.Header
	wroteHeader bool
	status      int
}

func (w *simpleResponseWriter) Header() http.Header {
	return w.header
}

func (w *simpleResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.conn.Write(b)
}

func (w *simpleResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status

	fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", status, http.StatusText(status))
	
	w.header.Set("Date", time.Now().Format(http.TimeFormat))
	w.header.Set("Connection", "close")
	
	for k, vv := range w.header {
		for _, v := range vv {
			fmt.Fprintf(w.conn, "%s: %s\r\n", k, v)
		}
	}
	
	fmt.Fprintf(w.conn, "\r\n")
}
