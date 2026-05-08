// Authentication API - Use v1 endpoints for authentication
import API_CONFIG from './config'

export const authAPI = {
  // Login - Use direct v1 endpoint
  login: async (credentials) => {
    try {
      const response = await fetch('http://192.168.0.2:8080/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify(credentials)
      })
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }
      
      const data = await response.json()
      return data.data || data
    } catch (error) {
      console.error('Auth login error:', error)
      throw error
    }
  },

  // Logout
  logout: async () => {
    return await API_CONFIG.request('/v1/auth/logout', {
      method: 'POST'
    })
  },

  // Refresh token
  refreshToken: async () => {
    return await API_CONFIG.request('/v1/auth/refresh', {
      method: 'POST'
    })
  },

  // Get current user
  getCurrentUser: async () => {
    return await API_CONFIG.request('/v1/auth/me')
  },
}
