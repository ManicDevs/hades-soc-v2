# Hades-V2 Deployment Guide

This guide covers production deployment of Hades-V2 in various environments.

## 🏗️ Prerequisites

### System Requirements
- **CPU**: 2 cores minimum (4+ recommended)
- **Memory**: 4GB RAM minimum (8GB+ recommended for large deployments)
- **Storage**: 20GB minimum (100GB+ recommended for logs and data)
- **Network**: HTTPS access, inbound ports 8443, 8080
- **OS**: Linux (Ubuntu 20.04+, RHEL 8+, CentOS 8+)

### Database Requirements
- **SQLite**: For small deployments (<100 users)
- **PostgreSQL**: For medium deployments (100-1000 users)
- **MySQL**: For large deployments (1000+ users)
- **Redis**: For session storage and caching (recommended)

### Security Requirements
- TLS/SSL certificates
- Firewall configuration
- Backup storage
- Audit logging storage

## 🚀 Quick Deployment

### 1. Binary Deployment
```bash
# Download and install
curl -L https://github.com/your-org/hades-v2/releases/latest/download/hades-linux-amd64 -o hades
chmod +x hades
sudo mv hades /opt/hades/

# Create user
sudo useradd -r -s /bin/false hades
sudo mkdir -p /opt/hades/{data,logs,config}
sudo chown -R hades:hades /opt/hades

# Configure
cd /opt/hades
sudo -u hades ./hades config wizard

# Initialize database
sudo -u hades ./hades migrate init
sudo -u hades ./hades migrate up

# Create admin user
sudo -u hades ./hades user create --username admin --email admin@company.com --role admin --password secure-password
```

### 2. Systemd Service
```bash
# Create service file
sudo tee /etc/systemd/system/hades.service > /dev/null <<EOF
[Unit]
Description=Hades-V2 Enterprise Security Framework
After=network.target

[Service]
Type=simple
User=hades
Group=hades
WorkingDirectory=/opt/hades
ExecStart=/opt/hades/hades web start --port 8443 --config /opt/hades/config/hades.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable hades
sudo systemctl start hades
```

## 🐳 Docker Deployment

### 1. Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hades ./cmd/hades

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -s /bin/sh hades

WORKDIR /app
COPY --from=builder /app/hades .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config ./config

RUN mkdir -p data logs
RUN chown -R hades:hades /app

USER hades
EXPOSE 8443 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8443/api/health || exit 1

CMD ["./hades", "web", "start", "--port", "8443"]
```

### 2. Docker Compose
```yaml
version: '3.8'

services:
  hades:
    build: .
    container_name: hades
    restart: unless-stopped
    ports:
      - "8443:8443"
      - "8080:8080"
    environment:
      - HADES_DB_TYPE=postgres
      - HADES_DB_HOST=postgres
      - HADES_DB_PORT=5432
      - HADES_DB_NAME=hades
      - HADES_DB_USER=hades
      - HADES_DB_PASSWORD=secure-password
      - HADES_REDIS_URL=redis://redis:6379
    volumes:
      - hades_data:/app/data
      - hades_logs:/app/logs
      - hades_config:/app/config
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8443/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  postgres:
    image: postgres:15-alpine
    container_name: hades-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: hades
      POSTGRES_USER: hades
      POSTGRES_PASSWORD: secure-password
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U hades"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: hades-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass redis-password
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  nginx:
    image: nginx:alpine
    container_name: hades-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./nginx/logs:/var/log/nginx
    depends_on:
      - hades

volumes:
  hades_data:
  hades_logs:
  hades_config:
  postgres_data:
  redis_data:

networks:
  default:
    name: hades-network
```

### 3. Nginx Configuration
```nginx
# nginx/nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream hades_backend {
        server hades:8443;
    }

    upstream hades_api {
        server hades:8080;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;

    server {
        listen 80;
        server_name hades.company.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name hades.company.com;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

        # Main application
        location / {
            proxy_pass http://hades_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # API endpoints
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://hades_api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Login endpoint with stricter rate limiting
        location /api/login {
            limit_req zone=login burst=5 nodelay;
            proxy_pass http://hades_api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## ☸️ Kubernetes Deployment

### 1. Namespace and ConfigMap
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: hades
---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hades-config
  namespace: hades
data:
  hades.yaml: |
    server:
      workers: 10
      queue_size: 50
      log_level: info
    database:
      type: postgres
      host: postgres-service
      port: 5432
      database: hades
      username: hades
    web:
      port: 8443
      enable_cors: true
    auth:
      session_timeout: 1440
      max_failed_attempts: 5
      lockout_duration: 15
      require_mfa: false
```

### 2. Secrets
```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: hades-secrets
  namespace: hades
type: Opaque
data:
  db-password: c2VjdXJlLXBhc3N3b3Jk  # base64 encoded
  jwt-secret: andl0LXNlY3JldC1rZXk=
  api-token: YXBpLXRva2VuLXNlY3VyZQ==
```

### 3. Deployment
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hades
  namespace: hades
  labels:
    app: hades
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: hades
  template:
    metadata:
      labels:
        app: hades
    spec:
      containers:
      - name: hades
        image: hades-v2:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8443
        - containerPort: 8080
        env:
        - name: HADES_DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: hades-secrets
              key: db-password
        - name: HADES_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: hades-secrets
              key: jwt-secret
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: data
          mountPath: /app/data
        - name: logs
          mountPath: /app/logs
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8443
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8443
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: hades-config
      - name: data
        persistentVolumeClaim:
          claimName: hades-data
      - name: logs
        persistentVolumeClaim:
          claimName: hades-logs
```

### 4. Services
```yaml
# k8s/services.yaml
apiVersion: v1
kind: Service
metadata:
  name: hades-service
  namespace: hades
spec:
  selector:
    app: hades
  ports:
  - name: web
    port: 8443
    targetPort: 8443
  - name: api
    port: 8080
    targetPort: 8080
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: hades
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
  type: ClusterIP
```

### 5. Ingress
```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hades-ingress
  namespace: hades
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - hades.company.com
    secretName: hades-tls
  rules:
  - host: hades.company.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: hades-service
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hades-service
            port:
              number: 8443
```

## 🔧 Configuration

### Environment Variables
```bash
# Database
HADES_DB_TYPE=postgres
HADES_DB_HOST=localhost
HADES_DB_PORT=5432
HADES_DB_NAME=hades
HADES_DB_USER=hades
HADES_DB_PASSWORD=secure-password

# Redis
HADES_REDIS_URL=redis://localhost:6379

# Web Server
HADES_WEB_PORT=8443
HADES_WEB_STATIC_PATH=/app/web/dashboard
HADES_WEB_CORS_ENABLED=true

# Authentication
HADES_AUTH_SESSION_TIMEOUT=1440
HADES_AUTH_MAX_FAILED_ATTEMPTS=5
HADES_AUTH_LOCKOUT_DURATION=15
HADES_AUTH_REQUIRE_MFA=false

# Logging
HADES_LOG_LEVEL=info
HADES_LOG_FILE=/app/logs/hades.log

# Performance
HADES_WORKERS=10
HADES_QUEUE_SIZE=50
```

### Configuration File
```yaml
# config/hades.yaml
server:
  workers: 10
  queue_size: 50
  log_level: info
  log_file: /app/logs/hades.log

database:
  type: postgres
  host: postgres-service
  port: 5432
  database: hades
  username: hades
  password: ${HADES_DB_PASSWORD}
  ssl_mode: require
  max_connections: 20

auth:
  session_timeout: 1440
  max_failed_attempts: 5
  lockout_duration: 15
  require_mfa: false
  jwt_secret: ${HADES_JWT_SECRET}

web:
  port: 8443
  static_path: /app/web/dashboard
  enable_cors: true

siem:
  provider: splunk
  endpoint: https://splunk.company.com:8088
  api_key: ${HADES_SPLUNK_API_KEY}
  index: hades_events
  batch_size: 100

encryption:
  algorithm: aes-256-gcm
  key_derivation: hkdf-sha256
```

## 🔒 Security Hardening

### 1. Network Security
```bash
# Firewall rules
ufw allow 22/tcp
ufw allow 443/tcp
ufw allow 8443/tcp
ufw deny 8080/tcp  # API only accessible internally
ufw enable
```

### 2. SSL/TLS Configuration
```bash
# Generate self-signed certificate (for testing)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/nginx/ssl/key.pem \
  -out /etc/nginx/ssl/cert.pem

# Or use Let's Encrypt
certbot --nginx -d hades.company.com
```

### 3. Database Security
```sql
-- Create dedicated user
CREATE USER hades_app WITH PASSWORD 'secure-password';
GRANT CONNECT ON DATABASE hades TO hades_app;
GRANT USAGE ON SCHEMA public TO hades_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO hades_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO hades_app;

-- Enable row-level security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_policy ON users FOR ALL TO hades_app USING (id = current_user_id());
```

## 📊 Monitoring

### 1. Prometheus Metrics
```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'hades'
    static_configs:
      - targets: ['hades:8443']
    metrics_path: /metrics
    scrape_interval: 30s
```

### 2. Grafana Dashboard
```json
{
  "dashboard": {
    "title": "Hades-V2 Monitoring",
    "panels": [
      {
        "title": "Active Users",
        "type": "stat",
        "targets": [
          {
            "expr": "hades_active_users_total"
          }
        ]
      },
      {
        "title": "Task Queue Depth",
        "type": "graph",
        "targets": [
          {
            "expr": "hades_task_queue_depth"
          }
        ]
      }
    ]
  }
}
```

### 3. Alerting Rules
```yaml
# monitoring/alerts.yml
groups:
  - name: hades
    rules:
      - alert: HadesDown
        expr: up{job="hades"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Hades-V2 is down"
          description: "Hades-V2 has been down for more than 1 minute"

      - alert: HadesHighMemoryUsage
        expr: hades_memory_usage > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Hades-V2 memory usage is above 80%"
```

## 🔄 Backup and Recovery

### 1. Database Backup
```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/hades"

# Create backup
pg_dump -h postgres -U hades hades > "$BACKUP_DIR/hades_$DATE.sql"

# Compress
gzip "$BACKUP_DIR/hades_$DATE.sql"

# Keep last 7 days
find "$BACKUP_DIR" -name "hades_*.sql.gz" -mtime +7 -delete
```

### 2. Configuration Backup
```bash
#!/bin/bash
# backup-config.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/hades"

# Backup configuration
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" /opt/hades/config

# Backup logs (last 24 hours)
find /opt/hades/logs -name "*.log" -mtime -1 -exec cp {} "$BACKUP_DIR/logs_$DATE/" \;
```

### 3. Recovery
```bash
#!/bin/bash
# restore.sh
BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Stop services
systemctl stop hades

# Restore database
gunzip -c "$BACKUP_FILE" | psql -h postgres -U hades hades

# Start services
systemctl start hades

# Verify
curl -f http://localhost:8443/api/health
```

## 🧪 Testing

### 1. Health Check
```bash
#!/bin/bash
# health-check.sh
API_URL="https://hades.company.com/api/health"

# Check API health
if curl -f "$API_URL" > /dev/null 2>&1; then
    echo "✅ API is healthy"
else
    echo "❌ API is unhealthy"
    exit 1
fi

# Check database connection
if pg_isready -h postgres -U hades > /dev/null 2>&1; then
    echo "✅ Database is accessible"
else
    echo "❌ Database is not accessible"
    exit 1
fi
```

### 2. Load Testing
```bash
#!/bin/bash
# load-test.sh
API_URL="https://hades.company.com/api/health"
CONCURRENT_USERS=50
DURATION=60s

# Run load test
hey -n 1000 -c "$CONCURRENT_USERS" -z "$DURATION" "$API_URL"
```

## 📚 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check database status
   pg_isready -h postgres -U hades
   
   # Check connection string
   psql -h postgres -U hades -d hades -c "SELECT 1;"
   ```

2. **High Memory Usage**
   ```bash
   # Check memory usage
   ps aux | grep hades
   
   # Adjust worker count
   ./hades config set server.workers 5
   ```

3. **Slow Response Times**
   ```bash
   # Check database performance
   psql -h postgres -U hades -d hades -c "SELECT * FROM pg_stat_activity;"
   
   # Check task queue
   curl http://localhost:8443/api/status
   ```

### Log Analysis
```bash
# View recent logs
tail -f /opt/hades/logs/hades.log

# Search for errors
grep -i error /opt/hades/logs/hades.log

# Analyze performance
grep -i "slow" /opt/hades/logs/hades.log
```

This deployment guide provides comprehensive instructions for deploying Hades-V2 in production environments with proper security, monitoring, and operational considerations.
