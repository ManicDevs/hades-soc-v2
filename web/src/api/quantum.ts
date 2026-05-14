// Quantum-Resistant Cryptography API
import API_CONFIG from './config'

export const quantumAPI = {
  // Get quantum algorithms
  getAlgorithms: async () => {
    return await API_CONFIG.request('/quantum/algorithms')
  },

  // Get quantum keys
  getKeys: async () => {
    return await API_CONFIG.request('/quantum/keys')
  },

  // Get quantum certificates
  getCertificates: async () => {
    return await API_CONFIG.request('/quantum/certificates')
  },

  // Get security metrics
  getMetrics: async () => {
    return await API_CONFIG.request('/quantum/metrics')
  },

  // Generate quantum key
  generateKey: async (algorithmId: string) => {
    return await API_CONFIG.request('/quantum/generate-key', {
      method: 'POST',
      body: JSON.stringify({ algorithm_id: algorithmId })
    })
  },

  // Rotate keys
  rotateKeys: async () => {
    return await API_CONFIG.request('/quantum/rotate-keys', {
      method: 'POST'
    })
  },

  // Verify quantum security
  verifySecurity: async () => {
    return await API_CONFIG.request('/quantum/verify', {
      method: 'POST'
    })
  },

  // Get algorithm details
  getAlgorithmDetails: async (algorithmId: string) => {
    return await API_CONFIG.request(`/quantum/algorithms/${algorithmId}`)
  }
}
