# Google Cloud Compute Engine Deployment Guide

Complete step-by-step guide to deploy Multi-Protocol Proxy on GCP. This guide covers both **SNI Proxy** (public, at `proxy.yourdomain.com`) and **Private HTTPS/SOCKS5 Proxy** (at `private.yourdomain.com`).

---

## Overview

| Mode | Domain | Ports | Authentication |
|------|--------|-------|----------------|
| SNI Proxy | `proxy.yourdomain.com` | 8443 | None (public) |
| Private Proxy | `private.yourdomain.com` | 8080, 1080 | Username/Password |

---

## Part 1: Create VM Instance

### Step 1.1: Open Compute Engine

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Navigate to **Compute Engine → VM Instances**
3. Click **Create Instance**

### Step 1.2: Configure Instance

| Setting | Value |
|---------|-------|
| **Name** | `multi-protocol-proxy` |
| **Region** | Choose closest to your users |
| **Zone** | Any available |
| **Machine type** | `e2-micro` (free tier) or `e2-small` |

### Step 1.3: Boot Disk

Click **Change**:
- **Operating system**: Ubuntu
- **Version**: Ubuntu 24.04 LTS
- **Size**: 10 GB (or 20 GB for logs)

Click **Select**

### Step 1.4: Firewall

Check both:
- ✅ **Allow HTTP traffic**
- ✅ **Allow HTTPS traffic**

### Step 1.5: Create

Click **Create** and wait for the instance to start (green checkmark).

---

## Part 2: Reserve Static IP

Prevents IP from changing on restart.

### Step 2.1: Reserve IP

1. Go to **VPC Network → IP Addresses**
2. Click **Reserve External Static Address**
3. Configure:
   - **Name**: `multi-protocol-proxy-ip`
   - **Region**: Same as your VM
   - **Attached to**: Select your VM instance
4. Click **Reserve**

**Note your static IP** - you'll need it for DNS.

---

## Part 3: Configure Firewall Rules

### Step 3.1: Open Firewall

1. Go to **VPC Network → Firewall**
2. Click **Create Firewall Rule**

### Step 3.2: Create Rules

#### For SNI Proxy
| Field | Value |
|-------|-------|
| **Name** | `allow-multi-protocol-proxy` |
| **Targets** | All instances in the network |
| **Source IP ranges** | `0.0.0.0/0` |
| **Protocols and ports** | TCP: `8443, 9090` |

#### For Private Proxy
| Field | Value |
|-------|-------|
| **Name** | `allow-private-proxy` |
| **Targets** | All instances in the network |
| **Source IP ranges** | `0.0.0.0/0` |
| **Protocols and ports** | TCP: `8080, 1080, 8443, 9090` |

Click **Create** for each rule.

---

## Part 4: Configure DNS

Add **A Records** in your DNS provider (Cloudflare, Namecheap, etc.):

**For SNI Proxy:**
```
proxy.yourdomain.com → YOUR_STATIC_IP
```

**For Private Proxy:**
```
private.yourdomain.com → YOUR_STATIC_IP
```

---

## Part 5: Build and Upload Binary

### Step 5.1: Build the Linux Binary

Open **Git Bash** or **WSL** in your project directory and use the build script:

```bash
# Build for Linux (production server)
./scripts/build.sh --os linux --arch amd64
```

This creates `build/multi-protocol-proxy-linux-amd64` ready for upload.

> **Tip:** Run `./scripts/build.sh --help` to see all build options including `--all` for all platforms.

### Step 5.2: Upload via Cloud Console

**Easiest method - use browser upload:**

1. Go to **Compute Engine → VM Instances**
2. Click **SSH** button next to your instance (opens browser terminal)
3. Click the **gear icon** (⚙️) → **Upload file**
4. Upload: `build/multi-protocol-proxy-linux-amd64`, `users.json`

**Alternative - use gcloud SCP:**
```bash
gcloud compute scp build/multi-protocol-proxy-linux-amd64 multi-protocol-proxy:~ --zone=YOUR_ZONE
gcloud compute scp users.json multi-protocol-proxy:~ --zone=YOUR_ZONE
```

---

## Part 6: SSH and Server Setup

### Step 6.1: Connect via SSH

**Option A: Browser SSH (Easiest)**
1. Go to **Compute Engine → VM Instances**
2. Click **SSH** button next to your instance

**Option B: gcloud CLI**
```bash
gcloud compute ssh multi-protocol-proxy --zone=YOUR_ZONE
```

### Step 6.2: Initial Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Create directories
sudo mkdir -p /opt/proxy
sudo mkdir -p /opt/proxy/certs

# Move uploaded files
sudo mv ~/multi-protocol-proxy-linux-amd64 /opt/proxy/multi-protocol-proxy
sudo mv ~/users.json /opt/proxy/
sudo chmod +x /opt/proxy/multi-protocol-proxy

# Create service user
sudo useradd -r -s /bin/false proxy
sudo chown -R proxy:proxy /opt/proxy
```

---

## Part 7: Get SSL Certificates

### Step 7.1: Install Certbot

```bash
sudo apt install -y certbot
```

### Step 7.2: Get Certificate

**For SNI Proxy:**
```bash
sudo certbot certonly --standalone -d proxy.yourdomain.com
```

**For Private Proxy:**
```bash
sudo certbot certonly --standalone -d private.yourdomain.com
```

### Step 7.3: Copy Certificates

**For SNI Proxy:**
```bash
sudo cp /etc/letsencrypt/live/proxy.yourdomain.com/fullchain.pem /opt/proxy/certs/cert.pem
sudo cp /etc/letsencrypt/live/proxy.yourdomain.com/privkey.pem /opt/proxy/certs/key.pem
sudo chown -R proxy:proxy /opt/proxy/certs
```

**For Private Proxy:**
```bash
sudo cp /etc/letsencrypt/live/private.yourdomain.com/fullchain.pem /opt/proxy/certs/cert.pem
sudo cp /etc/letsencrypt/live/private.yourdomain.com/privkey.pem /opt/proxy/certs/key.pem
sudo chown -R proxy:proxy /opt/proxy/certs
```

---

## Part 8: Configure the Proxy

### Step 8.1: Create Environment File

```bash
sudo nano /opt/proxy/.env
```

**For SNI Proxy (`proxy.yourdomain.com`):**
```bash
APP_ENV=production
PROXY_MODE=sni
DOMAIN=proxy.yourdomain.com

# TLS Certificates
CERT_FILE=/opt/proxy/certs/cert.pem
KEY_FILE=/opt/proxy/certs/key.pem

# Ports
LISTEN=:8443
METRICS_LISTEN=:9090
```

**For Private Proxy (`private.yourdomain.com`):**
```bash
APP_ENV=production
PROXY_MODE=https
DOMAIN=private.yourdomain.com

# TLS Certificates
CERT_FILE=/opt/proxy/certs/cert.pem
KEY_FILE=/opt/proxy/certs/key.pem

# Proxy Ports
HTTP_PROXY_PORT=:8080
HTTP_PROXY_TLS=true
HTTP_PROXY_TLS_PORT=:8443
SOCKS5_PORT=:1080

# Users
USERS_FILE=/opt/proxy/users.json
METRICS_LISTEN=:9090
```

Press `Ctrl+O` to save, `Ctrl+X` to exit.

### Step 8.2: Configure Users (Private Proxy Only)

```bash
sudo nano /opt/proxy/users.json
```

```json
{
  "users": [
    {
      "username": "tamecalm",
      "password_hash": "$2a$10$YOUR_BCRYPT_HASH_HERE",
      "rate_limit_rpm": 500,
      "enabled": true
    },
    {
      "username": "friend1",
      "password_hash": "$2a$10$THEIR_HASH_HERE",
      "rate_limit_rpm": 200,
      "enabled": true
    }
  ],
  "ip_whitelist": []
}
```

> **Generate password hash** on your local machine: `go run scripts/hash-password.go`

---

## Part 9: Create Systemd Service

```bash
sudo nano /etc/systemd/system/proxy.service
```

```ini
[Unit]
Description=Multi-Protocol Proxy Server
After=network.target

[Service]
Type=simple
User=proxy
Group=proxy
WorkingDirectory=/opt/proxy
EnvironmentFile=/opt/proxy/.env
ExecStart=/opt/proxy/multi-protocol-proxy
Restart=always
RestartSec=5

# Security
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/opt/proxy

[Install]
WantedBy=multi-user.target
```

---

## Part 10: Start and Enable Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable auto-start on boot
sudo systemctl enable proxy

# Start the service
sudo systemctl start proxy

# Check status
sudo systemctl status proxy
```

You should see: **Active: active (running)**

---

## Part 11: Verify Deployment

### Check Logs
```bash
sudo journalctl -u proxy -f
```

### Test SNI Proxy
Open a browser or use curl:
```bash
curl -v https://proxy.yourdomain.com:8443
```

Configure a client application to use the proxy domain `proxy.yourdomain.com` on port `8443`.

### Test Private Proxy
```bash
# Via domain
curl -x http://tamecalm:yourpassword@private.yourdomain.com:8080 https://httpbin.org/ip

# Via IP (also works)
curl -x http://tamecalm:yourpassword@YOUR_STATIC_IP:8080 https://httpbin.org/ip

# SOCKS5
curl --socks5 private.yourdomain.com:1080 --proxy-user tamecalm:yourpassword https://httpbin.org/ip
```

### Check Metrics
```bash
curl http://YOUR_STATIC_IP:9090/metrics | head -20
```

---

## Part 12: Auto-Renew Certificates

```bash
# Add renewal cron job
sudo crontab -e
```

Add this line:
```
0 3 * * * certbot renew --quiet && cp /etc/letsencrypt/live/*/fullchain.pem /opt/proxy/certs/cert.pem && cp /etc/letsencrypt/live/*/privkey.pem /opt/proxy/certs/key.pem && systemctl restart proxy
```

---

## Quick Reference

### SNI Proxy (proxy.yourdomain.com)
Point clients to:
```
Server: proxy.yourdomain.com
Port: 8443
```

### Private Proxy (private.yourdomain.com)

**Browser Settings:**
- HTTP Proxy: `private.yourdomain.com`
- Port: `8080`
- Enter username/password when prompted

**curl:**
```bash
curl -x http://user:pass@private.yourdomain.com:8080 https://example.com
```

**Mobile (WiFi Settings):**
- Proxy: Manual
- Server: `private.yourdomain.com`
- Port: `8080`
- Authentication: your username/password

**Direct IP:**
```bash
curl -x http://user:pass@YOUR_STATIC_IP:8080 https://example.com
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Service won't start | `sudo journalctl -u proxy -n 50` |
| Connection refused | Check firewall rules in VPC Network |
| Certificate error | Re-run certbot, verify DNS propagation |
| Auth failed | Check password hash in users.json |
| Port not reachable | Verify firewall rule includes correct port |

### Useful Commands

```bash
# View live logs
sudo journalctl -u proxy -f

# Restart service
sudo systemctl restart proxy

# Check listening ports
sudo ss -tlnp | grep -E '8080|1080|8443'

# Test locally
curl http://localhost:8080
```
