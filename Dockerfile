# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata curl

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hades ./cmd/hades

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN adduser -D -s /bin/sh hades

WORKDIR /app

# Copy binary and web assets
COPY --from=builder /app/hades .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config ./config

# Create necessary directories
RUN mkdir -p data logs && \
    chown -R hades:hades /app

# Switch to non-root user
USER hades

# Expose ports
EXPOSE 8443 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8443/api/health || exit 1

# Default command
CMD ["./hades", "web", "start", "--port", "8443"]
