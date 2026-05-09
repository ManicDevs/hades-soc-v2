# HADES-V2 Enterprise Port Configuration Guide

## Overview

This document outlines the enterprise-level port configuration for HADES-V2, designed for high-availability, security, and scalability in production environments.

## Port Architecture

### Primary Production Services

| Service | Port | Protocol | Access Level | Description |
|---------|------|----------|--------------|-------------|
| HADES API | 8443 | HTTPS | Internal + Load Balancer | Main API endpoints |
| HADES Dashboard | 8444 | HTTPS | Internal + Load Balancer | Web dashboard |
| HADES Admin Panel | 8445 | HTTPS | Restricted | Administrative interface |
| HADES Monitoring | 8446 | HTTPS | Internal | Monitoring endpoints |

### Load Balancer & Proxy

| Service | Port | Protocol | Access Level | Description |
|---------|------|----------|--------------|-------------|
| Nginx HTTP | 80 | HTTP | Public | HTTP redirect to HTTPS |
| Nginx HTTPS | 443 | HTTPS | Public | Main entry point |
| HAProxy Stats | 9000 | HTTP | Admin | Load balancer statistics |

### Internal Services (Non-Internet Facing)

| Service | Port | Protocol | Access Level | Description |
|---------|------|----------|--------------|-------------|
| PostgreSQL | 5432 | TCP | Internal | Primary database |
| Redis | 6379 | TCP | Internal | Cache and session storage |
| Sentinel | 2112 | TCP | Internal | Security monitoring |
| Uptime Kuma | 3001 | HTTP | Internal | Uptime monitoring |

### Development & Testing

| Service | Port | Protocol | Access Level | Description |
|---------|------|----------|--------------|-------------|
| Dev API | 8080 | HTTP | Internal | Development API |
| Dev Dashboard | 3000 | HTTP | Internal | Development dashboard |
| Test API | 8081 | HTTP | Internal | Testing API |
| Test Dashboard | 3001 | HTTP | Internal | Testing dashboard |

### Monitoring Stack

| Service | Port | Protocol | Access Level | Description |
|---------|------|----------|--------------|-------------|
| Grafana | 3002 | HTTP | Internal | Metrics visualization |
| Prometheus | 9090 | HTTP | Internal | Metrics collection |
| Jaeger | 16686 | HTTP | Internal | Distributed tracing |
| Elasticsearch | 9200 | HTTP | Internal | Log storage |
| Kibana | 5601 | HTTP | Internal | Log visualization |

## Security Configuration

### Firewall Rules

The enterprise configuration includes comprehensive firewall protection:

1. **Default Policy**: DROP all inbound traffic
2. **Allowed Services**: Only explicitly permitted ports
3. **Rate Limiting**: Protection against DDoS attacks
4. **Network Segmentation**: Internal vs external access control

### Access Control

- **Public Access**: Only ports 80 and 443 (via load balancer)
- **Internal Access**: Database and internal services (127.0.0.1, private networks)
- **Admin Access**: Restricted to management networks
- **Development Access**: Localhost only

### SSL/TLS Configuration

- **TLS 1.2+**: Modern encryption protocols only
- **Strong Ciphers**: AES-256-GCM with secure key exchange
- **HSTS**: HTTP Strict Transport Security enabled
- **Certificate Management**: Automated renewal and rotation

## Load Balancer Configuration

### Nginx Configuration

- **SSL Termination**: Offload SSL processing
- **Rate Limiting**: API endpoint protection
- **Health Checks**: Automatic service monitoring
- **WebSocket Support**: Real-time communication
- **Security Headers**: XSS, CSRF, and frame protection

### HAProxy Configuration

- **High Availability**: Multiple backend servers
- **Health Monitoring**: Service availability checks
- **Session Persistence**: User session maintenance
- **Statistics**: Performance monitoring dashboard
- **TCP Mode**: Internal service load balancing

## Deployment Architecture

### Production Deployment

```
Internet → Nginx (443) → HAProxy (8443) → HADES Services
                                    ↓
                              Monitoring Stack
                                    ↓
                              Database Layer
```

### Network Segmentation

1. **DMZ**: Load balancers and public-facing services
2. **Application Tier**: HADES services and APIs
3. **Data Tier**: Databases and storage
4. **Monitoring Tier**: Observability stack

## Port Management Best Practices

### 1. Port Standardization
- Use consistent port numbering across environments
- Document all port assignments
- Maintain port allocation registry

### 2. Security Hardening
- Implement firewall rules at multiple layers
- Use network segmentation for isolation
- Regular security audits of port usage

### 3. Monitoring & Alerting
- Monitor port availability and response times
- Alert on unauthorized port access attempts
- Track port utilization metrics

### 4. Disaster Recovery
- Document port failover procedures
- Maintain backup load balancer configurations
- Test port redundancy regularly

## Configuration Files

### Environment Variables
```bash
# Primary Services
HADES_API_PORT=8443
HADES_DASHBOARD_PORT=8444
HADES_ADMIN_PORT=8445
HADES_MONITORING_PORT=8446

# Load Balancer
NGINX_PORT=80
NGINX_SSL_PORT=443
HAPROXY_STATS_PORT=9000
```

### Docker Compose
```yaml
services:
  nginx:
    ports:
      - "80:80"
      - "443:443"
  hades-api:
    ports:
      - "8443:8443"
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**: Check for services already using assigned ports
2. **Firewall Rules**: Verify iptables configuration
3. **SSL Certificates**: Ensure certificates are valid and properly configured
4. **Load Balancer Health**: Check backend service availability

### Diagnostic Commands

```bash
# Check port usage
netstat -tlnp | grep :8443

# Test connectivity
curl -k https://localhost:8443/health

# Check firewall rules
iptables -L -n -v

# Monitor load balancer
curl http://localhost:9000/stats
```

## Migration Guide

### From Development to Production

1. **Port Mapping**: Update service configurations
2. **SSL Configuration**: Install production certificates
3. **Firewall Rules**: Apply enterprise security rules
4. **Load Balancer**: Configure production routing
5. **Monitoring**: Set up production monitoring

### Rollback Procedures

1. **Service Rollback**: Revert to previous port configurations
2. **Load Balancer**: Update routing rules
3. **Firewall**: Restore previous security rules
4. **Validation**: Test all service endpoints

## Compliance & Security

### Industry Standards
- **PCI DSS**: Secure payment card industry compliance
- **SOC 2**: Service organization controls
- **ISO 27001**: Information security management
- **GDPR**: Data protection regulations

### Audit Requirements
- **Port Access Logs**: Maintain comprehensive access logs
- **Change Management**: Document all port configuration changes
- **Security Reviews**: Regular security assessments
- **Penetration Testing**: Annual security testing

## Support & Maintenance

### Regular Maintenance
- **Certificate Renewal**: Quarterly SSL certificate checks
- **Port Audit**: Monthly port usage review
- **Security Updates**: Regular security patching
- **Performance Tuning**: Load balancer optimization

### Emergency Procedures
- **Port Failover**: Automatic failover procedures
- **Security Incident**: Response protocols for security breaches
- **Service Restoration**: Priority service restoration procedures
- **Communication**: Stakeholder notification procedures

---

## Contact Information

For questions or issues regarding enterprise port configuration:

- **Infrastructure Team**: infrastructure@hades-enterprise.com
- **Security Team**: security@hades-enterprise.com
- **DevOps Team**: devops@hades-enterprise.com

## Version History

- **v2.0**: Initial enterprise port configuration
- **v2.1**: Added monitoring stack integration
- **v2.2**: Enhanced security controls and compliance features
