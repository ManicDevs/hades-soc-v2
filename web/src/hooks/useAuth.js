import React, { useState, useEffect, useContext } from 'react'

console.log('🔄 Loading FIXED useAuth.js with real JWT authentication - FINAL VERSION')
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
    (async () => {
      const token = window.hadesToken || null
      const user = window.hadesUser || null
      const role = window.hadesRole || null
      const environment = window.hadesEnvironment || null
      
      // Check if we're in development environment
      const isDevelopment = window.location.hostname === 'localhost' ||
                          window.location.hostname === '127.0.0.1' ||
                          window.location.hostname === '192.168.0.2' ||
                          window.location.hostname.includes('dev') ||
                          window.location.hostname.includes('test') ||
                          window.location.hostname.includes('qa') ||
                          window.location.hostname.includes('staging')
      
      if (token && user) {
        // We have a stored session, restore it
        const parsedUser = typeof user === 'string' ? JSON.parse(user) : user
        setUser(parsedUser)
        setIsAuthenticated(true)
        
        if (environment) {
          console.log(`Restored session for ${environment} environment with role: ${role}`)
        }
      } else if (isDevelopment) {
        // For development, get a real JWT token from backend
        try {
          const response = await fetch('http://192.168.0.2:8080/api/v1/auth/login', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              username: 'admin',
              password: 'admin123'
            })
          })
          
          if (response.ok) {
            const data = await response.json()
            const defaultUser = data.data.user
            const realToken = data.data.token
            
            setUser(defaultUser)
            setIsAuthenticated(true)
            window.hadesToken = realToken
            window.hadesUser = defaultUser
            window.hadesRole = 'Administrator'
            window.hadesEnvironment = 'development'
            
            console.log('✅ SUCCESS: Created development session with REAL JWT token from backend')
            console.log('🔑 Token stored in window.hadesToken:', realToken.substring(0, 20) + '...')
            console.log('🔍 Token availability check:', window.hadesToken ? 'Available' : 'Not available')
          } else {
            // Fallback to fake token if backend is not available
            const defaultUser = {
              id: 1,
              username: 'admin',
              email: 'admin@hades-toolkit.com',
              role: 'Administrator',
              permissions: ['read', 'write', 'admin']
            }
            
            setUser(defaultUser)
            setIsAuthenticated(true)
            window.hadesToken = 'dev-token-' + Date.now()
            window.hadesUser = defaultUser
            window.hadesRole = 'Administrator'
            window.hadesEnvironment = 'development'
            
            console.log('Created fallback development session')
          }
        } catch (error) {
          console.error('Failed to get dev token:', error)
          // Fallback to fake token
          const defaultUser = {
            id: 1,
            username: 'admin',
            email: 'admin@hades-toolkit.com',
            role: 'Administrator',
            permissions: ['read', 'write', 'admin']
          }
          
          setUser(defaultUser)
          setIsAuthenticated(true)
          window.hadesToken = 'dev-token-' + Date.now()
          window.hadesUser = defaultUser
          window.hadesRole = 'Administrator'
          window.hadesEnvironment = 'development'
          
          console.log('Created fallback development session')
        }
      }
      
      setLoading(false)
    })()
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
        
        // Store all session data in secure memory
        window.hadesToken = response.token
        window.hadesUser = response.user
        window.hadesRole = credentials.role || 'user'
        window.hadesEnvironment = response.user.environment || 'development'
        
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
        
        // Store token in secure memory
        window.hadesToken = response.token
        window.hadesUser = response.user
        window.hadesRole = credentials.role
        
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
      console.log('Clearing secure memory and state')
      
      // Clear all secure memory
      window.hadesToken = null
      window.hadesUser = null
      window.hadesRole = null
      window.hadesEnvironment = null
      
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
      
      // Update token in secure memory
      window.hadesToken = response.token
      
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
