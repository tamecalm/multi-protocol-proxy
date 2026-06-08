# Metrics & Stats API

## Prometheus Metrics

**Endpoint:** `http://YOUR_EC2_IP:9090/metrics` (restricted to your IP in security group)

### HTTP Proxy Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `httpproxy_requests_total` | Counter | `username`, `method` | Total proxy requests |
| `httpproxy_active_connections` | Gauge | - | Current active connections |
| `httpproxy_bytes_total` | Counter | `username`, `direction` | Bytes transferred (upstream/downstream) |
| `httpproxy_duration_seconds` | Histogram | - | Request duration |
| `httpproxy_auth_failures_total` | Counter | `reason` | Auth failures by type |
| `httpproxy_rate_limited_total` | Counter | `username` | Rate limit hits |
| `httpproxy_errors_total` | Counter | `type` | Errors by type |

### SOCKS5 Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `socks5_connections_total` | Counter | `username` | Total connections |
| `socks5_active_connections` | Gauge | - | Active connections |
| `socks5_bytes_total` | Counter | `username`, `direction` | Bytes transferred |
| `socks5_duration_seconds` | Histogram | - | Connection duration |
| `socks5_auth_failures_total` | Counter | `reason` | Auth failures |
| `socks5_rate_limited_total` | Counter | `username` | Rate limits |
| `socks5_errors_total` | Counter | `type` | Errors |

### SNI Proxy Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `multiproxy_active_conns` | Gauge | - | Active connections |
| `multiproxy_relay_total` | Counter | `sni` | Relayed by SNI |
| `multiproxy_bytes_total` | Counter | `direction` | Bytes transferred |
| `multiproxy_errors_total` | Counter | `type` | Errors |

## JSON Stats API

### GET /api/stats

**URL:** `http://YOUR_EC2_IP:9090/api/stats`

**Response:**
```json
{
  "totalUsers": 1523,
  "activeConnections": 42,
  "uptimeSeconds": 86400,
  "dataThroughput": "15.2 MB/s",
  "latency": 18,
  "successRate": 99.8
}
```

### GET /api/history

**URL:** `http://YOUR_EC2_IP:9090/api/history`

24-hour historical data:
```json
[
  {"time": "00:00", "users": 150, "traffic": 1073741824},
  {"time": "01:00", "users": 120, "traffic": 858993459}
]
```

## Access on AWS EC2

The metrics port (9090) should be restricted in your security group:

| Type | Port | Source | Description |
|------|------|--------|-------------|
| Custom TCP | 9090 | My IP | Metrics access |

**SSH tunnel for secure access:**
```bash
ssh -i your-key.pem -L 9090:localhost:9090 ubuntu@YOUR_EC2_IP
# Then access: http://localhost:9090/metrics
```

## Grafana Integration

Example PromQL queries:

```promql
# Request rate
rate(httpproxy_requests_total[5m])

# Active connections
httpproxy_active_connections + socks5_active_connections

# Bandwidth
rate(httpproxy_bytes_total[5m])

# Error rate
rate(httpproxy_errors_total[5m]) / rate(httpproxy_requests_total[5m])
```
