# Multi-Protocol Proxy

<p align="center">
  <a href="https://github.com/tamecalm/multi-protocol-proxy/actions/workflows/ci.yml"><img src="https://github.com/tamecalm/multi-protocol-proxy/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8.svg?style=flat&logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License"></a>
  <img src="https://img.shields.io/badge/PRs-Welcome-brightgreen.svg" alt="PRs Welcome">
  <a href="https://twitter.com/tamecalm"><img src="https://img.shields.io/twitter/follow/tamecalm?style=social" alt="Twitter Follow"></a>
</p>

Multi-Protocol Proxy is a high-performance, container-friendly, and secure dual-mode proxy server written in Go. It multiplexes public SNI-based TLS routing with a private, authenticated HTTP/HTTPS/SOCKS5 forwarding proxy. Designed for deployment on virtual private servers (such as AWS EC2 or GCP Compute Engine), it provides fine-grained user access control, bandwidth shaping, and production-ready Prometheus observability out of the box.

---

## Key Features

- **Dual-Mode Routing**:
  - **SNI TLS Proxy (Default)**: Public SNI-based TLS routing to configured target servers.
  - **Private Proxy**: Authenticated HTTP (CONNECT) and SOCKS5 (RFC 1928/1929) proxy.
- **Proxy Auto-Config (PAC)**: Serving dynamic `.pac` files at `/proxy.pac` with support for access tokens and automated credential embedding.
- **Access Control & Security**:
  - Secure `bcrypt` password hashing.
  - IP Address whitelisting for both users and administrators.
  - Rate limiting per user using token bucket.
  - Expiry dates for user accounts.
- **Traffic Shaping**:
  - Max concurrent connection limits per user.
  - Bandwidth speed throttling (Mbps) per user.
  - Monthly bandwidth usage caps (GB) with automated calendar-month resets.
- **Observability**:
  - High-fidelity Prometheus metrics endpoint (`/metrics`).
  - Active connection tracker and historical logs API (`/api/stats`, `/api/history`).
- **Developer-First CLI**: Interactive build, run, and user administration shell scripts.

---

## System Architecture

```
                               ┌─────────────────────────────────────┐
                               │           VPS / VM Instance         │
                               │         (Ubuntu / Docker / OS)      │
                               ├─────────────────────────────────────┤
    Internet ─────────────────►│ ┌─────────────────────────────────┐ │
                               │ │      Multi-Protocol Proxy       │ │
                               │ ├─────────────────────────────────┤ │
                               │ │                                 │ │
   SNI TLS Mode (:8443) ──────►│ │  SNI Router (:8443)             │─┼─► Targets
   (Public bypass proxy)       │ │  • TLS termination              │ │   (e.g., Target Servers)
                               │ │  • SNI-based routing            │ │
                               │ │                                 │ │
                               │ ├─────────────────────────────────┤ │
                               │ │                                 │ │
   Private Mode (:8080/:1080) ►│ │  Private Proxy                  │─┼─► Internet
   (Authenticated proxy)       │ │  • HTTP/CONNECT proxy (:8080)   │ │
                               │ │  • SOCKS5 proxy (:1080)         │ │
                               │ │  • PAC file at /proxy.pac       │ │
                               │ │  • Bandwidth limit & throttle   │ │
                               │ │                                 │ │
                               │ └─────────────────────────────────┘ │
                               └─────────────────────────────────────┘
```

---

## Quick Start

### 1. Prerequisites
- [Go 1.25](https://go.dev/dl/) or higher.
- [Docker & Docker Compose](https://docs.docker.com/) (Optional, for containerized deployments).

### 2. Local Setup
Clone the repository and run the unified installer to fetch Go dependencies:

```bash
git clone https://github.com/tamecalm/multi-protocol-proxy.git
cd multi-protocol-proxy
./scripts/start.sh install
```

### 3. Run the Interactive CLI
The project includes an interactive terminal UI for common tasks. Simply run:

```bash
./scripts/start.sh
```

You will be presented with a menu:
```
  1) Install Dependencies   — Download Go modules
  2) Build                  — Build for current platform
  3) Build All              — Build for all platforms
  4) Run                    — Run the proxy
  5) Build & Run            — Build then run (dev mode)
  6) Platform Info          — Show system information
  q) Quit
```

---

## Deployment & Production Setup

### A. Run with Docker Compose (Recommended)
Make sure you copy the example configuration, customize `users.json`, and run:

```bash
cp users.example.json users.json
docker compose up -d --build
```

### B. Deployment on Linux (AWS EC2 / GCP)
1. **Build for target platform**:
   ```bash
   ./scripts/build.sh --os linux --arch amd64
   ```
2. **Move files to host**:
   Deploy the binary `build/multi-protocol-proxy-linux-amd64` along with `.env` and `users.json` into `/opt/proxy/`.
3. **Configure systemd service**:
   We recommend running the proxy as a hardened systemd service. An example systemd config:
   ```ini
   [Unit]
   Description=Multi-Protocol Proxy
   After=network.target

   [Service]
   Type=simple
   User=proxy
   WorkingDirectory=/opt/proxy
   ExecStart=/opt/proxy/multi-protocol-proxy
   Restart=always
   EnvironmentFile=/opt/proxy/.env

   [Install]
   WantedBy=multi-user.target
   ```

---

## Configuration Reference

The proxy loads configuration variables from the `.env` file in the working directory.

### Core Environment Settings
| Variable | Default | Description |
| --- | --- | --- |
| `APP_ENV` | `development` | `development` or `production` |
| `PROXY_MODE` | `sni` | `sni` (TLS SNI proxy) or `https` (HTTP/SOCKS5 proxy) |
| `DOMAIN` | `localhost:8443` | Public domain name for TLS certificates and PAC endpoint |
| `CERT_FILE` | `certs/dev/server.crt` | Path to TLS certificate |
| `KEY_FILE` | `certs/dev/server.key` | Path to TLS private key |
| `HTTP_PROXY_PORT` | `:8080` | Port for HTTP proxy forwarding |
| `HTTP_PROXY_TLS_PORT` | `:8443` | Port for HTTPS proxy forwarding (TLS CONNECT) |
| `SOCKS5_PORT` | `:1080` | Port for SOCKS5 proxy |
| `METRICS_LISTEN` | `:9090` | Port for metrics and Stats API server |
| `USERS_FILE` | `users.json` | Path to user configuration JSON |

For more details, see [docs/configuration/CONFIG.md](docs/configuration/CONFIG.md).

---

## User Management

When running in `https` (Private Proxy) mode, credentials are loaded from `users.json`. You can manage users using the interactive CLI script:

```bash
go run scripts/manage-users.go
```

This tool allows you to:
1. Add new users with custom rate limits, speed limits, bandwidth limits, and expiration dates.
2. List all users and check their active statuses.
3. Enable, disable, or delete users.
4. Modify speed/bandwidth limits for existing users.
5. Generate password hashes safely using bcrypt.

### users.json Schema Example
```json
{
  "users": [
    {
      "username": "customer1",
      "role": "user",
      "password_hash": "$2a$10$3Y...hash...",
      "rate_limit_rpm": 500,
      "enabled": true,
      "plan": "starter",
      "bandwidth_limit_gb": 50,
      "bandwidth_speed_mbps": 10,
      "max_connections": 5,
      "expires_at": "2026-12-31T00:00:00Z"
    }
  ],
  "ip_whitelist": ["192.168.1.0/24"]
}
```

---

## PAC (Proxy Auto-Config) API

The proxy hosts an automatic proxy configuration script at `/proxy.pac`. Clients can query this endpoint to auto-route traffic.

### Usage Examples
```bash
# Get basic PAC (browser prompts for password)
curl https://proxy.yourdomain.com/proxy.pac?user=customer1

# Get PAC with embedded credentials
curl "https://proxy.yourdomain.com/proxy.pac?user=customer1&pass=password123"

# PAC with secret token authentication (if PAC_TOKEN is configured)
curl "https://proxy.yourdomain.com/proxy.pac?token=abc1234&user=customer1"
```

For more details, see [docs/api/PAC.md](docs/api/PAC.md).

---

## Testing the Proxy

### Test HTTP/HTTPS Proxy (CONNECT)
```bash
curl -x http://customer1:password123@localhost:8080 https://httpbin.org/ip
```

### Test SOCKS5 Proxy
```bash
curl --socks5 localhost:1080 --proxy-user customer1:password123 https://httpbin.org/ip
```

---

## Observability & Prometheus Metrics

Observability endpoints are served on `METRICS_LISTEN` (default `:9090`).

- **Prometheus Metrics**: `http://localhost:9090/metrics`
- **Proxy Stats API**: `http://localhost:9090/api/stats`
- **History API**: `http://localhost:9090/api/history`
- **Bandwidth Usage API**: `http://localhost:9090/api/usage`

For a complete metrics list, see [docs/api/METRICS.md](docs/api/METRICS.md).

---

## Contributing

We welcome contributors! Check out our [Contributing Guidelines](CONTRIBUTING.md) to learn how to set up the project locally and submit changes.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
