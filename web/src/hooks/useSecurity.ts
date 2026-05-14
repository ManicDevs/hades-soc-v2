import { useState, useEffect } from 'react'
import { securityAPI } from '../api/security'

export const useSecurity = () => {
  const [policies, setPolicies] = useState<any[]>([])
  const [vulnerabilities, setVulnerabilities] = useState<any[]>([])
  const [securityScore, setSecurityScore] = useState<any>(null)
  const [auditLogs, setAuditLogs] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchSecurityData()
    
    // Set up periodic updates for security score
    const interval = setInterval(() => {
      fetchSecurityScore()
    }, 30000)

    return () => clearInterval(interval)
  }, [])

  const fetchSecurityData = async () => {
    console.log('fetchSecurityData starting')
    setLoading(true)
    setError(null)
    
    try {
      console.log('Calling security APIs...')
      const [policiesData, vulnerabilitiesData, scoreData] = await Promise.all([
        securityAPI.getPolicies(),
        securityAPI.getVulnerabilities(),
        securityAPI.getSecurityScore()
      ])
      
      console.log('Security API responses:', { policiesData, vulnerabilitiesData, scoreData })
      
      setPolicies(policiesData)
      setVulnerabilities(vulnerabilitiesData)
      setSecurityScore(scoreData)
    } catch (error) {
      console.error('Security data fetch error:', error)
      setError('Failed to fetch security data')
    } finally {
      setLoading(false)
    }
  }

  const fetchSecurityScore = async () => {
    try {
      const scoreData = await securityAPI.getSecurityScore()
      setSecurityScore(scoreData)
    } catch (error) {
      console.error('Security score fetch error:', error)
    }
  }

  const fetchAuditLogs = async (filters = {}) => {
    try {
      const logsData = await securityAPI.getAuditLogs(filters)
      setAuditLogs(logsData)
    } catch (error) {
      setError('Failed to fetch audit logs')
      console.error('Audit logs fetch error:', error)
    }
  }

  const updatePolicy = async (id: any, policyData: any) => {
    try {
      await securityAPI.updatePolicy(id, policyData)
      await fetchSecurityData()
    } catch (error) {
      setError('Failed to update security policy')
      throw error
    }
  }

  const updateVulnerability = async (id: any, status: any) => {
    try {
      await securityAPI.updateVulnerability(id, status)
      await fetchSecurityData()
    } catch (error) {
      setError('Failed to update vulnerability')
      throw error
    }
  }

  const runSecurityScan = async () => {
    try {
      await securityAPI.runSecurityScan()
      // Refresh data after scan
      setTimeout(() => {
        fetchSecurityData()
      }, 2000)
    } catch (error) {
      setError('Failed to run security scan')
      throw error
    }
  }

  const refreshData = () => {
    fetchSecurityData()
  }

  return {
    policies,
    vulnerabilities,
    securityScore,
    auditLogs,
    loading,
    error,
    updatePolicy,
    updateVulnerability,
    runSecurityScan,
    fetchAuditLogs,
    refreshData
  }
}

export default useSecurity
