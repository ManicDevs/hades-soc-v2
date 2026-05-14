/**
 * Core request function
 * Handles API requests with retry logic, failover, and error handling
 *
 * Security Headers:
 * - X-Content-Type-Options: Prevents MIME type sniffing
 * - X-XSS-Protection: Enables browser XSS filtering
 * - Content-Security-Policy: Server-side header (documented here for reference)
 *
 * Note: Content-Security-Policy must be configured on the server side.
 * The server should return appropriate CSP headers to complement these client-side protections.
 */

import { getAuthToken } from '../lib/authToken'
import { isDevelopment } from './environments'
import {
  getAPIEndpoints,
  setCurrentEndpointIndex
} from './endpoints'
import { getFallbackData } from './fallbacks'
import { getCSRFToken } from './csrf'

export interface RequestOptions extends RequestInit {
  retry?: number
  retryDelay?: number
  timeout?: number
}

export interface RequestConfig {
  method?: string
  headers?: Record<string, string>
  body?: string | FormData | null
  retry?: number
  retryDelay?: number
  timeout?: number
}

/**
 * Security headers for API requests
 * These headers help protect against XSS and content-type sniffing attacks
 */
export const SECURITY_HEADERS = {
  'X-Content-Type-Options': 'nosniff',
  'X-XSS-Protection': '1; mode=block'
} as const

/**
 * Content Security Policy guidance
 * The server should include these CSP directives:
 *
 * default-src 'self';
 * script-src 'self';
 * style-src 'self' 'unsafe-inline';
 * img-src 'self' data: https:;
 * connect-src 'self' wss: https:;
 * font-src 'self';
 * object-src 'none';
 * frame-ancestors 'none';
 *
 * This helps prevent XSS attacks by controlling which resources can be loaded.
 */

const DEFAULT_RETRY_COUNT = 3
const DEFAULT_RETRY_DELAY = 1000

/**
 * Gets default headers for API requests
 * @returns Headers object with content type and accept
 */
export function getDefaultHeaders(): Record<string, string> {
  return {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
}

/**
 * Gets authentication headers
 * @returns Headers object with authorization token if available
 */
export function getAuthHeaders(): Record<string, string> {
  const token = getAuthToken()
  return token ? { 'Authorization': `Bearer ${token}` } : {}
}

/**
 * Handles API response parsing and error checking
 * @param response - Fetch response object
 * @returns Parsed response data
 * @throws Error if response is not OK
 */
export async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`)
  }

  const data = await response.json()
  return (data as Record<string, unknown>).data as T ?? data as T
}

/**
 * Makes an API request with retry logic and endpoint failover
 * @param endpoint - API endpoint path
 * @param options - Request configuration options
 * @returns Response data or fallback data in development
 */
export async function request<T = unknown>(
  endpoint: string,
  options: RequestConfig = {}
): Promise<T> {
  const { method = 'GET', headers = {}, body = null, retry = DEFAULT_RETRY_COUNT } = options
  const retryDelay = options.retryDelay ?? DEFAULT_RETRY_DELAY

  const endpoints = getAPIEndpoints()
  let lastError: Error | null = null

  // Try each endpoint with failover
  for (let i = 0; i < endpoints.length; i++) {
    const baseURL = endpoints[i]
    const url = `${baseURL}${endpoint}`

    const authHeaders = getAuthHeaders()
    const csrfToken = getCSRFToken()

    const config: RequestInit = {
      method,
      headers: {
        ...getDefaultHeaders(),
        ...SECURITY_HEADERS,
        ...authHeaders,
        ...(csrfToken ? { 'X-CSRF-Token': csrfToken } : {}),
        ...headers
      },
      body: body as BodyInit | null,
      credentials: 'include'
    }

    try {
      const response = await fetch(url, config)

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()

      // Update current endpoint index on success
      setCurrentEndpointIndex(i)

      return (data as Record<string, unknown>).data as T ?? data as T
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error))

      // Retry logic for transient failures
      if (retry > 0 && isTransientError(lastError)) {
        await delay(retryDelay)
        continue
      }
    }
  }

  // All endpoints failed - fallback to development data
  if (isDevelopment()) {
    return getFallbackData(endpoint) as T
  }

  throw lastError || new Error('All API endpoints failed')
}

/**
 * Checks if an error is transient (network-related)
 * @param error - Error to check
 * @returns true if error is transient
 */
function isTransientError(error: Error): boolean {
  const message = error.message.toLowerCase()
  return (
    message.includes('network') ||
    message.includes('fetch') ||
    message.includes('timeout') ||
    message.includes('econnrefused') ||
    message.includes('ECONNREFUSED')
  )
}

/**
 * Delay utility for retry logic
 * @param ms - Milliseconds to delay
 */
function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * HTTP method shortcuts
 */
export const api = {
  get: <T = unknown>(endpoint: string, options?: RequestConfig) =>
    request<T>(endpoint, { ...options, method: 'GET' }),

  post: <T = unknown>(endpoint: string, body: unknown, options?: RequestConfig) =>
    request<T>(endpoint, { ...options, method: 'POST', body: JSON.stringify(body) }),

  put: <T = unknown>(endpoint: string, body: unknown, options?: RequestConfig) =>
    request<T>(endpoint, { ...options, method: 'PUT', body: JSON.stringify(body) }),

  patch: <T = unknown>(endpoint: string, body: unknown, options?: RequestConfig) =>
    request<T>(endpoint, { ...options, method: 'PATCH', body: JSON.stringify(body) }),

  delete: <T = unknown>(endpoint: string, options?: RequestConfig) =>
    request<T>(endpoint, { ...options, method: 'DELETE' })
}

export const RequestConfig = {
  getDefaultHeaders,
  getAuthHeaders,
  handleResponse,
  request,
  api
}

export default RequestConfig