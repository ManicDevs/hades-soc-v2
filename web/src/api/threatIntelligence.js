// AI-Powered Threat Intelligence API
import API_CONFIG from './config'

export const threatIntelligenceAPI = {
  // Get AI detected threats
  getThreats: async () => {
    return await API_CONFIG.request('/ai/threats')
  },

  // Get anomaly detection results
  getAnomalies: async () => {
    return await API_CONFIG.request('/ai/anomalies')
  },

  // Get ML predictions
  getPredictions: async () => {
    return await API_CONFIG.request('/ai/predictions')
  },

  // Get threat intelligence overview
  getOverview: async () => {
    return await API_CONFIG.request('/ai/overview')
  },

  // Analyze threat
  analyzeThreat: async (threatId) => {
    return await API_CONFIG.request('/ai/analyze', {
      method: 'POST',
      body: JSON.stringify({ threat_id: threatId })
    })
  },

  // Update threat intelligence
  updateIntelligence: async (data) => {
    return await API_CONFIG.request('/ai/update', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  }
}
