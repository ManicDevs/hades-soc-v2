// Hades Toolkit Configuration
export const config = {
  // Port allocation following hades-toolkit enterprise pattern
  ports: {
    web: 3000,           // Main web application
    api: 8080,           // API server
    auth: 8081,          // Authentication service
    dashboard: 8082,      // Dashboard service
    metrics: 9090,       // Metrics/monitoring
    admin: 8083,         // Admin interface
    customer: 8084,      // Customer portal
    license: 8085,       // License server
    repository: 8086,    // Repository service
    modules: 8087,       // Module management
    exploit: 8088,       // Exploit database
    monitor: 8089,       // Monitoring service
    siem: 8090,          // SIEM integration
    analytics: 8091,     // Analytics service
    notifications: 8092, // Notification service
    websocket: 8093,    // WebSocket service
    database: 5432,      // PostgreSQL database
    redis: 6379,         // Redis cache
    elasticsearch: 9200, // Search engine
  },
  
  // API endpoints
  api: {
    base: '/api',
    version: 'v2',
    endpoints: {
      auth: '/auth',
      users: '/users',
      modules: '/modules',
      exploits: '/exploits',
      dashboard: '/dashboard',
      analytics: '/analytics',
      admin: '/admin',
      customer: '/customer',
      license: '/license',
      repository: '/repository',
      siem: '/siem',
      metrics: '/metrics',
      notifications: '/notifications',
    }
  },
  
  // Authentication configuration
  auth: {
    tokenExpiry: '24h',
    refreshTokenExpiry: '7d',
    maxLoginAttempts: 5,
    lockoutDuration: '15m',
    sessionTimeout: '30m',
    mfaEnabled: true,
    passwordPolicy: {
      minLength: 8,
      requireUppercase: true,
      requireLowercase: true,
      requireNumbers: true,
      requireSpecialChars: true,
    }
  },
  
  // Database configuration
  database: {
    host: process.env.DB_HOST || 'localhost',
    port: process.env.DB_PORT || 5432,
    name: process.env.DB_NAME || 'hades_enterprise',
    user: process.env.DB_USER || 'hades',
    password: process.env.DB_PASSWORD || '',
    ssl: process.env.DB_SSL === 'true',
    poolSize: 20,
    timeout: 30000,
  },
  
  // Security configuration
  security: {
    cors: {
      origins: ['http://localhost:3000', 'http://localhost:8080'],
      credentials: true,
    },
    rateLimit: {
      windowMs: 15 * 60 * 1000, // 15 minutes
      max: 100, // limit each IP to 100 requests per windowMs
    },
    helmet: {
      contentSecurityPolicy: {
        directives: {
          defaultSrc: ["'self'"],
          styleSrc: ["'self'", "'unsafe-inline'"],
          scriptSrc: ["'self'"],
          imgSrc: ["'self'", "data:", "https:"],
        },
      },
    },
  },
  
  // Enterprise features
  enterprise: {
    multiTenant: true,
    auditLogging: true,
    compliance: {
      gdpr: true,
      hipaa: false,
      sox: false,
    },
    backup: {
      enabled: true,
      interval: '6h',
      retention: '30d',
    },
    monitoring: {
      enabled: true,
      metrics: true,
      alerts: true,
      healthChecks: true,
    },
  },
  
  // Development configuration
  development: {
    hotReload: true,
    debug: true,
    mockData: false,
    proxy: {
      '/api': 'http://localhost:8080',
      '/auth': 'http://localhost:8081',
    },
  },
  
  // Production configuration
  production: {
    minify: true,
    gzip: true,
    cache: true,
    ssl: true,
    cluster: true,
    workers: 4,
  }
}

export default config
