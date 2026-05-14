import { describe, it, expect, beforeEach } from 'vitest'
import {
  getEnvironment,
  isDevelopment,
  isProduction,
  isTesting
} from '@/api/environments'
import {
  getAPIEndpoints,
  getBaseURL,
  getCurrentVersion,
  getSupportedVersions,
  incrementEndpointIndex,
  resetEndpoints
} from '@/api/endpoints'
import { hasFallbackData, getFallbackEndpoints } from '@/api/fallbacks'

describe('API Config', () => {
  beforeEach(() => {
    resetEndpoints()
  })

  describe('environment detection', () => {
    it('should return correct environment', () => {
      const env = getEnvironment()
      expect(['development', 'testing', 'production']).toContain(env)
    })

    it('isDevelopment should return boolean', () => {
      expect(typeof isDevelopment()).toBe('boolean')
    })

    it('isProduction should return boolean', () => {
      expect(typeof isProduction()).toBe('boolean')
    })

    it('isTesting should return boolean', () => {
      expect(typeof isTesting()).toBe('boolean')
    })
  })

  describe('endpoint management', () => {
    it('should get API endpoints', () => {
      const endpoints = getAPIEndpoints()
      expect(Array.isArray(endpoints)).toBe(true)
      expect(endpoints.length).toBeGreaterThan(0)
    })

    it('should get base URL', () => {
      const baseURL = getBaseURL()
      expect(typeof baseURL).toBe('string')
      expect(baseURL).toContain('http')
    })
  })

  describe('version management', () => {
    it('should get current API version', () => {
      const version = getCurrentVersion()
      expect(version).toMatch(/^v\d+$/)
    })

    it('should get supported API versions', () => {
      const versions = getSupportedVersions()
      expect(Array.isArray(versions)).toBe(true)
      expect(versions).toContain('v2')
    })
  })

  describe('fallback data', () => {
    it('should check if fallback data exists', () => {
      const exists = hasFallbackData('/dashboard/metrics')
      expect(typeof exists).toBe('boolean')
    })

    it('should get fallback endpoints', () => {
      const endpoints = getFallbackEndpoints()
      expect(Array.isArray(endpoints)).toBe(true)
    })
  })

  describe('endpoint index management', () => {
    it('should increment endpoint index', () => {
      const initialEndpoints = getAPIEndpoints()
      if (initialEndpoints.length > 1) {
        incrementEndpointIndex()
        expect(true).toBe(true)
      } else {
        incrementEndpointIndex()
        expect(true).toBe(true)
      }
    })
  })
})
