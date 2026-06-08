# PAC Endpoint API

## Overview

The PAC (Proxy Auto-Config) endpoint serves JavaScript files that browsers and systems use to automatically configure proxy settings for `private.yourdomain.com`.

**Endpoint:** `https://private.yourdomain.com/proxy.pac`

## Usage Examples

### Basic PAC (Browser Prompts for Password)
```bash
curl https://private.yourdomain.com/proxy.pac?user=tamecalm
```

### PAC with Embedded Credentials
```bash
curl "https://private.yourdomain.com/proxy.pac?user=tamecalm&pass=yourpassword"
```

### With Access Token (if configured)
```bash
curl "https://private.yourdomain.com/proxy.pac?token=abc1234&user=tamecalm"
```

## Query Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `user` | Yes* | Username for the proxy. Required unless `PAC_DEFAULT_USER` is set |
| `pass` | No | Password to embed in PAC file. If omitted, browser will prompt |
| `token` | No | Required only if `PAC_TOKEN` environment variable is set |

## Response

### Headers
```http
Content-Type: application/x-ns-proxy-autoconfig
Cache-Control: no-cache, no-store, must-revalidate
Access-Control-Allow-Origin: *
```

### PAC File Content
```javascript
function FindProxyForURL(url, host) {
    // Don't proxy local addresses
    if (isPlainHostName(host) ||
        shExpMatch(host, "*.local") ||
        isInNet(host, "192.168.0.0", "255.255.0.0") ||
        isInNet(host, "10.0.0.0", "255.0.0.0") ||
        isInNet(host, "172.16.0.0", "255.240.0.0") ||
        host == "localhost" ||
        host == "127.0.0.1") {
        return "DIRECT";
    }
    
    // Route through private.yourdomain.com proxy
    return "PROXY tamecalm:password@private.yourdomain.com:8080; SOCKS5 tamecalm:password@private.yourdomain.com:1080; DIRECT";
}
```

## Client Configuration

### macOS
1. **System Preferences** → **Network** → Select connection → **Advanced**
2. **Proxies** tab → Check **Automatic Proxy Configuration**
3. URL: `https://private.yourdomain.com/proxy.pac?user=YOUR_USER`

### Windows 10/11
1. **Settings** → **Network & Internet** → **Proxy**
2. Enable **Use setup script**
3. Script address: `https://private.yourdomain.com/proxy.pac?user=YOUR_USER`

### iOS
1. **Settings** → **Wi-Fi** → Tap **(i)** on your network
2. **Configure Proxy** → **Automatic**
3. URL: `https://private.yourdomain.com/proxy.pac?user=YOUR_USER`

### Firefox
1. **Settings** → **Network Settings** → **Settings...**
2. Select **Automatic proxy configuration URL**
3. Enter: `https://private.yourdomain.com/proxy.pac?user=YOUR_USER`

### Chrome (uses system proxy)
Configure at OS level (Windows/macOS settings above)

## Error Responses

| Status | Cause |
|--------|-------|
| 401 | Invalid token (when `PAC_TOKEN` is configured) |
| 401 | Invalid credentials (when `pass` parameter provided but wrong) |
| 429 | Rate limit exceeded |

## Server Configuration

Add to `/opt/proxy/.env`:

```bash
# Enable PAC endpoint
PAC_ENABLED=true

# Optional: Require token for PAC access
PAC_TOKEN=

# Optional: Default user if no ?user= param
PAC_DEFAULT_USER=tamecalm

# Rate limit (requests per minute)
PAC_RATE_LIMIT_RPM=60
```

After changing, restart the service:
```bash
sudo systemctl restart proxy
```
