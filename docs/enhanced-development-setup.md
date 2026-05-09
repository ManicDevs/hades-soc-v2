# HADES-V2 Enhanced Development Environment Setup

## Overview

The HADES-V2 Enhanced Development Environment provides a comprehensive, production-ready development setup with advanced orchestration, security, monitoring, and developer tooling.

## Prerequisites

- Docker & Docker Compose
- 8GB+ RAM recommended
- 20GB+ available disk space
- Linux/macOS environment

## Quick Start

```bash
# Deploy the enhanced development environment
docker compose -f deploy/docker-compose.enhanced-dev.yml --env-file .env.dev up -d

# Check status
docker ps --filter "name=hades"

# View logs
docker compose -f deploy/docker-compose.enhanced-dev.yml logs -f
```

## Services Overview

### Core Services

| Service | Port | Description |
|---------|------|-------------|
| hades-api | 8080 | Main HADES API server |
| hades-dashboard | 3000 | React frontend dashboard |
| hades-admin | 8081 | Admin panel interface |
| hades-monitoring | 3001 | Monitoring service |

### Infrastructure Services

| Service | Port | Description |
|---------|------|-------------|
| postgres-dev | 5432 | PostgreSQL database |
| redis-dev | 6379 | Redis cache/messaging |
| traefik-dev | 8080,8443 | Reverse proxy & load balancer |

### Observability Stack

| Service | Port | Description |
|---------|------|-------------|
| prometheus-dev | 9090 | Metrics collection |
| grafana-dev | 3003 | Visualization dashboards |
| jaeger-dev | 16686 | Distributed tracing |

### Developer Tools

| Service | Port | Description |
|---------|------|-------------|
| dev-tools | 8084 | Development utilities & scripts |

## Environment Configuration

### Development Environment (.env.dev)

```bash
# Core Settings
ENVIRONMENT=dev
HOST_SEGMENT=hades
GIN_MODE=debug
LOG_LEVEL=debug

# Database Configuration
DATABASE_URL=postgresql://hades_dev_user:dev_password@postgres-dev:5432/hades_dev?sslmode=disable
REDIS_URL=redis://:redis-dev-password-2026@redis-dev:6379

# Security
HADES_JWT_SECRET=hades-dev-jwt-secret-2026
HADES_DB_ENCRYPTION_KEY=hades-dev-db-encryption-key-2026

# Development Features
ENABLE_HOT_RELOAD=true
ENABLE_DEV_TOOLS=true
ENABLE_PROFILING=true
ENABLE_METRICS=true
```

## Access Points

### Web Interfaces

- **Frontend Dashboard**: http://localhost:3000
- **Admin Panel**: http://localhost:8082
- **Grafana**: http://localhost:3003 (admin/admin)
- **Jaeger UI**: http://localhost:16686
- **Prometheus**: http://localhost:9090

### API Endpoints

- **HADES API**: http://localhost:8080
- **Health Checks**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics

## Development Features

### Hot Reload

All Go services use Air for automatic hot reloading during development:

```bash
# Monitor for file changes
# Automatic restart on code changes
# Live debugging support
```

### Debugging

- **API Server**: Debug port 8081
- **Profiling**: pprof endpoints enabled
- **Logging**: Debug level with structured output

### Monitoring

- **Metrics**: Prometheus collection from all services
- **Tracing**: Jaeger distributed tracing
- **Dashboards**: Pre-configured Grafana dashboards
- **Alerting**: Development alert rules

## Network Architecture

### Container Network

```
deploy_hades-dev-network (172.29.0.0/16)
├── hades-api-dev (172.29.0.x)
├── hades-dashboard-dev (172.29.0.x)
├── hades-admin-dev (172.29.0.x)
├── hades-monitoring-dev (172.29.0.x)
├── postgres-dev (172.29.0.9)
├── redis-dev (172.29.0.4)
├── prometheus-dev (172.29.0.3)
├── grafana-dev (172.29.0.5)
├── jaeger-dev (172.29.0.2)
└── traefik-dev (172.29.0.x)
```

### Service Discovery

- **Internal DNS**: Service names resolve within network
- **Health Checks**: All services expose /health endpoints
- **Load Balancing**: Traefik provides intelligent routing

## Security Configuration

### Authentication

- **Development Mode**: Authentication disabled by default
- **JWT Tokens**: Configurable secret keys
- **Session Management**: Redis-based sessions

### Network Security

- **Container Isolation**: Services in dedicated network
- **Port Mapping**: Only necessary ports exposed
- **Firewall Rules**: iptables for additional security

### Data Protection

- **Encryption**: Database encryption at rest
- **Secrets Management**: Environment-based configuration
- **Audit Logging**: Comprehensive activity logging

## Data Persistence

### Volumes

```yaml
volumes:
  postgres_dev_data:    # PostgreSQL data
  postgres_dev_logs:    # PostgreSQL logs
  redis_dev_data:       # Redis data
  redis_dev_logs:       # Redis logs
  prometheus_dev_data:  # Prometheus metrics
  grafana_dev_data:     # Grafana dashboards
```

### Backup & Recovery

```bash
# Manual backup
docker exec hades-postgres-dev pg_dump -U hades_dev_user hades_dev > backup.sql

# Automated backup script
./deploy/backup-cron.sh --dry-run
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**: Stop existing services on required ports
2. **Permission Issues**: Ensure Docker socket access
3. **Resource Limits**: Check memory/disk availability
4. **Network Issues**: Verify Docker network configuration

### Debug Commands

```bash
# Check container status
docker ps --filter "name=hades"

# View service logs
docker logs hades-api-dev

# Inspect network
docker network inspect deploy_hades-dev-network

# Execute in container
docker exec -it hades-api-dev sh
```

### Health Monitoring

```bash
# Check all services
curl http://localhost:8080/health
curl http://localhost:3003/api/health
docker exec hades-redis-dev redis-cli ping
docker exec hades-postgres-dev pg_isready
```

## Development Workflow

### 1. Environment Setup

```bash
# Clone repository
git clone <repository-url>
cd hades

# Configure environment
cp .env.dev.example .env.dev
# Edit .env.dev as needed
```

### 2. Start Development

```bash
# Start all services
docker compose -f deploy/docker-compose.enhanced-dev.yml --env-file .env.dev up -d

# Monitor startup
docker compose -f deploy/docker-compose.enhanced-dev.yml logs -f
```

### 3. Development

```bash
# Make code changes
# Services automatically reload

# View logs
docker logs -f hades-api-dev

# Debug endpoints
curl http://localhost:8080/debug/pprof/
```

### 4. Testing

```bash
# Run tests
docker compose -f deploy/docker-compose.test.yml --env-file .env.test up -d

# Load testing
docker compose -f deploy/docker-compose.test.yml up load-test
```

### 5. Cleanup

```bash
# Stop services
docker compose -f deploy/docker-compose.enhanced-dev.yml down

# Remove volumes (optional)
docker compose -f deploy/docker-compose.enhanced-dev.yml down -v
```

## Production Deployment

For production deployment, use the enterprise configuration:

```bash
# Production environment
docker compose -f deploy/docker-compose.enterprise.yml --env-file .env.production up -d
```

## Support & Documentation

- **API Documentation**: http://localhost:8080/docs
- **Metrics Dashboard**: http://localhost:3003
- **Tracing UI**: http://localhost:16686
- **Configuration Files**: `/deploy/` directory
- **Environment Templates**: `.env.*` files

## Performance Tuning

### Resource Allocation

- **Memory**: 2GB minimum per service
- **CPU**: 2 cores minimum for API services
- **Storage**: SSD recommended for database

### Optimization Tips

1. **Database**: Tune PostgreSQL settings in postgresql.conf
2. **Redis**: Configure maxmemory and eviction policies
3. **Prometheus**: Adjust retention periods and scrape intervals
4. **Grafana**: Optimize dashboard queries and caching

## Integration Points

### External Services

- **SIEM Integration**: Configurable webhook endpoints
- **EDR Connectors**: Plugin architecture for security tools
- **API Gateways**: Traefik integration for microservices
- **Message Queues**: Redis pub/sub for async processing

### Monitoring Integration

- **Prometheus**: Native metrics export
- **Grafana**: Custom dashboards and alerts
- **Jaeger**: Distributed tracing across services
- **ELK Stack**: Optional log aggregation

## Security Best Practices

### Development Security

1. **Secret Management**: Use environment variables for secrets
2. **Network Isolation**: Container networking prevents cross-talk
3. **Access Control**: Limit exposed ports and interfaces
4. **Audit Trails**: Enable comprehensive logging

### Production Hardening

1. **TLS/SSL**: Enable HTTPS for all services
2. **Authentication**: Implement proper auth mechanisms
3. **Firewall**: Configure iptables rules
4. **Monitoring**: Security-focused alerting rules

## Contributing

### Development Guidelines

1. **Code Style**: Follow Go and React best practices
2. **Testing**: Include unit and integration tests
3. **Documentation**: Update API docs and READMEs
4. **Security**: Follow security checklist

### Pull Request Process

1. **Fork** repository
2. **Feature branch** from develop
3. **Tests** passing
4. **Documentation** updated
5. **PR** submitted with description

---

**Last Updated**: 2026-05-09
**Version**: 2.0.0
**Environment**: Development
