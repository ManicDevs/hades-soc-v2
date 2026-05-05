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
  executePlaybook: async (incidentId, playbookId) => {
    return await API_CONFIG.request('/incident/execute-playbook', {
      method: 'POST',
      body: JSON.stringify({ incident_id: incidentId, playbook_id: playbookId })
    })
  },

  // Pause response
  pauseResponse: async (responseId) => {
    return await API_CONFIG.request('/incident/pause-response', {
      method: 'POST',
      body: JSON.stringify({ response_id: responseId })
    })
  },

  // Create incident
  createIncident: async (incidentData) => {
    return await API_CONFIG.request('/incident/incidents', {
      method: 'POST',
      body: JSON.stringify(incidentData)
    })
  },

  // Update incident status
  updateIncidentStatus: async (incidentId, status) => {
    return await API_CONFIG.request(`/incident/incidents/${incidentId}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status })
    })
  },

  // Get incident details
  getIncidentDetails: async (incidentId) => {
    return await API_CONFIG.request(`/incident/incidents/${incidentId}`)
  }
}
