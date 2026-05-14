# Multi-stage Dockerfile for Hades SOC V2.0 - Headless Sentinel
# Uses lightweight Alpine base to reduce attack surface
# Security-hardened with non-root user and seccomp profile

# Development build stage - Full development environment
FROM golang:1.25-alpine AS development

# Install development dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev \
    pkgconfig \
    curl \
    vim \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build hades binary for development
ARG VERSION=dev
ARG BUILD_TIME
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    -o hades \
    ./cmd/hades

# Install development tools
RUN go install -a github.com/air-verse/air@latest && \
    go install -a github.com/swaggo/swag/cmd/swag@latest

# Set environment for development
ENV GIN_MODE=debug
ENV HADES_ENV=development

# Expose ports
EXPOSE 8080 8081

# Development command
CMD ["./hades", "api"]

# Build stage - Go compiler with full toolchain
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev \
    pkgconfig \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Copy torgo for local replace directive
COPY torgo ./torgo

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build sentinel binary with security flags
ARG VERSION=dev
ARG BUILD_TIME
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    -o hades \
    ./cmd/hades

# Runtime stage - Minimal Alpine with security hardening
FROM alpine:3.19

# Install runtime dependencies only
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite \
    curl \
    && rm -rf /var/cache/apk/* \
    && addgroup -g 1001 -S hades \
    && adduser -u 1001 -S hades -G hades

# Set working directory
WORKDIR /app

# Create necessary directories with proper permissions
RUN mkdir -p /app/data /app/logs /app/config \
    && chown -R hades:hades /app

# Copy binary from builder stage
COPY --from=builder /build/hades /app/hades

# Copy configuration files
COPY --chown=hades:hades .env.production /app/.env

# Set ownership and permissions
RUN chmod +x /app/hades \
    && chmod 600 /app/.env

# Switch to non-root user
USER hades

# Expose metrics and health port
EXPOSE 2112

# Set environment variables for security
ENV GIN_MODE=release
ENV HADES_ENV=production

# Health check for Docker
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:2112/health || exit 1

# Default command
CMD ["/app/hades", "api"]
