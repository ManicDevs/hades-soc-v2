/**
 * Environment detection utilities
 * Provides functions to detect and check the current runtime environment
 */

export type Environment = 'development' | 'testing' | 'production'

/**
 * Detects the current environment based on hostname
 * @returns The detected environment type
 */
export function getEnvironment(): Environment {
  const hostname = window.location.hostname

  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return 'development'
  } else if (hostname.includes('dev')) {
    return 'development'
  } else if (hostname.includes('test')) {
    return 'testing'
  }
  return 'production'
}

/**
 * Checks if the current environment is development
 * @returns true if in development mode
 */
export function isDevelopment(): boolean {
  return getEnvironment() === 'development'
}

/**
 * Checks if the current environment is testing
 * @returns true if in testing mode
 */
export function isTesting(): boolean {
  return getEnvironment() === 'testing'
}

/**
 * Checks if the current environment is production
 * @returns true if in production mode
 */
export function isProduction(): boolean {
  return getEnvironment() === 'production'
}

/**
 * Environment configuration object
 */
export const EnvironmentConfig = {
  getEnvironment,
  isDevelopment,
  isTesting,
  isProduction
}

export default EnvironmentConfig