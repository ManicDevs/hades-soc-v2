// Zero-Trust Network Architecture API
import API_CONFIG from './config'

export const zeroTrustAPI = {
  // Get access policies
  getPolicies: async () => {
    return await API_CONFIG.request('/zerotrust/policies')
  },

  // Get access requests
  getAccessRequests: async () => {
    return await API_CONFIG.request('/zerotrust/access-requests')
  },

  // Get trust scores
  getTrustScores: async () => {
    return await API_CONFIG.request('/zerotrust/trust-scores')
  },

  // Get network segments
  getNetworkSegments: async (): Promise<any> => {
    return await API_CONFIG.request('/zerotrust/network-segments')
  },

  // Update policy
  updatePolicy: async (policyId: string, status: string): Promise<any> => {
    return await API_CONFIG.request('/zerotrust/policies/update', {
      method: 'POST',
      body: JSON.stringify({ policy_id: policyId, status })
    })
  },

  // Process access request
  processAccessRequest: async (requestId: string, decision: string): Promise<any> => {
    return await API_CONFIG.request('/zerotrust/access-requests', {
      method: 'POST',
      body: JSON.stringify({ request_id: requestId, decision })
    })
  },

  // Create policy
  createPolicy: async (policyData: Record<string, any>) => {
    return await API_CONFIG.request('/zerotrust/policies/create', {
      method: 'POST',
      body: JSON.stringify(policyData)
    })
  },

  // Update trust score
  updateTrustScore: async (userId: string, score: number) => {
    return await API_CONFIG.request('/zerotrust/trust-scores/update', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, score })
    })
  }
}
