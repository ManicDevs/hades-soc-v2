# HADES-V2 Dockerfile - Pure Go Edition (No CGO)
# Multi-stage build for minimal, fully cross-platform binaries

# ============= Build Stage =============
FROM golang:1.25-alpine AS builder

# Install build dependencies (no CGO needed)
RUN apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates

WORKDIR /build

# Copy go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO disabled for full cross-platform compatibility
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=unknown
ARG BUILD_TIME=unknown

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

RUN go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -trimpath \
    -o /build/hades \
    ./cmd/hades

# Verify the binary is statically linked
RUN echo "=== Binary Info ===" \
    && file /build/hades \
    && echo "=== Checking for dynamic dependencies ===" \
    && (ldd /build/hades 2>/dev/null | head -5 || echo "Statically linked (no ldd output)")

# ============= Runtime Stage =============
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1000 hades && adduser -u 1000 -G hades -s /bin/sh -D hades

# Copy binary from builder
COPY --from=builder /build/hades /usr/local/bin/hades
RUN chmod +x /usr/local/bin/hades

# Create directory structure
RUN mkdir -p /opt/hades/{config,data,logs} \
    && chown -R hades:hades /opt/hades

# Copy configuration template (if exists)
COPY --chown=hades:hades deploy/hades.yaml /etc/hades/hades.yaml 2>/dev/null || true

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8443/api/health || exit 1

# Expose ports
EXPOSE 8443 8080 2222

# Volume mounts
VOLUME ["/opt/hades/data", "/opt/hades/config", "/var/log/hades"]

USER hades

WORKDIR /opt/hades

ENTRYPOINT ["hades"]
CMD ["api", "start", "--port", "8443"]