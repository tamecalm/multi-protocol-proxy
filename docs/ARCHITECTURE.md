# Multi-Protocol Proxy Architecture

## Overview

Multi-Protocol Proxy is a multi-mode proxy server deployed on virtual servers supporting:
- **SNI TLS Proxy Mode** (`proxy.yourdomain.com`) - Public TLS proxy with SNI routing
- **Private Proxy Mode** (`private.yourdomain.com`) - Authenticated HTTPS/SOCKS5 proxy

## Deployment Architecture

```
                    ┌─────────────────────────────────────┐
                    │           VPS / VM Instance         │
                    │         Ubuntu / Docker / OS        │
                    │         Elastic IP assigned         │
                    ├─────────────────────────────────────┤
    Internet ──────►│ ┌─────────────────────────────────┐ │
                    │ │      Multi-Protocol Proxy       │ │
                    │ │    /opt/proxy/mp-proxy          │ │
                    │ ├─────────────────────────────────┤ │
                    │ │                                 │ │
 proxy.yourdomain.   │ │  SNI Mode (:8443)              │──► Target Servers
    com:8443        │ │  • TLS termination              │    (SNI-matched target)
                    │ │  • SNI-based routing            │    
                    │ │  • No authentication            │    
                    │ │                                 │ │
                    │ ├─────────────────────────────────┤ │
                    │ │                                 │ │
private.yourdomain. │ │  Private Mode (:8080/:1080)     │──► Internet
 com:8080/:1080     │ │  • HTTP/CONNECT proxy           │
                    │ │  • SOCKS5 proxy                 │
                    │ │  • username:password auth       │
                    │ │  • PAC file at /proxy.pac       │
                    │ │                                 │ │
                    │ └─────────────────────────────────┘ │
                    │                                     │
                    │  /opt/proxy/                        │
                    │  ├── mp-proxy (binary)              │
                    │  ├── users.json (credentials)      │
                    │  ├── .env (configuration)          │
                    │  └── certs/                         │
                    │      ├── cert.pem (Let's Encrypt)  │
                    │      └── key.pem                    │
                    └─────────────────────────────────────┘
```

## Component Overview

### Entry Points

| Port | Domain | Service | Auth |
|------|--------|---------|------|
| 8443 | `proxy.yourdomain.com` | SNI TLS Proxy | None |
| 8080 | `private.yourdomain.com` | HTTP Proxy + PAC | Basic Auth |
| 1080 | `private.yourdomain.com` | SOCKS5 Proxy | User/Pass |
| 9090 | Internal | Metrics & Stats API | None |

### Core Modules

| Module | Path | Purpose |
|--------|------|---------|
| `proxy/` | `internal/proxy/` | SNI TLS proxy with host routing |
| `httpproxy/` | `internal/httpproxy/` | HTTP/HTTPS forward proxy |
| `socks5/` | `internal/socks5/` | SOCKS5 proxy (RFC 1928/1929) |
| `pac/` | `internal/pac/` | PAC file generation & serving |
| `auth/` | `internal/auth/` | User authentication & rate limiting |
| `config/` | `internal/config/` | Configuration loading |
| `ui/` | `internal/ui/` | CLI formatting & logging |

## Directory Structure

```
/opt/proxy/                    # Deployment location
├── mp-proxy                   # Compiled binary
├── users.json                 # User credentials
├── .env                       # Environment config
└── certs/
    ├── cert.pem              # Let's Encrypt fullchain
    └── key.pem               # Let's Encrypt private key
```

```
Source-Code/                   # Source code
├── cmd/proxy/main.go         # Entry point (mode switching)
├── internal/
│   ├── auth/                 # Authentication
│   ├── config/               # Configuration
│   ├── httpproxy/            # HTTP proxy
│   ├── pac/                  # PAC file serving
│   ├── proxy/                # SNI proxy
│   ├── socks5/               # SOCKS5 proxy
│   └── ui/                   # CLI formatting
├── docs/                     # Documentation
├── scripts/                  # Build & utility scripts
└── users.json                # User template
```

## Request Flow

### SNI Proxy Mode (Default)
```
Client App → TLS:8443 → SNI Detection → Route to target servers → Relay data
```

### HTTPS Proxy Mode
```
Browser → HTTP:8080 → Proxy-Authorization → Validate user → CONNECT → Target
                  ↓
            /proxy.pac → Return PAC file
```

### SOCKS5 Mode
```
Client → TCP:1080 → Negotiate auth → User/Pass → Connect → Target → Relay
```

## Security

- **TLS 1.2+** for all encrypted connections
- **Let's Encrypt** certificates with auto-renewal
- **bcrypt** password hashing (cost 10+)
- **Rate limiting** per user (token bucket)
- **IP whitelisting** (optional)
- **Systemd hardening** (NoNewPrivileges, ProtectSystem)

## Key Features

| Feature | Description |
|---------|-------------|
| Dual-mode | SNI TLS proxy or general HTTPS/SOCKS5 |
| PAC Support | Dynamic `/proxy.pac` with credential embedding |
| Metrics | Prometheus metrics at `:9090/metrics` |
| Stats API | JSON stats at `:9090/api/stats` |
| Auto-renew | Let's Encrypt certificate auto-renewal |
| Graceful shutdown | Clean connection draining on SIGTERM |
