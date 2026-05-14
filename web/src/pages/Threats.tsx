import { AlertTriangle, Shield, Activity, Clock, MapPin, AlertCircle } from 'lucide-react'
import { useThreats } from '../hooks/useThreats'

interface ThreatsProps {
  user: any
}

function Threats() {
  const { threats, stats, loading, error, updateThreatStatus, getThreatDetails, updateFilters, refreshData } = useThreats()

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading threat data...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={refreshData} className="hades-button-primary">
            Retry
          </button>
        </div>
      </div>
    )
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

  const getStatusIcon = (status) => {
    switch (status) {
      case 'blocked':
        return <Shield className="w-4 h-4 text-green-400" />
      case 'monitoring':
        return <Activity className="w-4 h-4 text-yellow-400" />
      case 'resolved':
        return <AlertCircle className="w-4 h-4 text-blue-400" />
      default:
        return <AlertTriangle className="w-4 h-4 text-red-400" />
    }
  }

  const formatTime = (timestamp) => {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Threat Intelligence</h1>
        <p className="text-gray-400">Real-time threat detection and analysis</p>
      </div>

      {/* Threat Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="hades-card p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Total Threats</p>
              <p className="text-2xl font-bold text-white mt-1">{stats?.total_threats?.toLocaleString() || 1247}</p>
            </div>
            <AlertTriangle className="w-8 h-8 text-yellow-400" />
          </div>
        </div>
        <div className="hades-card p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Blocked</p>
              <p className="text-2xl font-bold text-green-400 mt-1">{stats?.blocked_threats?.toLocaleString() || 1198}</p>
            </div>
            <Shield className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="hades-card p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Active</p>
              <p className="text-2xl font-bold text-yellow-400 mt-1">{stats?.active_threats || 3}</p>
            </div>
            <Activity className="w-8 h-8 text-yellow-400" />
          </div>
        </div>
        <div className="hades-card p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Critical</p>
              <p className="text-2xl font-bold text-red-400 mt-1">{stats?.critical_threats || 1}</p>
            </div>
            <AlertTriangle className="w-8 h-8 text-red-400" />
          </div>
        </div>
      </div>

      {/* Threat Feed */}
      <div className="hades-card p-6">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-white">Recent Threats</h2>
          <div className="flex space-x-2">
            <button onClick={refreshData} className="hades-button-secondary text-sm">Refresh</button>
            <button className="hades-button-primary text-sm">Export</button>
          </div>
        </div>

        <div className="space-y-4">
          {threats?.map((threat) => (
            <div key={threat.id} className={`p-4 rounded-lg border ${getSeverityColor(threat.severity)}`}>
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-3 mb-2">
                    <h3 className="text-white font-medium">{threat.title}</h3>
                    <span className={`px-2 py-1 rounded text-xs font-medium capitalize ${getSeverityColor(threat.severity)}`}>
                      {threat.severity}
                    </span>
                    <div className="flex items-center space-x-1">
                      {getStatusIcon(threat.status)}
                      <span className="text-gray-400 text-sm capitalize">{threat.status}</span>
                    </div>
                  </div>
                  <p className="text-gray-400 text-sm mb-2">{threat.description}</p>
                  <div className="flex items-center space-x-4 text-xs text-gray-500">
                    <div className="flex items-center space-x-1">
                      <MapPin className="w-3 h-3" />
                      <span>{threat.source}</span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <Clock className="w-3 h-3" />
                      <span>{formatTime(threat.timestamp)}</span>
                    </div>
                  </div>
                </div>
                <div className="flex space-x-2 ml-4">
                  <button className="text-hades-primary hover:text-blue-400 text-sm">Details</button>
                  <button className="text-gray-400 hover:text-white text-sm">Action</button>
                </div>
              </div>
            </div>
          )) || [
            <div key="1" className={`p-4 rounded-lg border text-red-400 bg-red-900/20 border-red-500/20`}>
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-3 mb-2">
                    <h3 className="text-white font-medium">Trojan.Dropper Detected</h3>
                    <span className="px-2 py-1 rounded text-xs font-medium capitalize text-red-400 bg-red-900/20 border-red-500/20">
                      critical
                    </span>
                    <div className="flex items-center space-x-1">
                      <Shield className="w-4 h-4 text-green-400" />
                      <span className="text-gray-400 text-sm capitalize">blocked</span>
                    </div>
                  </div>
                  <p className="text-gray-400 text-sm mb-2">Malicious payload detected and blocked at network perimeter</p>
                  <div className="flex items-center space-x-4 text-xs text-gray-500">
                    <div className="flex items-center space-x-1">
                      <MapPin className="w-3 h-3" />
                      <span>192.168.1.105</span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <Clock className="w-3 h-3" />
                      <span>2 hours ago</span>
                    </div>
                  </div>
                </div>
                <div className="flex space-x-2 ml-4">
                  <button className="text-hades-primary hover:text-blue-400 text-sm">Details</button>
                  <button className="text-gray-400 hover:text-white text-sm">Action</button>
                </div>
              </div>
            </div>
          ]}
        </div>
      </div>

      {/* Threat Map Placeholder */}
      <div className="mt-6 hades-card p-6">
        <h2 className="text-xl font-semibold text-white mb-4">Global Threat Map</h2>
        <div className="h-64 flex items-center justify-center bg-gray-700/30 rounded-lg border border-gray-600">
          <div className="text-center">
            <MapPin className="w-12 h-12 text-hades-primary mx-auto mb-2" />
            <p className="text-gray-400">Threat geographic distribution</p>
            <p className="text-gray-500 text-sm">Interactive map would be rendered here</p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Threats
