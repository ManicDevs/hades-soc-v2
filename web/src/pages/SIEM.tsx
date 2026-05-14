import { useState, useEffect } from 'react'
import { Shield, Activity, AlertTriangle, Database, TrendingUp, Filter, Eye, Settings } from 'lucide-react'

function SIEM() {
  const [events, setEvents] = useState<any[]>([])
  const [alerts, setAlerts] = useState<any[]>([])
  const [correlations, setCorrelations] = useState<any[]>([])
  const [threatFeeds, setThreatFeeds] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedEvent, setSelectedEvent] = useState<any>(null)
  const [filters, setFilters] = useState({
    severity: 'all',
    source: 'all',
    timeframe: '1h'
  })

  useEffect(() => {
    fetchSIEMData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchSIEMData()
    }, 5000)

    return () => clearInterval(interval)
  }, [])

  const fetchSIEMData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [eventsData, alertsData, correlationsData, feedsData] = await Promise.all([
        API_CONFIG.request('/siem/events'),
        API_CONFIG.request('/siem/alerts'),
        API_CONFIG.request('/siem/correlations'),
        API_CONFIG.request('/siem/threat-feeds')
      ])
      
      setEvents(eventsData.events || [])
      setAlerts(alertsData.alerts || [])
      setCorrelations(correlationsData.correlations || [])
      setThreatFeeds(feedsData.feeds || [])
    } catch (error) {
      setError('Failed to fetch SIEM data')
    } finally {
      setLoading(false)
    }
  }

  const acknowledgeAlert = async (alertId) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/siem/alerts/acknowledge', {
        method: 'POST',
        body: JSON.stringify({ alert_id: alertId })
      })
      
      fetchSIEMData()
    } catch (error) {
      console.error('Failed to acknowledge alert:', error)
      setError('Failed to acknowledge alert')
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
      case 'info':
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'new':
        return 'text-red-400'
      case 'investigating':
        return 'text-yellow-400'
      case 'acknowledged':
        return 'text-blue-400'
      case 'resolved':
        return 'text-green-400'
      default:
        return 'text-gray-400'
    }
  }

  const filteredEvents = events.filter(event => {
    if (filters.severity !== 'all' && event.severity !== filters.severity) return false
    if (filters.source !== 'all' && event.source !== filters.source) return false
    return true
  })

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading SIEM data...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchSIEMData} className="hades-button-primary">
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
          Advanced SIEM Integration
        </h1>
        <p className="text-gray-400">Real-time event correlation and threat intelligence integration</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Events Processed</p>
              <p className="text-2xl font-bold text-white">{events.length}</p>
            </div>
            <Activity className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Active Alerts</p>
              <p className="text-2xl font-bold text-white">{alerts.filter(a => a.status === 'new').length}</p>
            </div>
            <AlertTriangle className="w-8 h-8 text-red-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Correlations</p>
              <p className="text-2xl font-bold text-white">{correlations.length}</p>
            </div>
            <Database className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Threat Feeds</p>
              <p className="text-2xl font-bold text-white">{threatFeeds.length}</p>
            </div>
            <TrendingUp className="w-8 h-8 text-blue-400" />
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-gray-800 rounded-lg border border-gray-700 p-4 mb-6">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold text-white flex items-center">
            <Filter className="mr-2 text-hades-primary" />
            Event Filters
          </h2>
          <div className="flex items-center space-x-4">
            <select
              value={filters.severity}
              onChange={(e) => setFilters({...filters, severity: e.target.value})}
              className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
            >
              <option value="all">All Severities</option>
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
              <option value="info">Info</option>
            </select>
            <select
              value={filters.source}
              onChange={(e) => setFilters({...filters, source: e.target.value})}
              className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
            >
              <option value="all">All Sources</option>
              <option value="firewall">Firewall</option>
              <option value="ids">IDS/IPS</option>
              <option value="endpoint">Endpoint</option>
              <option value="network">Network</option>
            </select>
            <select
              value={filters.timeframe}
              onChange={(e) => setFilters({...filters, timeframe: e.target.value})}
              className="px-3 py-2 bg-gray-700 border border-gray-600 rounded text-white"
            >
              <option value="1h">Last Hour</option>
              <option value="6h">Last 6 Hours</option>
              <option value="24h">Last 24 Hours</option>
              <option value="7d">Last 7 Days</option>
            </select>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Security Events */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Activity className="mr-2 text-hades-primary" />
              Security Events
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {filteredEvents.slice(0, 20).map((event, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{event.event_type}</h3>
                    <p className="text-gray-400 text-sm mt-1">{event.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getSeverityColor(event.severity)}`}>
                        {event.severity}
                      </span>
                      <span className="text-xs text-gray-400">
                        {event.source}
                      </span>
                      <span className="text-xs text-gray-400">
                        {new Date(event.timestamp).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                  <button 
                    onClick={() => setSelectedEvent(event)}
                    className="ml-3 p-2 text-gray-400 hover:text-white"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* SIEM Alerts */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <AlertTriangle className="mr-2 text-red-400" />
              SIEM Alerts
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {alerts.map((alert, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{alert.title}</h3>
                    <p className="text-gray-400 text-sm mt-1">{alert.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getSeverityColor(alert.severity)}`}>
                        {alert.severity}
                      </span>
                      <span className={`text-xs ${getStatusColor(alert.status)}`}>
                        {alert.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {alert.event_count} events
                      </span>
                    </div>
                  </div>
                  {alert.status === 'new' && (
                    <button 
                      onClick={() => acknowledgeAlert(alert.id)}
                      className="ml-3 p-2 text-blue-400 hover:text-blue-300"
                      title="Acknowledge"
                    >
                      <Settings className="w-4 h-4" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Event Correlations */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Database className="mr-2 text-green-400" />
              Event Correlations
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {correlations.map((correlation, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{correlation.rule_name}</h3>
                    <p className="text-gray-400 text-sm">{correlation.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getSeverityColor(correlation.severity)}`}>
                        {correlation.severity}
                      </span>
                      <span className="text-xs text-gray-400">
                        {correlation.event_count} events
                      </span>
                      <span className="text-xs text-gray-400">
                        Confidence: {correlation.confidence}%
                      </span>
                    </div>
                  </div>
                  <Database className="w-5 h-5 text-green-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Threat Intelligence Feeds */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <TrendingUp className="mr-2 text-blue-400" />
              Threat Intelligence Feeds
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {threatFeeds.map((feed, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{feed.name}</h3>
                    <p className="text-gray-400 text-sm">{feed.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${feed.status === 'active' ? 'text-green-400 bg-green-900/20 border-green-500/20' : 'text-gray-400 bg-gray-900/20 border-gray-500/20'}`}>
                        {feed.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {feed.indicators?.length || 0} indicators
                      </span>
                      <span className="text-xs text-gray-400">
                        Updated: {new Date(feed.last_updated).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <TrendingUp className="w-5 h-5 text-blue-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Event Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Eye className="mr-2 text-hades-primary" />
              Event Details
            </h2>
          </div>
          <div className="p-4">
            {selectedEvent ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedEvent.event_type}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedEvent.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Severity</p>
                    <span className={`px-2 py-1 rounded text-sm ${getSeverityColor(selectedEvent.severity)}`}>
                      {selectedEvent.severity}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Source</p>
                    <p className="text-white">{selectedEvent.source}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Timestamp</p>
                    <p className="text-white">{new Date(selectedEvent.timestamp).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Source IP</p>
                    <p className="text-white">{selectedEvent.source_ip || 'N/A'}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Destination IP</p>
                    <p className="text-white">{selectedEvent.destination_ip || 'N/A'}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Raw Data</p>
                    <div className="mt-2 p-3 bg-gray-900 rounded border border-gray-700">
                      <pre className="text-xs text-gray-300 font-mono">
                        {JSON.stringify(selectedEvent.raw_data, null, 2)}
                      </pre>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Eye className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select an event to view detailed information</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default SIEM
