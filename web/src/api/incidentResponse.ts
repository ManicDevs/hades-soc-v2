// Automated Incident Response API
import API_CONFIG from "./config";

export const incidentResponseAPI = {
  // Get incidents
  getIncidents: async (): Promise<unknown> => {
    return await API_CONFIG.request("/incident/incidents");
  },

  // Get response playbooks
  getPlaybooks: async (): Promise<unknown> => {
    return await API_CONFIG.request("/incident/playbooks");
  },

  // Get active responses
  getActiveResponses: async (): Promise<unknown> => {
    return await API_CONFIG.request("/incident/active-responses");
  },

  // Get response actions
  getResponseActions: async (): Promise<unknown> => {
    return await API_CONFIG.request("/incident/response-actions");
  },

  // Execute playbook
  executePlaybook: async (
    incidentId: string,
    playbookId: string,
  ): Promise<unknown> => {
    return await API_CONFIG.request("/incident/execute-playbook", {
      method: "POST",
      body: JSON.stringify({
        incident_id: incidentId,
        playbook_id: playbookId,
      }),
    });
  },

  // Pause response
  pauseResponse: async (responseId: string): Promise<unknown> => {
    return await API_CONFIG.request("/incident/pause-response", {
      method: "POST",
      body: JSON.stringify({ response_id: responseId }),
    });
  },

  // Create incident
  createIncident: async (
    incidentData: Record<string, unknown>,
  ): Promise<unknown> => {
    return await API_CONFIG.request("/incident/create", {
      method: "POST",
      body: JSON.stringify(incidentData),
    });
  },

  // Update incident status
  updateIncidentStatus: async (
    incidentId: string,
    status: string,
  ): Promise<unknown> => {
    return await API_CONFIG.request(`/incident/${incidentId}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
  },

  // Get incident details
  getIncidentDetails: async (incidentId: string): Promise<unknown> => {
    return await API_CONFIG.request(`/incident/incidents/${incidentId}`);
  },

  // Get incident timeline
  getIncidentTimeline: async (incidentId: string): Promise<unknown> => {
    return await API_CONFIG.request(`/incident/${incidentId}/timeline`);
  },
};
