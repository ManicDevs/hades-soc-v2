// Advanced Analytics API
import API_CONFIG from './config'

export const analyticsAPI = {
  // Get analytics overview
  getOverview: async () => {
    return await API_CONFIG.request('/analytics/overview')
  },

  // Get ML insights
  getMLInsights: async () => {
    return await API_CONFIG.request('/analytics/ml-insights')
  },

  // Get predictions
  getPredictions: async () => {
    return await API_CONFIG.request('/analytics/predictions')
  },

  // Get performance metrics
  getMetrics: async () => {
    return await API_CONFIG.request('/analytics/metrics')
  },

  // Get trend analysis
  getTrends: async (timeframe) => {
    return await API_CONFIG.request(`/analytics/trends?timeframe=${timeframe}`)
  },

  // Generate report
  generateReport: async (params) => {
    return await API_CONFIG.request('/analytics/reports', {
      method: 'POST',
      body: JSON.stringify(params)
    })
  },

  // Export analytics data
  exportData: async (format) => {
    return await API_CONFIG.request(`/analytics/export?format=${format}`)
  }
}
