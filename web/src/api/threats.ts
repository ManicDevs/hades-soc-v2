// Threats API
import API_CONFIG from './config'

export const threatsAPI = {
  // Get all threats
  getThreats: async (filters = {}) => {
    const params = new URLSearchParams(filters)
    return await API_CONFIG.request(`/threats?${params}`, {
      method: 'GET'
    })
  },

  // Get threat by ID
  getThreat: async (id: string) => {
    return await API_CONFIG.request(`/threats/${id}`, {
      method: 'GET'
    })
  },

  // Update threat status
  updateThreatStatus: async (id: string, status: string) => {
    return await API_CONFIG.request(`/threats/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    })
  },

  // Get threat statistics
  getThreatStats: async () => {
    return await API_CONFIG.request('/threats/stats', {
      method: 'GET'
    })
  },

  // Block threat
  blockThreat: async (id: string) => {
    return await API_CONFIG.request(`/threats/${id}/block`, {
      method: 'POST'
    })
  },
}
