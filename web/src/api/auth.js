// Authentication API
import API_CONFIG from './config'

export const authAPI = {
  // Login
  login: async (credentials) => {
    return await API_CONFIG.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials)
    })
  },

  // Logout
  logout: async () => {
    return await API_CONFIG.request('/auth/logout', {
      method: 'POST'
    })
  },

  // Refresh token
  refreshToken: async () => {
    return await API_CONFIG.request('/auth/refresh', {
      method: 'POST'
    })
  },

  // Get current user
  getCurrentUser: async () => {
    return await API_CONFIG.request('/auth/me')
  },
}
