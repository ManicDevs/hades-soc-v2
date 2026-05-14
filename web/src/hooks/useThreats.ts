import { useState, useEffect } from 'react'
import { threatsAPI } from '../api/threats'

export const useThreats = () => {
  const [threats, setThreats] = useState<any[]>([])
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filters, setFilters] = useState({
    severity: 'all',
    status: 'all',
    type: 'all'
  })

  useEffect(() => {
    fetchThreats()
    fetchThreatStats()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchThreatStats()
    }, 10000)

    return () => clearInterval(interval)
  }, [filters])

  const fetchThreats = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const threatsData = await threatsAPI.getThreats(filters)
      // Extract threats array from nested API response structure
      setThreats(threatsData?.data?.data || [])
    } catch (error) {
      setError('Failed to fetch threats')
      console.error('Threats fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchThreatStats = async () => {
    try {
      // Use the main threats endpoint which includes stats in metadata
      const threatsData = await threatsAPI.getThreats(filters)
      // Extract stats from the metadata
      const stats = {
        total_threats: threatsData?.data?.metadata?.total_threats || 0,
        by_severity: threatsData?.data?.data?.reduce((acc: any, threat: any) => {
          acc[threat.severity] = (acc[threat.severity] || 0) + 1
          return acc
        }, {}) || {},
        by_status: threatsData?.data?.data?.reduce((acc: any, threat: any) => {
          acc[threat.status] = (acc[threat.status] || 0) + 1
          return acc
        }, {}) || {},
        by_type: threatsData?.data?.data?.reduce((acc: any, threat: any) => {
          acc[threat.type] = (acc[threat.type] || 0) + 1
          return acc
        }, {}) || {}
      }
      setStats(stats)
    } catch (error) {
      console.error('Threat stats fetch error:', error)
    }
  }

  const updateThreatStatus = async (id: any, status: any) => {
    try {
      await threatsAPI.updateThreatStatus(id, status)
      // Refresh threats list
      await fetchThreats()
      await fetchThreatStats()
    } catch (error) {
      setError('Failed to update threat status')
      throw error
    }
  }

  const getThreatDetails = async (id: any) => {
    try {
      return await threatsAPI.getThreat(id)
    } catch (error) {
      setError('Failed to fetch threat details')
      throw error
    }
  }

  const updateFilters = (newFilters: any) => {
    setFilters(prev => ({ ...prev, ...newFilters }))
  }

  const refreshData = () => {
    fetchThreats()
    fetchThreatStats()
  }

  return {
    threats,
    stats,
    loading,
    error,
    filters,
    updateThreatStatus,
    getThreatDetails,
    updateFilters,
    refreshData
  }
}

export default useThreats
