import { useState, useEffect } from 'react'
import { threatsAPI } from '../api/threats'

export const useThreats = () => {
  const [threats, setThreats] = useState([])
  const [stats, setStats] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
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
      setThreats(threatsData)
    } catch (error) {
      setError('Failed to fetch threats')
      console.error('Threats fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchThreatStats = async () => {
    try {
      const statsData = await threatsAPI.getThreatStats()
      setStats(statsData)
    } catch (error) {
      console.error('Threat stats fetch error:', error)
    }
  }

  const updateThreatStatus = async (id, status) => {
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

  const getThreatDetails = async (id) => {
    try {
      return await threatsAPI.getThreat(id)
    } catch (error) {
      setError('Failed to fetch threat details')
      throw error
    }
  }

  const updateFilters = (newFilters) => {
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
