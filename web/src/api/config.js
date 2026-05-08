import { getAuthToken } from '../lib/authToken'

// Centralized API Configuration with Distributed Support
export const API_CONFIG = {
  // Environment detection
  getEnvironment: () => {
    const hostname = window.location.hostname
    
    if (hostname === 'localhost' || hostname === '127.0.0.1') {
      return 'development'
    } else if (hostname.includes('dev')) {
      return 'development'
    } else if (hostname.includes('test')) {
      return 'testing'
    }
    return 'production'
  },

  // Distributed API endpoints with failover
  getAPIEndpoints: () => {
    const env = API_CONFIG.getEnvironment()
    const apiBase = import.meta.env.VITE_API_BASE_URL
    
    const endpoints = {
      development: [
        apiBase || 'http://localhost:8080/api/v2'
      ],
      testing: [
        apiBase || 'http://localhost:8080/api/v2'
      ],
      production: [
        apiBase || `${window.location.protocol}//${window.location.host}/api/v2`,
      ]
    }
    
    return endpoints[env] || endpoints.development
  },

  // Base URL with load balancing
  getBaseURL: () => {
    const endpoints = API_CONFIG.getAPIEndpoints()
    const currentIndex = API_CONFIG.getCurrentEndpointIndex()
    return endpoints[currentIndex % endpoints.length]
  },

  // Load balancing index
  currentEndpointIndex: 0,
  getCurrentEndpointIndex: () => API_CONFIG.currentEndpointIndex,
  incrementEndpointIndex: () => {
    API_CONFIG.currentEndpointIndex = (API_CONFIG.currentEndpointIndex + 1) % API_CONFIG.getAPIEndpoints().length
  },

  // API version management
  getCurrentVersion: () => 'v2',
  getSupportedVersions: () => ['v1', 'v2', 'v3'],
  
  // Request configuration
  getDefaultHeaders: () => ({
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }),

  getAuthHeaders: () => {
    const token = getAuthToken()
    return token ? { 'Authorization': `Bearer ${token}` } : {}
  },

  // Error handling
  handleResponse: async (response) => {
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`)
    }
    
    const data = await response.json()
    return data.data || data
  },

  // Request wrapper with distributed failover
  request: async (endpoint, options = {}) => {
    const currentEnv = API_CONFIG.getEnvironment()
    console.log(`API REQUEST START - Endpoint: ${endpoint}`)
    const endpoints = API_CONFIG.getAPIEndpoints()
    console.log(`Available endpoints:`, endpoints)
    let lastError = null
    
    // Try each endpoint with failover
    for (let i = 0; i < endpoints.length; i++) {
      const baseURL = endpoints[i]
      const url = `${baseURL}${endpoint}`
      
      console.log(`Attempting API request to: ${url}`)
      
      const authHeaders = API_CONFIG.getAuthHeaders()
      const config = {
        ...options,
        headers: {
          ...API_CONFIG.getDefaultHeaders(),
          ...authHeaders,
          ...options.headers
        }
      }

      // Debug: Show exactly what headers are being sent
      console.log('🔍 Request headers being sent:', config.headers)
      console.log('🔍 Request method:', options.method || 'GET')
      console.log('🔍 Request body:', options.body || 'No body')

      try {
        const response = await fetch(url, config)
        console.log(`API Response status for ${url}: ${response.status}`)
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        
        const data = await response.json()
        console.log(`API Success for ${endpoint}:`, data)
        
        // Update current endpoint index on success
        API_CONFIG.currentEndpointIndex = i
        return data.data || data
      } catch (error) {
        console.warn(`API Error for ${endpoint} on ${baseURL}:`, error)
        lastError = error
        continue
      }
    }
    
    // All endpoints failed
    console.error(`All API endpoints failed for ${endpoint}:`, lastError)
    
    // In development, provide fallback data as final fallback
    console.log(`Current environment: ${currentEnv}`)
    if (currentEnv === 'development') {
      console.log(`Using fallback data for ${endpoint} (final fallback)`)
      const fallbackData = API_CONFIG.getFallbackData(endpoint)
      console.log(`Fallback data for ${endpoint}:`, fallbackData)
      return fallbackData
    } else {
      console.log(`Not using fallback data - environment is: ${currentEnv}`)
    }
    
    throw lastError || new Error('All API endpoints failed')
  },

  // Hot-swap configuration
  hotSwapConfig: async (newConfig) => {
    try {
      // Validate new configuration
      if (!newConfig.endpoints || !Array.isArray(newConfig.endpoints)) {
        throw new Error('Invalid configuration: endpoints must be an array')
      }
      
      // Update configuration
      const env = API_CONFIG.getEnvironment()
      API_CONFIG.getAPIEndpoints = () => newConfig.endpoints
      
      console.log('Configuration hot-swapped successfully')
      return true
    } catch (error) {
      console.error('Failed to hot-swap configuration:', error)
      return false
    }
  },

  // Get system status
  getSystemStatus: async () => {
    const endpoints = API_CONFIG.getAPIEndpoints()
    const status = {
      environment: API_CONFIG.getEnvironment(),
      endpoints: [],
      healthy: false
    }
    
    // Check each endpoint
    for (const endpoint of endpoints) {
      try {
        const response = await fetch(`${endpoint}/health`, {
          headers: API_CONFIG.getDefaultHeaders(),
          method: 'GET'
        })
        
        if (response.ok) {
          status.endpoints.push({
            url: endpoint,
            status: 'healthy',
            responseTime: response.headers.get('x-response-time') || 'unknown'
          })
          status.healthy = true
        } else {
          status.endpoints.push({
            url: endpoint,
            status: 'unhealthy',
            error: `HTTP ${response.status}`
          })
        }
      } catch (error) {
        status.endpoints.push({
          url: endpoint,
          status: 'unhealthy',
          error: error.message
        })
      }
    }
    
    return status
  },

  // Fallback data for development
  getFallbackData: (endpoint) => {
    console.log(`getFallbackData called for endpoint: ${endpoint}`)
    const fallbacks = {
      '/dashboard/metrics': {
        securityScore: 98,
        activeThreats: 3,
        blockedAttacks: 127,
        systemHealth: 95,
        activeUsers: 4
      },
      '/dashboard/activity': [
        { id: 1, type: 'threat', message: 'Malware attack blocked', time: '2 min ago', severity: 'high' },
        { id: 2, type: 'user', message: 'New user registered', time: '15 min ago', severity: 'low' },
        { id: 3, type: 'system', message: 'Security scan completed', time: '1 hour ago', severity: 'medium' }
      ],
      '/dashboard/status': {
        api: 'online',
        database: 'connected',
        workers: '5/5 active',
        memory: '68.7%',
        cpu: '12.4%'
      },
      '/dashboard/security': {
        threats: { critical: 1, high: 2, medium: 5, low: 8 },
        policies: { active: 12, warning: 2, inactive: 1 },
        vulnerabilities: { open: 4, inProgress: 2, resolved: 15 }
      },
      '/users': [
        { id: 1, username: 'admin', email: 'admin@hades-toolkit.com', role: 'Administrator', status: 'active' },
        { id: 2, username: 'analyst', email: 'analyst@hades-toolkit.com', role: 'Security Analyst', status: 'active' }
      ],
      '/users/roles': [
        { id: 1, name: 'Administrator', permissions: ['read', 'write', 'admin'] },
        { id: 2, name: 'Security Analyst', permissions: ['read', 'write'] },
        { id: 3, name: 'User', permissions: ['read'] }
      ],
      '/threats': {
        data: [
          {
            id: 1,
            type: 'malware',
            severity: 'critical',
            title: 'Advanced Persistent Threat Detected',
            source: {
              ip_address: '192.168.1.105',
              country: 'Unknown',
              asn: 'AS12345',
              domain: 'malicious.example.com',
              url: 'http://malicious.example.com/payload'
            },
            status: 'blocked',
            timestamp: new Date().toISOString(),
            description: 'Sophisticated APT with multiple attack vectors detected and blocked',
            impact: {
              risk_score: 95,
              affected_assets: ['web-server', 'database', 'file-server'],
              business_impact: 'High - Potential data breach',
              data_classification: 'Confidential'
            }
          },
          {
            id: 2,
            type: 'phishing',
            severity: 'high',
            title: 'Targeted Phishing Campaign',
            source: {
              ip_address: '203.0.113.45',
              country: 'US',
              asn: 'AS6789',
              domain: 'suspicious-corp.com',
              url: ''
            },
            status: 'monitoring',
            timestamp: new Date().toISOString(),
            description: 'Sophisticated phishing campaign targeting executive team',
            impact: {
              risk_score: 85,
              affected_assets: ['email-server', 'user-accounts'],
              business_impact: 'Medium - Potential credential theft',
              data_classification: 'Internal'
            }
          }
        ],
        pagination: {
          page: 1,
          page_size: 20,
          total: 2,
          total_pages: 1,
          has_next: false,
          has_prev: false
        }
      },
      '/threats/stats': {
        total_threats: 2,
        by_severity: {
          critical: 1,
          high: 1,
          medium: 0,
          low: 0
        },
        by_status: {
          blocked: 1,
          monitoring: 1,
          investigating: 0,
          resolved: 0
        },
        by_type: {
          malware: 1,
          phishing: 1,
          ddos: 0,
          sql_injection: 0
        }
      },
      '/kubernetes/clusters': {
        clusters: [
          {
            id: 'cluster-1',
            name: 'production-cluster',
            status: 'running',
            nodes: 5,
            version: 'v1.28.0',
            region: 'us-west-2',
            created_at: '2026-01-15T10:00:00Z',
            resources: {
              cpu: '16 cores',
              memory: '64GB',
              storage: '500GB'
            }
          },
          {
            id: 'cluster-2',
            name: 'staging-cluster',
            status: 'running',
            nodes: 3,
            version: 'v1.27.0',
            region: 'us-east-1',
            created_at: '2026-02-20T14:30:00Z',
            resources: {
              cpu: '8 cores',
              memory: '32GB',
              storage: '200GB'
            }
          }
        ]
      },
      '/kubernetes/deployments': {
        deployments: [
          {
            id: 'deploy-1',
            name: 'hades-api',
            namespace: 'default',
            replicas: 3,
            ready: 3,
            status: 'running',
            image: 'hades-toolkit/api:v2.1.0',
            created_at: '2026-03-10T09:15:00Z'
          },
          {
            id: 'deploy-2',
            name: 'hades-web',
            namespace: 'default',
            replicas: 2,
            ready: 2,
            status: 'running',
            image: 'hades-toolkit/web:v1.8.0',
            created_at: '2026-03-10T09:20:00Z'
          }
        ]
      },
      '/kubernetes/services': {
        services: [
          {
            id: 'svc-1',
            name: 'hades-api-service',
            namespace: 'default',
            type: 'ClusterIP',
            cluster_ip: '10.96.0.100',
            ports: ['80:8080'],
            selector: 'app=hades-api'
          },
          {
            id: 'svc-2',
            name: 'hades-web-service',
            namespace: 'default',
            type: 'LoadBalancer',
            external_ip: '203.0.113.45',
            ports: ['80:3000'],
            selector: 'app=hades-web'
          }
        ]
      },
      '/kubernetes/autoscalers': {
        autoscalers: [
          {
            id: 'hpa-1',
            name: 'hades-api-hpa',
            namespace: 'default',
            min_replicas: 2,
            max_replicas: 10,
            current_replicas: 3,
            target_cpu: 70,
            status: 'active'
          },
          {
            id: 'hpa-2',
            name: 'hades-web-hpa',
            namespace: 'default',
            min_replicas: 1,
            max_replicas: 5,
            current_replicas: 2,
            target_cpu: 80,
            status: 'active'
          }
        ]
      },
      '/threat/vulnerabilities': {
        vulnerabilities: [
          {
            id: 'vuln-1',
            name: 'Critical SQL Injection Vulnerability',
            severity: 'critical',
            cvss_score: 9.8,
            status: 'open',
            discovered_at: '2026-04-15T10:30:00Z',
            affected_systems: ['web-api', 'database'],
            description: 'SQL injection vulnerability in authentication endpoint'
          },
          {
            id: 'vuln-2',
            name: 'Cross-Site Scripting (XSS)',
            severity: 'high',
            cvss_score: 7.5,
            status: 'mitigated',
            discovered_at: '2026-04-10T14:20:00Z',
            affected_systems: ['web-dashboard'],
            description: 'Reflected XSS in search functionality'
          }
        ]
      },
      '/threat/mitigations': {
        mitigations: [
          {
            id: 'mit-1',
            name: 'Input Validation Enhancement',
            type: 'technical',
            status: 'implemented',
            effectiveness: 95,
            implemented_at: '2026-04-16T09:00:00Z',
            description: 'Enhanced input validation for all user inputs'
          },
          {
            id: 'mit-2',
            name: 'Security Training Program',
            type: 'administrative',
            status: 'in_progress',
            effectiveness: 80,
            implemented_at: '2026-04-01T08:00:00Z',
            description: 'Comprehensive security awareness training for all staff'
          }
        ]
      },
      '/threat/simulations': {
        simulations: [
          {
            id: 'sim-1',
            name: 'APT29 Advanced Persistent Threat Simulation',
            type: 'red_team',
            status: 'completed',
            duration: 72,
            success_rate: 85,
            conducted_at: '2026-04-01T00:00:00Z',
            description: 'Simulated APT29 attack scenario with lateral movement'
          },
          {
            id: 'sim-2',
            name: 'Ransomware Attack Simulation',
            type: 'purple_team',
            status: 'scheduled',
            duration: 48,
            success_rate: 0,
            scheduled_at: '2026-05-01T00:00:00Z',
            description: 'Ransomware attack simulation with defense validation'
          }
        ]
      },
      '/threat/models': {
        models: [
          {
            id: 'model-1',
            name: 'MITRE ATT&CK Framework',
            version: 'v12.1',
            coverage: 95,
            last_updated: '2026-04-15T00:00:00Z',
            techniques_count: 195,
            description: 'Comprehensive MITRE ATT&CK threat model'
          },
          {
            id: 'model-2',
            name: 'Custom Enterprise Threat Model',
            version: 'v2.0',
            coverage: 88,
            last_updated: '2026-04-10T00:00:00Z',
            techniques_count: 172,
            description: 'Enterprise-specific threat model tailored to organization'
          }
        ]
      },
      '/threat/scenarios': {
        scenarios: [
          {
            id: 'scenario-1',
            name: 'Supply Chain Attack Scenario',
            complexity: 'high',
            likelihood: 'medium',
            impact: 'critical',
            duration: 168,
            description: 'Simulated supply chain compromise with third-party vendor'
          },
          {
            id: 'scenario-2',
            name: 'Insider Threat Scenario',
            complexity: 'medium',
            likelihood: 'low',
            impact: 'high',
            duration: 24,
            description: 'Malicious insider data exfiltration scenario'
          }
        ]
      },
      '/incident/playbooks': {
        playbooks: [
          {
            id: 'playbook-1',
            name: 'Ransomware Response Playbook',
            category: 'malware',
            severity: 'critical',
            status: 'active',
            last_updated: '2026-04-15T00:00:00Z',
            steps: 12,
            estimated_duration: 240,
            description: 'Comprehensive ransomware incident response procedures'
          },
          {
            id: 'playbook-2',
            name: 'Data Breach Response Playbook',
            category: 'data_loss',
            severity: 'high',
            status: 'active',
            last_updated: '2026-04-10T00:00:00Z',
            steps: 15,
            estimated_duration: 480,
            description: 'Data breach containment and notification procedures'
          }
        ]
      },
      '/incident/incidents': {
        incidents: [
          {
            id: 'incident-1',
            title: 'Critical Ransomware Attack',
            severity: 'critical',
            status: 'active',
            created_at: '2026-04-15T14:30:00Z',
            affected_systems: ['file-server', 'backup-system'],
            assigned_to: 'incident-response-team',
            description: 'Ransomware attack detected on critical infrastructure'
          },
          {
            id: 'incident-2',
            title: 'Suspicious Login Activity',
            severity: 'medium',
            status: 'investigating',
            created_at: '2026-04-16T09:15:00Z',
            affected_systems: ['auth-system'],
            assigned_to: 'security-analyst',
            description: 'Multiple failed login attempts detected from unusual locations'
          }
        ]
      },
      '/incident/active-responses': {
        active_responses: [
          {
            id: 'response-1',
            incident_id: 'incident-1',
            action: 'network_isolation',
            status: 'in_progress',
            started_at: '2026-04-15T15:00:00Z',
            estimated_completion: '2026-04-15T16:30:00Z',
            description: 'Isolating affected systems from network'
          },
          {
            id: 'response-2',
            incident_id: 'incident-1',
            action: 'backup_restoration',
            status: 'pending',
            started_at: null,
            estimated_completion: '2026-04-16T12:00:00Z',
            description: 'Restoring systems from clean backups'
          }
        ]
      },
      '/incident/response-actions': {
        response_actions: [
          {
            id: 'action-1',
            name: 'Isolate Compromised System',
            category: 'containment',
            automation_level: 'semi_automated',
            execution_time: 5,
            description: 'Network isolation of compromised endpoints'
          },
          {
            id: 'action-2',
            name: 'Collect Forensic Evidence',
            category: 'investigation',
            automation_level: 'manual',
            execution_time: 120,
            description: 'Memory and disk image collection for forensic analysis'
          },
          {
            id: 'action-3',
            name: 'Notify Stakeholders',
            category: 'communication',
            automation_level: 'automated',
            execution_time: 2,
            description: 'Automatic notification to incident response team and management'
          }
        ]
      },
      '/siem/events': {
        events: [
          {
            id: 'event-1',
            timestamp: '2026-04-16T10:30:00Z',
            severity: 'high',
            source: 'firewall',
            event_type: 'intrusion_attempt',
            details: 'Suspicious traffic detected from external IP',
            status: 'investigating'
          },
          {
            id: 'event-2',
            timestamp: '2026-04-16T10:25:00Z',
            severity: 'medium',
            source: 'ids',
            event_type: 'malware_detection',
            details: 'Potential malware signature detected',
            status: 'contained'
          },
          {
            id: 'event-3',
            timestamp: '2026-04-16T10:20:00Z',
            severity: 'low',
            source: 'authentication',
            event_type: 'failed_login',
            details: 'Multiple failed login attempts detected',
            status: 'monitoring'
          }
        ]
      },
      '/siem/alerts': {
        alerts: [
          {
            id: 'alert-1',
            title: 'Critical Security Alert',
            severity: 'critical',
            created_at: '2026-04-16T10:30:00Z',
            source: 'siem',
            status: 'active',
            description: 'Critical security incident requiring immediate attention'
          },
          {
            id: 'alert-2',
            title: 'Anomaly Detection Alert',
            severity: 'medium',
            created_at: '2026-04-16T10:15:00Z',
            source: 'behavioral_analysis',
            status: 'investigating',
            description: 'Unusual user activity pattern detected'
          }
        ]
      },
      '/siem/correlations': {
        correlations: [
          {
            id: 'correlation-1',
            name: 'Lateral Movement Detection',
            confidence: 85,
            events_count: 5,
            time_window: 300,
            description: 'Correlated events indicating potential lateral movement'
          },
          {
            id: 'correlation-2',
            name: 'Data Exfiltration Pattern',
            confidence: 92,
            events_count: 8,
            time_window: 600,
            description: 'Pattern consistent with data exfiltration activity'
          }
        ]
      },
      '/siem/threat-feeds': {
        threat_feeds: [
          {
            id: 'feed-1',
            name: 'Malware Intelligence Feed',
            source: 'vendor_intelligence',
            last_updated: '2026-04-16T09:00:00Z',
            indicators_count: 1547,
            status: 'active',
            description: 'Real-time malware indicators and signatures'
          },
          {
            id: 'feed-2',
            name: 'Threat Actor Intelligence',
            source: 'threat_intel',
            last_updated: '2026-04-16T08:30:00Z',
            indicators_count: 892,
            status: 'active',
            description: 'Known threat actor TTPs and infrastructure'
          }
        ]
      },
      '/analytics/overview': {
        overview: {
          total_events: 1547,
          critical_alerts: 23,
          ml_accuracy: 94.5,
          prediction_confidence: 87.2,
          last_updated: '2026-04-16T10:30:00Z'
        }
      },
      '/analytics/ml-insights': {
        insights: [
          {
            id: 'insight-1',
            type: 'anomaly_detection',
            confidence: 92.3,
            description: 'Unusual traffic patterns detected in network segment',
            severity: 'high',
            timestamp: '2026-04-16T10:15:00Z'
          },
          {
            id: 'insight-2',
            type: 'behavioral_analysis',
            confidence: 88.7,
            description: 'User behavior deviation from baseline patterns',
            severity: 'medium',
            timestamp: '2026-04-16T09:45:00Z'
          }
        ]
      },
      '/analytics/predictions': {
        predictions: [
          {
            id: 'pred-1',
            type: 'threat_prediction',
            probability: 0.87,
            timeframe: '24h',
            description: 'High probability of ransomware activity',
            confidence: 85.2
          },
          {
            id: 'pred-2',
            type: 'resource_usage',
            probability: 0.92,
            timeframe: '6h',
            description: 'Expected spike in security event volume',
            confidence: 91.8
          }
        ]
      },
      '/analytics/metrics': {
        metrics: {
          detection_rate: 94.5,
          false_positive_rate: 3.2,
          response_time: 0.8,
          coverage: 98.7,
          efficiency: 91.3
        }
      },
      '/analytics/trends': {
        trends: [
          {
            metric: 'threat_volume',
            direction: 'increasing',
            change_percentage: 12.5,
            period: '7d',
            significance: 'high'
          },
          {
            metric: 'detection_accuracy',
            direction: 'stable',
            change_percentage: 0.8,
            period: '30d',
            significance: 'low'
          }
        ]
      },
      '/analytics/reports': {
        reports: [
          {
            id: 'report-1',
            title: 'Weekly Security Analytics Report',
            generated_at: '2026-04-15T00:00:00Z',
            format: 'pdf',
            size: '2.4MB'
          }
        ]
      },
      '/analytics/export': {
        export_data: {
          format: 'csv',
          size: '15.2MB',
          records: 1547,
          generated_at: '2026-04-16T10:30:00Z'
        }
      },
      '/ai/threats': {
        threats: [
          {
            id: 'ai-threat-1',
            name: 'AI-Detected Advanced Threat',
            confidence: 94.2,
            severity: 'critical',
            detected_at: '2026-04-16T10:30:00Z',
            description: 'Advanced persistent threat detected by ML algorithms'
          }
        ]
      },
      '/ai/anomalies': {
        anomalies: [
          {
            id: 'anomaly-1',
            type: 'network_behavior',
            severity: 'high',
            confidence: 89.7,
            detected_at: '2026-04-16T10:25:00Z',
            description: 'Unusual network traffic patterns detected'
          }
        ]
      },
      '/ai/predictions': {
        predictions: [
          {
            id: 'ai-pred-1',
            type: 'threat_probability',
            probability: 0.87,
            confidence: 91.3,
            timeframe: '24h',
            description: 'High probability of security incident'
          }
        ]
      },
      '/ai/overview': {
        overview: {
          total_threats: 127,
          detected_threats: 124,
          blocked_threats: 119,
          accuracy: 97.6,
          last_updated: '2026-04-16T10:30:00Z'
        }
      },
      '/threat-hunting/hunts': {
        hunts: [
          {
            id: 'hunt-1',
            name: 'APT29 lateral movement hunt',
            status: 'active',
            started_at: '2026-04-16T08:00:00Z',
            progress: 67,
            threats_found: 3,
            description: 'Hunting for APT29 lateral movement techniques'
          },
          {
            id: 'hunt-2',
            name: 'Ransomware precursor hunt',
            status: 'completed',
            started_at: '2026-04-15T14:00:00Z',
            completed_at: '2026-04-15T18:30:00Z',
            progress: 100,
            threats_found: 7,
            description: 'Proactive hunt for ransomware precursors'
          }
        ]
      },
      '/threat-hunting/threats': {
        threats: [
          {
            id: 'hunt-threat-1',
            hunt_id: 'hunt-1',
            type: 'lateral_movement',
            severity: 'high',
            detected_at: '2026-04-16T09:45:00Z',
            description: 'Potential lateral movement activity detected'
          }
        ]
      },
      '/threat-hunting/indicators': {
        indicators: [
          {
            id: 'indicator-1',
            type: 'ioc',
            value: 'malicious-domain.com',
            confidence: 94.2,
            source: 'threat_hunt',
            description: 'Malicious domain identified during hunt'
          }
        ]
      },
      '/threat-hunting/start': {
        status: 'started',
        hunt_id: 'hunt-3',
        message: 'Threat hunt initiated successfully'
      },
      '/threat-hunting/stop': {
        status: 'stopped',
        hunt_id: 'hunt-1',
        message: 'Threat hunt stopped successfully'
      },
      '/threat-hunting/create': {
        hunt_id: 'hunt-3',
        status: 'created',
        message: 'New threat hunt created successfully'
      },
      '/threat-hunting/results/hunt-1': {
        results: [
          {
            id: 'result-1',
            type: 'threat',
            severity: 'medium',
            timestamp: '2026-04-16T09:30:00Z',
            description: 'Suspicious process execution detected'
          }
        ]
      },
      '/blockchain/audit-logs': {
        audit_logs: [
          {
            id: 'audit-1',
            timestamp: '2026-04-16T10:30:00Z',
            action: 'user_login',
            user: 'admin',
            resource: 'security_dashboard',
            hash: '0xabc123...',
            block_hash: '0xdef456...'
          }
        ]
      },
      '/blockchain/blocks': {
        blocks: [
          {
            id: 'block-1',
            hash: '0xdef456...',
            previous_hash: '0xabc123...',
            timestamp: '2026-04-16T10:30:00Z',
            transactions_count: 5,
            size: '2.1KB'
          }
        ]
      },
      '/blockchain/transactions': {
        transactions: [
          {
            id: 'tx-1',
            hash: '0xghi789...',
            block_hash: '0xdef456...',
            timestamp: '2026-04-16T10:30:00Z',
            type: 'audit_log',
            status: 'confirmed'
          }
        ]
      },
      '/blockchain/integrity': {
        integrity: {
          is_valid: true,
          last_verified: '2026-04-16T10:30:00Z',
          total_blocks: 1547,
          chain_hash: '0xchain123...'
        }
      },
      '/blockchain/verify': {
        verification: {
          status: 'valid',
          verified_at: '2026-04-16T10:30:00Z',
          confidence: 99.9
        }
      },
      '/blockchain/audit-log': {
        status: 'added',
        log_id: 'audit-2',
        hash: '0xxyz789...'
      },
      '/blockchain/blocks/block-1': {
        block: {
          id: 'block-1',
          hash: '0xdef456...',
          previous_hash: '0xabc123...',
          timestamp: '2026-04-16T10:30:00Z',
          nonce: 12345,
          transactions: ['tx-1', 'tx-2']
        }
      },
      '/blockchain/transactions/tx-1': {
        transaction: {
          id: 'tx-1',
          hash: '0xghi789...',
          block_hash: '0xdef456...',
          timestamp: '2026-04-16T10:30:00Z',
          from: 'system',
          to: 'blockchain',
          data: 'audit_log_entry'
        }
      },
      '/zerotrust/policies': {
        policies: [
          {
            id: 'policy-1',
            name: 'Admin Access Policy',
            status: 'active',
            created_at: '2026-04-15T08:00:00Z',
            rules_count: 12,
            description: 'Administrative access control policy'
          }
        ]
      },
      '/zerotrust/access-requests': {
        access_requests: [
          {
            id: 'request-1',
            user: 'analyst1',
            resource: 'threat_intelligence',
            status: 'pending',
            requested_at: '2026-04-16T10:15:00Z',
            trust_score: 78.5
          }
        ]
      },
      '/zerotrust/trust-scores': {
        trust_scores: [
          {
            user_id: 'admin',
            score: 95.2,
            last_updated: '2026-04-16T10:30:00Z',
            factors: ['authentication', 'behavior', 'location']
          }
        ]
      },
      '/zerotrust/network-segments': {
        segments: [
          {
            id: 'segment-1',
            name: 'Security Operations',
            trust_level: 'high',
            devices_count: 47,
            isolation_status: 'isolated'
          }
        ]
      },
      '/zerotrust/policies/update': {
        status: 'updated',
        policy_id: 'policy-1',
        message: 'Policy updated successfully'
      },
      '/zerotrust/access-process': {
        status: 'processed',
        request_id: 'request-1',
        decision: 'approved',
        message: 'Access request processed successfully'
      },
      '/zerotrust/policies/create': {
        status: 'created',
        policy_id: 'policy-2',
        message: 'New policy created successfully'
      },
      '/zerotrust/trust-scores/update': {
        status: 'updated',
        user_id: 'analyst1',
        new_score: 82.3,
        message: 'Trust score updated successfully'
      },
      '/quantum/algorithms': {
        algorithms: [
          {
            id: 'algo-1',
            name: 'Kyber-768',
            type: 'key_exchange',
            security_level: 128,
            status: 'active',
            description: 'Post-quantum key exchange algorithm'
          }
        ]
      },
      '/quantum/keys': {
        keys: [
          {
            id: 'key-1',
            algorithm_id: 'algo-1',
            created_at: '2026-04-16T09:00:00Z',
            expires_at: '2026-05-16T09:00:00Z',
            status: 'active',
            key_size: 2400
          }
        ]
      },
      '/quantum/certificates': {
        certificates: [
          {
            id: 'cert-1',
            key_id: 'key-1',
            issued_at: '2026-04-16T09:05:00Z',
            expires_at: '2026-05-16T09:05:00Z',
            status: 'valid',
            issuer: 'hades_quantum_ca'
          }
        ]
      },
      '/quantum/metrics': {
        metrics: {
          total_keys: 47,
          active_keys: 42,
          expired_keys: 5,
          security_score: 98.7,
          last_rotation: '2026-04-15T00:00:00Z'
        }
      },
      '/quantum/generate-key': {
        status: 'generated',
        key_id: 'key-2',
        algorithm_id: 'algo-1',
        message: 'New quantum key generated successfully'
      },
      '/quantum/rotate-keys': {
        status: 'rotated',
        rotated_keys: 5,
        message: 'Key rotation completed successfully'
      },
      '/quantum/verify': {
        status: 'verified',
        security_level: 'high',
        confidence: 99.9,
        message: 'Quantum security verification completed'
      },
      '/quantum/algorithms/algo-1': {
        algorithm: {
          id: 'algo-1',
          name: 'Kyber-768',
          type: 'key_exchange',
          security_level: 128,
          key_size: 2400,
          description: 'Post-quantum key exchange algorithm',
          performance: {
            keygen_time: '2.3ms',
            encap_time: '1.8ms',
            decap_time: '1.6ms'
          }
        }
      },
      '/security/policies': {
        policies: [
          {
            id: 'policy-1',
            name: 'Access Control Policy',
            type: 'access_control',
            status: 'active',
            severity: 'high',
            created_at: '2026-04-15T08:00:00Z',
            description: 'Comprehensive access control and authentication policies'
          },
          {
            id: 'policy-2',
            name: 'Data Protection Policy',
            type: 'data_protection',
            status: 'active',
            severity: 'critical',
            created_at: '2026-04-10T14:00:00Z',
            description: 'Data encryption and protection requirements'
          },
          {
            id: 'policy-3',
            name: 'Network Security Policy',
            type: 'network_security',
            status: 'active',
            severity: 'high',
            created_at: '2026-04-12T10:30:00Z',
            description: 'Network segmentation and firewall policies'
          }
        ]
      },
      '/security/vulnerabilities': {
        vulnerabilities: [
          {
            id: 'vuln-1',
            name: 'SQL Injection Vulnerability',
            severity: 'critical',
            cvss_score: 9.8,
            status: 'open',
            discovered_at: '2026-04-15T10:30:00Z',
            affected_systems: ['web-api', 'database'],
            description: 'SQL injection vulnerability in authentication endpoint'
          },
          {
            id: 'vuln-2',
            name: 'Cross-Site Scripting (XSS)',
            severity: 'high',
            cvss_score: 7.5,
            status: 'mitigated',
            discovered_at: '2026-04-10T14:20:00Z',
            affected_systems: ['web-dashboard'],
            description: 'Reflected XSS in search functionality'
          },
          {
            id: 'vuln-3',
            name: 'Outdated OpenSSL Version',
            severity: 'medium',
            cvss_score: 5.3,
            status: 'open',
            discovered_at: '2026-04-08T09:15:00Z',
            affected_systems: ['api-server'],
            description: 'OpenSSL version requires security update'
          }
        ]
      },
      '/security/score': {
        security_score: {
          overall_score: 87.5,
          last_updated: '2026-04-16T10:30:00Z',
          components: {
            access_control: 92.3,
            data_protection: 85.7,
            network_security: 89.1,
            vulnerability_management: 82.9
          },
          trend: 'improving',
          recommendations: [
            'Update OpenSSL to latest version',
            'Implement additional input validation',
            'Review access control policies'
          ]
        }
      },
      '/security/vulnerabilities?': {
        vulnerabilities: [
          {
            id: 'vuln-1',
            name: 'SQL Injection Vulnerability',
            severity: 'critical',
            cvss_score: 9.8,
            status: 'open',
            discovered_at: '2026-04-15T10:30:00Z',
            affected_systems: ['web-api', 'database'],
            description: 'SQL injection vulnerability in authentication endpoint'
          },
          {
            id: 'vuln-2',
            name: 'Cross-Site Scripting (XSS)',
            severity: 'high',
            cvss_score: 7.5,
            status: 'mitigated',
            discovered_at: '2026-04-10T14:20:00Z',
            affected_systems: ['web-dashboard'],
            description: 'Reflected XSS in search functionality'
          },
          {
            id: 'vuln-3',
            name: 'Outdated OpenSSL Version',
            severity: 'medium',
            cvss_score: 5.3,
            status: 'open',
            discovered_at: '2026-04-08T09:15:00Z',
            affected_systems: ['api-server'],
            description: 'OpenSSL version requires security update'
          }
        ]
      }
    }
    
    const result = fallbacks[endpoint] || null
    console.log(`getFallbackData result for ${endpoint}:`, result)
    return result
  }
}

export default API_CONFIG
