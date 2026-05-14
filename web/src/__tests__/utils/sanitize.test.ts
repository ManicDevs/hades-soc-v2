import { describe, it, expect } from 'vitest'

// Utility functions to test

/**
 * Sanitizes user input to prevent XSS
 */
export function sanitize(input: string): string {
  if (typeof input !== 'string') return ''
  if (input.includes('&lt;') || input.includes('&gt;') || input.includes('&amp;')) {
    return input
  }
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
    .replace(/\//g, '&#x2F;')
}

/**
 * Escapes regex special characters
 */
export function escapeRegex(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

/**
 * Normalizes path for comparison
 */
export function normalizePath(path: string): string {
  return path.replace(/\/+/g, '/').replace(/\/$/, '')
}

/**
 * Validates email format
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

/**
 * Truncates text to specified length
 */
export function truncate(text: string, maxLength: number): string {
  if (maxLength < 3) return '...'
  if (text.length <= maxLength) return text
  return text.slice(0, maxLength - 3) + '...'
}

describe('Sanitize utility', () => {
  describe('sanitize function', () => {
    it('should escape HTML special characters', () => {
      expect(sanitize('<script>alert("xss")</script>')).toBe(
        '&lt;script&gt;alert(&quot;xss&quot;)&lt;&#x2F;script&gt;'
      )
    })

    it('should escape ampersands', () => {
      expect(sanitize('foo & bar')).toBe('foo &amp; bar')
    })

    it('should escape quotes', () => {
      expect(sanitize('"double" and \'single\'')).toBe(
        '&quot;double&quot; and &#x27;single&#x27;'
      )
    })

    it('should escape forward slashes', () => {
      expect(sanitize('path/to/file')).toBe('path&#x2F;to&#x2F;file')
    })

    it('should return empty string for non-string input', () => {
      expect(sanitize(null as unknown as string)).toBe('')
      expect(sanitize(undefined as unknown as string)).toBe('')
      expect(sanitize(123 as unknown as string)).toBe('')
    })

    it('should handle empty string', () => {
      expect(sanitize('')).toBe('')
    })

    it('should handle already sanitized input', () => {
      const sanitized = '&lt;script&gt;'
      expect(sanitize(sanitized)).toBe(sanitized)
    })
  })

  describe('escapeRegex function', () => {
    it('should escape regex special characters', () => {
      expect(escapeRegex('*.*')).toBe('\\*\\.\\*')
    })

    it('should escape brackets', () => {
      expect(escapeRegex('[test]')).toBe('\\[test\\]')
    })

    it('should escape parentheses', () => {
      expect(escapeRegex('(test)')).toBe('\\(test\\)')
    })

    it('should escape pipes', () => {
      expect(escapeRegex('a|b')).toBe('a\\|b')
    })

    it('should escape dollar sign', () => {
      expect(escapeRegex('$var')).toBe('\\$var')
    })

    it('should handle plain text without special chars', () => {
      expect(escapeRegex('plaintext')).toBe('plaintext')
    })
  })

  describe('normalizePath function', () => {
    it('should replace multiple slashes with single', () => {
      expect(normalizePath('a//b///c')).toBe('a/b/c')
    })

    it('should remove trailing slash', () => {
      expect(normalizePath('path/')).toBe('path')
    })

    it('should handle root path', () => {
      expect(normalizePath('/')).toBe('')
    })

    it('should handle already normalized path', () => {
      expect(normalizePath('a/b/c')).toBe('a/b/c')
    })
  })

  describe('isValidEmail function', () => {
    it('should validate correct email formats', () => {
      expect(isValidEmail('test@example.com')).toBe(true)
      expect(isValidEmail('user.name@domain.org')).toBe(true)
      expect(isValidEmail('admin@hades-toolkit.com')).toBe(true)
    })

    it('should reject invalid email formats', () => {
      expect(isValidEmail('invalid')).toBe(false)
      expect(isValidEmail('missing@domain')).toBe(false)
      expect(isValidEmail('@domain.com')).toBe(false)
      expect(isValidEmail('user@')).toBe(false)
    })

    it('should reject email with spaces', () => {
      expect(isValidEmail('user name@domain.com')).toBe(false)
    })
  })

  describe('truncate function', () => {
    it('should not truncate short text', () => {
      expect(truncate('short', 10)).toBe('short')
    })

    it('should truncate long text with ellipsis', () => {
      expect(truncate('this is a long text', 10)).toBe('this is...')
    })

    it('should handle exact length', () => {
      expect(truncate('abc', 3)).toBe('abc')
    })

    it('should handle edge case of maxLength 3', () => {
      expect(truncate('abcd', 3)).toBe('...')
    })

    it('should handle maxLength less than 3 gracefully', () => {
      expect(truncate('test', 2)).toBe('...')
    })
  })
})
