import { describe, expect, it } from 'vitest'
import { clearAuthToken, getAuthToken, setAuthToken } from './authToken'

describe('authToken store', () => {
  it('stores and clears token in memory', () => {
    clearAuthToken()
    expect(getAuthToken()).toBeNull()

    setAuthToken('abc123')
    expect(getAuthToken()).toBe('abc123')

    clearAuthToken()
    expect(getAuthToken()).toBeNull()
  })
})
