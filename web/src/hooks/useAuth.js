import React, { useState, useEffect, useContext } from 'react'
import { authAPI } from '../api/auth'

const AuthContext = React.createContext()

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  
  return context
}

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Check for existing session on mount
  React.useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('hades-token')
      const user = localStorage.getItem('hades-user')
      const role = localStorage.getItem('hades-role')
      const environment = localStorage.getItem('hades-environment')
      
      // Check if we're in development environment
      const isDevelopment = window.location.hostname === 'localhost' ||
                          window.location.hostname === '127.0.0.1' ||
                          window.location.hostname.includes('dev') ||
                          window.location.hostname.includes('test') ||
                          window.location.hostname.includes('qa') ||
                          window.location.hostname.includes('staging')
      
      if (token && user) {
        // We have a stored session, restore it
        const parsedUser = JSON.parse(user)
        setUser(parsedUser)
        setIsAuthenticated(true)
        
        if (environment) {
          console.log(`Restored session for ${environment} environment with role: ${role}`)
        }
      } else if (isDevelopment) {
        // For development, create a default session
        const defaultUser = {
          id: 1,
          username: 'admin',
          email: 'admin@hades-toolkit.com',
          role: 'Administrator',
          permissions: ['read', 'write', 'admin']
        }
        
        setUser(defaultUser)
        setIsAuthenticated(true)
        localStorage.setItem('hades-token', 'dev-token-' + Date.now())
        localStorage.setItem('hades-user', JSON.stringify(defaultUser))
        localStorage.setItem('hades-role', 'Administrator')
        localStorage.setItem('hades-environment', 'development')
        
        console.log('Created default development session')
      }
      
      setLoading(false)
    }

    checkAuth()
  }, [])

  const login = async (credentials) => {
    setLoading(true)
    setError(null)
    
    try {
      // Check if this is a development environment login with role-based data
      const isDevelopment = window.location.hostname === 'localhost' ||
                           window.location.hostname === '127.0.0.1' ||
                           window.location.hostname.includes('dev') ||
                           window.location.hostname.includes('test') ||
                           window.location.hostname.includes('qa') ||
                           window.location.hostname.includes('staging')
      
      if (isDevelopment && credentials.user && credentials.token) {
        // This is a development login with realistic data
        const response = credentials
        
        // Store all session data
        localStorage.setItem('hades-token', response.token)
        localStorage.setItem('hades-user', JSON.stringify(response.user))
        localStorage.setItem('hades-role', credentials.role || 'user')
        localStorage.setItem('hades-environment', response.user.environment || 'development')
        
        setUser(response.user)
        setIsAuthenticated(true)
        
        return response
      } else {
        // Production authentication (or fallback)
        const loginCredentials = {
          ...credentials,
          role: credentials.role || 'user'
        }
        
        const response = await authAPI.login(loginCredentials)
        
        // Store token
        localStorage.setItem('hades-token', response.token)
        localStorage.setItem('hades-user', JSON.stringify(response.user))
        localStorage.setItem('hades-role', credentials.role)
        
        setUser(response.user)
        setIsAuthenticated(true)
        
        return response
      }
    } catch (error) {
      setError('Invalid credentials. Please try again.')
      throw error
    }
  }

  const logout = async () => {
    console.log('Logout function called')
    setLoading(true)
    
    try {
      // Call logout API if available (only in production)
      const isDevelopment = window.location.hostname === 'localhost' ||
                           window.location.hostname === '127.0.0.1' ||
                           window.location.hostname.includes('dev') ||
                           window.location.hostname.includes('test') ||
                           window.location.hostname.includes('qa') ||
                           window.location.hostname.includes('staging')
      
      console.log('Is development environment:', isDevelopment)
      
      if (!isDevelopment) {
        await authAPI.logout()
      }
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      console.log('Clearing localStorage and state')
      
      // Clear all local storage
      localStorage.removeItem('hades-token')
      localStorage.removeItem('hades-user')
      localStorage.removeItem('hades-role')
      localStorage.removeItem('hades-environment')
      
      // Clear authentication state
      setUser(null)
      setIsAuthenticated(false)
      setLoading(false)
      
      console.log('Redirecting to login page')
      
      // Force redirect to login page using window.location for full page reload
      window.location.replace('/login')
    }
  }

  const refreshToken = async () => {
    try {
      const response = await authAPI.refreshToken()
      
      // Update token
      localStorage.setItem('hades-token', response.token)
      
      return response
    } catch (error) {
      // Refresh failed, logout user
      await logout()
      throw error
    }
  }

  const value = {
    user,
    isAuthenticated,
    loading,
    error,
    login,
    logout,
    refreshData: refreshToken,
    setUser,
    setIsAuthenticated,
    setLoading,
  }

  return React.createElement(
    AuthContext.Provider,
    { value: value },
    children
  )
}

export default useAuth
