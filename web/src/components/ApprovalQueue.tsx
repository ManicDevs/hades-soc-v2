import { useState, useEffect } from 'react'
import { useWebSocket } from '../hooks/useWebSocket'
import { API_CONFIG } from '../api/config'

const ApprovalQueue = () => {
  const [pendingActions, setPendingActions] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [processing, setProcessing] = useState(new Set<string>())

  // WebSocket connection for real-time updates
  const { lastMessage } = useWebSocket()

  // Fetch pending actions on component mount
  useEffect(() => {
    fetchPendingActions()
    
    // Set up polling for pending actions updates
    const interval = setInterval(() => {
      fetchPendingActions()
    }, 10000) // Update every 10 seconds

    return () => clearInterval(interval)
  }, [])

  // Handle WebSocket messages
  useEffect(() => {
    if (lastMessage) {
      try {
        const message = JSON.parse(lastMessage.data)
        
        if (message.type === 'GOVERNOR_INTERCEPT') {
          // Add new pending action to the queue
          setPendingActions(prev => [message.data, ...prev])
        } else if (message.type === 'GOVERNOR_APPROVAL_UPDATE') {
          // Remove approved/denied action from the queue
          setPendingActions(prev => prev.filter(action => action.action_id !== message.data.action_id))
          // Remove from processing set
          setProcessing(prev => {
            const newSet = new Set(prev)
            newSet.delete(message.data.action_id)
            return newSet
          })
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }
  }, [lastMessage])

  const fetchPendingActions = async () => {
    try {
      const data = await API_CONFIG.request('/governor/pending')
      setPendingActions(data || [])
      setError(null)
    } catch (err) {
      console.error('Failed to fetch pending actions:', err)
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const handleApprove = async (actionId: string) => {
    if (processing.has(actionId)) return
    
    setProcessing(prev => new Set(prev).add(actionId))
    
    try {
      await API_CONFIG.request(`/governor/approve/${actionId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: 'approved',
          user_id: 'current_user', // TODO: Get from auth context
          reason: 'Approved by security analyst'
        })
      })

      // Remove from pending list immediately (WebSocket will also update)
      setPendingActions(prev => prev.filter(action => action.action_id !== actionId))
      
    } catch (err) {
      console.error('Failed to approve action:', err)
      setError(err instanceof Error ? err.message : 'Unknown error')
      // Remove from processing set on error
      setProcessing(prev => {
        const newSet = new Set(prev)
        newSet.delete(actionId)
        return newSet
      })
    }
  }

  const handleDeny = async (actionId: string) => {
    if (processing.has(actionId)) return
    
    setProcessing(prev => new Set(prev).add(actionId))
    
    try {
      await API_CONFIG.request(`/governor/approve/${actionId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: 'denied',
          user_id: 'current_user', // TODO: Get from auth context
          reason: 'Denied by security analyst - potential risk too high'
        })
      })

      // Remove from pending list immediately (WebSocket will also update)
      setPendingActions(prev => prev.filter(action => action.action_id !== actionId))
      
    } catch (err) {
      console.error('Failed to deny action:', err)
      setError(err instanceof Error ? err.message : 'Unknown error')
      // Remove from processing set on error
      setProcessing(prev => {
        const newSet = new Set(prev)
        newSet.delete(actionId)
        return newSet
      })
    }
  }

  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString()
  }

  const getSeverityColor = (actionName: string) => {
    const highRiskActions = ['block_ip', 'isolate_node', 'reset_credentials', 'firewall_drop']
    if (highRiskActions.some(risk => actionName.toLowerCase().includes(risk))) {
      return 'border-red-500 bg-red-500/10'
    }
    return 'border-yellow-500 bg-yellow-500/10'
  }

  const getSeverityBadgeColor = (actionName: string) => {
    const highRiskActions = ['block_ip', 'isolate_node', 'reset_credentials', 'firewall_drop']
    if (highRiskActions.some(risk => actionName.toLowerCase().includes(risk))) {
      return 'bg-red-500 text-red-900'
    }
    return 'bg-yellow-500 text-yellow-900'
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
          <h3 className="text-red-400 font-semibold mb-2">Error Loading Approval Queue</h3>
          <p className="text-red-300">{error}</p>
          <button 
            onClick={fetchPendingActions}
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
        <h2 className="text-2xl font-bold text-white">Approval Queue</h2>
        <div className="flex items-center space-x-2">
          <div className={`w-3 h-3 rounded-full ${pendingActions.length > 0 ? 'bg-yellow-500 animate-pulse' : 'bg-gray-500'}`}></div>
          <span className="text-gray-400 text-sm">
            {pendingActions.length > 0 ? `${pendingActions.length} pending` : 'No pending actions'}
          </span>
        </div>
      </div>

      {pendingActions.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-gray-500 text-lg mb-2">No Actions Require Approval</div>
          <div className="text-gray-400 text-sm">All automated actions have been processed</div>
        </div>
      ) : (
        <div className="space-y-4">
          {pendingActions.map((action) => (
            <div 
              key={action.action_id} 
              className={`border rounded-lg p-6 transition-all duration-200 ${getSeverityColor(action.action_name)}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  {/* Action Header */}
                  <div className="flex items-center space-x-3 mb-3">
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${getSeverityBadgeColor(action.action_name)}`}>
                      {action.requires_approval ? 'MANUAL ACK REQUIRED' : 'PENDING'}
                    </span>
                    <span className="text-gray-400 text-xs">
                      {formatTime(action.timestamp)}
                    </span>
                  </div>

                  {/* Action Details */}
                  <div className="space-y-2">
                    <div>
                      <span className="text-gray-400 text-sm">Action:</span>
                      <span className="text-white font-medium ml-2">{action.action_name}</span>
                    </div>
                    <div>
                      <span className="text-gray-400 text-sm">Target:</span>
                      <span className="text-white ml-2">{action.target}</span>
                    </div>
                    {action.requester && (
                      <div>
                        <span className="text-gray-400 text-sm">Requester:</span>
                        <span className="text-white ml-2">{action.requester}</span>
                      </div>
                    )}
                    {action.reasoning && (
                      <div>
                        <span className="text-gray-400 text-sm">Reasoning:</span>
                        <p className="text-gray-300 mt-1 text-sm">{action.reasoning}</p>
                      </div>
                    )}
                    {action.block_reason && (
                      <div className="mt-3 p-3 bg-yellow-500/20 border border-yellow-500 rounded">
                        <span className="text-yellow-400 text-sm font-medium">⚠️ {action.block_reason}</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex flex-col space-y-2 ml-6">
                  <button
                    onClick={() => handleApprove(action.action_id)}
                    disabled={processing.has(action.action_id)}
                    className={`px-4 py-2 rounded-lg font-medium transition-all duration-200 ${
                      processing.has(action.action_id)
                        ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                        : 'bg-green-500 hover:bg-green-600 text-white'
                    }`}
                  >
                    {processing.has(action.action_id) ? (
                      <div className="flex items-center space-x-2">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                        <span>Processing...</span>
                      </div>
                    ) : (
                      <div className="flex items-center space-x-2">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                        <span>Authorize</span>
                      </div>
                    )}
                  </button>
                  
                  <button
                    onClick={() => handleDeny(action.action_id)}
                    disabled={processing.has(action.action_id)}
                    className={`px-4 py-2 rounded-lg font-medium transition-all duration-200 ${
                      processing.has(action.action_id)
                        ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                        : 'bg-red-500 hover:bg-red-600 text-white'
                    }`}
                  >
                    {processing.has(action.action_id) ? (
                      <div className="flex items-center space-x-2">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                        <span>Processing...</span>
                      </div>
                    ) : (
                      <div className="flex items-center space-x-2">
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                        <span>Deny</span>
                      </div>
                    )}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Queue Summary */}
      {pendingActions.length > 0 && (
        <div className="mt-8 p-4 bg-gray-800 rounded-lg border border-gray-700">
          <div className="flex items-center justify-between">
            <div className="text-gray-400 text-sm">
              Queue Summary: {pendingActions.length} action{pendingActions.length !== 1 ? 's' : ''} awaiting review
            </div>
            <button 
              onClick={fetchPendingActions}
              className="px-3 py-1 bg-blue-500 hover:bg-blue-600 text-white text-sm rounded transition-colors"
            >
              Refresh Queue
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default ApprovalQueue
