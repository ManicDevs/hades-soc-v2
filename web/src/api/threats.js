// Threats API
import API_CONFIG from './config'

export const threatsAPI = {
  // Get all threats
  getThreats: async (filters = {}) => {
    const params = new URLSearchParams(filters)
    return await API_CONFIG.request(`/threats?${params}`)
  },

  // Get threat by ID
  getThreat: async (id) => {
    return await API_CONFIG.request(`/threats/${id}`)
  },

  // Update threat status
  updateThreat: async (id, status) => {
    return await API_CONFIG.request(`/threats/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    })
  },

  // Get threat statistics
  getThreatStats: async () => {
    return await API_CONFIG.request('/threats/stats')
  },

  // Block threat
  blockThreat: async (id) => {
    return await API_CONFIG.request(`/threats/${id}/block`, {
      method: 'POST'
    })
  },
}
