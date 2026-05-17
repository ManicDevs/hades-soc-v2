import { useState, useEffect } from "react";
import { usersAPI } from "../api/users";
import type { User, UserFilters, UserStats } from "../types/models";

export const useUsers = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [roles, setRoles] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<UserFilters>({
    search: "",
    role: "all",
    status: "all",
  });

  useEffect(() => {
    fetchUsers();
    fetchUserStats();
    fetchUserRoles();
  }, [filters]);

  const fetchUsers = async () => {
    setLoading(true);
    setError(null);

    try {
      const usersData = (await usersAPI.getUsers(filters)) as
        | User[]
        | undefined;
      const list: User[] = usersData ?? [];
      setUsers(list);

      // Calculate stats from users data
      const calculatedStats: UserStats = {
        total_users: list.length || 0,
        active_users: list.filter((u) => u.status === "active").length || 0,
        inactive_users: list.filter((u) => u.status === "inactive").length || 0,
        by_role: list.reduce((acc: Record<string, number>, user: User) => {
          const role = user.role || "unknown";
          acc[role] = (acc[role] || 0) + 1;
          return acc;
        }, {}),
        by_status: list.reduce((acc: Record<string, number>, user: User) => {
          const st = user.status || "unknown";
          acc[st] = (acc[st] || 0) + 1;
          return acc;
        }, {}),
      };
      setStats(calculatedStats);

      // Extract unique roles from users data
      const uniqueRoles = [
        ...new Set(list.map((user) => user.role || "")),
      ].filter((r) => r);
      setRoles(uniqueRoles);
    } catch (err) {
      setError("Failed to fetch users");
      console.error("Users fetch error:", err);
    } finally {
      setLoading(false);
    }
  };

  const fetchUserStats = async () => {
    // Stats are now calculated in fetchUsers to avoid broken endpoint
  };

  const fetchUserRoles = async () => {
    // Roles are now extracted in fetchUsers to avoid broken endpoint
  };

  const createUser = async (userData: Partial<User>) => {
    try {
      await usersAPI.createUser(userData as Record<string, unknown>);
      await fetchUsers();
      await fetchUserStats();
    } catch (err) {
      setError("Failed to create user");
      throw err;
    }
  };

  const updateUser = async (id: string | number, userData: Partial<User>) => {
    try {
      await usersAPI.updateUser(
        String(id),
        userData as Record<string, unknown>,
      );
      await fetchUsers();
    } catch (err) {
      setError("Failed to update user");
      throw err;
    }
  };

  const deleteUser = async (id: string | number) => {
    try {
      await usersAPI.deleteUser(String(id));
      await fetchUsers();
      await fetchUserStats();
    } catch (err) {
      setError("Failed to delete user");
      throw err;
    }
  };

  const getUserDetails = async (id: string | number) => {
    try {
      return (await usersAPI.getUser(String(id))) as User;
    } catch (err) {
      setError("Failed to fetch user details");
      throw err;
    }
  };

  const updateFilters = (newFilters: Partial<UserFilters>) => {
    setFilters((prev) => ({ ...prev, ...newFilters }));
  };

  const refreshData = () => {
    fetchUsers();
    fetchUserStats();
    fetchUserRoles();
  };

  return {
    users,
    stats,
    roles,
    loading,
    error,
    filters,
    createUser,
    updateUser,
    deleteUser,
    getUserDetails,
    updateFilters,
    refreshData,
  };
};

export default useUsers;
