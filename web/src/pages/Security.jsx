import React from 'react'
import { Shield, Lock, Eye, EyeOff, CheckCircle, AlertTriangle, XCircle } from 'lucide-react'
import { useSecurity } from '../hooks/useSecurity'

function Security({ user }) {
  console.log('Security component rendering - user:', user)
  try {
    console.log('About to call useSecurity hook...')
    const securityData = useSecurity()
    console.log('useSecurity hook returned:', securityData)
    const { policies, vulnerabilities, securityScore, auditLogs, loading, error, updatePolicy, updateVulnerability, runSecurityScan, refreshData } = securityData
    console.log('Security hook destructured data:', { policies, vulnerabilities, securityScore, loading, error })

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading security data...</p>
        </div>
      </div>
    )
  }

  if (error) {
    console.log('Security component error state:', error)
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          {refreshData && (
            <button onClick={refreshData} className="hades-button-primary">
              Retry
            </button>
          )}
        </div>
      </div>
    )
  }

  const getStatusIcon = (status) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="w-4 h-4 text-green-400" />
      case 'warning':
        return <AlertTriangle className="w-4 h-4 text-yellow-400" />
      default:
        return <XCircle className="w-4 h-4 text-red-400" />
    }
  }

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'critical':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      case 'high':
        return 'text-orange-400 bg-orange-900/20 border-orange-500/20'
      case 'medium':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      case 'low':
        return 'text-blue-400 bg-blue-900/20 border-blue-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Security Management</h1>
        <p className="text-gray-400">Configure and monitor security policies</p>
      </div>

      {/* Security Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <div className="hades-card p-6 text-center">
          <Shield className="w-12 h-12 text-green-400 mx-auto mb-3" />
          <h3 className="text-2xl font-bold text-white mb-1">{securityScore?.overall_score || 98}%</h3>
          <p className="text-gray-400 text-sm">Security Score</p>
        </div>
        <div className="hades-card p-6 text-center">
          <Lock className="w-12 h-12 text-yellow-400 mx-auto mb-3" />
          <h3 className="text-2xl font-bold text-white mb-1">{policies?.length || 4}</h3>
          <p className="text-gray-400 text-sm">Active Policies</p>
        </div>
        <div className="hades-card p-6 text-center">
          <AlertTriangle className="w-12 h-12 text-red-400 mx-auto mb-3" />
          <h3 className="text-2xl font-bold text-white mb-1">{vulnerabilities?.filter(v => v.status === 'open').length || 2}</h3>
          <p className="text-gray-400 text-sm">Open Vulnerabilities</p>
        </div>
      </div>

      {/* Security Policies */}
      <div className="hades-card p-6 mb-6">
        <h2 className="text-xl font-semibold text-white mb-4">Security Policies</h2>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="text-left py-3 px-4 text-gray-400 font-medium">Policy Name</th>
                <th className="text-left py-3 px-4 text-gray-400 font-medium">Status</th>
                <th className="text-left py-3 px-4 text-gray-400 font-medium">Last Updated</th>
                <th className="text-left py-3 px-4 text-gray-400 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {policies?.map((policy) => (
                <tr key={policy.id} className="border-b border-gray-800">
                  <td className="py-3 px-4 text-white">{policy.name}</td>
                  <td className="py-3 px-4">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(policy.status)}
                      <span className="text-gray-300 capitalize">{policy.status}</span>
                    </div>
                  </td>
                  <td className="py-3 px-4 text-gray-400">{new Date(policy.lastUpdated || policy.created_at).toLocaleString()}</td>
                  <td className="py-3 px-4">
                    <button className="text-hades-primary hover:text-blue-400 text-sm">Edit</button>
                  </td>
                </tr>
              )) || [
                <tr key="1" className="border-b border-gray-800">
                  <td className="py-3 px-4 text-white">Password Policy</td>
                  <td className="py-3 px-4">
                    <div className="flex items-center space-x-2">
                      <CheckCircle className="w-4 h-4 text-green-400" />
                      <span className="text-gray-300 capitalize">active</span>
                    </div>
                  </td>
                  <td className="py-3 px-4 text-gray-400">2 hours ago</td>
                  <td className="py-3 px-4">
                    <button className="text-hades-primary hover:text-blue-400 text-sm">Edit</button>
                  </td>
                </tr>
              ]}
            </tbody>
          </table>
        </div>
      </div>

      {/* Vulnerabilities */}
      <div className="hades-card p-6">
        <h2 className="text-xl font-semibold text-white mb-4">Vulnerability Assessment</h2>
        <div className="space-y-3">
          {vulnerabilities?.map((vuln) => (
            <div key={vuln.id} className={`p-4 rounded-lg border ${getSeverityColor(vuln.severity)}`}>
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-white font-medium">{vuln.name || vuln.title}</h3>
                  <p className="text-gray-400 text-sm mt-1">Affected: {vuln.affected_systems?.join(', ') || vuln.affected}</p>
                </div>
                <div className="flex items-center space-x-3">
                  <span className={`px-2 py-1 rounded text-xs font-medium capitalize ${getSeverityColor(vuln.severity)}`}>
                    {vuln.severity}
                  </span>
                  <span className="text-gray-400 text-sm capitalize">{vuln.status}</span>
                </div>
              </div>
            </div>
          )) || [
            <div key="1" className={`p-4 rounded-lg border text-red-400 bg-red-900/20 border-red-500/20`}>
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-white font-medium">Outdated SSL Certificate</h3>
                  <p className="text-gray-400 text-sm mt-1">Affected: web-server</p>
                </div>
                <div className="flex items-center space-x-3">
                  <span className="px-2 py-1 rounded text-xs font-medium capitalize text-red-400 bg-red-900/20 border-red-500/20">
                    critical
                  </span>
                  <span className="text-gray-400 text-sm capitalize">open</span>
                </div>
              </div>
            </div>
          ]}
        </div>
      </div>
    </div>
  )
  } catch (error) {
    console.error('Security component error:', error)
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">Security component error: {error.message}</p>
          <p className="text-gray-400 text-sm">Please check the console for more details.</p>
        </div>
      </div>
    )
  }
}

export default Security
