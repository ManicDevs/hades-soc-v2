/**
 * Hot-swap configuration logic
 * Allows runtime configuration updates without application restart
 */

import { setCustomEndpoints, resetEndpoints } from './endpoints'
import { refreshCSRFToken } from './csrf'

export interface HotSwapConfig {
  endpoints?: string[]
  apiVersion?: string
  timeout?: number
}

export interface ValidationResult {
  valid: boolean
  errors: string[]
}

/**
 * Validates a new configuration
 * @param config - Configuration to validate
 * @returns Validation result with errors if any
 */
export function validateConfig(config: HotSwapConfig): ValidationResult {
  const errors: string[] = []

  // Validate endpoints if provided
  if (config.endpoints !== undefined) {
    if (!Array.isArray(config.endpoints)) {
      errors.push('Invalid configuration: endpoints must be an array')
    } else if (config.endpoints.length === 0) {
      errors.push('Invalid configuration: endpoints array cannot be empty')
    } else {
      // Validate each endpoint URL
      for (let i = 0; i < config.endpoints.length; i++) {
        const endpoint = config.endpoints[i]
        if (typeof endpoint !== 'string' || endpoint.length === 0) {
          errors.push(`Invalid configuration: endpoint at index ${i} must be a non-empty string`)
        } else if (!isValidUrl(endpoint)) {
          errors.push(`Invalid configuration: endpoint URL "${endpoint}" is not valid`)
        }
      }
    }
  }

  // Validate apiVersion if provided
  if (config.apiVersion !== undefined) {
    if (typeof config.apiVersion !== 'string' || config.apiVersion.length === 0) {
      errors.push('Invalid configuration: apiVersion must be a non-empty string')
    }
  }

  // Validate timeout if provided
  if (config.timeout !== undefined) {
    if (typeof config.timeout !== 'number' || config.timeout <= 0) {
      errors.push('Invalid configuration: timeout must be a positive number')
    }
  }

  return {
    valid: errors.length === 0,
    errors
  }
}

/**
 * Validates a URL string
 * @param url - URL to validate
 * @returns true if URL is valid
 */
function isValidUrl(url: string): boolean {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

/**
 * Updates the configuration with new values
 * @param newConfig - New configuration to apply
 * @returns true if configuration was successfully updated
 */
export async function updateConfiguration(newConfig: HotSwapConfig): Promise<boolean> {
  try {
    // Validate the new configuration
    const validation = validateConfig(newConfig)
    if (!validation.valid) {
      return false
    }

    // Update endpoints if provided
    if (newConfig.endpoints) {
      setCustomEndpoints(newConfig.endpoints)
    }

    // Optionally refresh CSRF token on configuration change
    if (newConfig.endpoints) {
      refreshCSRFToken()
    }

    return true
  } catch {
    return false
  }
}

/**
 * Resets configuration to defaults
 */
export function resetConfiguration(): void {
  resetEndpoints()
}

/**
 * Gets current hot-swap configuration state
 * @returns Current configuration object
 */
export function getCurrentConfig(): HotSwapConfig {
  return {
    endpoints: [],
    apiVersion: 'v1',
    timeout: 5000
  }
}

export const HotSwapConfig = {
  validateConfig,
  updateConfiguration,
  resetConfiguration,
  getCurrentConfig
}

export default HotSwapConfig