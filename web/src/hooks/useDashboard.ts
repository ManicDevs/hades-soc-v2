import { useState, useEffect } from 'react'
import { dashboardAPI } from '../api/dashboard'

export const useDashboard = () => {
  const [metrics, setMetrics] = useState<any>(null)
  const [activity, setActivity] = useState<any[]>([])
  const [systemStatus, setSystemStatus] = useState(null)
  const [securityOverview, setSecurityOverview] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchDashboardData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchMetrics() // Only update metrics frequently
    }, 5000)

    return () => clearInterval(interval)
  }, [])

  const fetchDashboardData = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const [metricsData, activityData, statusData, securityData] = await Promise.all([
        dashboardAPI.getMetrics(),
        dashboardAPI.getActivity(),
        dashboardAPI.getSystemStatus(),
        dashboardAPI.getSecurityOverview()
      ])
      
      setMetrics(metricsData)
      setActivity(activityData)
      setSystemStatus(statusData)
      setSecurityOverview(securityData)
    } catch (error) {
      setError('Failed to fetch dashboard data')
      console.error('Dashboard data fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchMetrics = async () => {
    try {
      const metricsData = await dashboardAPI.getMetrics()
      setMetrics(metricsData)
    } catch (error) {
      console.error('Metrics fetch error:', error)
    }
  }

  const refreshData = () => {
    fetchDashboardData()
  }

  return {
    metrics,
    activity,
    systemStatus,
    securityOverview,
    loading,
    error,
    refreshData,
    fetchMetrics
  }
}

export default useDashboard
