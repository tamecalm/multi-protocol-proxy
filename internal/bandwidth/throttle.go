package bandwidth

import (
	"net"
	"sync"
	"time"
)

type ThrottledConn struct {
	net.Conn
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 
	lastRefill time.Time
}

func NewThrottledConn(conn net.Conn, speedMbps int) net.Conn {
	if speedMbps <= 0 {
		return conn 
	}

	bytesPerSec := float64(speedMbps) * 1024 * 1024 / 8 

	// Allow burst up to 1 second of bandwidth
	maxTokens := bytesPerSec

	return &ThrottledConn{
		Conn:       conn,
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: bytesPerSec,
		lastRefill: time.Now(),
	}
}

func (tc *ThrottledConn) Read(b []byte) (int, error) {
	tc.waitForTokens(len(b))
	n, err := tc.Conn.Read(b)
	if n > 0 {
		tc.consumeTokens(n)
	}
	return n, err
}

func (tc *ThrottledConn) Write(b []byte) (int, error) {
	tc.waitForTokens(len(b))
	n, err := tc.Conn.Write(b)
	if n > 0 {
		tc.consumeTokens(n)
	}
	return n, err
}

func (tc *ThrottledConn) waitForTokens(needed int) {
	for {
		tc.mu.Lock()
		tc.refill()
		if tc.tokens >= 1 {
			tc.mu.Unlock()
			return
		}
		deficit := float64(needed) - tc.tokens
		if deficit < 1 {
			deficit = 1
		}
		waitDuration := time.Duration(deficit / tc.refillRate * float64(time.Second))
		if waitDuration < time.Millisecond {
			waitDuration = time.Millisecond
		}
		if waitDuration > 100*time.Millisecond {
			waitDuration = 100 * time.Millisecond
		}
		tc.mu.Unlock()
		time.Sleep(waitDuration)
	}
}

func (tc *ThrottledConn) consumeTokens(n int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tokens -= float64(n)
}

func (tc *ThrottledConn) refill() {
	now := time.Now()
	elapsed := now.Sub(tc.lastRefill).Seconds()
	tc.tokens += elapsed * tc.refillRate
	if tc.tokens > tc.maxTokens {
		tc.tokens = tc.maxTokens
	}
	tc.lastRefill = now
}
