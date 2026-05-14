// Authentication API - Use v1 endpoints for authentication
import API_CONFIG from './config'

export const authAPI = {
  // Login - Use direct v1 endpoint
  login: async (credentials: { username: string; password: string }) => {
    try {
      const baseURL = API_CONFIG.getBaseURL().replace('/api/v2', '')
      const response = await fetch(`${baseURL}/api/v1/auth/login`, {
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
