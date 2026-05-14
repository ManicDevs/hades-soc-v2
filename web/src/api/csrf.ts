/**
 * CSRF token management
 * Handles CSRF token generation, retrieval, and refresh
 */

// In-memory CSRF token storage
let csrfToken: string | null = null

/**
 * Generates a CSRF token
 * @returns Generated token string
 */
function generateCSRFToken(): string {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * Gets the current CSRF token
 * Generates a new one if not exists
 * @returns Current CSRF token
 */
export function getCSRFToken(): string | null {
  if (csrfToken === null) {
    csrfToken = generateCSRFToken()
  }
  return csrfToken
}

/**
 * Sets a CSRF token
 * @param token - Token to set
 */
export function setCSRFToken(token: string): void {
  csrfToken = token
}

/**
 * Clears the CSRF token
 */
export function clearCSRFToken(): void {
  csrfToken = null
}

/**
 * Refreshes the CSRF token
 * Generates a new token and returns it
 * @returns New CSRF token
 */
export function refreshCSRFToken(): string {
  csrfToken = generateCSRFToken()
  return csrfToken
}

/**
 * Checks if a CSRF token is valid
 * @param token - Token to validate
 * @returns true if token is valid format
 */
export function isValidCSRFToken(token: string | null): boolean {
  if (!token || typeof token !== 'string') {
    return false
  }
  // Token should be at least 32 hex characters (64 chars for 32 bytes)
  return token.length >= 64 && /^[0-9a-f]+$/.test(token)
}

/**
 * Gets the CSRF header name
 * @returns Header name for CSRF token
 */
export function getCSRFHeaderName(): string {
  return 'X-CSRF-Token'
}

/**
 * Gets the CSRF token for header inclusion
 * @returns Token formatted for header
 */
export function getCSRFHeaderValue(): string | null {
  const token = getCSRFToken()
  return token || null
}

export const CSRFConfig = {
  getCSRFToken,
  setCSRFToken,
  clearCSRFToken,
  refreshCSRFToken,
  isValidCSRFToken,
  getCSRFHeaderName,
  getCSRFHeaderValue
}

export default CSRFConfig