import { useState, useEffect } from 'react'
import { AlertTriangle, PlayCircle, PauseCircle, Clock, FileText, Settings, Eye } from 'lucide-react'

function IncidentResponse() {
  const [incidents, setIncidents] = useState<any[]>([])
  const [playbooks, setPlaybooks] = useState<any[]>([])
  const [activeResponses, setActiveResponses] = useState<any[]>([])
  const [responseActions, setResponseActions] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedIncident, setSelectedIncident] = useState<any>(null)

  useEffect(() => {
    fetchIncidentResponseData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchIncidentResponseData()
    }, 8000)

    return () => clearInterval(interval)
  }, [])

  const fetchIncidentResponseData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [incidentsData, playbooksData, responsesData, actionsData] = await Promise.all([
        API_CONFIG.request('/incident/incidents'),
        API_CONFIG.request('/incident/playbooks'),
        API_CONFIG.request('/incident/active-responses'),
        API_CONFIG.request('/incident/response-actions')
      ])
      
      setIncidents(incidentsData.incidents || [])
      setPlaybooks(playbooksData.playbooks || [])
      setActiveResponses(responsesData.responses || [])
      setResponseActions(actionsData.actions || [])
    } catch (error) {
      setError('Failed to fetch incident response data')
      console.error('Incident response fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const executePlaybook = async (incidentId: any, playbookId) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/incident/execute-playbook', {
        method: 'POST',
        body: JSON.stringify({ incident_id: incidentId, playbook_id: playbookId })
      })
      
      // Refresh data
      fetchIncidentResponseData()
    } catch (error) {
      console.error('Failed to execute playbook:', error)
    }
  }

  const pauseResponse = async (responseId: any) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/incident/pause-response', {
        method: 'POST',
        body: JSON.stringify({ response_id: responseId })
      })
      
      // Refresh data
      fetchIncidentResponseData()
    } catch (error) {
      console.error('Failed to pause response:', error)
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'new':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      case 'investigating':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      case 'responding':
        return 'text-blue-400 bg-blue-900/20 border-blue-500/20'
      case 'resolved':
        return 'text-green-400 bg-green-900/20 border-green-500/20'
      case 'closed':
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getPriorityColor = (priority: any) => {
    switch (priority) {
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

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading incident response...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchIncidentResponseData} className="hades-button-primary">
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
          <AlertTriangle className="mr-3 text-hades-primary" />
          Automated Incident Response
        </h1>
        <p className="text-gray-400">Automated response orchestration with playbooks and actions</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Active Incidents</p>
              <p className="text-2xl font-bold text-white">{incidents.filter(i => i.status !== 'resolved' && i.status !== 'closed').length}</p>
            </div>
            <AlertTriangle className="w-8 h-8 text-red-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Active Responses</p>
              <p className="text-2xl font-bold text-white">{activeResponses.filter(r => r.status === 'running').length}</p>
            </div>
            <PlayCircle className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Playbooks</p>
              <p className="text-2xl font-bold text-white">{playbooks.length}</p>
            </div>
            <FileText className="w-8 h-8 text-blue-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Response Rate</p>
              <p className="text-2xl font-bold text-white">
                {incidents.length > 0 ? Math.round((activeResponses.length / incidents.length) * 100) : 0}%
              </p>
            </div>
            <Settings className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Incidents */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <AlertTriangle className="mr-2 text-red-400" />
              Security Incidents
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {incidents.map((incident, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{incident.title}</h3>
                    <p className="text-gray-400 text-sm mt-1">{incident.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(incident.status)}`}>
                        {incident.status}
                      </span>
                      <span className={`text-xs ${getPriorityColor(incident.priority)}`}>
                        {incident.priority}
                      </span>
                      <span className="text-xs text-gray-400">
                        {new Date(incident.created_at).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                  <button 
                    onClick={() => setSelectedIncident(incident)}
                    className="ml-3 p-2 text-gray-400 hover:text-white"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Active Responses */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <PlayCircle className="mr-2 text-green-400" />
              Active Responses
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {activeResponses.map((response, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{response.playbook_name}</h3>
                    <p className="text-gray-400 text-sm">{response.incident_title}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${response.status === 'running' ? 'text-green-400 bg-green-900/20 border-green-500/20' : response.status === 'paused' ? 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20' : 'text-gray-400 bg-gray-900/20 border-gray-500/20'}`}>
                        {response.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {response.completed_steps}/{response.total_steps} steps
                      </span>
                      <span className="text-xs text-gray-400">
                        {Math.round((response.completed_steps / response.total_steps) * 100)}%
                      </span>
                    </div>
                    <div className="mt-2">
                      <div className="w-full bg-gray-700 rounded-full h-2">
                        <div 
                          className="bg-green-400 h-2 rounded-full" 
                          style={{ width: `${(response.completed_steps / response.total_steps) * 100}%` }}
                        ></div>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    {response.status === 'running' && (
                      <button 
                        onClick={() => pauseResponse(response.id)}
                        className="p-2 text-yellow-400 hover:text-yellow-300"
                        title="Pause"
                      >
                        <PauseCircle className="w-4 h-4" />
                      </button>
                    )}
                    {response.status === 'paused' && (
                      <button 
                        onClick={() => executePlaybook(response.incident_id, response.playbook_id)}
                        className="p-2 text-green-400 hover:text-green-300"
                        title="Resume"
                      >
                        <PlayCircle className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Response Playbooks */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <FileText className="mr-2 text-blue-400" />
              Response Playbooks
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {playbooks.map((playbook, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{playbook.name}</h3>
                    <p className="text-gray-400 text-sm">{playbook.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        {playbook.steps?.length || 0} steps
                      </span>
                      <span className={`text-xs ${getPriorityColor(playbook.priority)}`}>
                        {playbook.priority}
                      </span>
                      <span className="text-xs text-gray-400">
                        {playbook.triggers?.length || 0} triggers
                      </span>
                    </div>
                  </div>
                  <FileText className="w-5 h-5 text-blue-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Response Actions */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Settings className="mr-2 text-hades-primary" />
              Response Actions
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {responseActions.slice().reverse().map((action, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{action.action_type}</h3>
                    <p className="text-gray-400 text-sm">{action.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${action.status === 'completed' ? 'text-green-400 bg-green-900/20 border-green-500/20' : action.status === 'failed' ? 'text-red-400 bg-red-900/20 border-red-500/20' : 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'}`}>
                        {action.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {new Date(action.executed_at).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                  <Settings className="w-5 h-5 text-hades-primary" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Incident Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Eye className="mr-2 text-hades-primary" />
              Incident Details
            </h2>
          </div>
          <div className="p-4">
            {selectedIncident ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedIncident.title}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedIncident.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Status</p>
                    <span className={`px-2 py-1 rounded text-sm ${getStatusColor(selectedIncident.status)}`}>
                      {selectedIncident.status}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Priority</p>
                    <span className={`text-lg font-medium ${getPriorityColor(selectedIncident.priority)}`}>
                      {selectedIncident.priority}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Created</p>
                    <p className="text-white">{new Date(selectedIncident.created_at).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Assigned To</p>
                    <p className="text-white">{selectedIncident.assigned_to || 'Unassigned'}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Available Playbooks</p>
                    <div className="space-y-2">
                      {playbooks
                        .filter(pb => pb.triggers?.some(trigger => 
                          selectedIncident.tags?.includes(trigger) || 
                          selectedIncident.type === trigger
                        ))
                        .map(playbook => (
                          <div key={playbook.id} className="flex items-center justify-between p-2 bg-gray-900 rounded border border-gray-700">
                            <div>
                              <p className="text-white font-medium">{playbook.name}</p>
                              <p className="text-gray-400 text-sm">{playbook.description}</p>
                            </div>
                            <button 
                              onClick={() => executePlaybook(selectedIncident.id, playbook.id)}
                              className="hades-button-primary"
                            >
                              Execute
                            </button>
                          </div>
                        ))}
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Timeline</p>
                    <div className="space-y-2 mt-2">
                      {selectedIncident.timeline?.map((event: any, index) => (
                        <div key={index} className="flex items-start space-x-3">
                          <Clock className="w-4 h-4 text-gray-400 mt-1" />
                          <div>
                            <p className="text-white text-sm">{event.action}</p>
                            <p className="text-gray-400 text-xs">{new Date(event.timestamp).toLocaleString()}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <AlertTriangle className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select an incident to view detailed information</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default IncidentResponse
