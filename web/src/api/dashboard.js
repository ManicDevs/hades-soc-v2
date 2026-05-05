// Dashboard API
import API_CONFIG from './config'

export const dashboardAPI = {
  // Get dashboard metrics
  getMetrics: async () => {
    return await API_CONFIG.request('/dashboard/metrics')
  },

  // Get recent activity
  getActivity: async () => {
    return await API_CONFIG.request('/dashboard/activity')
  },

  // Get system status
  getSystemStatus: async () => {
    return await API_CONFIG.request('/dashboard/status')
  },

  // Get security overview
  getSecurityOverview: async () => {
    return await API_CONFIG.request('/dashboard/security')
  },
}
