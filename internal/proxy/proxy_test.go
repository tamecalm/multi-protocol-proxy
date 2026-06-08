package proxy

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"
	"multi-protocol-proxy/internal/config"
)

func TestProxyRedirection(t *testing.T) {
	certFile, keyFile := writeTempCertFiles(t)

	targetCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatal(err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{targetCert},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	mockServerAddr := ln.Addr().String()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				tlsConn, ok := c.(*tls.Conn)
				if ok {
					if err := tlsConn.Handshake(); err != nil {
						return
					}
				}
				buf := make([]byte, 1024)
				n, _ := c.Read(buf)
				if n > 0 {
					c.Write([]byte("MOCK_PROXY_RESPONSE: " + string(buf[:n])))
				}
			}(conn)
		}
	}()

	cfg := &config.Config{
		Listen:        "127.0.0.1:0",
		TimeoutSec:    2,
		MaxConns:      10,
		MetricsListen: ":0",
		Hosts: map[string]string{
			"localhost": mockServerAddr,
		},
		CertFile: certFile,
		KeyFile:  keyFile,
		Env: &config.EnvConfig{
			Env: config.Development,
		},
	}

	srv := NewServer(cfg)
	
	fmt.Println("Starting proxy server...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	if srv.ln == nil {
		t.Fatal("Server listener not initialized")
	}
	proxyAddr := srv.ln.Addr().String()
	fmt.Println("Proxy listening on:", proxyAddr)

	time.Sleep(100 * time.Millisecond)
	conf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "localhost",
	}
	conn, err := tls.Dial("tcp", proxyAddr, conf)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	innerConf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "localhost",
	}
	innerConn := tls.Client(conn, innerConf)
	if err := innerConn.Handshake(); err != nil {
		t.Fatalf("Inner TLS handshake failed: %s", err)
	}
	defer innerConn.Close()

	payload := "Hello Proxy"
	fmt.Fprint(innerConn, payload)
	
	resp := make([]byte, 1024)
	n, err := innerConn.Read(resp)
	if err != nil {
		t.Fatal(err)
	}

	expected := "MOCK_PROXY_RESPONSE: " + payload
	if string(resp[:n]) != expected {
		t.Errorf("Expected %q, got %q", expected, string(resp[:n]))
	}
}

func writeTempCertFiles(t *testing.T) (certFile string, keyFile string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Multi-Protocol Proxy Test"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "127.0.0.1"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %s", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("unable to marshal private key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	// Create temp files
	fCert, err := os.CreateTemp("", "test-cert-*.crt")
	if err != nil {
		t.Fatalf("failed to create temp cert file: %s", err)
	}
	defer fCert.Close()
	fCert.Write(certPEM)

	fKey, err := os.CreateTemp("", "test-key-*.key")
	if err != nil {
		t.Fatalf("failed to create temp key file: %s", err)
	}
	defer fKey.Close()
	fKey.Write(keyPEM)

	t.Cleanup(func() {
		os.Remove(fCert.Name())
		os.Remove(fKey.Name())
	})

	return fCert.Name(), fKey.Name()
}
