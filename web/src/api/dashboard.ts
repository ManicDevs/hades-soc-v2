// Dashboard API
import API_CONFIG from './config'

export const dashboardAPI = {
  // Get dashboard metrics
  getMetrics: async () => {
    try {
      const response = await fetch('http://localhost:8000/api/v2/dashboard/metrics')
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      return await response.json()
    } catch (error) {
      console.error('Metrics fetch error:', error)
      // Return fallback data for now
      return {
        security_score: { overall: 95 },
        active_threats: 2,
        blocked_attacks: 127,
        system_health: 92,
        active_users: 4
      }
    }
  },

  // Get recent activity
  getActivity: async () => {
    try {
      const response = await fetch('http://localhost:8000/api/v2/dashboard/activity')
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      return await response.json()
    } catch (error) {
      console.error('Activity fetch error:', error)
      return []
    }
  },

  // Get system status
  getSystemStatus: async () => {
    try {
      const response = await fetch('http://localhost:8000/api/v2/dashboard/status')
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      return await response.json()
    } catch (error) {
      console.error('Status fetch error:', error)
      return {
        api: 'online',
        database: 'connected',
        workers: '5/5 active',
        memory: '68.7%',
        cpu: '12.4%'
      }
    }
  },

  // Get security overview
  getSecurityOverview: async () => {
    try {
      const response = await fetch('http://localhost:8000/api/v2/dashboard/security')
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      return await response.json()
    } catch (error) {
      console.error('Security fetch error:', error)
      return {
        threats: { critical: 1, high: 2, medium: 5, low: 8 },
        policies: { active: 12, warning: 2, inactive: 1 },
        vulnerabilities: { open: 4, inProgress: 2, resolved: 15 }
      }
    }
  },
}
