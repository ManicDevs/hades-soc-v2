// Threat Hunting API
import API_CONFIG from "./config";

export const threatHuntingAPI = {
  // Get threat hunts
  getHunts: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat-hunting/hunts");
  },

  // Get detected threats
  getThreats: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat-hunting/threats");
  },

  // Get threat indicators
  getIndicators: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat-hunting/indicators");
  },

  // Start threat hunt
  startHunt: async (huntId: string): Promise<unknown> => {
    return await API_CONFIG.request("/threat-hunting/start", {
      method: "POST",
      body: JSON.stringify({ hunt_id: huntId }),
    });
  },

  // Stop threat hunt
  stopHunt: async (huntId: string): Promise<unknown> => {
    return await API_CONFIG.request("/threat-hunting/stop", {
      method: "POST",
      body: JSON.stringify({ hunt_id: huntId }),
    });
  },

  // Create hunt
  createHunt: async (huntData: Record<string, unknown>): Promise<unknown> => {
    return await API_CONFIG.request("/threat-hunting/create", {
      method: "POST",
      body: JSON.stringify(huntData),
    });
  },

  // Get hunt results
  getHuntResults: async (huntId: string): Promise<unknown> => {
    return await API_CONFIG.request(`/threat-hunting/results/${huntId}`);
  },
};
