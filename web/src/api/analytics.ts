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
  getMetrics: async (): Promise<any> => {
    return await API_CONFIG.request('/analytics/metrics')
  },

  // Get trend analysis
  getTrends: async (timeframe: string): Promise<any> => {
    return await API_CONFIG.request(`/analytics/trends?timeframe=${timeframe}`)
  },

  // Get detailed analytics
  getDetailedAnalytics: async (params: Record<string, string>): Promise<any> => {
    const queryString = new URLSearchParams(params).toString()
    return await API_CONFIG.request(`/analytics/detailed?${queryString}`)
  },

  // Generate report
  generateReport: async (params: Record<string, string>): Promise<any> => {
    return await API_CONFIG.request('/analytics/reports', {
      method: 'POST',
      body: JSON.stringify(params)
    })
  },

  // Export analytics data
  exportData: async (format: string): Promise<any> => {
    return await API_CONFIG.request(`/analytics/export?format=${format}`)
  }
}
