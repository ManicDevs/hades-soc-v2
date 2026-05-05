// Threat Modeling and Attack Simulation API
import API_CONFIG from './config'

export const threatModelingAPI = {
  // Get threat models
  getModels: async () => {
    return await API_CONFIG.request('/threat/models')
  },

  // Get attack scenarios
  getScenarios: async () => {
    return await API_CONFIG.request('/threat/scenarios')
  },

  // Get vulnerabilities
  getVulnerabilities: async () => {
    return await API_CONFIG.request('/threat/vulnerabilities')
  },

  // Get mitigations
  getMitigations: async () => {
    return await API_CONFIG.request('/threat/mitigations')
  },

  // Get simulations
  getSimulations: async () => {
    return await API_CONFIG.request('/threat/simulations')
  },

  // Run simulation
  runSimulation: async (scenarioId) => {
    return await API_CONFIG.request('/threat/simulations', {
      method: 'POST',
      body: JSON.stringify({ scenario_id: scenarioId })
    })
  },

  // Create threat model
  createModel: async (modelData) => {
    return await API_CONFIG.request('/threat/models', {
      method: 'POST',
      body: JSON.stringify(modelData)
    })
  },

  // Create attack scenario
  createScenario: async (scenarioData) => {
    return await API_CONFIG.request('/threat/scenarios', {
      method: 'POST',
      body: JSON.stringify(scenarioData)
    })
  },

  // Update vulnerability status
  updateVulnerability: async (vulnId, status) => {
    return await API_CONFIG.request(`/threat/vulnerabilities/${vulnId}`, {
      method: 'PUT',
      body: JSON.stringify({ status })
    })
  },

  // Get simulation results
  getSimulationResults: async (simulationId) => {
    return await API_CONFIG.request(`/threat/simulations/${simulationId}`)
  }
}
