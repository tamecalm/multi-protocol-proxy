package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username     string `json:"username"`
	Role         string `json:"role"`
	PasswordHash string `json:"password_hash"`
	RateLimitRPM int    `json:"rate_limit_rpm"`
	Enabled      bool   `json:"enabled"`
	Plan               string `json:"plan,omitempty"`                
	BandwidthLimitGB   int    `json:"bandwidth_limit_gb,omitempty"`  
	BandwidthSpeedMbps int    `json:"bandwidth_speed_mbps,omitempty"` 
	MaxConnections     int    `json:"max_connections,omitempty"`     
	ExpiresAt          string `json:"expires_at,omitempty"`          
}

type UsersConfig struct {
	Users         []User   `json:"users"`
	IPWhitelist   []string `json:"ip_whitelist"`    
	SuperAdminIPs []string `json:"super_admin_ips"` 
}

type UserStore struct {
	mu             sync.RWMutex
	users          map[string]*User
	ipWhitelist    []*net.IPNet
	superAdminIPs  []*net.IPNet
	superAdminUser *User 
	rateLimiter    *RateLimiter
	credCacheMu sync.RWMutex
	credCache   map[string]credCacheEntry
}

const credCacheTTL = 5 * time.Minute

type credCacheEntry struct {
	user       *User
	validUntil time.Time
}

func NewUserStore(configPath string) (*UserStore, error) {
	store := &UserStore{
		users:         make(map[string]*User),
		ipWhitelist:   make([]*net.IPNet, 0),
		superAdminIPs: make([]*net.IPNet, 0),
		rateLimiter:   NewRateLimiter(),
		credCache:     make(map[string]credCacheEntry),
	}

	if err := store.LoadFromFile(configPath); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *UserStore) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read users file: %w", err)
	}

	var cfg UsersConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse users file: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.users = make(map[string]*User)
	for i := range cfg.Users {
		user := &cfg.Users[i]
		if user.Enabled {
			s.users[strings.ToLower(user.Username)] = user
			// Initialize rate limiter for user
			if user.RateLimitRPM > 0 {
				s.rateLimiter.SetLimit(user.Username, user.RateLimitRPM)
			}
		}
	}

	s.superAdminUser = nil
	for _, user := range s.users {
		if strings.ToLower(user.Role) == "super_admin" {
			s.superAdminUser = user
			break
		}
	}

	s.ipWhitelist = make([]*net.IPNet, 0, len(cfg.IPWhitelist))
	for _, cidr := range cfg.IPWhitelist {
		ipNet, err := parseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid IP whitelist entry '%s': %w", cidr, err)
		}
		s.ipWhitelist = append(s.ipWhitelist, ipNet)
	}

	s.superAdminIPs = make([]*net.IPNet, 0, len(cfg.SuperAdminIPs))
	for _, cidr := range cfg.SuperAdminIPs {
		ipNet, err := parseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid super_admin_ips entry '%s': %w", cidr, err)
		}
		s.superAdminIPs = append(s.superAdminIPs, ipNet)
	}

	s.InvalidateAllCredentials()

	return nil
}

func (s *UserStore) ValidateCredentials(username, password string) (*User, bool) {
	passHash := sha256.Sum256([]byte(password))
	cacheKey := strings.ToLower(username) + ":" + hex.EncodeToString(passHash[:])

	s.credCacheMu.RLock()
	if entry, ok := s.credCache[cacheKey]; ok && time.Now().Before(entry.validUntil) {
		s.credCacheMu.RUnlock()
		return entry.user, true
	}
	s.credCacheMu.RUnlock()

	s.mu.RLock()
	user, exists := s.users[strings.ToLower(username)]
	s.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, false
	}

	s.credCacheMu.Lock()
	s.credCache[cacheKey] = credCacheEntry{
		user:       user,
		validUntil: time.Now().Add(credCacheTTL),
	}
	s.credCacheMu.Unlock()

	return user, true
}

func (s *UserStore) InvalidateUser(username string) {
	s.credCacheMu.Lock()
	defer s.credCacheMu.Unlock()

	prefix := strings.ToLower(username) + ":"
	for key := range s.credCache {
		if strings.HasPrefix(key, prefix) {
			delete(s.credCache, key)
		}
	}
}

func (s *UserStore) InvalidateAllCredentials() {
	s.credCacheMu.Lock()
	defer s.credCacheMu.Unlock()

	s.credCache = make(map[string]credCacheEntry)
}

func (s *UserStore) CheckIPAllowed(ipStr string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.ipWhitelist) == 0 {
		return true
	}

	host := ipStr
	if strings.Contains(ipStr, ":") {
		var err error
		host, _, err = net.SplitHostPort(ipStr)
		if err != nil {
			// Might be IPv6 without port
			host = ipStr
		}
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// Check against whitelist
	for _, ipNet := range s.ipWhitelist {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

func (s *UserStore) CheckRateLimit(username string) bool {
	s.mu.RLock()
	user, exists := s.users[strings.ToLower(username)]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if user.RateLimitRPM <= 0 {
		return true
	}

	return s.rateLimiter.Allow(username)
}

func (s *UserStore) IsSuperAdminIP(ipStr string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.superAdminUser == nil || len(s.superAdminIPs) == 0 {
		return nil, false
	}

	ip := parseIP(ipStr)
	if ip == nil {
		return nil, false
	}

	for _, ipNet := range s.superAdminIPs {
		if ipNet.Contains(ip) {
			return s.superAdminUser, true
		}
	}

	return nil, false
}

func (s *UserStore) GetUserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

func (s *UserStore) GetUser(username string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[strings.ToLower(username)]
}

func (s *UserStore) CheckExpiry(username string) bool {
	s.mu.RLock()
	user, exists := s.users[strings.ToLower(username)]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if user.ExpiresAt == "" {
		return true 
	}

	expiryTime, err := time.Parse(time.RFC3339, user.ExpiresAt)
	if err != nil {
		return true
	}

	return time.Now().Before(expiryTime)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func parseCIDR(cidr string) (*net.IPNet, error) {
	if !strings.Contains(cidr, "/") {
		if strings.Contains(cidr, ":") {
			cidr = cidr + "/128" // IPv6
		} else {
			cidr = cidr + "/32" // IPv4
		}
	}
	_, ipNet, err := net.ParseCIDR(cidr)
	return ipNet, err
}

func parseIP(ipStr string) net.IP {
	host := ipStr
	if strings.Contains(ipStr, ":") {
		var err error
		host, _, err = net.SplitHostPort(ipStr)
		if err != nil {
			host = ipStr 
		}
	}
	return net.ParseIP(host)
}
