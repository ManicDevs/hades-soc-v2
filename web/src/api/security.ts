// Security API
import API_CONFIG from './config'

export const securityAPI = {
  // Get security policies
  getPolicies: async () => {
    return await API_CONFIG.request('/security/policies')
  },

  // Update security policy
  updatePolicy: async (id: string, policyData: Record<string, any>) => {
    return await API_CONFIG.request(`/security/policies/${id}`, {
      method: 'PUT',
      body: JSON.stringify(policyData)
    })
  },

  // Get vulnerabilities
  getVulnerabilities: async (filters = {}) => {
    const params = new URLSearchParams(filters)
    return await API_CONFIG.request(`/security/vulnerabilities?${params}`)
  },

  // Update vulnerability status
  updateVulnerability: async (id: string, status: string) => {
    return await API_CONFIG.request(`/security/vulnerabilities/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    })
  },

  // Get security score
  getSecurityScore: async () => {
    return await API_CONFIG.request('/security/score')
  },

  // Run security scan
  runSecurityScan: async () => {
    const base = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    const response = await fetch(`${base}/security/scan`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('hades-token')}`,
      },
    })
    
    if (!response.ok) {
      throw new Error('Failed to run security scan')
    }
    
    return response.json()
  },

  // Get audit logs
  getAuditLogs: async (filters = {}) => {
    const params = new URLSearchParams(filters)
    const base = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    const response = await fetch(`${base}/security/audit-logs?${params}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('hades-token')}`,
      },
    })
    
    if (!response.ok) {
      throw new Error('Failed to fetch audit logs')
    }
    
    return response.json()
  },
}
