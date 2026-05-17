import { useState, useEffect } from "react";
import { securityAPI } from "../api/security";
import type {
  Policy,
  Vulnerability,
  AuditLog,
  SecurityScore,
} from "../types/models";

export const useSecurity = () => {
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [vulnerabilities, setVulnerabilities] = useState<Vulnerability[]>([]);
  const [securityScore, setSecurityScore] = useState<SecurityScore | null>(
    null,
  );
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchSecurityData();

    // Set up periodic updates for security score
    const interval = setInterval(() => {
      fetchSecurityScore();
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  const fetchSecurityData = async () => {
    console.log("fetchSecurityData starting");
    setLoading(true);
    setError(null);

    try {
      console.log("Calling security APIs...");
      const [policiesData, vulnerabilitiesData, scoreData] = await Promise.all([
        securityAPI.getPolicies(),
        securityAPI.getVulnerabilities(),
        securityAPI.getSecurityScore(),
      ]);

      console.log("Security API responses:", {
        policiesData,
        vulnerabilitiesData,
        scoreData,
      });

      setPolicies(policiesData as Policy[]);
      setVulnerabilities(vulnerabilitiesData as Vulnerability[]);
      setSecurityScore(scoreData as SecurityScore);
    } catch (err) {
      console.error("Security data fetch error:", err);
      setError("Failed to fetch security data");
    } finally {
      setLoading(false);
    }
  };

  const fetchSecurityScore = async () => {
    try {
      const scoreData = await securityAPI.getSecurityScore();
      setSecurityScore(scoreData as SecurityScore);
    } catch (err) {
      console.error("Security score fetch error:", err);
    }
  };

  const fetchAuditLogs = async (filters: Record<string, string> = {}) => {
    try {
      const logsData = await securityAPI.getAuditLogs(filters);
      setAuditLogs(logsData as AuditLog[]);
    } catch (err) {
      setError("Failed to fetch audit logs");
      console.error("Audit logs fetch error:", err);
    }
  };

  const updatePolicy = async (
    id: string | number,
    policyData: Partial<Policy>,
  ) => {
    try {
      await securityAPI.updatePolicy(
        String(id),
        policyData as Record<string, unknown>,
      );
      await fetchSecurityData();
    } catch (err) {
      setError("Failed to update security policy");
      throw err;
    }
  };

  const updateVulnerability = async (id: string | number, status: string) => {
    try {
      await securityAPI.updateVulnerability(String(id), status);
      await fetchSecurityData();
    } catch (err) {
      setError("Failed to update vulnerability");
      throw err;
    }
  };

  const runSecurityScan = async () => {
    try {
      await securityAPI.runSecurityScan();
      // Refresh data after scan
      setTimeout(() => {
        fetchSecurityData();
      }, 2000);
    } catch (err) {
      setError("Failed to run security scan");
      throw err;
    }
  };

  const refreshData = () => {
    fetchSecurityData();
  };

  return {
    policies,
    vulnerabilities,
    securityScore,
    auditLogs,
    loading,
    error,
    updatePolicy,
    updateVulnerability,
    runSecurityScan,
    fetchAuditLogs,
    refreshData,
  };
};

export default useSecurity;
