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
        apiBase || 'http://localhost:8000/api/v2'
      ],
      testing: [
        apiBase || 'http://localhost:8000/api/v2'
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
    return endpoints[API_CONFIG.currentEndpointIndex]
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
  handleResponse: async (response: Response) => {
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`)
    }
    
    const data = await response.json()
    return data.data || data
  },

  // Request wrapper with distributed failover
  request: async (endpoint: string, options: RequestInit = {}) => {
    const currentEnv = API_CONFIG.getEnvironment()
    const endpoints = API_CONFIG.getAPIEndpoints()
    let lastError = null
    
    // Try each endpoint with failover
    for (let i = 0; i < endpoints.length; i++) {
      const baseURL = endpoints[i]
      const url = `${baseURL}${endpoint}`
      
      const authHeaders = API_CONFIG.getAuthHeaders()
      const config = {
        ...options,
        headers: {
          ...API_CONFIG.getDefaultHeaders(),
          ...authHeaders,
          ...options.headers
        }
      }

      try {
        const response = await fetch(url, config as RequestInit)
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        
        const data = await response.json()
        
        // Update current endpoint index on success
        API_CONFIG.currentEndpointIndex = i
        return data.data || data
      } catch (error) {
        lastError = error as any
        continue
      }
    }
    
    // All endpoints failed
    console.error(`All API endpoints failed for ${endpoint}:`, lastError)
    
    throw lastError || new Error('All API endpoints failed')
  },

  // Hot-swap configuration
  hotSwapConfig: async (newConfig: { endpoints: string[] }) => {
    try {
      // Validate new configuration
      if (!newConfig.endpoints || !Array.isArray(newConfig.endpoints)) {
        throw new Error('Invalid configuration: endpoints must be an array')
      }
      
      // Update configuration
      API_CONFIG.getAPIEndpoints = () => newConfig.endpoints
      
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
      endpoints: [] as Array<{ url: string; status: string; responseTime?: string; error?: string }>,
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
          error: error instanceof Error ? error.message : 'Unknown error'
        })
      }
    }
    
    return status
  },

  }

export default API_CONFIG