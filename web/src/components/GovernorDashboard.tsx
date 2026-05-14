import { useState, useEffect } from 'react'
import { useWebSocket } from '../hooks/useWebSocket'

const GovernorDashboard = () => {
  const [governorStats, setGovernorStats] = useState<any>(null)
  const [governorStatus, setGovernorStatus] = useState<any>(null)
  const [interceptMessages, setInterceptMessages] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // WebSocket connection for real-time updates
  const { lastMessage } = useWebSocket()

  // Fetch governor stats on component mount
  useEffect(() => {
    fetchGovernorStats()
    fetchGovernorStatus()
    
    // Set up polling for stats updates
    const interval = setInterval(() => {
      fetchGovernorStats()
      fetchGovernorStatus()
    }, 30000) // Update every 30 seconds

    return () => clearInterval(interval)
  }, [])

  // Handle WebSocket messages
  useEffect(() => {
    if (lastMessage) {
      try {
        const message = JSON.parse(lastMessage.data)
        
        if (message.type === 'GOVERNOR_INTERCEPT') {
          // Add new intercept message
          setInterceptMessages(prev => [message.data, ...prev.slice(0, 9)]) // Keep last 10
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }
  }, [lastMessage])

  const fetchGovernorStats = async () => {
    try {
      const response = await fetch('/api/v2/governor/stats')
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      setGovernorStats(data)
      setError(null)
    } catch (err) {
      console.error('Failed to fetch governor stats:', err)
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const fetchGovernorStatus = async () => {
    try {
      const response = await fetch('/api/v2/governor/status')
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      setGovernorStatus(data)
    } catch (err) {
      console.error('Failed to fetch governor status:', err)
    }
  }

  const formatTimeRemaining = (seconds: number) => {
    if (!seconds || seconds <= 0) return '0s'
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`
    } else {
      return `${secs}s`
    }
  }

  const getBlockStatusColor = (remaining: number, max: number) => {
    const percentage = (remaining / max) * 100
    if (percentage > 60) return 'text-green-400'
    if (percentage > 30) return 'text-yellow-400'
    return 'text-red-400'
  }

  const getProgressBarColor = (remaining: number, max: number) => {
    const percentage = (remaining / max) * 100
    if (percentage > 60) return 'bg-green-500'
    if (percentage > 30) return 'bg-yellow-500'
    return 'bg-red-500'
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-500/20 border border-red-500 rounded-lg p-4">
          <h3 className="text-red-400 font-semibold mb-2">Error Loading Governor Data</h3>
          <p className="text-red-300">{error}</p>
          <button 
            onClick={fetchGovernorStats}
            className="mt-4 px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-white">Safety Governor</h2>
        <div className="flex items-center space-x-2">
          <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
          <span className="text-green-400 text-sm">Active</span>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Blocks Remaining */}
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <h3 className="text-gray-400 text-sm font-medium mb-2">Blocks Remaining</h3>
          <div className="flex items-baseline space-x-2">
            <span className={`text-3xl font-bold ${getBlockStatusColor(governorStatus?.remaining_blocks || 0, governorStatus?.max_blocks_per_hour || 5)}`}>
              {governorStatus?.remaining_blocks || 0}
            </span>
            <span className="text-gray-400 text-sm">/ {governorStatus?.max_blocks_per_hour || 5}</span>
          </div>
          <div className="mt-3">
            <div className="w-full bg-gray-700 rounded-full h-2">
              <div 
                className={`h-2 rounded-full transition-all duration-300 ${getProgressBarColor(governorStatus?.remaining_blocks || 0, governorStatus?.max_blocks_per_hour || 5)}`}
                style={{ width: `${((governorStatus?.remaining_blocks || 0) / (governorStatus?.max_blocks_per_hour || 5)) * 100}%` }}
              ></div>
            </div>
          </div>
        </div>

        {/* Time Until Reset */}
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <h3 className="text-gray-400 text-sm font-medium mb-2">Time Until Reset</h3>
          <div className="flex items-baseline space-x-2">
            <span className="text-3xl font-bold text-blue-400">
              {formatTimeRemaining(governorStats?.time_until_reset_seconds || 0)}
            </span>
          </div>
          <p className="text-gray-500 text-xs mt-2">Next reset window</p>
        </div>

        {/* Current Block Count */}
        <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
          <h3 className="text-gray-400 text-sm font-medium mb-2">Current Block Count</h3>
          <div className="flex items-baseline space-x-2">
            <span className="text-3xl font-bold text-purple-400">
              {governorStatus?.current_block_count || 0}
            </span>
          </div>
          <p className="text-gray-500 text-xs mt-2">Blocks used this hour</p>
        </div>
      </div>

      {/* Manual ACK Intercepts */}
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <h3 className="text-white text-lg font-semibold mb-4">Manual ACK Required</h3>
        {interceptMessages.length === 0 ? (
          <div className="text-center py-8">
            <div className="text-gray-500 text-sm">No manual approvals pending</div>
          </div>
        ) : (
          <div className="space-y-3">
            {interceptMessages.map((message, index) => (
              <div key={index} className="bg-yellow-500/20 border border-yellow-500 rounded-lg p-4">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-2">
                      <span className="bg-yellow-500 text-yellow-900 text-xs px-2 py-1 rounded font-medium">
                        MANUAL ACK REQUIRED
                      </span>
                      <span className="text-gray-400 text-xs">
                        {new Date(message.timestamp * 1000).toLocaleTimeString()}
                      </span>
                    </div>
                    <h4 className="text-white font-medium mb-1">{message.action_name}</h4>
                    <p className="text-gray-300 text-sm mb-2">Target: {message.target}</p>
                    <p className="text-gray-400 text-sm">{message.reasoning}</p>
                    <p className="text-yellow-400 text-sm mt-2">{message.block_reason}</p>
                  </div>
                  <div className="flex space-x-2 ml-4">
                    <button className="px-3 py-1 bg-green-500 hover:bg-green-600 text-white text-sm rounded transition-colors">
                      Approve
                    </button>
                    <button className="px-3 py-1 bg-red-500 hover:bg-red-600 text-white text-sm rounded transition-colors">
                      Deny
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Additional Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <h4 className="text-gray-400 text-xs font-medium mb-1">Approved (24h)</h4>
          <p className="text-white text-xl font-semibold">{governorStats?.approved_last_24h || 0}</p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <h4 className="text-gray-400 text-xs font-medium mb-1">Blocked (24h)</h4>
          <p className="text-white text-xl font-semibold">{governorStats?.blocked_last_24h || 0}</p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <h4 className="text-gray-400 text-xs font-medium mb-1">Manual ACK (24h)</h4>
          <p className="text-white text-xl font-semibold">{governorStats?.manual_ack_last_24h || 0}</p>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <h4 className="text-gray-400 text-xs font-medium mb-1">Total Actions (24h)</h4>
          <p className="text-white text-xl font-semibold">{governorStats?.total_last_24h || 0}</p>
        </div>
      </div>
    </div>
  )
}

export default GovernorDashboard
