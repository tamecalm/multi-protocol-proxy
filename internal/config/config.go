package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Listen        string            `json:"listen"`
	CertFile      string            `json:"cert_file"`
	KeyFile       string            `json:"key_file"`
	TimeoutSec    int               `json:"timeout_sec"`
	MaxConns      int               `json:"max_conns"`
	MetricsListen string            `json:"metrics_listen"`
	Hosts         map[string]string `json:"hosts"`
	
	Env *EnvConfig `json:"-"`
}

func Load() *Config {
	cfg := &Config{
		Listen:        ":8443",
		TimeoutSec:    300,
		MaxConns:      1000,
		MetricsListen: ":9090",
		CertFile:      "certs/dev/server.crt",
		KeyFile:       "certs/dev/server.key",
		Hosts:         make(map[string]string),
		Env:           LoadEnv(),
	}

	if file, err := os.Open("config.json"); err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(cfg)
	}

	if certFile := os.Getenv("CERT_FILE"); certFile != "" {
		cfg.CertFile = certFile
	}
	if keyFile := os.Getenv("KEY_FILE"); keyFile != "" {
		cfg.KeyFile = keyFile
	}
	if metricsListen := os.Getenv("METRICS_LISTEN"); metricsListen != "" {
		cfg.MetricsListen = metricsListen
	}

	cleaned := make(map[string]string)
	for k, v := range cfg.Hosts {
		cleaned[strings.ToLower(strings.TrimSpace(k))] = v
	}
	cfg.Hosts = cleaned

	return cfg
}

func (c *Config) Validate() error {
	var errs []string

	if c.Listen == "" {
		errs = append(errs, "listen address is required")
	}

	if _, err := os.Stat(c.CertFile); os.IsNotExist(err) {
		errs = append(errs, fmt.Sprintf("certificate file not found: %s", c.CertFile))
	}
	if _, err := os.Stat(c.KeyFile); os.IsNotExist(err) {
		errs = append(errs, fmt.Sprintf("key file not found: %s", c.KeyFile))
	}

	if c.TimeoutSec <= 0 {
		errs = append(errs, "timeout_sec must be positive")
	}
	if c.MaxConns <= 0 {
		errs = append(errs, "max_conns must be positive")
	}

	if len(c.Hosts) == 0 {
		errs = append(errs, "at least one host mapping is required")
	}

	if len(errs) > 0 {
		return errors.New("config validation failed:\n  - " + strings.Join(errs, "\n  - "))
	}

	return nil
}
