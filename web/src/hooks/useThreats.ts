import { useState, useEffect } from "react";
import { threatsAPI } from "../api/threats";
import type { Threat, ThreatFilters, ThreatStats } from "../types/models";

export const useThreats = () => {
  const [threats, setThreats] = useState<Threat[]>([]);
  const [stats, setStats] = useState<ThreatStats | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<ThreatFilters>({
    severity: "all",
    status: "all",
    type: "all",
  });

  useEffect(() => {
    fetchThreats();
    fetchThreatStats();

    // Set up real-time updates
    const interval = setInterval(() => {
      fetchThreatStats();
    }, 10000);

    return () => clearInterval(interval);
  }, [filters]);

  const fetchThreats = async () => {
    setLoading(true);
    setError(null);

    try {
      const threatsData = (await threatsAPI.getThreats(filters)) as any;
      // Extract threats array from nested API response structure
      const items: Threat[] =
        threatsData?.data?.data || threatsData?.data || [];
      setThreats(items);
    } catch (err) {
      setError("Failed to fetch threats");
      console.error("Threats fetch error:", err);
    } finally {
      setLoading(false);
    }
  };

  const fetchThreatStats = async () => {
    try {
      // Use the main threats endpoint which includes stats in metadata
      const threatsData = (await threatsAPI.getThreats(filters)) as any;
      // Extract stats from the metadata
      const computed: ThreatStats = {
        total_threats: threatsData?.data?.metadata?.total_threats || 0,
        by_severity:
          threatsData?.data?.data?.reduce(
            (acc: Record<string, number>, threat: any) => {
              acc[threat.severity] = (acc[threat.severity] || 0) + 1;
              return acc;
            },
            {},
          ) || {},
        by_status:
          threatsData?.data?.data?.reduce(
            (acc: Record<string, number>, threat: any) => {
              acc[threat.status] = (acc[threat.status] || 0) + 1;
              return acc;
            },
            {},
          ) || {},
        by_type:
          threatsData?.data?.data?.reduce(
            (acc: Record<string, number>, threat: any) => {
              acc[threat.type] = (acc[threat.type] || 0) + 1;
              return acc;
            },
            {},
          ) || {},
      };
      setStats(computed);
    } catch (err) {
      console.error("Threat stats fetch error:", err);
    }
  };

  const updateThreatStatus = async (id: string | number, status: string) => {
    try {
      await threatsAPI.updateThreatStatus(id, status);
      // Refresh threats list
      await fetchThreats();
      await fetchThreatStats();
    } catch (err) {
      setError("Failed to update threat status");
      throw err;
    }
  };

  const getThreatDetails = async (id: string | number) => {
    try {
      return await threatsAPI.getThreat(id);
    } catch (err) {
      setError("Failed to fetch threat details");
      throw err;
    }
  };

  const updateFilters = (newFilters: Partial<ThreatFilters>) => {
    setFilters((prev) => ({ ...prev, ...newFilters }));
  };

  const refreshData = () => {
    fetchThreats();
    fetchThreatStats();
  };

  return {
    threats,
    stats,
    loading,
    error,
    filters,
    updateThreatStatus,
    getThreatDetails,
    updateFilters,
    refreshData,
  };
};

export default useThreats;
