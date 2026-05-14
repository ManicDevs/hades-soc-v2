import { useState, useEffect, useRef, useCallback, useMemo, RefObject } from 'react'

export interface VirtualListOptions<T> {
  items: T[]
  itemHeight: number
  overscan?: number
  containerRef?: RefObject<HTMLElement>
}

export interface VirtualListResult<T> {
  virtualItems: { item: T; index: number; style: React.CSSProperties }[]
  totalHeight: number
  scrollToIndex: (index: number) => void
  containerProps: {
    onScroll: () => void
    ref: RefObject<HTMLElement | null>
    style: React.CSSProperties
  }
}

export function useVirtualList<T>({
  items,
  itemHeight,
  overscan = 3,
  containerRef: externalContainerRef
}: VirtualListOptions<T>): VirtualListResult<T> {
  const internalContainerRef = useRef<HTMLElement>(null)
  const containerRef = externalContainerRef || internalContainerRef
  const [scrollTop, setScrollTop] = useState(0)
  const [containerHeight, setContainerHeight] = useState(0)

  const totalHeight = items.length * itemHeight

  const handleScroll = useCallback(() => {
    if (containerRef.current) {
      setScrollTop(containerRef.current.scrollTop)
    }
  }, [containerRef])

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const updateHeight = () => {
      setContainerHeight(container.clientHeight)
    }

    updateHeight()
    container.addEventListener('scroll', handleScroll, { passive: true })
    window.addEventListener('resize', updateHeight)

    return () => {
      container.removeEventListener('scroll', handleScroll)
      window.removeEventListener('resize', updateHeight)
    }
  }, [containerRef, handleScroll])

  const virtualItems = useMemo(() => {
    const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan)
    const endIndex = Math.min(
      items.length - 1,
      Math.ceil((scrollTop + containerHeight) / itemHeight) + overscan
    )

    const result: { item: T; index: number; style: React.CSSProperties }[] = []
    for (let i = startIndex; i <= endIndex; i++) {
      const item = items[i]
      if (item !== undefined) {
        result.push({
          item,
          index: i,
          style: {
            position: 'absolute',
            top: i * itemHeight,
            left: 0,
            right: 0,
            height: itemHeight
          }
        })
      }
    }
    return result
  }, [items, itemHeight, scrollTop, containerHeight, overscan])

  const scrollToIndex = useCallback(
    (index: number) => {
      if (containerRef.current) {
        containerRef.current.scrollTop = index * itemHeight
      }
    },
    [containerRef, itemHeight]
  )

  const containerProps = useMemo(
    () => ({
      onScroll: handleScroll,
      ref: containerRef,
      style: {
        overflow: 'auto',
        position: 'relative' as const
      } as React.CSSProperties
    }),
    [handleScroll, containerRef]
  )

  return {
    virtualItems,
    totalHeight,
    scrollToIndex,
    containerProps
  }
}

export default useVirtualList
