// AI-Powered Threat Intelligence API
import API_CONFIG from "./config";

export const threatIntelligenceAPI = {
  // Get AI detected threats
  getThreats: async (): Promise<unknown> => {
    return await API_CONFIG.request("/ai/threats");
  },

  // Get anomaly detection results
  getAnomalies: async (): Promise<unknown> => {
    return await API_CONFIG.request("/ai/anomalies");
  },

  // Get ML predictions
  getPredictions: async (): Promise<unknown> => {
    return await API_CONFIG.request("/ai/predictions");
  },

  // Get threat intelligence overview
  getOverview: async (): Promise<unknown> => {
    return await API_CONFIG.request("/ai/overview");
  },

  // Analyze threat
  analyzeThreat: async (threatId: string): Promise<unknown> => {
    return await API_CONFIG.request("/ai/analyze", {
      method: "POST",
      body: JSON.stringify({ threat_id: threatId }),
    });
  },

  // Update threat intelligence
  updateIntelligence: async (
    data: Record<string, unknown>,
  ): Promise<unknown> => {
    return await API_CONFIG.request("/ai/intelligence", {
      method: "PUT",
      body: JSON.stringify(data),
    });
  },
};
