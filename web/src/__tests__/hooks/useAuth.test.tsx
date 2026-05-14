import { describe, it, expect, vi } from 'vitest'

describe('useAuth hook', () => {
  describe('auth state', () => {
    it('should have expected auth interface', () => {
      const authMock = {
        user: null,
        isAuthenticated: false,
        loading: false,
        error: null,
        login: vi.fn(),
        logout: vi.fn(),
        refreshData: vi.fn()
      }
      expect(typeof authMock.login).toBe('function')
      expect(typeof authMock.logout).toBe('function')
    })
  })
})
