// Users API
import API_CONFIG from './config'

export const usersAPI = {
  // Get all users
  getUsers: async (filters = {}) => {
    const params = new URLSearchParams(filters)
    return await API_CONFIG.request(`/users?${params}`)
  },

  // Get user by ID
  getUser: async (id) => {
    return await API_CONFIG.request(`/users/${id}`)
  },

  // Create user
  createUser: async (userData) => {
    return await API_CONFIG.request('/users', {
      method: 'POST',
      body: JSON.stringify(userData)
    })
  },

  // Update user
  updateUser: async (id, userData) => {
    return await API_CONFIG.request(`/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(userData)
    })
  },

  // Delete user
  deleteUser: async (id) => {
    return await API_CONFIG.request(`/users/${id}`, {
      method: 'DELETE'
    })
  },

  // Get user statistics
  getUserStats: async () => {
    return await API_CONFIG.request('/users/stats')
  },

  // Get user roles
  getUserRoles: async () => {
    return await API_CONFIG.request('/users/roles')
  },
}
