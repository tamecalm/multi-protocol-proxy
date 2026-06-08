package pac

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"multi-protocol-proxy/internal/auth"
	"multi-protocol-proxy/internal/ui"
)

type Config struct {
	Enabled        bool
	ProxyHost      string 
	HTTPPort       string 
	SOCKS5Port     string 
	Token          string 
	DefaultUser    string 
	RateLimitRPM   int    
}

type Handler struct {
	config    *Config
	userStore *auth.UserStore

	rateMu      sync.Mutex
	rateTokens  map[string]int
	rateWindow  map[string]time.Time
}

func NewHandler(cfg *Config, userStore *auth.UserStore) *Handler {
	return &Handler{
		config:     cfg,
		userStore:  userStore,
		rateTokens: make(map[string]int),
		rateWindow: make(map[string]time.Time),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)

	if h.config.RateLimitRPM > 0 && !h.checkRateLimit(clientIP) {
		ui.LogStatus("warn", "PAC rate limited: "+clientIP)
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	if h.config.Token != "" {
		token := r.URL.Query().Get("token")
		if token != h.config.Token {
			ui.LogStatus("warn", "PAC invalid token from: "+clientIP)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	username := r.URL.Query().Get("user")
	if username == "" {
		username = h.config.DefaultUser
	}

	if username == "" {
		h.sendErrorPAC(w, "No user specified. Use ?user=USERNAME")
		return
	}

	
	password := r.URL.Query().Get("pass")
	if password == "" {
		h.sendPACWithPlaceholder(w, username)
		return
	}

	_, valid := h.userStore.ValidateCredentials(username, password)
	if !valid {
		ui.LogStatus("warn", "PAC invalid credentials for user: "+username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	pac := h.generatePAC(username, password)
	h.sendPAC(w, pac)

	ui.LogStatus("info", "PAC served for user: "+username+" from "+clientIP)
}

func (h *Handler) generatePAC(username, password string) string {
	proxyURL := fmt.Sprintf("%s:%s@%s:%s",
		username, password, h.config.ProxyHost, h.config.HTTPPort)

	socks5URL := fmt.Sprintf("%s:%s@%s:%s",
		username, password, h.config.ProxyHost, h.config.SOCKS5Port)

	return fmt.Sprintf(`function FindProxyForURL(url, host) {
    // Don't proxy local addresses
    if (isPlainHostName(host) ||
        shExpMatch(host, "*.local") ||
        isInNet(host, "192.168.0.0", "255.255.0.0") ||
        isInNet(host, "10.0.0.0", "255.0.0.0") ||
        isInNet(host, "172.16.0.0", "255.240.0.0") ||
        host == "localhost" ||
        host == "127.0.0.1") {
        return "DIRECT";
    }
    
    // Route everything else through proxy
    // Primary: HTTP/HTTPS proxy, Fallback: SOCKS5
    return "PROXY %s; SOCKS5 %s; DIRECT";
}
`, proxyURL, socks5URL)
}

func (h *Handler) sendPACWithPlaceholder(w http.ResponseWriter, username string) {
	pac := fmt.Sprintf(`function FindProxyForURL(url, host) {
    // PAC file for user: %s
    // Note: This PAC requires authentication. Your browser/system will prompt for password.
    
    // Don't proxy local addresses
    if (isPlainHostName(host) ||
        shExpMatch(host, "*.local") ||
        isInNet(host, "192.168.0.0", "255.255.0.0") ||
        isInNet(host, "10.0.0.0", "255.0.0.0") ||
        isInNet(host, "172.16.0.0", "255.240.0.0") ||
        host == "localhost" ||
        host == "127.0.0.1") {
        return "DIRECT";
    }
    
    // Route everything else through proxy (credentials required separately)
    return "PROXY %s:%s; SOCKS5 %s:%s; DIRECT";
}
`, username, h.config.ProxyHost, h.config.HTTPPort, h.config.ProxyHost, h.config.SOCKS5Port)

	h.sendPAC(w, pac)
}

func (h *Handler) sendErrorPAC(w http.ResponseWriter, message string) {
	pac := fmt.Sprintf(`// Error: %s
function FindProxyForURL(url, host) {
    return "DIRECT";
}
`, message)

	h.sendPAC(w, pac)
}

func (h *Handler) sendPAC(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func (h *Handler) checkRateLimit(clientIP string) bool {
	h.rateMu.Lock()
	defer h.rateMu.Unlock()

	now := time.Now()
	windowStart, exists := h.rateWindow[clientIP]

	if !exists || now.Sub(windowStart) > time.Minute {
		h.rateWindow[clientIP] = now
		h.rateTokens[clientIP] = 1
		return true
	}

	if h.rateTokens[clientIP] < h.config.RateLimitRPM {
		h.rateTokens[clientIP]++
		return true
	}

	return false
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

func GenerateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:7], nil
}
