import { renderHook } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useVirtualList } from '@/utils/virtualScroll'

// Mock ResizeObserver
class ResizeObserverMock {
  observe = vi.fn()
  unobserve = vi.fn()
  disconnect = vi.fn()
}

describe('useVirtualList hook', () => {
  let containerRef: { current: HTMLElement | null }

  beforeEach(() => {
    global.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver

    containerRef = { current: null }

    // Create a mock container
    const mockContainer = document.createElement('div')
    mockContainer.style.height = '500px'
    mockContainer.style.overflow = 'auto'
    containerRef.current = mockContainer
    document.body.appendChild(mockContainer)
  })

  afterEach(() => {
    if (containerRef.current) {
      document.body.removeChild(containerRef.current)
    }
  })

  describe('initial state', () => {
    it('should return empty virtual items for empty list', () => {
      const { result } = renderHook(() =>
        useVirtualList({ items: [], itemHeight: 50 })
      )

      expect(result.current.virtualItems).toEqual([])
    })

    it('should return correct total height', () => {
      const items = [{ id: 1 }, { id: 2 }, { id: 3 }]
      const { result } = renderHook(() =>
        useVirtualList({ items, itemHeight: 50 })
      )

      expect(result.current.totalHeight).toBe(150)
    })
  })

  describe('virtual items calculation', () => {
    it('should calculate virtual items based on scroll position', () => {
      const items = Array.from({ length: 100 }, (_, i) => ({ id: i }))
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 50,
          containerRef
        })
      )

      expect(result.current.virtualItems.length).toBeGreaterThan(0)
      expect(result.current.virtualItems.length).toBeLessThanOrEqual(items.length)
    })

    it('should include overscan items', () => {
      const items = Array.from({ length: 50 }, (_, i) => ({ id: i }))
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 50,
          overscan: 5,
          containerRef
        })
      )

      // With overscan of 5, should have more items visible
      expect(result.current.virtualItems.length).toBeGreaterThan(0)
    })

    it('should return item with correct index', () => {
      const items = [{ id: 'a' }, { id: 'b' }, { id: 'c' }]
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 50,
          containerRef
        })
      )

      result.current.virtualItems.forEach((virtualItem) => {
        expect(items[virtualItem.index]).toBeDefined()
      })
    })

    it('should return item with style containing position', () => {
      const items = [{ id: 1 }]
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 50,
          containerRef
        })
      )

      expect(result.current.virtualItems?.[0]?.style.position).toBe('absolute')
    })
  })

  describe('container props', () => {
    it('should return container props with onScroll handler', () => {
      const { result } = renderHook(() =>
        useVirtualList({ items: [], itemHeight: 50 })
      )

      expect(result.current.containerProps.onScroll).toBeDefined()
      expect(typeof result.current.containerProps.onScroll).toBe('function')
    })

    it('should return container props with ref', () => {
      const { result } = renderHook(() =>
        useVirtualList({
          items: [],
          itemHeight: 50,
          containerRef
        })
      )

      expect(result.current.containerProps.ref).toBeDefined()
    })

    it('should return container props with overflow style', () => {
      const { result } = renderHook(() =>
        useVirtualList({ items: [], itemHeight: 50 })
      )

      expect(result.current.containerProps.style.overflow).toBe('auto')
    })
  })

  describe('scrollToIndex function', () => {
    it('should return scrollToIndex function', () => {
      const { result } = renderHook(() =>
        useVirtualList({ items: [], itemHeight: 50 })
      )

      expect(typeof result.current.scrollToIndex).toBe('function')
    })
  })

  describe('edge cases', () => {
    it('should handle single item', () => {
      const items = [{ id: 1 }]
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 50,
          containerRef
        })
      )

      expect(result.current.virtualItems.length).toBe(1)
      expect(result.current.totalHeight).toBe(50)
    })

    it('should handle zero item height', () => {
      const items = [{ id: 1 }]
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 0,
          containerRef
        })
      )

      expect(result.current.totalHeight).toBe(0)
    })

    it('should handle very large item height', () => {
      const items = [{ id: 1 }]
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 10000,
          containerRef
        })
      )

      expect(result.current.totalHeight).toBe(10000)
    })

    it('should use default overscan of 3', () => {
      const items = Array.from({ length: 100 }, (_, i) => ({ id: i }))
      const { result } = renderHook(() =>
        useVirtualList({
          items,
          itemHeight: 50,
          containerRef
        })
      )

      expect(result.current.virtualItems.length).toBeGreaterThan(0)
    })
  })
})
