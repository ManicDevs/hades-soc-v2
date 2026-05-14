/**
 * React hook for XSS sanitization using DOMPurify
 * Provides sanitization functions with hook-based API
 */

import { useCallback, useMemo } from 'react'
import {
  sanitizeHTML,
  sanitizeInput,
  sanitizeObject,
  stripHTML,
  sanitizeURL,
  createSanitizer
} from '../utils/sanitize'
import type { Config } from 'dompurify'

/**
 * Hook providing sanitization utilities for React components
 * @returns Object containing sanitization functions
 */
export function useSanitize() {
  /**
   * Sanitizes HTML string with safe defaults
   */
  const sanitizeHtml = useCallback((html: string): string => {
    return sanitizeHTML(html)
  }, [])

  /**
   * Sanitizes HTML with custom configuration
   */
  const sanitizeHtmlWithConfig = useCallback(
    (html: string, config: Config): string => {
      return sanitizeHTML(html, config)
    },
    []
  )

  /**
   * Sanitizes plain text input
   */
  const sanitize = useCallback((input: string): string => {
    return sanitizeInput(input)
  }, [])

  /**
   * Recursively sanitizes object properties
   */
  const sanitizeObj = useCallback(
    <T extends Record<string, unknown>>(obj: T, keys?: string[]): T => {
      return sanitizeObject(obj, keys)
    },
    []
  )

  /**
   * Strips all HTML tags from a string
   */
  const strip = useCallback((html: string): string => {
    return stripHTML(html)
  }, [])

  /**
   * Sanitizes a URL
   */
  const sanitizeUrl = useCallback((url: string): string => {
    return sanitizeURL(url)
  }, [])

  /**
   * Creates a custom sanitizer with given config
   */
  const createCustomSanitizer = useCallback((config: Config) => {
    return createSanitizer(config)
  }, [])

  /**
   * Async sanitization for large content
   */
  const sanitizeAsync = useCallback(async (html: string): Promise<string> => {
    // For large content, use setTimeout to prevent blocking
    return new Promise(resolve => {
      setTimeout(() => {
        resolve(sanitizeHTML(html))
      }, 0)
    })
  }, [])

  /**
   * Batch sanitization for arrays
   */
  const sanitizeArray = useCallback(
    (items: string[]): string[] => {
      return items.map(item => sanitizeInput(item))
    },
    [sanitizeInput]
  )

  return useMemo(
    () => ({
      sanitizeHtml,
      sanitizeHtmlWithConfig,
      sanitize,
      sanitizeObj,
      strip,
      sanitizeUrl,
      createCustomSanitizer,
      sanitizeAsync,
      sanitizeArray
    }),
    [
      sanitizeHtml,
      sanitizeHtmlWithConfig,
      sanitize,
      sanitizeObj,
      strip,
      sanitizeUrl,
      createCustomSanitizer,
      sanitizeAsync,
      sanitizeArray
    ]
  )
}

export default useSanitize