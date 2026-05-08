import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setAuthToken } from '../lib/authToken'
import { API_CONFIG } from './config'

describe('API_CONFIG', () => {
  beforeEach(() => {
    API_CONFIG.currentEndpointIndex = 0
    vi.unstubAllEnvs()
    setAuthToken(null)
  })

  it('uses env API base when provided', () => {
    vi.stubEnv('VITE_API_BASE_URL', 'http://api.test.local/api/v2')
    const endpoints = API_CONFIG.getAPIEndpoints()
    expect(endpoints[0]).toBe('http://api.test.local/api/v2')
  })

  it('adds bearer auth header from in-memory token store', () => {
    setAuthToken('token-123')
    expect(API_CONFIG.getAuthHeaders()).toEqual({
      Authorization: 'Bearer token-123'
    })
  })
})
