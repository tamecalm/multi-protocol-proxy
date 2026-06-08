# User Management

## Overview

Users are stored in `/opt/proxy/users.json` on your EC2 instance. The proxy uses bcrypt password hashing for secure credential storage.

## users.json Structure

```json
{
  "users": [
    {
      "username": "tamecalm",
      "password_hash": "$2a$10$v.YxawzckkkBZtE13QFhl.P7nF1DCsNswkgSQu4S3JOYuJWxKSuJO",
      "rate_limit_rpm": 500,
      "enabled": true
    },
    {
      "username": "friend1",
      "password_hash": "$2a$10$THEIR_BCRYPT_HASH",
      "rate_limit_rpm": 200,
      "enabled": true
    }
  ],
  "ip_whitelist": []
}
```

| Field | Type | Description |
|-------|------|-------------|
| `username` | string | Unique username (case-insensitive) |
| `password_hash` | string | bcrypt hash (cost 10+) |
| `rate_limit_rpm` | int | Requests per minute (0 = unlimited) |
| `enabled` | bool | Account active status |
| `ip_whitelist` | array | CIDR ranges to allow (empty = all) |

---

## Creating Password Hashes

### Option 1: Using the included script

```bash
# On your local machine
go run scripts/hash-password.go
# Enter password when prompted
# Copy the hash to users.json
```

### Option 2: Using htpasswd

```bash
htpasswd -nbB username password | cut -d: -f2
```

### Option 3: Using Python

```python
import bcrypt
password = b"your-password"
hash = bcrypt.hashpw(password, bcrypt.gensalt(rounds=10))
print(hash.decode())
```

---

## Adding a New User

1. **Generate password hash locally:**
   ```bash
   go run scripts/hash-password.go
   ```

2. **SSH to EC2 and edit users.json:**
   ```bash
   ssh -i your-key.pem ubuntu@YOUR_EC2_IP
   sudo nano /opt/proxy/users.json
   ```

3. **Add the new user:**
   ```json
   {
     "username": "newuser",
     "password_hash": "$2a$10$YOUR_GENERATED_HASH",
     "rate_limit_rpm": 100,
     "enabled": true
   }
   ```

4. **Restart the proxy:**
   ```bash
   sudo systemctl restart proxy
   ```

---

## IP Whitelisting

Restrict proxy access to specific IPs:

```json
{
  "ip_whitelist": [
    "203.0.113.50",
    "192.168.1.0/24",
    "10.0.0.0/8"
  ]
}
```

- Empty array = allow all IPs
- Supports CIDR notation
- Single IPs auto-convert to /32

---

## Rate Limiting

Each user has an individual rate limit:

| Setting | Effect |
|---------|--------|
| `"rate_limit_rpm": 500` | 500 requests per minute |
| `"rate_limit_rpm": 0` | Unlimited |

Users exceeding their limit receive HTTP 429 (Too Many Requests).

---

## Disabling a User

Set `enabled` to `false`:

```json
{
  "username": "blocked-user",
  "enabled": false
}
```

Restart proxy for changes to take effect.

---

## Security Best Practices

1. **Strong passwords** - Use 16+ characters
2. **Unique hashes** - Generate fresh hash per user
3. **Rate limits** - Set appropriate limits per user
4. **Regular rotation** - Change passwords periodically
5. **Monitor logs** - Check `journalctl -u proxy` for auth failures
