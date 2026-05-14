/**
 * System status checking
 * Provides health check utilities and system status monitoring
 */

import { getEnvironment } from './environments'
import { getAPIEndpoints } from './endpoints'
import { getDefaultHeaders } from './request'

export interface EndpointStatus {
  url: string
  status: 'healthy' | 'unhealthy'
  responseTime?: string
  error?: string
}

export interface SystemStatus {
  environment: string
  endpoints: EndpointStatus[]
  healthy: boolean
}

/**
 * Checks health of a single endpoint
 * @param endpoint - Endpoint URL to check
 * @returns Promise with endpoint status
 */
async function checkEndpointHealth(endpoint: string): Promise<EndpointStatus> {
  try {
    const response = await fetch(`${endpoint}/health`, {
      headers: getDefaultHeaders(),
      method: 'GET'
    })

    if (response.ok) {
      const responseTime = response.headers.get('x-response-time')
      return {
        url: endpoint,
        status: 'healthy' as const,
        ...(responseTime && { responseTime })
      }
    }

    return {
      url: endpoint,
      status: 'unhealthy' as const,
      error: `HTTP ${response.status}`
    }
  } catch (error) {
    return {
      url: endpoint,
      status: 'unhealthy',
      error: error instanceof Error ? error.message : 'Unknown error'
    }
  }
}

/**
 * Gets system status across all configured endpoints
 * @returns System status object with endpoint health information
 */
export async function getSystemStatus(): Promise<SystemStatus> {
  const endpoints = getAPIEndpoints()
  const environment = getEnvironment()

  const status: SystemStatus = {
    environment,
    endpoints: [],
    healthy: false
  }

  // Check each endpoint in parallel
  const results = await Promise.all(endpoints.map(checkEndpointHealth))

  status.endpoints = results

  // System is healthy if at least one endpoint is healthy
  status.healthy = results.some(result => result.status === 'healthy')

  return status
}

/**
 * Checks if the system is healthy
 * @returns Promise that resolves to true if system is healthy
 */
export async function isSystemHealthy(): Promise<boolean> {
  const status = await getSystemStatus()
  return status.healthy
}

/**
 * Gets the number of healthy endpoints
 * @returns Promise with count of healthy endpoints
 */
export async function getHealthyEndpointCount(): Promise<number> {
  const status = await getSystemStatus()
  return status.endpoints.filter(e => e.status === 'healthy').length
}

/**
 * Gets detailed status for a specific endpoint
 * @param endpoint - Endpoint URL to check
 * @returns Endpoint status object
 */
export async function getEndpointStatus(endpoint: string): Promise<EndpointStatus> {
  return checkEndpointHealth(endpoint)
}

export const StatusConfig = {
  getSystemStatus,
  isSystemHealthy,
  getHealthyEndpointCount,
  getEndpointStatus
}

export default StatusConfig