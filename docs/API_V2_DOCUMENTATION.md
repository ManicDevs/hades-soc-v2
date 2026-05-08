# Hades Toolkit API v2 Documentation

## Overview

The Hades Toolkit API v2 provides enhanced functionality for enterprise security management with advanced versioning, comprehensive error handling, and real-time analytics.

## Base URL

```
http://localhost:8080/api/v2
```

## Authentication

All API endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Versioning Strategies

### 1. URL Path Versioning (Recommended)
```
/api/v1/users     # Version 1
/api/v2/users     # Version 2
/api/v3/users     # Version 3 (Beta)
```

### 2. Header Versioning
```
GET /api/users
Headers:
  API-Version: v2
```

### 3. Query Parameter Versioning
```
GET /api/users?version=v2
```

### 4. Content-Type Versioning
```
GET /api/users
Headers:
  Content-Type: application/vnd.hades.v2+json
  Accept: application/vnd.hades.v2+json
```

## Enhanced Features in v2

### 1. Rich Data Models
- Nested relationships and metadata
- Enhanced user profiles with sessions
- Detailed threat intelligence with timelines
- Advanced security scoring with historical data

### 2. Pagination and Filtering
```bash
GET /api/v2/threats?page=1&page_size=20&severity=high&status=active
```

### 3. Real-time Analytics
```bash
GET /api/v2/analytics
```

### 4. Webhook Integration
```bash
GET /api/v2/webhooks
POST /api/v2/webhooks
```

## Endpoints

### Authentication

#### Enhanced Login
```bash
POST /api/v2/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123",
  "remember": true,
  "client_info": {
    "user_agent": "Mozilla/5.0...",
    "ip_address": "192.168.1.100",
    "platform": "web"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@hades-toolkit.com",
      "role": "Administrator",
      "status": "active",
      "last_login": "2026-05-01T20:30:00Z",
      "permissions": ["read", "write", "admin"],
      "profile": {
        "first_name": "Admin",
        "last_name": "User",
        "avatar": "/avatars/admin.png",
        "department": "Security",
        "location": "HQ",
        "bio": "System administrator",
        "last_seen": "2026-05-01T20:30:00Z"
      },
      "sessions": [
        {
          "id": "session_1234567890",
          "ip_address": "192.168.1.100",
          "user_agent": "Mozilla/5.0...",
          "created_at": "2026-05-01T20:30:00Z",
          "expires_at": "2026-05-02T20:30:00Z",
          "active": true
        }
      ],
      "preferences": {
        "theme": "dark",
        "notifications": true,
        "language": "en",
        "timezone": "UTC"
      },
      "created_at": "2025-05-01T20:30:00Z",
      "updated_at": "2026-05-01T20:30:00Z"
    },
    "expires_at": "2026-05-02T20:30:00Z",
    "session_id": "session_1234567890"
  }
}
```

### Dashboard

#### Enhanced Metrics
```bash
GET /api/v2/dashboard/metrics
```

**Response:**
```json
{
  "success": true,
  "data": {
    "security_score": {
      "overall": 98,
      "categories": {
        "authentication": 95,
        "authorization": 98,
        "encryption": 100,
        "monitoring": 96,
        "compliance": 99
      },
      "factors": [
        {
          "name": "Password Strength",
          "impact": 95,
          "description": "Strong password policies",
          "trend": "stable"
        }
      ],
      "history": [
        {
          "date": "2026-04-24T20:30:00Z",
          "score": 95
        }
      ]
    },
    "active_threats": 3,
    "blocked_attacks": 1247,
    "system_health": {
      "status": "healthy",
      "uptime": "24h0m0s",
      "services": {
        "api_server": "running",
        "database": "operational",
        "cache": "operational",
        "queue": "operational",
        "monitoring": "active",
        "backup_service": "scheduled"
      },
      "performance": {
        "cpu": 45.2,
        "memory": 67.8,
        "disk": 23.4,
        "network": 12.1
      },
      "resources": {
        "total_cpu": 8,
        "total_memory": 16384,
        "total_disk": 1000000,
        "used_cpu": 4,
        "used_memory": 11100,
        "used_disk": 234000
      }
    },
    "active_users": 24,
    "analytics": {
      "requests_per_second": 145.7,
      "response_time": "120ms",
      "error_rate": 0.02,
      "top_endpoints": [
        {
          "path": "/api/v2/dashboard/metrics",
          "requests": 1247,
          "avg_response": 45.2,
          "error_rate": 0.01
        }
      ],
      "user_activity": [
        {
          "user_id": "1",
          "activity": "login",
          "timestamp": "2026-05-01T20:25:00Z",
          "duration": "2s"
        }
      ]
    },
    "trends": [
      {
        "metric": "threats_blocked",
        "values": [45, 52, 48, 61, 58, 72, 69, 78],
        "labels": ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun", "Today"],
        "direction": "up"
      }
    ],
    "alerts": [
      {
        "id": "alert-001",
        "type": "security",
        "severity": "medium",
        "title": "Unusual login pattern detected",
        "description": "Multiple failed login attempts from unknown IP",
        "timestamp": "2026-05-01T20:00:00Z",
        "status": "open",
        "assignee": "security-team",
        "metadata": {
          "source_ip": "203.0.113.45",
          "user_id": "unknown",
          "attempts": "5"
        }
      }
    ]
  }
}
```

### Threats

#### Enhanced Threat Intelligence
```bash
GET /api/v2/threats?page=1&page_size=20&severity=high&status=active
```

**Response:**
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": 1,
        "type": "malware",
        "severity": "critical",
        "title": "Advanced Persistent Threat Detected",
        "source": {
          "ip_address": "192.168.1.105",
          "country": "Unknown",
          "asn": "AS12345",
          "domain": "malicious.example.com",
          "url": "http://malicious.example.com/payload"
        },
        "status": "blocked",
        "timestamp": "2026-05-01T18:31:10Z",
        "description": "Sophisticated APT with multiple attack vectors detected and blocked",
        "impact": {
          "risk_score": 95,
          "affected_assets": ["web-server", "database", "file-server"],
          "business_impact": "High - Potential data breach",
          "data_classification": "Confidential"
        },
        "mitigation": {
          "actions": ["Blocked IP address", "Updated firewall rules", "Isolated affected systems"],
          "automated": true,
          "completed_at": "2026-05-01T19:31:10Z",
          "assigned_to": "security-automation",
          "priority": "critical"
        },
        "related_entities": [
          {
            "type": "user",
            "id": "user-123",
            "name": "jsmith"
          },
          {
            "type": "asset",
            "id": "asset-456",
            "name": "web-server-01"
          }
        ],
        "timeline": [
          {
            "event": "initial_detection",
            "timestamp": "2026-05-01T18:31:10Z",
            "user": "system",
            "details": "Anomaly detected in network traffic"
          },
          {
            "event": "analysis_started",
            "timestamp": "2026-05-01T18:37:10Z",
            "user": "analyst-1",
            "details": "Security analyst began investigation"
          }
        ],
        "tags": ["apt", "malware", "blocked", "automated-response"],
        "metadata": {
          "attack_vector": "phishing",
          "malware_family": "APT-29",
          "confidence": 0.95,
          "false_positive": false
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 2,
      "total_pages": 1,
      "has_next": false,
      "has_prev": false
    },
    "metadata": {
      "total_threats": 2,
      "filtered_count": 2,
      "search_time": "268ns",
      "cache_hit": false,
      "request_id": "req_1777663870899122794"
    }
  }
}
```

### Analytics

#### API Analytics
```bash
GET /api/v2/analytics
```

**Response:**
```json
{
  "success": true,
  "data": {
    "api_metrics": {
      "requests_per_second": 145.7,
      "average_response_time": "120ms",
      "error_rate": 0.02,
      "uptime": "99.9%"
    },
    "top_endpoints": [
      {
        "path": "/api/v2/dashboard/metrics",
        "requests": 1247,
        "avg_response": 45.2
      },
      {
        "path": "/api/v2/threats",
        "requests": 892,
        "avg_response": 67.8
      },
      {
        "path": "/api/v2/users",
        "requests": 456,
        "avg_response": 34.1
      }
    ],
    "user_analytics": {
      "active_users": 24,
      "total_sessions": 156,
      "avg_session_duration": "45m"
    },
    "security_metrics": {
      "blocked_requests": 1247,
      "failed_authentications": 23,
      "suspicious_activities": 5
    },
    "performance": {
      "cpu_usage": 45.2,
      "memory_usage": 67.8,
      "disk_usage": 23.4,
      "network_io": 12.1
    }
  }
}
```

### Webhooks

#### List Webhooks
```bash
GET /api/v2/webhooks
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "webhook-001",
      "name": "Threat Alert Webhook",
      "url": "https://customer.example.com/webhooks/threats",
      "events": ["threat.created", "threat.updated", "threat.resolved"],
      "active": true,
      "created_at": "2026-04-01T20:31:10Z",
      "last_triggered": "2026-05-01T18:31:10Z"
    },
    {
      "id": "webhook-002",
      "name": "Security Score Webhook",
      "url": "https://monitoring.example.com/webhooks/security",
      "events": ["security.score.changed"],
      "active": true,
      "created_at": "2026-04-16T20:31:10Z",
      "last_triggered": "2026-05-01T19:31:10Z"
    }
  ]
}
```

#### Create Webhook
```bash
POST /api/v2/webhooks
Content-Type: application/json

{
  "name": "Custom Alert Webhook",
  "url": "https://example.com/webhooks/alerts",
  "events": ["threat.created", "user.login.failed"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "webhook-123",
    "name": "Custom Alert Webhook",
    "url": "https://example.com/webhooks/alerts",
    "events": ["threat.created", "user.login.failed"],
    "active": true,
    "created_at": "2026-05-01T20:31:10Z",
    "last_triggered": null
  }
}
```

## Error Handling

### Enhanced Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request format",
    "details": "Field 'username' is required",
    "field": "username"
  },
  "request": {
    "method": "POST",
    "path": "/api/v2/auth/login",
    "headers": {
      "content-type": "application/json",
      "user-agent": "Mozilla/5.0..."
    },
    "timestamp": "2026-05-01T20:31:10Z",
    "request_id": "req_1777663870899122794"
  },
  "system": {
    "version": "2.0.0",
    "build": "20240501-001",
    "timestamp": "2026-05-01T20:31:10Z",
    "trace_id": "trace_1777663870899122794"
  }
}
```

### Common Error Codes

- `VALIDATION_ERROR` - Request validation failed
- `AUTHENTICATION_FAILED` - Invalid credentials
- `AUTHORIZATION_FAILED` - Insufficient permissions
- `RESOURCE_NOT_FOUND` - Resource does not exist
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_SERVER_ERROR` - Server error

## Version Discovery

### Get Available Versions
```bash
GET /api/versions
```

**Response:**
```json
{
  "success": true,
  "data": {
    "versions": {
      "v1": {
        "version": "v1",
        "status": "active",
        "released": "2025-05-01T20:30:00Z",
        "endpoints": ["/auth", "/dashboard", "/threats", "/users", "/security"],
        "features": [
          "JWT Authentication",
          "Basic CRUD operations",
          "Mock data responses",
          "Simple error handling"
        ]
      },
      "v2": {
        "version": "v2",
        "status": "active",
        "released": "2026-04-01T20:30:00Z",
        "endpoints": ["/auth", "/dashboard", "/threats", "/users", "/security", "/analytics", "/webhooks"],
        "features": [
          "Enhanced JWT with refresh tokens",
          "Advanced filtering and pagination",
          "Real-time WebSocket support",
          "Rate limiting and quotas",
          "Advanced error responses",
          "Request/response compression",
          "API analytics and metrics",
          "Webhook integrations",
          "GraphQL endpoints",
          "OpenAPI 3.0 documentation"
        ]
      },
      "v3": {
        "version": "v3",
        "status": "beta",
        "released": "2026-04-24T20:30:00Z",
        "endpoints": ["/auth", "/dashboard", "/threats", "/users", "/security", "/analytics", "/webhooks", "/ml", "/automation"],
        "features": [
          "All v2 features",
          "Machine learning threat detection",
          "Automated response workflows",
          "Advanced analytics dashboard",
          "Multi-tenant support",
          "GraphQL subscriptions",
          "Event-driven architecture"
        ]
      }
    },
    "default_version": "v2",
    "supported_versions": ["v1", "v2", "v3"],
    "versioning_strategies": {
      "url_path": "/api/v{version}/endpoint",
      "header": "API-Version: v{version}",
      "query_param": "?version=v{version}",
      "content_type": "application/vnd.hades.v{version}+json"
    },
    "deprecation_policy": "6-month deprecation, 12-month sunset"
  }
}
```

## Deprecation and Migration

### Deprecation Headers

When using a deprecated version, the API returns deprecation headers:

```
X-API-Deprecated: true
X-API-Sunset: 2026-12-01T00:00:00Z
X-API-Supported-Versions: v1,v2,v3
```

### Migration Guide

1. **From v1 to v2**:
   - Enhanced authentication with session management
   - Richer data models with nested relationships
   - New analytics and webhook endpoints
   - Improved error handling with detailed context

2. **From v2 to v3** (Future):
   - Machine learning integration
   - Automated response workflows
   - GraphQL support
   - Multi-tenant architecture

## Rate Limiting

API v2 includes rate limiting with the following quotas:

- **Default**: 1000 requests per hour
- **Authenticated**: 5000 requests per hour
- **Premium**: 10000 requests per hour

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1714560600
```

## SDK Examples

### JavaScript/Node.js

```javascript
// Using fetch with version header
const response = await fetch('http://localhost:8080/api/dashboard/metrics', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'API-Version': 'v2',
    'Content-Type': 'application/vnd.hades.v2+json'
  }
});

const data = await response.json();
console.log(data.data.security_score.overall);
```

### Python

```python
import requests

headers = {
    'Authorization': f'Bearer {token}',
    'API-Version': 'v2',
    'Content-Type': 'application/vnd.hades.v2+json'
}

response = requests.get('http://localhost:8080/api/v2/dashboard/metrics', headers=headers)
data = response.json()
print(data['data']['security_score']['overall'])
```

### curl

```bash
# URL path versioning
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v2/dashboard/metrics

# Header versioning
curl -H "Authorization: Bearer $TOKEN" \
     -H "API-Version: v2" \
     http://localhost:8080/api/dashboard/metrics

# Content-Type versioning
curl -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/vnd.hades.v2+json" \
     http://localhost:8080/api/dashboard/metrics
```

## Testing

### Health Check
```bash
curl http://localhost:8080/api/v2/health
```

### Version Discovery
```bash
curl http://localhost:8080/api/versions
```

### Authentication Test
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin123"}' \
     http://localhost:8080/api/v2/auth/login
```

## Support

- **Documentation**: http://localhost:8080/api/docs
- **Swagger UI**: http://localhost:8080/api/swagger
- **Support Email**: api@hades-toolkit.com
- **Status Page**: https://status.hades-toolkit.com
