# HTTP/HTTPS and SOCKS5 Proxy Configuration

This document describes the **HTTP/HTTPS/SOCKS5 private proxy mode** in the Multi-Protocol Proxy codebase.

## Quick Start

### 1. Generate Password Hash

```bash
# Run the password hash generator
go run scripts/hash-password.go
# Enter your password when prompted, copy the hash
```

### 2. Configure Users

Edit `users.json`:
```json
{
  "users": [
    {
      "username": "admin",
      "password_hash": "$2a$10$YOUR_HASH_HERE",
      "rate_limit_rpm": 500,
      "enabled": true
    }
  ],
  "ip_whitelist": []
}
```

### 3. Start in HTTPS Proxy Mode

```bash
# Set environment and run
PROXY_MODE=https ./multi-protocol-proxy

# Or use the env file
cp env.https.example .env
# Edit .env with your settings
./multi-protocol-proxy
```

### 4. Test the Proxy

```bash
# HTTP Proxy with curl
curl -x http://admin:password@localhost:8080 http://httpbin.org/ip

# HTTPS via CONNECT
curl -x http://admin:password@localhost:8080 https://httpbin.org/ip

# SOCKS5
curl --socks5 localhost:1080 --proxy-user admin:password https://httpbin.org/ip
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   multi-protocol-proxy                      │
├────────────────────────┬────────────────────────────────────┤
│  PROXY_MODE=sni        │     PROXY_MODE=https               │
│  (Default)             │                                    │
├────────────────────────┼────────────────────────────────────┤
│  SNI TLS Proxy         │  HTTP Proxy      │  SOCKS5 Proxy  │
│  Port 8443             │  Port 8080/8443  │  Port 1080     │
│  SNI-based routing     │  CONNECT method  │  RFC 1928/1929 │
│  No authentication     │  Basic auth      │  User/pass auth│
└────────────────────────┴──────────────────┴─────────────────┘
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PROXY_MODE` | `sni` | `sni` (SNI routing) or `https` (Private Proxy) |
| `HTTP_PROXY_PORT` | `:8080` | HTTP proxy listen port |
| `HTTP_PROXY_TLS` | `true` | Enable TLS for HTTP proxy |
| `HTTP_PROXY_TLS_PORT` | `:8443` | HTTPS proxy listen port |
| `SOCKS5_PORT` | `:1080` | SOCKS5 listen port |
| `USERS_FILE` | `users.json` | Path to user credentials |
| `PAC_ENABLED` | `true` | Enable PAC endpoint at `/proxy.pac` |
| `PAC_TOKEN` | *(empty)* | Optional secret token for PAC access |
| `PAC_DEFAULT_USER` | *(empty)* | Default username for PAC |
| `PAC_RATE_LIMIT_RPM` | `60` | PAC endpoint rate limit |

---

## PAC (Proxy Auto-Config)

Enable automatic proxy configuration for browsers and systems:

### Get PAC File
```bash
# Basic PAC (browser will prompt for password)
curl https://private.yourdomain.com/proxy.pac?user=admin

# PAC with embedded credentials
curl "https://private.yourdomain.com/proxy.pac?user=admin&pass=yourpassword"
```

### Configure Clients
- **macOS**: System Preferences → Network → Proxies → Automatic Proxy Configuration
- **Windows**: Settings → Proxy → Use setup script
- **Firefox**: Settings → Network Settings → Automatic proxy configuration URL

Enter: `https://private.yourdomain.com/proxy.pac?user=YOUR_USER`

See [api/PAC.md](api/PAC.md) for full documentation.

---

## User Configuration

### users.json Format

```json
{
  "users": [
    {
      "username": "john",
      "password_hash": "$2a$10$...",
      "rate_limit_rpm": 500,
      "enabled": true
    }
  ],
  "ip_whitelist": ["192.168.1.0/24", "10.0.0.5"]
}
```

- **rate_limit_rpm**: Requests per minute (0 = unlimited)
- **ip_whitelist**: CIDR ranges allowed (empty = allow all)

---

## Prometheus Metrics

### HTTP Proxy Metrics
- `httpproxy_requests_total{user, method}`
- `httpproxy_bytes_total{user, direction}`
- `httpproxy_active_connections`
- `httpproxy_auth_failures_total{type}`
- `httpproxy_rate_limited_total{user}`

### SOCKS5 Metrics
- `socks5_connections_total{user}`
- `socks5_bytes_total{user, direction}`
- `socks5_active_connections`
- `socks5_auth_failures_total{type}`

---

## Deployment

### Deploy as HTTPS Proxy (Separate VPS)

```bash
# 1. Copy binary and config
scp build/multi-protocol-proxy-linux-amd64 user@vps:/opt/proxy/multi-protocol-proxy
scp users.json user@vps:/opt/proxy/
scp env.https.example user@vps:/opt/proxy/.env

# 2. Configure TLS certificates
# Place cert.pem and key.pem in certs/ directory

# 3. Start the proxy
ssh user@vps
cd /opt/proxy
PROXY_MODE=https ./multi-protocol-proxy
```

### Systemd Service

```ini
[Unit]
Description=HTTPS/SOCKS5 Proxy
After=network.target

[Service]
Type=simple
User=proxy
WorkingDirectory=/opt/proxy
ExecStart=/opt/proxy/multi-protocol-proxy
Restart=always
Environment=PROXY_MODE=https
Environment=APP_ENV=production

[Install]
WantedBy=multi-user.target
```

---

## Client Configuration

### Browser (Firefox)
1. Settings → Network Settings → Manual proxy
2. HTTP Proxy: `your-vps-ip`, Port: `8080`
3. Check "Use this proxy for all protocols"
4. Enter username/password when prompted

### Mobile (Android/iOS)
1. WiFi Settings → Proxy → Manual
2. Host: `your-vps-ip`, Port: `8080`
3. Username/Password as configured

### CLI Tools
```bash
# Environment variables
export http_proxy=http://user:pass@vps-ip:8080
export https_proxy=http://user:pass@vps-ip:8080

# Or per-command
curl -x http://user:pass@vps-ip:8080 https://example.com
```
