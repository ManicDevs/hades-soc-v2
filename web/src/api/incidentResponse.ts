// Automated Incident Response API
import API_CONFIG from './config'

export const incidentResponseAPI = {
  // Get incidents
  getIncidents: async () => {
    return await API_CONFIG.request('/incident/incidents')
  },

  // Get response playbooks
  getPlaybooks: async () => {
    return await API_CONFIG.request('/incident/playbooks')
  },

  // Get active responses
  getActiveResponses: async () => {
    return await API_CONFIG.request('/incident/active-responses')
  },

  // Get response actions
  getResponseActions: async () => {
    return await API_CONFIG.request('/incident/response-actions')
  },

  // Execute playbook
  executePlaybook: async (incidentId: string, playbookId: string) => {
    return await API_CONFIG.request('/incident/execute-playbook', {
      method: 'POST',
      body: JSON.stringify({ incident_id: incidentId, playbook_id: playbookId })
    })
  },

  // Pause response
  pauseResponse: async (responseId: string) => {
    return await API_CONFIG.request('/incident/pause-response', {
      method: 'POST',
      body: JSON.stringify({ response_id: responseId })
    })
  },

  // Create incident
  createIncident: async (incidentData: Record<string, any>) => {
    return await API_CONFIG.request('/incident/create', {
      method: 'POST',
      body: JSON.stringify(incidentData)
    })
  },

  // Update incident status
  updateIncidentStatus: async (incidentId: string, status: string) => {
    return await API_CONFIG.request(`/incident/${incidentId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status })
    })
  },

  // Get incident details
  getIncidentDetails: async (incidentId: string) => {
    return await API_CONFIG.request(`/incident/incidents/${incidentId}`)
  },

  // Get incident timeline
  getIncidentTimeline: async (incidentId: string) => {
    return await API_CONFIG.request(`/incident/${incidentId}/timeline`)
  }
}
