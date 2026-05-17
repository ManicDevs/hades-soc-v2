// Advanced SIEM Integration API
import API_CONFIG from "./config";

export const siemAPI = {
  // Get security events
  getEvents: async (): Promise<unknown> => {
    return await API_CONFIG.request("/siem/events");
  },

  // Get SIEM alerts
  getAlerts: async (): Promise<unknown> => {
    return await API_CONFIG.request("/siem/alerts");
  },

  // Get event correlations
  getCorrelations: async (): Promise<unknown> => {
    return await API_CONFIG.request("/siem/correlations");
  },

  // Get threat intelligence feeds
  getThreatFeeds: async (): Promise<unknown> => {
    return await API_CONFIG.request("/siem/threat-feeds");
  },

  // Acknowledge alert
  acknowledgeAlert: async (alertId: string): Promise<unknown> => {
    return await API_CONFIG.request("/siem/alerts/acknowledge", {
      method: "POST",
      body: JSON.stringify({ alert_id: alertId }),
    });
  },

  // Create correlation rule
  createCorrelation: async (
    ruleData: Record<string, unknown>,
  ): Promise<unknown> => {
    return await API_CONFIG.request("/siem/correlations/create", {
      method: "POST",
      body: JSON.stringify(ruleData),
    });
  },

  // Update threat feed
  updateThreatFeed: async (
    feedId: string,
    feedData: Record<string, unknown>,
  ): Promise<unknown> => {
    return await API_CONFIG.request(`/siem/threat-feeds/${feedId}`, {
      method: "PUT",
      body: JSON.stringify(feedData),
    });
  },

  // Get event details
  getEventDetails: async (eventId: string): Promise<unknown> => {
    return await API_CONFIG.request(`/siem/events/${eventId}`);
  },

  // Export events
  exportEvents: async (
    format: string,
    filters: Record<string, string>,
  ): Promise<unknown> => {
    const params = new URLSearchParams(filters);
    return await API_CONFIG.request(
      `/siem/events/export?format=${format}&${params}`,
    );
  },
};
