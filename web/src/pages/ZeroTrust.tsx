import { useState, useEffect } from 'react'
import { Shield, Lock, Key, User, Network, CheckCircle, AlertTriangle, Settings, Eye } from 'lucide-react'

function ZeroTrust() {
  const [policies, setPolicies] = useState<any[]>([])
  const [accessRequests, setAccessRequests] = useState<any[]>([])
  const [trustScores, setTrustScores] = useState<any[]>([])
  const [networkSegments, setNetworkSegments] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<any>(null)
  const [selectedPolicy, setSelectedPolicy] = useState<any>(null)

  useEffect(() => {
    fetchZeroTrustData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchZeroTrustData()
    }, 10000)

    return () => clearInterval(interval)
  }, [])

  const fetchZeroTrustData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [policiesData, requestsData, scoresData, segmentsData] = await Promise.all([
        API_CONFIG.request('/zerotrust/policies'),
        API_CONFIG.request('/zerotrust/access-requests'),
        API_CONFIG.request('/zerotrust/trust-scores'),
        API_CONFIG.request('/zerotrust/network-segments')
      ])
      
      setPolicies(policiesData.policies || [])
      setAccessRequests(requestsData.requests || [])
      setTrustScores(scoresData.scores || [])
      setNetworkSegments(segmentsData.segments || [])
    } catch (error) {
      setError('Failed to fetch zero-trust data')
      console.error('Zero-trust fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const updatePolicy = async (policyId: any, status) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/zerotrust/policies/update', {
        method: 'POST',
        body: JSON.stringify({ policy_id: policyId, status })
      })
      
      // Refresh data
      fetchZeroTrustData()
    } catch (error) {
      console.error('Failed to update policy:', error)
    }
  }

  const processAccessRequest = async (requestId: any, decision) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/zerotrust/access-process', {
        method: 'POST',
        body: JSON.stringify({ request_id: requestId, decision })
      })
      
      // Refresh data
      fetchZeroTrustData()
    } catch (error) {
      console.error('Failed to process access request:', error)
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'active':
        return 'text-green-400 bg-green-900/20 border-green-500/20'
      case 'pending':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      case 'denied':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      case 'inactive':
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getTrustColor = (score) => {
    if (score >= 80) return 'text-green-400'
    if (score >= 60) return 'text-yellow-400'
    if (score >= 40) return 'text-orange-400'
    return 'text-red-400'
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading zero-trust data...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchZeroTrustData} className="hades-button-primary">
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-white mb-2 flex items-center">
          <Shield className="mr-3 text-hades-primary" />
          Zero-Trust Network Architecture
        </h1>
        <p className="text-gray-400">Policy-based access control and trust management system</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Active Policies</p>
              <p className="text-2xl font-bold text-white">{policies.filter(p => p.status === 'active').length}</p>
            </div>
            <Lock className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Access Requests</p>
              <p className="text-2xl font-bold text-white">{accessRequests.filter(r => r.status === 'pending').length}</p>
            </div>
            <Key className="w-8 h-8 text-yellow-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Avg Trust Score</p>
              <p className={`text-2xl font-bold ${getTrustColor(trustScores.length > 0 ? Math.round(trustScores.reduce((acc, s) => acc + s.score, 0) / trustScores.length) : 0)}`}>
                {trustScores.length > 0 ? Math.round(trustScores.reduce((acc, s) => acc + s.score, 0) / trustScores.length) : 0}%
              </p>
            </div>
            <User className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Network Segments</p>
              <p className="text-2xl font-bold text-white">{networkSegments.length}</p>
            </div>
            <Network className="w-8 h-8 text-blue-400" />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Zero-Trust Policies */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Lock className="mr-2 text-green-400" />
              Access Policies
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {policies.map((policy, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{policy.name}</h3>
                    <p className="text-gray-400 text-sm mt-1">{policy.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(policy.status)}`}>
                        {policy.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {policy.resources?.length || 0} resources
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button 
                      onClick={() => updatePolicy(policy.id, policy.status === 'active' ? 'inactive' : 'active')}
                      className="p-2 text-gray-400 hover:text-white"
                      title="Toggle Policy"
                    >
                      <Settings className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={() => setSelectedPolicy(policy)}
                      className="p-2 text-gray-400 hover:text-white"
                    >
                      <Eye className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Access Requests */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Key className="mr-2 text-yellow-400" />
              Access Requests
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {accessRequests.map((request, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{request.user}</h3>
                    <p className="text-gray-400 text-sm">{request.resource}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(request.status)}`}>
                        {request.status}
                      </span>
                      <span className={`text-xs ${getTrustColor(request.trust_score)}`}>
                        Trust: {request.trust_score}%
                      </span>
                    </div>
                  </div>
                  {request.status === 'pending' && (
                    <div className="flex items-center space-x-2">
                      <button 
                        onClick={() => processAccessRequest(request.id, 'approve')}
                        className="p-2 text-green-400 hover:text-green-300"
                        title="Approve"
                      >
                        <CheckCircle className="w-4 h-4" />
                      </button>
                      <button 
                        onClick={() => processAccessRequest(request.id, 'deny')}
                        className="p-2 text-red-400 hover:text-red-300"
                        title="Deny"
                      >
                        <AlertTriangle className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Trust Scores */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <User className="mr-2 text-hades-primary" />
              Trust Scores
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {trustScores.map((score, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{score.user}</h3>
                    <p className="text-gray-400 text-sm">{score.role}</p>
                    <div className="mt-2">
                      <div className="w-full bg-gray-700 rounded-full h-2">
                        <div 
                          className={`h-2 rounded-full ${score.score >= 80 ? 'bg-green-400' : score.score >= 60 ? 'bg-yellow-400' : score.score >= 40 ? 'bg-orange-400' : 'bg-red-400'}`}
                          style={{ width: `${score.score}%` }}
                        ></div>
                      </div>
                      <p className={`text-sm mt-1 ${getTrustColor(score.score)}`}>
                        {score.score}% trust score
                      </p>
                    </div>
                  </div>
                  <Shield className={`w-5 h-5 ${getTrustColor(score.score)}`} />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Network Segments */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Network className="mr-2 text-blue-400" />
              Network Segments
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {networkSegments.map((segment, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{segment.name}</h3>
                    <p className="text-gray-400 text-sm">{segment.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(segment.status)}`}>
                        {segment.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {segment.devices?.length || 0} devices
                      </span>
                      <span className="text-xs text-gray-400">
                        {segment.policies?.length || 0} policies
                      </span>
                    </div>
                  </div>
                  <Network className="w-5 h-5 text-blue-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Policy Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Settings className="mr-2 text-hades-primary" />
              Policy Details
            </h2>
          </div>
          <div className="p-4">
            {selectedPolicy ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedPolicy.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedPolicy.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Status</p>
                    <span className={`px-2 py-1 rounded text-sm ${getStatusColor(selectedPolicy.status)}`}>
                      {selectedPolicy.status}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Resources</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedPolicy.resources?.map((resource, index) => (
                        <li key={index}>{resource}</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Conditions</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedPolicy.conditions?.map((condition, index) => (
                        <li key={index}>{condition}</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Actions</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedPolicy.actions?.map((action, index) => (
                        <li key={index}>{action}</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Created</p>
                    <p className="text-white">{new Date(selectedPolicy.created_at).toLocaleString()}</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Settings className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select a policy to view detailed information</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default ZeroTrust
