package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"multi-protocol-proxy/internal/auth"
	"multi-protocol-proxy/internal/bandwidth"
	"multi-protocol-proxy/internal/config"
	"multi-protocol-proxy/internal/httpproxy"
	"multi-protocol-proxy/internal/proxy"
	"multi-protocol-proxy/internal/socks5"
	"multi-protocol-proxy/internal/ui"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ui.PrintBanner()

	cfg := config.Load()

	if cfg.Env.IsDevelopment() {
		ui.LogStatus("info", "Environment: "+ui.Warn("DEVELOPMENT"))
		ui.LogStatus("info", "Domain: "+cfg.Env.Domain)
	} else {
		ui.LogStatus("info", "Environment: "+ui.Success("PRODUCTION"))
		ui.LogStatus("info", "Domain: "+cfg.Env.Domain)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	switch cfg.Env.ProxyMode {
	case "https", "http", "general":
		runHTTPSProxyMode(ctx, cfg)
	default:
		runSNIProxyMode(ctx, cfg)
	}
}

func runSNIProxyMode(ctx context.Context, cfg *config.Config) {
	ui.LogStatus("info", "Proxy Mode: "+ui.Success("SNI"))

	if err := cfg.Validate(); err != nil {
		ui.LogStatus("error", err.Error())
		os.Exit(1)
	}

	metrics := proxy.NewMetricsServer(cfg.MetricsListen, nil)
	metrics.Start()
	go func() {
		<-ctx.Done()
		ui.LogGracefulShutdown()
		metrics.Shutdown(context.Background())
	}()

	srv := proxy.NewServer(cfg)

	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-sighup:
				ui.LogStatus("info", "SIGHUP received, reloading certificates...")
				if err := srv.Reload(); err != nil {
					ui.LogStatus("error", "Reload failed: "+err.Error())
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	if err := srv.Start(ctx); err != nil {
		ui.LogStatus("error", "Server failed: "+err.Error())
		log.Fatal(err)
	}
}

func runHTTPSProxyMode(ctx context.Context, cfg *config.Config) {
	ui.LogStatus("info", "Proxy Mode: "+ui.Success("HTTPS/SOCKS5"))

	userStore, err := auth.NewUserStore(cfg.Env.UsersFile)
	if err != nil {
		ui.LogStatus("error", "Failed to load users: "+err.Error())
		os.Exit(1)
	}
	ui.LogStatus("info", "Loaded "+itoa(userStore.GetUserCount())+" users from "+cfg.Env.UsersFile)

	usageFile := filepath.Join(filepath.Dir(cfg.Env.UsersFile), "bandwidth_usage.json")
	bwTracker := bandwidth.NewTracker(usageFile)
	defer bwTracker.Stop()
	ui.LogStatus("info", "Bandwidth tracker active → "+usageFile)

	// Start metrics server with /api/usage endpoint
	usageHandler := bandwidth.UsageHandler(bwTracker, cfg.Env.AllowedOrigin)
	metrics := proxy.NewMetricsServer(cfg.MetricsListen, usageHandler)
	metrics.Start()
	go func() {
		<-ctx.Done()
		ui.LogGracefulShutdown()
		metrics.Shutdown(context.Background())
	}()

	httpSrv := httpproxy.NewServer(cfg, userStore, bwTracker)

	socks5Srv := socks5.NewServer(cfg, userStore, bwTracker)

	go func() {
		if err := socks5Srv.Start(ctx); err != nil {
			ui.LogStatus("error", "SOCKS5 server failed: "+err.Error())
		}
	}()

	if err := httpSrv.Start(ctx); err != nil {
		ui.LogStatus("error", "HTTP proxy failed: "+err.Error())
		log.Fatal(err)
	}
}

// itoa is a simple int to string helper
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
