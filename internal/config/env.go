package config

import (
	"os"
	"strings"
)

type Environment string

const (
	Development Environment = "development"
	Production Environment = "production"
)

type EnvConfig struct {
	Env Environment
	Domain        string
	APIDomain     string
	BaseURL       string
	AllowedOrigin string
	Debug bool
	LogLevel string
	NgrokEnabled bool
	NgrokDomain  string
	ProxyMode       string 
	HTTPProxyPort   string 
	HTTPProxyTLS    bool   
	HTTPProxyTLSPort string 
	SOCKS5Port      string 
	UsersFile       string 
	PACEnabled      bool    
	PACToken        string  
	PACDefaultUser  string 
	PACRateLimitRPM int    
}

func LoadEnv() *EnvConfig {
	env := getEnvOrDefault("APP_ENV", "development")

	cfg := &EnvConfig{
		Env:      Environment(strings.ToLower(env)),
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
	}

	switch cfg.Env {
	case Production:
		cfg.Domain = getEnvOrDefault("DOMAIN", "proxy.yourdomain.com")
		cfg.APIDomain = getEnvOrDefault("API_DOMAIN", "api."+cfg.Domain)
		cfg.BaseURL = getEnvOrDefault("BASE_URL", "https://"+cfg.Domain)
		cfg.AllowedOrigin = getEnvOrDefault("ALLOWED_ORIGIN", "*")
		cfg.Debug = getEnvOrDefault("DEBUG", "false") == "true"
		cfg.NgrokEnabled = false 
		if cfg.LogLevel == "info" {
			cfg.LogLevel = "info"
		}
	default:
		cfg.Env = Development

		cfg.NgrokEnabled = getEnvOrDefault("NGROK_ENABLED", "true") == "true"
		cfg.NgrokDomain = getEnvOrDefault("NGROK_DOMAIN", "")

		if cfg.NgrokEnabled && cfg.NgrokDomain != "" {
			cfg.Domain = getEnvOrDefault("DOMAIN", cfg.NgrokDomain)
		} else {
			cfg.Domain = getEnvOrDefault("DOMAIN", "localhost:8443")
		}

		cfg.BaseURL = getEnvOrDefault("BASE_URL", "https://"+cfg.Domain)
		cfg.APIDomain = cfg.Domain
		cfg.AllowedOrigin = "*"
		cfg.Debug = getEnvOrDefault("DEBUG", "true") == "true"
		if cfg.LogLevel == "info" {
			cfg.LogLevel = "debug"
		}
	}

	cfg.ProxyMode = strings.ToLower(getEnvOrDefault("PROXY_MODE", "sni"))
	cfg.HTTPProxyPort = getEnvOrDefault("HTTP_PROXY_PORT", ":8080")
	cfg.HTTPProxyTLS = getEnvOrDefault("HTTP_PROXY_TLS", "true") == "true"
	cfg.HTTPProxyTLSPort = getEnvOrDefault("HTTP_PROXY_TLS_PORT", ":8443")
	cfg.SOCKS5Port = getEnvOrDefault("SOCKS5_PORT", ":1080")
	cfg.UsersFile = getEnvOrDefault("USERS_FILE", "users.json")

	// Load PAC configuration
	cfg.PACEnabled = getEnvOrDefault("PAC_ENABLED", "true") == "true"
	cfg.PACToken = getEnvOrDefault("PAC_TOKEN", "") 
	cfg.PACDefaultUser = getEnvOrDefault("PAC_DEFAULT_USER", "")
	cfg.PACRateLimitRPM = parseIntOrDefault(getEnvOrDefault("PAC_RATE_LIMIT_RPM", "60"), 60)

	return cfg
}

func (e *EnvConfig) IsDevelopment() bool {
	return e.Env == Development
}

func (e *EnvConfig) IsProduction() bool {
	return e.Env == Production
}

func (e Environment) String() string {
	return string(e)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntOrDefault(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	result := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return defaultValue
		}
	}
	return result
}
