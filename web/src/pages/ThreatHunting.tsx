import { useState, useEffect } from 'react'
import { Search, Target, Shield, Activity, AlertTriangle, Zap, Play } from 'lucide-react'

function ThreatHunting() {
  const [hunts, setHunts] = useState<any[]>([])
  const [activeHunt, setActiveHunt] = useState<any>(null)
  const [threats, setThreats] = useState<any[]>([])
  const [indicators, setIndicators] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedHunt, setSelectedHunt] = useState<any>(null)

  useEffect(() => {
    fetchThreatHuntingData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchThreatHuntingData()
    }, 8000)

    return () => clearInterval(interval)
  }, [])

  const fetchThreatHuntingData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [huntsData, threatsData, indicatorsData] = await Promise.all([
        API_CONFIG.request('/threat-hunting/hunts'),
        API_CONFIG.request('/threat-hunting/threats'),
        API_CONFIG.request('/threat-hunting/indicators')
      ])
      
      setHunts(huntsData.hunts || [])
      setThreats(threatsData.threats || [])
      setIndicators(indicatorsData.indicators || [])
    } catch (error) {
      setError('Failed to fetch threat hunting data')
      console.error('Threat hunting fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const startHunt = async (huntId) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/threat-hunting/start', {
        method: 'POST',
        body: JSON.stringify({ hunt_id: huntId })
      })
      
      // Refresh data
      fetchThreatHuntingData()
    } catch (error) {
      console.error('Failed to start hunt:', error)
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'active':
        return 'text-green-400 bg-green-900/20 border-green-500/20'
      case 'completed':
        return 'text-blue-400 bg-blue-900/20 border-blue-500/20'
      case 'failed':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      case 'pending':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'critical':
        return 'text-red-400'
      case 'high':
        return 'text-orange-400'
      case 'medium':
        return 'text-yellow-400'
      case 'low':
        return 'text-blue-400'
      default:
        return 'text-gray-400'
    }
  }

  const filteredHunts = hunts.filter(hunt => 
    hunt.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    hunt.description.toLowerCase().includes(searchQuery.toLowerCase())
  )

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading threat hunting...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchThreatHuntingData} className="hades-button-primary">
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
          <Target className="mr-3 text-hades-primary" />
          Real-Time Threat Hunting Automation
        </h1>
        <p className="text-gray-400">Automated threat hunting with real-time detection and response</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Active Hunts</p>
              <p className="text-2xl font-bold text-white">{hunts.filter(h => h.status === 'active').length}</p>
            </div>
            <Search className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Threats Found</p>
              <p className="text-2xl font-bold text-white">{threats.length}</p>
            </div>
            <AlertTriangle className="w-8 h-8 text-red-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Indicators</p>
              <p className="text-2xl font-bold text-white">{indicators.length}</p>
            </div>
            <Shield className="w-8 h-8 text-yellow-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Success Rate</p>
              <p className="text-2xl font-bold text-white">
                {hunts.length > 0 ? Math.round((hunts.filter(h => h.status === 'completed').length / hunts.length) * 100) : 0}%
              </p>
            </div>
            <Target className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
      </div>

      {/* Search Bar */}
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
          <input
            type="text"
            placeholder="Search threat hunts..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:border-hades-primary"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Active Hunts */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Search className="mr-2 text-hades-primary" />
              Threat Hunts
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {filteredHunts.map((hunt, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{hunt.name}</h3>
                    <p className="text-gray-400 text-sm mt-1">{hunt.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(hunt.status)}`}>
                        {hunt.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {hunt.duration || 'N/A'}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    {hunt.status === 'pending' && (
                      <button 
                        onClick={() => startHunt(hunt.id)}
                        className="p-2 text-green-400 hover:text-green-300"
                        title="Start Hunt"
                      >
                        <Play className="w-4 h-4" />
                      </button>
                    )}
                    <button 
                      onClick={() => setSelectedHunt(hunt)}
                      className="p-2 text-gray-400 hover:text-white"
                    >
                      <Target className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Detected Threats */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <AlertTriangle className="mr-2 text-red-400" />
              Detected Threats
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {threats.map((threat, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{threat.type}</h3>
                    <p className="text-gray-400 text-sm">{threat.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`text-xs ${getSeverityColor(threat.severity)}`}>
                        {threat.severity}
                      </span>
                      <span className="text-xs text-gray-400">
                        {new Date(threat.detected_at).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                  <Zap className="w-5 h-5 text-red-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Threat Indicators */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-yellow-400" />
              Threat Indicators
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {indicators.map((indicator, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{indicator.type}</h3>
                    <p className="text-gray-400 text-sm">{indicator.value}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        Confidence: {indicator.confidence}%
                      </span>
                      <span className="text-xs text-gray-400">
                        Source: {indicator.source}
                      </span>
                    </div>
                  </div>
                  <Shield className="w-5 h-5 text-yellow-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Hunt Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Activity className="mr-2 text-hades-primary" />
              Hunt Details
            </h2>
          </div>
          <div className="p-4">
            {selectedHunt ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedHunt.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedHunt.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Status</p>
                    <span className={`px-2 py-1 rounded text-sm ${getStatusColor(selectedHunt.status)}`}>
                      {selectedHunt.status}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Progress</p>
                    <div className="mt-2">
                      <div className="w-full bg-gray-700 rounded-full h-2">
                        <div 
                          className="bg-hades-primary h-2 rounded-full" 
                          style={{ width: `${selectedHunt.progress || 0}%` }}
                        ></div>
                      </div>
                      <p className="text-sm text-gray-400 mt-1">{selectedHunt.progress || 0}%</p>
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Hunt Parameters</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedHunt.parameters?.map((param, index) => (
                        <li key={index}>{param}</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Results</p>
                    <p className="text-white">{selectedHunt.results || 'No results yet'}</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Target className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select a hunt to view detailed information</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default ThreatHunting
