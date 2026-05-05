import React, { useState } from 'react'
import { Shield, Eye, EyeOff, Lock, Mail, ChevronDown, User, Crown, ShieldCheck, Key } from 'lucide-react'

function Login({ onLogin }) {
  const [credentials, setCredentials] = useState({ username: '', password: '', role: 'user' })
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')
  const [showRoleMenu, setShowRoleMenu] = useState(false)

  // Environment detection - only show role selection in non-production environments
  const isDevelopmentEnvironment = window.location.hostname === 'localhost' ||
                                  window.location.hostname === '127.0.0.1' ||
                                  window.location.hostname.includes('dev') ||
                                  window.location.hostname.includes('test') ||
                                  window.location.hostname.includes('qa') ||
                                  window.location.hostname.includes('staging')

  // Get environment type for realistic data generation
  const getEnvironmentType = () => {
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
      return 'development'
    } else if (window.location.hostname.includes('dev')) {
      return 'development'
    } else if (window.location.hostname.includes('test')) {
      return 'testing'
    } else if (window.location.hostname.includes('qa')) {
      return 'qa'
    } else if (window.location.hostname.includes('staging')) {
      return 'staging'
    }
    return 'production'
  }

  const environmentType = getEnvironmentType()

  const roles = [
    { id: 'user', name: 'User', icon: User, description: 'Standard user access', tier: 'Basic' },
    { id: 'analyst', name: 'Security Analyst', icon: ShieldCheck, description: 'Security analysis tools', tier: 'Professional' },
    { id: 'engineer', name: 'Security Engineer', icon: Key, description: 'Full system access', tier: 'Enterprise' },
    { id: 'admin', name: 'Administrator', icon: Crown, description: 'Complete administrative access', tier: 'Premium' }
  ]

  const selectedRole = roles.find(role => role.id === credentials.role)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setIsLoading(true)
    setError('')

    try {
      // In development environments, allow role-based testing login
      if (isDevelopmentEnvironment) {
        // Create realistic user data based on environment type
        const realisticUser = createRealisticUser(credentials, selectedRole, environmentType)
        
        // Store user data
        localStorage.setItem('hades-token', realisticUser.token)
        localStorage.setItem('hades-user', JSON.stringify(realisticUser.user))
        localStorage.setItem('hades-role', credentials.role)
        localStorage.setItem('hades-environment', environmentType)
        
        // Call onLogin with realistic response
        await onLogin({ ...credentials, ...realisticUser })
      } else {
        // Production environment - normal authentication
        await onLogin(credentials)
      }
    } catch (err) {
      setError(`Invalid credentials for ${selectedRole?.name || 'selected role'}. Please try again.`)
    } finally {
      setIsLoading(false)
    }
  }

  const createRealisticUser = (credentials, role, envType) => {
    const envData = getEnvironmentSpecificData(envType)
    const userId = envData.userIdBase + (credentials.username?.length || 0)
    
    return {
      token: `${envData.tokenPrefix}-${Date.now()}-${userId}`,
      user: {
        id: userId,
        username: credentials.username || envData.defaultUsername,
        email: `${credentials.username || envData.defaultUsername}@${envData.emailDomain}`,
        role: role?.name || 'User',
        status: 'active',
        lastLogin: new Date(),
        permissions: getPermissionsForRole(credentials.role),
        department: envData.department,
        location: envData.location,
        employeeId: envData.employeeIdPrefix + String(userId).padStart(6, '0'),
        manager: envData.manager,
        joinDate: envData.joinDate,
        securityClearance: role?.tier || 'Basic',
        sessionTimeout: envData.sessionTimeout,
        ipAddress: envData.ipAddress,
        userAgent: navigator.userAgent,
        environment: envType,
        region: envData.region,
        datacenter: envData.datacenter,
        cluster: envData.cluster,
        lastActivity: new Date(),
        preferences: {
          theme: 'dark',
          language: 'en-US',
          timezone: envData.timezone,
          notifications: true,
          twoFactorEnabled: envType !== 'development'
        }
      }
    }
  }

  const getEnvironmentSpecificData = (envType) => {
    const environments = {
      development: {
        userIdBase: 1000,
        defaultUsername: 'devuser',
        emailDomain: 'dev.hades-toolkit.com',
        department: 'Engineering',
        location: 'San Francisco, CA',
        employeeIdPrefix: 'DEV-',
        manager: 'John Smith (Engineering Lead)',
        joinDate: new Date('2024-01-15'),
        sessionTimeout: 7200, // 2 hours
        ipAddress: '192.168.1.100',
        timezone: 'America/Los_Angeles',
        region: 'us-west-2',
        datacenter: 'dev-cluster-1',
        cluster: 'development',
        tokenPrefix: 'DEV'
      },
      testing: {
        userIdBase: 2000,
        defaultUsername: 'testuser',
        emailDomain: 'test.hades-toolkit.com',
        department: 'Quality Assurance',
        location: 'Austin, TX',
        employeeIdPrefix: 'TST-',
        manager: 'Sarah Johnson (QA Manager)',
        joinDate: new Date('2024-02-01'),
        sessionTimeout: 3600, // 1 hour
        ipAddress: '10.0.0.50',
        timezone: 'America/Chicago',
        region: 'us-central-1',
        datacenter: 'test-cluster-2',
        cluster: 'testing',
        tokenPrefix: 'TST'
      },
      qa: {
        userIdBase: 3000,
        defaultUsername: 'qauser',
        emailDomain: 'qa.hades-toolkit.com',
        department: 'Quality Assurance',
        location: 'Denver, CO',
        employeeIdPrefix: 'QA-',
        manager: 'Mike Wilson (QA Lead)',
        joinDate: new Date('2024-01-20'),
        sessionTimeout: 1800, // 30 minutes
        ipAddress: '172.16.0.25',
        timezone: 'America/Denver',
        region: 'us-west-1',
        datacenter: 'qa-cluster-1',
        cluster: 'quality-assurance',
        tokenPrefix: 'QA'
      },
      staging: {
        userIdBase: 4000,
        defaultUsername: 'stageuser',
        emailDomain: 'stage.hades-toolkit.com',
        department: 'Staging Operations',
        location: 'New York, NY',
        employeeIdPrefix: 'STG-',
        manager: 'Emily Davis (Staging Manager)',
        joinDate: new Date('2024-03-01'),
        sessionTimeout: 2400, // 40 minutes
        ipAddress: '10.1.0.75',
        timezone: 'America/New_York',
        region: 'us-east-1',
        datacenter: 'staging-cluster-1',
        cluster: 'staging',
        tokenPrefix: 'STG'
      }
    }
    
    return environments[envType] || environments.development
  }

  const getPermissionsForRole = (roleId) => {
    const permissions = {
      'user': ['read', 'dashboard'],
      'analyst': ['read', 'write', 'security', 'reports'],
      'engineer': ['read', 'write', 'admin', 'security', 'reports', 'config'],
      'admin': ['read', 'write', 'admin', 'security', 'reports', 'config', 'users']
    }
    return permissions[roleId] || ['read']
  }

  // Close dropdown when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event) => {
      if (showRoleMenu && !event.target.closest('.relative')) {
        setShowRoleMenu(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [showRoleMenu])

  return (
    <div className="min-h-screen bg-hades-dark flex items-center justify-center px-4">
      <div className="max-w-md w-full">
        {/* Logo and Title */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-hades-primary rounded-2xl mb-4">
            <Shield className="w-8 h-8 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-white mb-2">Hades Toolkit</h1>
          <p className="text-gray-400">Enterprise Security Platform</p>
        </div>

        {/* Login Form */}
        <div className="hades-card p-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Username */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Username
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  required
                  value={credentials.username}
                  onChange={(e) => setCredentials({ ...credentials, username: e.target.value })}
                  className="hades-input pl-10 w-full"
                  placeholder="Enter your username"
                />
              </div>
            </div>

            {/* Role Selection - Only show in development/testing environments */}
            {isDevelopmentEnvironment && (
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  Test Role Selection
                  <span className="ml-2 text-xs px-2 py-1 bg-blue-900/30 text-blue-400 rounded">
                    Testing Only
                  </span>
                </label>
                <div className="relative">
                  <button
                    type="button"
                    onClick={() => setShowRoleMenu(!showRoleMenu)}
                    className="hades-input pl-10 pr-10 w-full flex items-center justify-between hover:bg-gray-800 transition-colors"
                  >
                    <div className="flex items-center">
                      {selectedRole && React.createElement(selectedRole.icon, { className: "w-4 h-4 text-gray-400 mr-3" })}
                      <span className="text-gray-300">{selectedRole?.name || 'Select test role'}</span>
                    </div>
                    <ChevronDown className={`w-4 h-4 text-gray-400 transition-transform ${showRoleMenu ? 'rotate-180' : ''}`} />
                  </button>

                  {/* Role Dropdown Menu */}
                  {showRoleMenu && (
                    <div className="absolute z-10 w-full mt-2 bg-gray-800 border border-gray-700 rounded-lg shadow-lg">
                      <div className="py-2">
                        {roles.map((role) => {
                          const Icon = role.icon
                          return (
                            <button
                              key={role.id}
                              type="button"
                              onClick={() => {
                                setCredentials({ ...credentials, role: role.id })
                                setShowRoleMenu(false)
                              }}
                              className={`w-full px-4 py-3 flex items-start space-x-3 hover:bg-gray-700 transition-colors ${
                                credentials.role === role.id ? 'bg-gray-700' : ''
                              }`}
                            >
                              <Icon className="w-5 h-5 text-gray-400 mt-0.5 flex-shrink-0" />
                              <div className="text-left flex-1">
                                <div className="flex items-center justify-between">
                                  <span className="text-white font-medium">{role.name}</span>
                                  <span className={`text-xs px-2 py-1 rounded ${
                                    role.tier === 'Basic' ? 'bg-green-900/30 text-green-400' :
                                    role.tier === 'Professional' ? 'bg-blue-900/30 text-blue-400' :
                                    role.tier === 'Enterprise' ? 'bg-purple-900/30 text-purple-400' :
                                    'bg-orange-900/30 text-orange-400'
                                  }`}>
                                    {role.tier}
                                  </span>
                                </div>
                                <p className="text-gray-400 text-sm mt-1">{role.description}</p>
                              </div>
                            </button>
                          )
                        })}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Password */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type={showPassword ? 'text' : 'password'}
                  required
                  value={credentials.password}
                  onChange={(e) => setCredentials({ ...credentials, password: e.target.value })}
                  className="hades-input pl-10 pr-10 w-full"
                  placeholder="Enter your password"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
            </div>

            {/* Error Message */}
            {error && (
              <div className="bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded-md text-sm">
                {error}
              </div>
            )}

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading}
              className="hades-button-primary w-full py-3 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          {/* Role Information - Only show in development environments */}
          {isDevelopmentEnvironment && (
            <div className="mt-6 p-4 bg-gray-800/50 rounded-lg border border-gray-700">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-300">Test Role Selected:</span>
                <span className={`text-xs px-2 py-1 rounded ${
                  selectedRole?.tier === 'Basic' ? 'bg-green-900/30 text-green-400' :
                  selectedRole?.tier === 'Professional' ? 'bg-blue-900/30 text-blue-400' :
                  selectedRole?.tier === 'Enterprise' ? 'bg-purple-900/30 text-purple-400' :
                  'bg-orange-900/30 text-orange-400'
                }`}>
                  {selectedRole?.tier || 'Basic'}
                </span>
              </div>
              <p className="text-gray-400 text-xs">
                {selectedRole?.description || 'Select a test role above'}
              </p>
              <p className="text-blue-400 text-xs mt-2">
                💡 This is a testing environment - any credentials will work
              </p>
            </div>
          )}

          {/* Additional Options */}
          <div className="mt-4 text-center">
            {isDevelopmentEnvironment ? (
              <>
                <p className="text-blue-400 text-sm">
                  Development Environment - Testing Mode Active
                </p>
                <p className="text-gray-500 text-xs mt-2">
                  Role-based testing enabled - Select a role to test different access levels
                </p>
              </>
            ) : (
              <>
                <p className="text-gray-400 text-sm">
                  Valid credentials required for access
                </p>
                <p className="text-gray-500 text-xs mt-2">
                  Contact your administrator for login credentials
                </p>
              </>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-gray-500 text-xs">
            © 2024 Hades Toolkit. All rights reserved.
          </p>
        </div>
      </div>
    </div>
  )
}

export default Login
