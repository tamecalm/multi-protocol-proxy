# ==========================================
# Builder Stage
# ==========================================
FROM golang:1.25.4-alpine AS builder

# Install system dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build static binary
# CGO_ENABLED=0 compiles all dependencies statically into the binary
# -ldflags="-s -w" shrinks the binary size by stripping debug symbols
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/multi-protocol-proxy \
    ./cmd/proxy

# ==========================================
# Final Run Stage
# ==========================================
FROM alpine:3.19

# Install certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S proxygroup && adduser -S proxyuser -G proxygroup

# Set working directory
WORKDIR /opt/proxy

# Copy binary from builder
COPY --from=builder /app/multi-protocol-proxy /opt/proxy/multi-protocol-proxy
COPY --from=builder /app/config.json /opt/proxy/config.json
COPY --from=builder /app/users.example.json /opt/proxy/users.json

# Ensure correct permissions
RUN chown -R proxyuser:proxygroup /opt/proxy

# Use non-root user
USER proxyuser

# Expose proxy ports
# 8080: HTTP Proxy / PAC
# 8443: HTTPS Proxy / SNI TLS Proxy
# 1080: SOCKS5 Proxy
# 9090: Prometheus Metrics & Stats API
EXPOSE 8080 8443 1080 9090

# Set environment variables defaults
ENV APP_ENV=production \
    PROXY_MODE=sni \
    HTTP_PROXY_PORT=:8080 \
    HTTP_PROXY_TLS=true \
    HTTP_PROXY_TLS_PORT=:8443 \
    SOCKS5_PORT=:1080 \
    METRICS_LISTEN=:9090 \
    USERS_FILE=/opt/proxy/users.json

# Run the proxy
ENTRYPOINT ["/opt/proxy/multi-protocol-proxy"]
