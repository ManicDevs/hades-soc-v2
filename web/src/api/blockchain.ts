// Blockchain Audit Logging API
import API_CONFIG from "./config";

export const blockchainAPI = {
  // Get audit logs
  getAuditLogs: async () => {
    return await API_CONFIG.request("/blockchain/audit-logs");
  },

  // Get blockchain blocks
  getBlocks: async () => {
    return await API_CONFIG.request("/blockchain/blocks");
  },

  // Get transactions
  getTransactions: async () => {
    return await API_CONFIG.request("/blockchain/transactions");
  },

  // Get chain integrity
  getIntegrity: async () => {
    return await API_CONFIG.request("/blockchain/integrity");
  },

  // Verify chain integrity
  verifyIntegrity: async () => {
    return await API_CONFIG.request("/blockchain/verify", {
      method: "POST",
    });
  },

  // Add audit log
  addAuditLog: async (logData: Record<string, unknown>): Promise<unknown> => {
    return await API_CONFIG.request("/blockchain/audit-log", {
      method: "POST",
      body: JSON.stringify(logData),
    });
  },

  // Get block details
  getBlockDetails: async (blockId: string): Promise<unknown> => {
    return await API_CONFIG.request(`/blockchain/blocks/${blockId}`);
  },

  // Get transaction details
  getTransactionDetails: async (txId: string): Promise<unknown> => {
    return await API_CONFIG.request(`/blockchain/transactions/${txId}`);
  },
};
