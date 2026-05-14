import { describe, it, expect } from 'vitest'
import {
  getDefaultHeaders,
  getAuthHeaders
} from '@/api/request'

describe('Request API', () => {
  describe('getDefaultHeaders', () => {
    it('should return correct default headers', () => {
      const headers = getDefaultHeaders()

      expect(headers).toEqual({
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      })
    })
  })

  describe('getAuthHeaders', () => {
    it('should return object type', () => {
      const headers = getAuthHeaders()
      expect(typeof headers).toBe('object')
    })
  })
})
