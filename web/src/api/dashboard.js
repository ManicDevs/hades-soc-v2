// Dashboard API
import API_CONFIG from './config'

export const dashboardAPI = {
  // Get dashboard metrics
  getMetrics: async () => {
    return await API_CONFIG.request('/dashboard/metrics', {
      method: 'GET'
    })
  },

  // Get recent activity
  getActivity: async () => {
    return await API_CONFIG.request('/dashboard/activity', {
      method: 'GET'
    })
  },

  // Get system status
  getSystemStatus: async () => {
    return await API_CONFIG.request('/dashboard/status', {
      method: 'GET'
    })
  },

  // Get security overview
  getSecurityOverview: async () => {
    return await API_CONFIG.request('/dashboard/security', {
      method: 'GET'
    })
  },
}
