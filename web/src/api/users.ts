// Users API
import API_CONFIG from "./config";

export const usersAPI = {
  // Get all users
  getUsers: async (
    filters: Record<string, string> = {} as Record<string, string>,
  ): Promise<unknown> => {
    const params = new URLSearchParams(filters);
    return await API_CONFIG.request(`/users?${params}`, {
      method: "GET",
    });
  },

  // Get user by ID
  getUser: async (id: string): Promise<unknown> => {
    return await API_CONFIG.request(`/users/${id}`, {
      method: "GET",
    });
  },

  // Create user
  createUser: async (userData: Record<string, unknown>): Promise<unknown> => {
    return await API_CONFIG.request("/users", {
      method: "POST",
      body: JSON.stringify(userData),
    });
  },

  // Update user
  updateUser: async (
    id: string,
    userData: Record<string, unknown>,
  ): Promise<unknown> => {
    return await API_CONFIG.request(`/users/${id}`, {
      method: "PUT",
      body: JSON.stringify(userData),
    });
  },

  // Delete user
  deleteUser: async (id: string): Promise<unknown> => {
    return await API_CONFIG.request(`/users/${id}`, {
      method: "DELETE",
    });
  },

  // Get user statistics
  getUserStats: async (): Promise<unknown> => {
    return await API_CONFIG.request("/users/stats", {
      method: "GET",
    });
  },

  // Get user roles
  getUserRoles: async (): Promise<unknown> => {
    return await API_CONFIG.request("/users/roles", {
      method: "GET",
    });
  },
};
