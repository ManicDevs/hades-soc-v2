/**
 * API endpoint configuration
 * Manages distributed API endpoints with load balancing
 */

import { getEnvironment } from './environments'

export interface Endpoint {
  url: string
  weight?: number
}

export interface EndpointConfig {
  development: string[]
  testing: string[]
  production: string[]
}

// In-memory endpoint storage for hot-swap support
let customEndpoints: string[] | null = null

/**
 * Gets the configured API endpoints for the current environment
 * @returns Array of endpoint URLs
 */
export function getAPIEndpoints(): string[] {
  // Return custom endpoints if hot-swapped
  if (customEndpoints !== null) {
    return customEndpoints
  }

  const env = getEnvironment()
  const apiBase = import.meta.env.VITE_API_BASE_URL

  const endpoints: EndpointConfig = {
    development: [
      apiBase || 'http://localhost:8080/api/v2'
    ],
    testing: [
      apiBase || 'http://localhost:8080/api/v2'
    ],
    production: [
      apiBase || `${window.location.protocol}//${window.location.host}/api/v2`
    ]
  }

  return endpoints[env] || endpoints.development
}

/**
 * Gets the base URL with load balancing
 * @returns The current endpoint URL based on load balancing index
 */
export function getBaseURL(): string {
  const endpoints = getAPIEndpoints()
  const currentIndex = getCurrentEndpointIndex()
  const endpoint = endpoints[currentIndex % endpoints.length]
  return endpoint || ''
}

// Load balancing index state
let currentEndpointIndex = 0

/**
 * Gets the current endpoint index for load balancing
 * @returns Current endpoint index
 */
export function getCurrentEndpointIndex(): number {
  return currentEndpointIndex
}

/**
 * Increments the endpoint index for load balancing
 * Moves to the next endpoint in a round-robin fashion
 */
export function incrementEndpointIndex(): void {
  const endpoints = getAPIEndpoints()
  currentEndpointIndex = (currentEndpointIndex + 1) % endpoints.length
}

/**
 * Sets the current endpoint index
 * @param index - The new endpoint index
 */
export function setCurrentEndpointIndex(index: number): void {
  currentEndpointIndex = index
}

/**
 * Sets custom endpoints (for hot-swap)
 * @param endpoints - Array of custom endpoint URLs
 */
export function setCustomEndpoints(endpoints: string[]): void {
  customEndpoints = endpoints
}

/**
 * Resets to default endpoints
 */
export function resetEndpoints(): void {
  customEndpoints = null
}

/**
 * Gets the current API version
 * @returns API version string
 */
export function getCurrentVersion(): string {
  return 'v2'
}

/**
 * Gets the list of supported API versions
 * @returns Array of supported versions
 */
export function getSupportedVersions(): string[] {
  return ['v1', 'v2', 'v3']
}

export const EndpointConfig = {
  getAPIEndpoints,
  getBaseURL,
  getCurrentEndpointIndex,
  incrementEndpointIndex,
  setCurrentEndpointIndex,
  setCustomEndpoints,
  resetEndpoints,
  getCurrentVersion,
  getSupportedVersions
}

export default EndpointConfig