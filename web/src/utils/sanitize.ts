/**
 * XSS Sanitization utilities using DOMPurify
 * Provides various sanitization functions for different use cases
 */

import DOMPurify, { Config } from 'dompurify'

// Allowed HTML tags for safe rendering
const ALLOWED_TAGS = [
  'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
  'p', 'br', 'hr',
  'ul', 'ol', 'li',
  'a', 'img',
  'strong', 'em', 'b', 'i', 'u', 's', 'code', 'pre', 'blockquote',
  'table', 'thead', 'tbody', 'tr', 'th', 'td',
  'div', 'span', 'section', 'article', 'header', 'footer'
]

// Allowed attributes for HTML elements
const ALLOWED_ATTR = [
  'href', 'src', 'alt', 'title', 'class', 'id',
  'target', 'rel', 'width', 'height', 'style'
]

// Default DOMPurify configuration
const defaultConfig: Config = {
  ALLOWED_TAGS,
  ALLOWED_ATTR,
  ALLOW_DATA_ATTR: false,
  ADD_ATTR: ['target'],
  FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form', 'input'],
  FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'onfocus', 'onblur']
}

/**
 * Gets a configured DOMPurify instance
 * @returns Configured DOMPurify instance
 */
function getSanitizer(): typeof DOMPurify {
  if (typeof window !== 'undefined') {
    return DOMPurify
  }
  return DOMPurify
}

/**
 * Sanitizes an HTML string
 * @param html - The HTML string to sanitize
 * @param config - Optional custom configuration
 * @returns Sanitized HTML string
 */
export function sanitizeHTML(html: string, config?: Config): string {
  if (!html || typeof html !== 'string') {
    return ''
  }

  const sanitizer = getSanitizer()
  const mergedConfig = config ? { ...defaultConfig, ...config } : defaultConfig
  return sanitizer.sanitize(html, mergedConfig)
}

/**
 * Sanitizes plain text input for form fields
 * @param input - The input string to sanitize
 * @returns Sanitized string safe for text display
 */
export function sanitizeInput(input: string): string {
  if (!input || typeof input !== 'string') {
    return ''
  }

  const sanitizer = getSanitizer()
  return sanitizer.sanitize(input, {
    ALLOWED_TAGS: [],
    ALLOWED_ATTR: [],
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form'],
    FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'onfocus', 'onblur']
  }).trim()
}

/**
 * Recursively sanitizes all string properties in an object
 * @param obj - The object to sanitize
 * @param keys - Optional array of keys to sanitize (sanitizes all if not provided)
 * @returns Sanitized object
 */
export function sanitizeObject<T extends Record<string, unknown>>(
  obj: T,
  keys?: string[]
): T {
  if (!obj || typeof obj !== 'object') {
    return obj
  }

  const sanitized = { ...obj }

  for (const key of Object.keys(sanitized)) {
    const value = sanitized[key]

    // If keys array is provided, only sanitize those keys
    if (keys && !keys.includes(key)) {
      continue
    }

    if (typeof value === 'string') {
      (sanitized as Record<string, unknown>)[key] = sanitizeInput(value)
    } else if (Array.isArray(value)) {
      (sanitized as Record<string, unknown>)[key] = value.map(item =>
        typeof item === 'string' ? sanitizeInput(item) : item
      )
    } else if (value && typeof value === 'object' && !Array.isArray(value)) {
      (sanitized as Record<string, unknown>)[key] = sanitizeObject(
        value as Record<string, unknown>
      )
    }
  }

  return sanitized
}

/**
 * Strips all HTML tags from a string
 * @param html - The HTML string to strip
 * @returns Plain text
 */
export function stripHTML(html: string): string {
  if (!html || typeof html !== 'string') {
    return ''
  }

  const sanitizer = getSanitizer()
  return sanitizer.sanitize(html, {
    ALLOWED_TAGS: [],
    ALLOWED_ATTR: []
  }).trim()
}

/**
 * Sanitizes a URL
 * @param url - The URL to sanitize
 * @returns Sanitized URL or empty string if invalid
 */
export function sanitizeURL(url: string): string {
  if (!url || typeof url !== 'string') {
    return ''
  }

  try {
    const parsed = new URL(url)
    // Only allow http and https protocols
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
      return ''
    }
    return parsed.href
  } catch {
    // If URL parsing fails, check if it's a relative path
    if (url.startsWith('/') || url.startsWith('./') || url.startsWith('../')) {
      return url
    }
    return ''
  }
}

/**
 * Creates a sanitizer with custom configuration
 * @param config - Custom DOMPurify configuration
 * @returns Configured sanitizer function
 */
export function createSanitizer(config: Config): (html: string) => string {
  return (html: string) => sanitizeHTML(html, config)
}