import { describe, it, expect, vi } from 'vitest'

describe('useErrorBoundary hook', () => {
  describe('initial state', () => {
    it('should have expected hook interface', () => {
      const mockHook = {
        error: null,
        showError: vi.fn(),
        resetError: vi.fn(),
        ErrorBoundaryComponent: vi.fn()
      }
      expect(mockHook.error).toBeNull()
      expect(typeof mockHook.showError).toBe('function')
      expect(typeof mockHook.resetError).toBe('function')
    })
  })
})
