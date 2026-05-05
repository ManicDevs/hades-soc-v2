// Advanced SIEM Integration API
import API_CONFIG from './config'

export const siemAPI = {
  // Get security events
  getEvents: async () => {
    return await API_CONFIG.request('/siem/events')
  },

  // Get SIEM alerts
  getAlerts: async () => {
    return await API_CONFIG.request('/siem/alerts')
  },

  // Get event correlations
  getCorrelations: async () => {
    return await API_CONFIG.request('/siem/correlations')
  },

  // Get threat intelligence feeds
  getThreatFeeds: async () => {
    return await API_CONFIG.request('/siem/threat-feeds')
  },

  // Acknowledge alert
  acknowledgeAlert: async (alertId) => {
    return await API_CONFIG.request('/siem/alerts/acknowledge', {
      method: 'POST',
      body: JSON.stringify({ alert_id: alertId })
    })
  },

  // Create correlation rule
  createCorrelation: async (ruleData) => {
    return await API_CONFIG.request('/siem/correlations/create', {
      method: 'POST',
      body: JSON.stringify(ruleData)
    })
  },

  // Update threat feed
  updateThreatFeed: async (feedId, feedData) => {
    return await API_CONFIG.request(`/siem/threat-feeds/${feedId}`, {
      method: 'PUT',
      body: JSON.stringify(feedData)
    })
  },

  // Get event details
  getEventDetails: async (eventId) => {
    return await API_CONFIG.request(`/siem/events/${eventId}`)
  },

  // Export SIEM data
  exportData: async (format, filters) => {
    return await API_CONFIG.request('/siem/export', {
      method: 'POST',
      body: JSON.stringify({ format, filters })
    })
  }
}
