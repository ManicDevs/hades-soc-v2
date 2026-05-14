import React, { useState, useContext } from 'react'

import { authAPI } from '../api/auth'

// Type definitions
interface User {
  id: number | string
  username: string
  email: string
  role: string
  permissions: string[]
}

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  loading: boolean
  login: (credentials: { username: string; password: string }) => Promise<void>
  logout: () => Promise<void>
  setUser: (user: User | null) => void
  setIsAuthenticated: (auth: boolean) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

// Global window type declarations
declare global {
  interface Window {
    hadesToken: string | null
    hadesUser: User | null
    hadesRole: string | null
    hadesEnvironment: string | null
  }
  
  interface ImportMetaEnv {
    VITE_API_BASE_URL: string
    VITE_WS_BASE_URL?: string
  }
  
  interface ImportMeta {
    env: ImportMetaEnv
  }
}

const AuthContext = React.createContext<AuthContextType | undefined>(undefined)

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  
  return context
}

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Check for existing session on mount
  React.useEffect(() => {
    (async () => {
      const token = window.hadesToken || null
      const user = window.hadesUser || null
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
          // Session restored silently
        }
      } else if (isDevelopment) {
        // For development, get a real JWT token from backend
        const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
        try {
          const response = await fetch(`${apiUrl}/api/v1/auth/login`, {
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
            
            // Development session created with real JWT token
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
            
            // Fallback development session created
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
          
          // Fallback development session created
        }
      }
      
      setLoading(false)
    })()
  }, [])

  const login = async (credentials: any) => {
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
    setLoading(true)
    
    try {
      // Call logout API if available (only in production)
      const isDevelopment = window.location.hostname === 'localhost' ||
                           window.location.hostname === '127.0.0.1' ||
                           window.location.hostname.includes('dev') ||
                           window.location.hostname.includes('test') ||
                           window.location.hostname.includes('qa') ||
                           window.location.hostname.includes('staging')
      
      if (!isDevelopment) {
        await authAPI.logout()
      }
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      // Clear secure memory and state
      window.hadesToken = null
      window.hadesUser = null
      window.hadesRole = null
      window.hadesEnvironment = null
      
      // Clear authentication state
      setUser(null)
      setIsAuthenticated(false)
      setLoading(false)
      
      // Redirect to login page
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
    setError,
  }

  return React.createElement(
    AuthContext.Provider,
    { value: value },
    children
  )
}

export default useAuth
