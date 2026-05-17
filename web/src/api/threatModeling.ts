// Threat Modeling and Attack Simulation API
import API_CONFIG from "./config";

export const threatModelingAPI = {
  // Get threat models
  getModels: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat/models");
  },

  // Get attack scenarios
  getScenarios: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat/scenarios");
  },

  // Get vulnerabilities
  getVulnerabilities: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat/vulnerabilities");
  },

  // Get mitigations
  getMitigations: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat/mitigations");
  },

  // Get simulations
  getSimulations: async (): Promise<unknown> => {
    return await API_CONFIG.request("/threat/simulations");
  },

  // Run simulation
  runSimulation: async (scenarioId: string): Promise<unknown> => {
    return await API_CONFIG.request("/threat/simulations", {
      method: "POST",
      body: JSON.stringify({ scenario_id: scenarioId }),
    });
  },

  // Create threat model
  createModel: async (modelData: Record<string, unknown>): Promise<unknown> => {
    return await API_CONFIG.request("/threat/models", {
      method: "POST",
      body: JSON.stringify(modelData),
    });
  },

  // Create attack scenario
  createScenario: async (
    scenarioData: Record<string, unknown>,
  ): Promise<unknown> => {
    return await API_CONFIG.request("/threat/scenarios", {
      method: "POST",
      body: JSON.stringify(scenarioData),
    });
  },

  // Update vulnerability
  updateVulnerability: async (
    vulnId: string,
    status: string,
  ): Promise<unknown> => {
    return await API_CONFIG.request(`/threat/vulnerabilities/${vulnId}`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
  },

  // Get simulation results
  getSimulationResults: async (simulationId: string): Promise<unknown> => {
    return await API_CONFIG.request(
      `/threat/simulations/${simulationId}/results`,
    );
  },
};
