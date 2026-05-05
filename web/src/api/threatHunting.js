// Threat Hunting API
import API_CONFIG from './config'

export const threatHuntingAPI = {
  // Get threat hunts
  getHunts: async () => {
    return await API_CONFIG.request('/threat-hunting/hunts')
  },

  // Get detected threats
  getThreats: async () => {
    return await API_CONFIG.request('/threat-hunting/threats')
  },

  // Get threat indicators
  getIndicators: async () => {
    return await API_CONFIG.request('/threat-hunting/indicators')
  },

  // Start threat hunt
  startHunt: async (huntId) => {
    return await API_CONFIG.request('/threat-hunting/start', {
      method: 'POST',
      body: JSON.stringify({ hunt_id: huntId })
    })
  },

  // Stop threat hunt
  stopHunt: async (huntId) => {
    return await API_CONFIG.request('/threat-hunting/stop', {
      method: 'POST',
      body: JSON.stringify({ hunt_id: huntId })
    })
  },

  // Create hunt
  createHunt: async (huntData) => {
    return await API_CONFIG.request('/threat-hunting/create', {
      method: 'POST',
      body: JSON.stringify(huntData)
    })
  },

  // Get hunt results
  getHuntResults: async (huntId) => {
    return await API_CONFIG.request(`/threat-hunting/results/${huntId}`)
  }
}
