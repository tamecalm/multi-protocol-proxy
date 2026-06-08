# Multi-Protocol Proxy Documentation

## Quick Links

| Topic | Document |
|-------|----------|
| **Getting Started** | [HTTPS_PROXY.md](HTTPS_PROXY.md) |
| **Architecture** | [ARCHITECTURE.md](ARCHITECTURE.md) |
| **AWS Deployment** | [aws/DEPLOYMENT.md](aws/DEPLOYMENT.md) |
| **GCP Deployment** | [gcp/DEPLOYMENT.md](gcp/DEPLOYMENT.md) |

## Example Ports & Endpoints

| Endpoint | Default URL/Port | Auth |
|----------|-----|------|
| SNI TLS Proxy | `proxy.yourdomain.com:8443` | None |
| HTTP Proxy | `private.yourdomain.com:8080` | Basic Auth |
| SOCKS5 Proxy | `private.yourdomain.com:1080` | User/Pass |
| PAC File | `https://private.yourdomain.com/proxy.pac` | Optional Token |
| Metrics Exporter | `http://localhost:9090/metrics` | None |

## Documentation Index

### Core Guides
- [HTTPS_PROXY.md](HTTPS_PROXY.md) - HTTP/HTTPS/SOCKS5 proxy quick start
- [ARCHITECTURE.md](ARCHITECTURE.md) - System design and components

### API Reference
- [api/PAC.md](api/PAC.md) - PAC (Proxy Auto-Config) endpoint
- [api/METRICS.md](api/METRICS.md) - Prometheus metrics and stats API

### Configuration
- [configuration/CONFIG.md](configuration/CONFIG.md) - Environment variables reference
- [configuration/USERS.md](configuration/USERS.md) - User management and passwords

### Deployment
- [aws/DEPLOYMENT.md](aws/DEPLOYMENT.md) - AWS EC2 deployment
- [gcp/DEPLOYMENT.md](gcp/DEPLOYMENT.md) - Google Cloud deployment
- [nginx/api-config.md](nginx/api-config.md) - Nginx reverse proxy

## Directory Structure

```
docs/
├── README.md              # This file
├── ARCHITECTURE.md        # System architecture
├── HTTPS_PROXY.md         # Quick start guide
├── api/
│   ├── PAC.md            # PAC endpoint
│   └── METRICS.md        # Metrics API
├── configuration/
│   ├── CONFIG.md         # Config reference
│   └── USERS.md          # User management
├── aws/
│   └── DEPLOYMENT.md     # AWS EC2 guide
├── gcp/
│   └── DEPLOYMENT.md     # GCP guide
└── nginx/
    └── api-config.md     # Nginx config
```
