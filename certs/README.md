# TLS Certificates

This directory contains TLS certificates for the Multi-Protocol Proxy.

> **⚠️ Security**: Certificate files (`.crt`, `.key`, `.pem`) are gitignored and should never be committed to version control.

---

## Development Certificates

For local development, use self-signed certificates.

### Generate Self-Signed Certificate

```bash
cd certs/dev

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key \
  -out server.crt \
  -subj "/C=US/ST=Development/L=Local/O=MultiProtocolProxy/CN=localhost"
```

### Generate with Interactive Prompts

```bash
cd certs/dev

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key \
  -out server.crt
```

You'll be prompted for:
- **Country Name**: e.g., `US`
- **State**: e.g., `California`
- **Locality**: e.g., `San Francisco`
- **Organization**: e.g., `My Company`
- **Common Name**: e.g., `localhost` or `127.0.0.1`
- **Email**: e.g., `dev@example.com`

### Generate with Subject Alternative Names (SAN)

For testing with multiple domains/IPs:

```bash
cd certs/dev

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key \
  -out server.crt \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:127.0.0.1"
```

---

## Production Certificates

For production, use certificates from a trusted Certificate Authority (CA).

### Option 1: Let's Encrypt (Recommended)

Free, automated, and trusted certificates.

```bash
# Install certbot
sudo apt install certbot  # Debian/Ubuntu
sudo dnf install certbot  # Fedora/RHEL

# Generate certificate (standalone mode)
sudo certbot certonly --standalone \
  -d proxy.yourdomain.com \
  --agree-tos \
  --email admin@yourdomain.com

# Certificates are saved to:
# /etc/letsencrypt/live/proxy.yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/proxy.yourdomain.com/privkey.pem

# Copy to prod folder
sudo cp /etc/letsencrypt/live/proxy.yourdomain.com/fullchain.pem certs/prod/server.crt
sudo cp /etc/letsencrypt/live/proxy.yourdomain.com/privkey.pem certs/prod/server.key
sudo chown $(whoami) certs/prod/*
```

### Option 2: Cloudflare Origin Certificates

If using Cloudflare as your CDN/proxy:

1. Go to **Cloudflare Dashboard** → **SSL/TLS** → **Origin Server**
2. Click **Create Certificate**
3. Choose **RSA (2048)** key type
4. Add your domain(s)
5. Download the certificate and private key
6. Save as `certs/prod/server.crt` and `certs/prod/server.key`

### Option 3: Commercial CA

Purchase from providers like DigiCert, Comodo, or Sectigo:

1. Generate a CSR (Certificate Signing Request):
   ```bash
   openssl req -new -newkey rsa:2048 -nodes \
     -keyout certs/prod/server.key \
     -out certs/prod/server.csr \
     -subj "/C=US/ST=State/L=City/O=Company/CN=proxy.yourdomain.com"
   ```

2. Submit the CSR to your CA provider
3. Download the issued certificate
4. Save as `certs/prod/server.crt`

---

## Configuration

Update `config.json` to point to the correct certificates:

### Development

```json
{
  "cert_file": "certs/dev/server.crt",
  "key_file": "certs/dev/server.key"
}
```

### Production

```json
{
  "cert_file": "certs/prod/server.crt",
  "key_file": "certs/prod/server.key"
}
```

---

## Verification

### Check Certificate Details

```bash
openssl x509 -in certs/dev/server.crt -text -noout
```

### Verify Certificate and Key Match

```bash
# These should output the same hash
openssl x509 -noout -modulus -in certs/dev/server.crt | openssl md5
openssl rsa -noout -modulus -in certs/dev/server.key | openssl md5
```

### Test TLS Connection

```bash
openssl s_client -connect localhost:8443 -servername localhost
```

---

## Security Best Practices

1. **Never commit certificates** — Keep `.crt`, `.key`, `.pem` files out of version control
2. **Restrict file permissions** — `chmod 600` for private keys
3. **Rotate regularly** — Renew certificates before expiration
4. **Use strong keys** — Minimum 2048-bit RSA or 256-bit ECDSA
5. **Monitor expiration** — Set up alerts for certificate expiry
