# Configuration Reference

## Environment Variables

Configure in `/opt/proxy/.env` on your EC2 instance.

### Core Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | `development` or `production` |
| `PROXY_MODE` | `sni` | `sni` for SNI proxy, `https` for private proxy |
| `DOMAIN` | `localhost` | Your domain (e.g., `private.yourdomain.com`) |
| `DEBUG` | `false` | Enable debug logging |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

### TLS Certificates

| Variable | Default | Description |
|----------|---------|-------------|
| `CERT_FILE` | `certs/dev/server.crt` | Path to certificate (Let's Encrypt fullchain.pem) |
| `KEY_FILE` | `certs/dev/server.key` | Path to private key (Let's Encrypt privkey.pem) |

### Proxy Ports

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PROXY_PORT` | `:8080` | HTTP proxy listen port |
| `HTTP_PROXY_TLS` | `true` | Enable HTTPS proxy |
| `HTTP_PROXY_TLS_PORT` | `:8443` | HTTPS proxy listen port |
| `SOCKS5_PORT` | `:1080` | SOCKS5 listen port |
| `METRICS_LISTEN` | `:9090` | Metrics server port |

### Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `USERS_FILE` | `users.json` | Path to user credentials file |

### PAC Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PAC_ENABLED` | `true` | Enable PAC endpoint at `/proxy.pac` |
| `PAC_TOKEN` | *(empty)* | Secret token for PAC access (optional) |
| `PAC_DEFAULT_USER` | *(empty)* | Default username for PAC requests |
| `PAC_RATE_LIMIT_RPM` | `60` | Rate limit for PAC endpoint |

---

## Production Configuration

### SNI Proxy (`proxy.yourdomain.com`)

`/opt/proxy/.env`:
```bash
APP_ENV=production
PROXY_MODE=sni
DOMAIN=proxy.yourdomain.com

CERT_FILE=/opt/proxy/certs/cert.pem
KEY_FILE=/opt/proxy/certs/key.pem

LISTEN=:8443
METRICS_LISTEN=:9090
```

### Private Proxy (`private.yourdomain.com`)

`/opt/proxy/.env`:
```bash
APP_ENV=production
PROXY_MODE=https
DOMAIN=private.yourdomain.com

CERT_FILE=/opt/proxy/certs/cert.pem
KEY_FILE=/opt/proxy/certs/key.pem

HTTP_PROXY_PORT=:8080
HTTP_PROXY_TLS=true
HTTP_PROXY_TLS_PORT=:8443
SOCKS5_PORT=:1080

USERS_FILE=/opt/proxy/users.json
METRICS_LISTEN=:9090

# PAC settings
PAC_ENABLED=true
PAC_DEFAULT_USER=tamecalm
PAC_RATE_LIMIT_RPM=60
```

---

## Apply Configuration

After editing `/opt/proxy/.env`:

```bash
sudo systemctl restart proxy
sudo systemctl status proxy
```

Check logs if issues:
```bash
sudo journalctl -u proxy -f
```
