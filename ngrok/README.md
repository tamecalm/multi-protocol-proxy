# Ngrok Development Setup

Some client applications (like the Signal messenger client) do not support `localhost` for testing, so we use ngrok to expose your local development server with a public HTTPS URL.

> **📋 Quick Start**: Expose your local proxy to a public URL in 5 minutes for client testing.

---

## Prerequisites

- Active ngrok account ([sign up free](https://dashboard.ngrok.com/signup))
- ngrok CLI installed

---

## Installation

### Windows

```powershell
# Using Chocolatey
choco install ngrok

# Or using Scoop
scoop install ngrok

# Or download directly from https://ngrok.com/download
```

### macOS

```bash
# Using Homebrew
brew install ngrok/ngrok/ngrok
```

### Linux

```bash
# Debian/Ubuntu
curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc | \
  sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null && \
  echo "deb https://ngrok-agent.s3.amazonaws.com buster main" | \
  sudo tee /etc/apt/sources.list.d/ngrok.list && \
  sudo apt update && sudo apt install ngrok

# Or use snap
sudo snap install ngrok
```

---

## Setup

### 1. Configure Authtoken

Get your authtoken from the [ngrok dashboard](https://dashboard.ngrok.com/get-started/your-authtoken):

```bash
ngrok config add-authtoken YOUR_AUTHTOKEN
```

### 2. Copy Configuration Template

```bash
cd ngrok
cp ngrok.yml.example ngrok.yml
# Edit ngrok.yml with your settings
```

### 3. Start ngrok Tunnel

```bash
# Using config file
ngrok start --config ngrok.yml multi-protocol-proxy

# Or quick start (without config file)
ngrok tls 8443
```

### 4. Copy the Public URL

ngrok will display something like:

```
Forwarding   https://abc123xyz.ngrok.io -> localhost:8443
```

Use `abc123xyz.ngrok.io` as your domain in `.env`:

```bash
NGROK_ENABLED=true
NGROK_DOMAIN=abc123xyz.ngrok.io
DOMAIN=abc123xyz.ngrok.io
BASE_URL=https://abc123xyz.ngrok.io
```

---

## Reserved Domains (Recommended)

Free ngrok accounts get random URLs that change on every restart. For consistent testing:

### Option 1: Reserved Subdomain (Paid)

1. Go to [ngrok Reserved Domains](https://dashboard.ngrok.com/cloud-edge/domains)
2. Reserve a subdomain like `multi-protocol-proxy.ngrok.io`
3. Update `ngrok.yml`:
   ```yaml
   tunnels:
     multi-protocol-proxy:
       proto: tls
       addr: 8443
       domain: multi-protocol-proxy.ngrok.io
   ```

### Option 2: Custom Domain (Paid)

Use your own domain:

1. Reserve a custom domain in the ngrok dashboard
2. Add a CNAME record: `proxy-dev.yourdomain.com` → `your-tunnel.ngrok.io`
3. Configure in `ngrok.yml`:
   ```yaml
   tunnels:
     multi-protocol-proxy:
       proto: tls
       addr: 8443
       hostname: proxy-dev.yourdomain.com
   ```

---

## Development Workflow

### Standard Workflow

```bash
# Terminal 1: Start ngrok
cd multi-protocol-proxy/ngrok
ngrok start --config ngrok.yml multi-protocol-proxy

# Terminal 2: Start the proxy server
# (copy ngrok URL to .env first)
cd multi-protocol-proxy
./multi-protocol-proxy
```

### Environment Setup

Your `.env` should look like:

```bash
APP_ENV=development
DEBUG=true
LOG_LEVEL=debug

# Ngrok Configuration
NGROK_ENABLED=true
NGROK_DOMAIN=your-subdomain.ngrok.io

# Domain (uses ngrok domain in development)
DOMAIN=your-subdomain.ngrok.io
BASE_URL=https://your-subdomain.ngrok.io
```

---

## Web Inspection UI

ngrok provides a web interface for inspecting traffic:

- **URL**: http://127.0.0.1:4040
- **Features**:
  - View all requests/responses
  - Replay requests
  - Debug connection issues

---

## Troubleshooting

### "Tunnel session failed: Your account is limited"

**Cause**: Free tier limits reached.

**Solutions**:
- Wait for limit reset
- Upgrade to paid plan
- Use a different ngrok account

### "Failed to start tunnel: address already in use"

**Cause**: Another ngrok process is running.

**Solution**:
```bash
# Windows
taskkill /F /IM ngrok.exe

# macOS/Linux
pkill ngrok
```

### "TLS handshake error"

**Cause**: Certificate mismatch.

**Solutions**:
1. Ensure the proxy has valid dev certificates:
   ```bash
   cd certs/dev
   openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
     -keyout server.key -out server.crt \
     -subj "/CN=localhost"
     ```
2. Restart both ngrok and the proxy server

### Connection Drops After 2 Hours

**Cause**: Free tier session timeout.

**Solution**: Restart ngrok or upgrade to paid plan for persistent sessions.

---

## Security Notes

> **⚠️ Warning**: ngrok exposes your local server to the internet.

- Never run with production data
- Use ngrok's IP restrictions for sensitive testing
- Monitor the inspection UI (http://127.0.0.1:4040) for unexpected traffic
- Consider using ngrok's OAuth or IP policies for additional protection

---

## Useful Commands

```bash
# Check ngrok status
ngrok diagnose

# View active tunnels
ngrok api tunnels list

# Update ngrok
ngrok update

# Check version
ngrok version
```

---

## Resources

- [ngrok Documentation](https://ngrok.com/docs)
- [ngrok TLS Tunnels](https://ngrok.com/docs/tls)
- [ngrok Configuration](https://ngrok.com/docs/agent/config)
- [ngrok Pricing](https://ngrok.com/pricing)
