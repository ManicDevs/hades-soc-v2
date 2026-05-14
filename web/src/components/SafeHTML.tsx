/**
 * SafeHTML Component
 * Renders sanitized HTML content safely in React
 */

import React, { useMemo } from 'react'
import { sanitizeHTML } from '../utils/sanitize'

interface SafeHTMLProps {
  /** The HTML content to render */
  html: string
  /** Optional CSS class name */
  className?: string
  /** Fallback content when HTML is empty or invalid */
  fallback?: React.ReactNode
}

/**
 * Safely renders HTML content with XSS protection
 * Uses DOMPurify internally to sanitize all HTML
 */
export const SafeHTML: React.FC<SafeHTMLProps> = ({
  html,
  className,
  fallback = null
}) => {
  const sanitizedContent = useMemo(() => {
    if (!html || typeof html !== 'string') {
      return null
    }
    return sanitizeHTML(html)
  }, [html])

  if (!sanitizedContent) {
    return <>{fallback}</>
  }

  return (
    <div
      className={className}
      dangerouslySetInnerHTML={{ __html: sanitizedContent }}
    />
  )
}

/**
 * Inline version of SafeHTML for use in text flows
 */
export const SafeSpan: React.FC<SafeHTMLProps> = ({
  html,
  className,
  fallback = null
}) => {
  const sanitizedContent = useMemo(() => {
    if (!html || typeof html !== 'string') {
      return null
    }
    return sanitizeHTML(html)
  }, [html])

  if (!sanitizedContent) {
    return <span className={className}>{fallback}</span>
  }

  return (
    <span
      className={className}
      dangerouslySetInnerHTML={{ __html: sanitizedContent }}
    />
  )
}

export default SafeHTML