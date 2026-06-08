# AWS EC2 Deployment Guide

Complete step-by-step guide to deploy Multi-Protocol Proxy on AWS EC2. This guide covers both **SNI Proxy** (public, at `proxy.yourdomain.com`) and **Private HTTPS/SOCKS5 Proxy** (at `private.yourdomain.com`).

---

## Overview

| Mode | Domain | Ports | Authentication |
|------|--------|-------|----------------|
| SNI Proxy | `proxy.yourdomain.com` | 8443 | None (public) |
| Private Proxy | `private.yourdomain.com` | 8080, 1080 | Username/Password |

---

## Part 1: Create EC2 Instance

### Step 1.1: Launch Instance

1. Open [AWS Console → EC2](https://console.aws.amazon.com/ec2)
2. Click **Launch Instance**

### Step 1.2: Configure Instance

| Setting | Value |
|---------|-------|
| **Name** | `multi-protocol-proxy` |
| **AMI** | Ubuntu Server 24.04 LTS |
| **Instance type** | `t3.micro` (or `t3.small` for better performance) |
| **Key pair** | Create new → Download `.pem` file → Save securely |

### Step 1.3: Network Settings

Click **Edit** and configure:

- ✅ **Allow SSH traffic** from My IP
- ✅ **Allow HTTPS traffic** from the internet
- ✅ **Allow HTTP traffic** from the internet

### Step 1.4: Storage

- Keep default 8 GB or increase to 20 GB for logs

### Step 1.5: Launch

Click **Launch Instance** → Wait for "Running" status

---

## Part 2: Configure Security Group

### Step 2.1: Find Security Group

1. Go to **EC2 → Instances**
2. Click your instance
3. Click the **Security** tab
4. Click the security group link (e.g., `sg-xxxxx`)

### Step 2.2: Add Inbound Rules

Click **Edit inbound rules** → **Add rule** for each:

#### For SNI Proxy (`proxy.yourdomain.com`)
| Type | Port | Source | Description |
|------|------|--------|-------------|
| Custom TCP | 8443 | 0.0.0.0/0 | SNI TLS Proxy |
| Custom TCP | 9090 | My IP | Metrics (optional) |

#### For Private Proxy (`private.yourdomain.com`)
| Type | Port | Source | Description |
|------|------|--------|-------------|
| Custom TCP | 8080 | 0.0.0.0/0 | HTTP Proxy |
| Custom TCP | 1080 | 0.0.0.0/0 | SOCKS5 Proxy |
| Custom TCP | 8443 | 0.0.0.0/0 | HTTPS Proxy (TLS) |
| Custom TCP | 9090 | My IP | Metrics |

Click **Save rules**

---

## Part 3: Assign Elastic IP (Recommended)

Static IP prevents IP changes on restart.

1. Go to **EC2 → Elastic IPs**
2. Click **Allocate Elastic IP address** → **Allocate**
3. Select the new IP → **Actions → Associate Elastic IP address**
4. Select your instance → **Associate**

**Note your Elastic IP** - you'll use this for DNS.

---

## Part 4: Configure DNS

### For SNI Proxy
Add an **A Record** in your DNS provider:
```
proxy.yourdomain.com → YOUR_ELASTIC_IP
```

### For Private Proxy
Add an **A Record**:
```
private.yourdomain.com → YOUR_ELASTIC_IP
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

### Step 5.2: Upload Files

Use your preferred method:

**Option A: Using SCP (Git Bash or WSL)**
```bash
EC2_IP="YOUR_ELASTIC_IP"
KEY="path/to/your-key.pem"

scp -i $KEY build/multi-protocol-proxy-linux-amd64 ubuntu@$EC2_IP:~
scp -i $KEY users.json ubuntu@$EC2_IP:~
```

**Option B: Using WinSCP**
1. Download [WinSCP](https://winscp.net/)
2. Connect with: Host = Elastic IP, Username = ubuntu, Key = your .pem
3. Drag and drop: `build/multi-protocol-proxy-linux-amd64`, `users.json`

**Option C: Using FileZilla**
1. Edit → Settings → SFTP → Add key file (.pem)
2. Connect: sftp://ubuntu@YOUR_ELASTIC_IP
3. Upload files from `build/` folder

---

## Part 6: SSH and Server Setup

### Step 6.1: Connect via SSH

**Windows (PowerShell):**
```powershell
ssh -i "path\to\your-key.pem" ubuntu@YOUR_ELASTIC_IP
```

**If permission error on Windows:**
```powershell
icacls "path\to\your-key.pem" /inheritance:r /grant:r "$($env:USERNAME):(R)"
```

### Step 6.2: Initial Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Create directories
sudo mkdir -p /opt/proxy
sudo mkdir -p /opt/proxy/certs

# Move files
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

> **Generate password hash** locally: `go run scripts/hash-password.go`

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
NoNewPrivileges=false   # Must be false to allow capabilities
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
ProtectSystem=strict
ReadWritePaths=/opt/proxy
ReadOnlyPaths=/opt

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

---

## Part 11: Verify Deployment

### Check Logs
```bash
sudo journalctl -u proxy -f
```

### Test SNI Proxy
From your local machine:
```bash
# Verify the TLS listener is responding
curl -v https://proxy.yourdomain.com:8443
```

Configure client apps to route TLS connections through `proxy.yourdomain.com` on port `8443`.

### Test Private Proxy
```bash
# HTTP Proxy
curl -x http://tamecalm:yourpassword@private.yourdomain.com:8080 https://httpbin.org/ip

# SOCKS5
curl --socks5 private.yourdomain.com:1080 --proxy-user tamecalm:yourpassword https://httpbin.org/ip

# Direct IP also works
curl -x http://tamecalm:yourpassword@YOUR_ELASTIC_IP:8080 https://httpbin.org/ip
```

### Check Metrics
```bash
curl http://YOUR_ELASTIC_IP:9090/metrics | head -20
```

---

## Part 12: Auto-Renew Certificates

```bash
# Test renewal
sudo certbot renew --dry-run

# Add cron job for auto-renewal
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

**Browser:**
- HTTP Proxy: `private.yourdomain.com:8080`
- Username/Password when prompted

**curl:**
```bash
curl -x http://user:pass@private.yourdomain.com:8080 https://example.com
```

**Mobile:**
- WiFi Settings → Proxy → Manual
- Server: `private.yourdomain.com`
- Port: `8080`

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Service won't start | `sudo journalctl -u proxy -n 50` |
| Connection refused | Check security group rules |
| Certificate error | Re-run certbot, check file permissions |
| Auth failed | Verify password hash in users.json |
