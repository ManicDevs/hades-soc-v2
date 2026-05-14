import { useState, useCallback, KeyboardEvent as ReactKeyboardEvent } from 'react'

/**
 * useKeyboardNav - Hook for managing keyboard navigation in lists, menus, grids
 * Supports: Arrow keys, Enter/Space activation, Tab management
 */

interface UseKeyboardNavOptions<T> {
  items: T[]
  initialIndex?: number
  onSelect?: (item: T, index: number) => void
  onEnter?: (item: T, index: number) => void
  orientation?: 'vertical' | 'horizontal' | 'both'
  loop?: boolean
  disabled?: boolean
}

interface KeyboardNavState {
  currentIndex: number
  focusedIndex: number
}

/**
 * useKeyboardNav - Main hook for keyboard navigation
 */
export function useKeyboardNav<T>(options: UseKeyboardNavOptions<T>) {
  const { items, initialIndex = 0, onSelect, onEnter, orientation = 'vertical', loop = true, disabled = false } = options

  const [state, setState] = useState<KeyboardNavState>({
    currentIndex: initialIndex,
    focusedIndex: -1,
  })

  const _getNextIndex = useCallback((currentIndex: number, direction: 'next' | 'prev'): number => {
    const totalItems = items.length
    if (totalItems === 0) return -1

    if (loop) {
      if (direction === 'next') {
        return (currentIndex + 1) % totalItems
      }
      return (currentIndex - 1 + totalItems) % totalItems
    }

    if (direction === 'next') {
      return Math.min(currentIndex + 1, totalItems - 1)
    }
    return Math.max(currentIndex - 1, 0)
  }, [items.length, loop])

  const getAdjacentIndex = useCallback((currentIndex: number, direction: 'up' | 'down' | 'left' | 'right'): number => {
    const totalItems = items.length
    if (totalItems === 0) return -1

    if (orientation === 'horizontal' || orientation === 'both') {
      if (direction === 'left') {
        return loop
          ? (currentIndex - 1 + totalItems) % totalItems
          : Math.max(currentIndex - 1, 0)
      }
      if (direction === 'right') {
        return loop
          ? (currentIndex + 1) % totalItems
          : Math.min(currentIndex + 1, totalItems - 1)
      }
    }

    if (orientation === 'vertical' || orientation === 'both') {
      if (direction === 'up') {
        return loop
          ? (currentIndex - 1 + totalItems) % totalItems
          : Math.max(currentIndex - 1, 0)
      }
      if (direction === 'down') {
        return loop
          ? (currentIndex + 1) % totalItems
          : Math.min(currentIndex + 1, totalItems - 1)
      }
    }

    return currentIndex
  }, [items.length, loop, orientation])

  const handleKeyDown = useCallback((e: ReactKeyboardEvent<HTMLElement>) => {
    if (disabled) return

    const { key } = e

    // Tab management - allow natural tab flow
    if (key === 'Tab') {
      return
    }

    // Arrow key navigation
    if (key === 'ArrowUp' || key === 'ArrowDown' || key === 'ArrowLeft' || key === 'ArrowRight') {
      e.preventDefault()

      const direction = key.replace('Arrow', '').toLowerCase() as 'up' | 'down' | 'left' | 'right'
      const nextIndex = getAdjacentIndex(state.currentIndex, direction)

      setState(prev => ({
        ...prev,
        currentIndex: nextIndex,
        focusedIndex: nextIndex,
      }))

      // Call onSelect if provided
      if (onSelect && items[nextIndex]) {
        onSelect(items[nextIndex], nextIndex)
      }
    }

    // Home key - jump to first item
    if (key === 'Home') {
      e.preventDefault()
      setState(prev => ({
        ...prev,
        currentIndex: 0,
        focusedIndex: 0,
      }))
      if (onSelect && items[0]) {
        onSelect(items[0], 0)
      }
    }

    // End key - jump to last item
    if (key === 'End') {
      e.preventDefault()
      const lastIndex = items.length - 1
      setState(prev => ({
        ...prev,
        currentIndex: lastIndex,
        focusedIndex: lastIndex,
      }))
      if (onSelect && items[lastIndex]) {
        onSelect(items[lastIndex], lastIndex)
      }
    }

    // Enter/Space activation
    if ((key === 'Enter' || key === ' ') && state.currentIndex >= 0) {
      e.preventDefault()
      if (onEnter && items[state.currentIndex] !== undefined) {
        const item = items[state.currentIndex]
        if (item !== undefined) {
          onEnter(item, state.currentIndex)
        }
      }
    }
  }, [disabled, state.currentIndex, items, getAdjacentIndex, onSelect, onEnter])

  const setCurrentIndex = useCallback((index: number) => {
    if (index >= 0 && index < items.length) {
      setState(prev => ({
        ...prev,
        currentIndex: index,
        focusedIndex: index,
      }))
    }
  }, [items.length])

  const reset = useCallback(() => {
    setState({
      currentIndex: initialIndex,
      focusedIndex: -1,
    })
  }, [initialIndex])

  return {
    currentIndex: state.currentIndex,
    focusedIndex: state.focusedIndex,
    handleKeyDown,
    setCurrentIndex,
    reset,
  }
}

/**
 * useListNavigation - Simplified hook for list navigation
 */
interface ListItem {
  id: string | number
  [key: string]: unknown
}

interface UseListNavigationOptions {
  items: ListItem[]
  onSelect?: (item: ListItem) => void
  onActivate?: (item: ListItem) => void
  loop?: boolean
}

export function useListNavigation(options: UseListNavigationOptions) {
  const { items, onSelect, onActivate, loop = true } = options

  const [activeIndex, setActiveIndex] = useState(-1)

  const handleKeyDown = useCallback((e: ReactKeyboardEvent<HTMLElement>) => {
    const { key } = e
    const itemCount = items.length

    if (itemCount === 0) return

    switch (key) {
      case 'ArrowDown': {
        e.preventDefault()
        setActiveIndex(prev => {
          const next = loop ? (prev + 1) % itemCount : Math.min(prev + 1, itemCount - 1)
          const newItem = items[next]
          if (newItem && onSelect) {
            onSelect(newItem)
          }
          return next
        })
        break
      }
      case 'ArrowUp': {
        e.preventDefault()
        setActiveIndex(prev => {
          const next = loop ? (prev - 1 + itemCount) % itemCount : Math.max(prev - 1, 0)
          const newItem = items[next]
          if (newItem && onSelect) {
            onSelect(newItem)
          }
          return next
        })
        break
      }
      case 'Home': {
        e.preventDefault()
        setActiveIndex(0)
        if (items[0] && onSelect) {
          onSelect(items[0])
        }
        break
      }
      case 'End': {
        e.preventDefault()
        const lastIndex = itemCount - 1
        setActiveIndex(lastIndex)
        if (items[lastIndex] && onSelect) {
          onSelect(items[lastIndex])
        }
        break
      }
      case 'Enter':
      case ' ': {
        e.preventDefault()
        if (activeIndex >= 0 && items[activeIndex] && onActivate) {
          onActivate(items[activeIndex])
        }
        break
      }
    }
  }, [items, activeIndex, onSelect, onActivate, loop])

  const handleItemClick = useCallback((index: number) => {
    setActiveIndex(index)
    if (items[index] && onSelect) {
      onSelect(items[index])
    }
  }, [items, onSelect])

  return {
    activeIndex,
    handleKeyDown,
    handleItemClick,
    setActiveIndex,
  }
}

/**
 * useMenuKeyboardNav - Specialized hook for menu navigation
 */
interface MenuItem {
  label: string
  disabled?: boolean
  onClick?: () => void
  children?: MenuItem[]
}

interface UseMenuKeyboardNavOptions {
  items: MenuItem[]
  onSelect?: (item: MenuItem, index: number) => void
  onEscape?: () => void
}

export function useMenuKeyboardNav(options: UseMenuKeyboardNavOptions) {
  const { items, onSelect, onEscape } = options
  const [focusedIndex, setFocusedIndex] = useState(-1)
  const [expandedIndex, setExpandedIndex] = useState<number | null>(null)

  const handleKeyDown = useCallback((e: ReactKeyboardEvent<HTMLElement>) => {
    const { key } = e

    switch (key) {
      case 'ArrowDown': {
        e.preventDefault()
        setFocusedIndex(prev => {
          let next = prev + 1
          // Skip disabled items
          while (next < items.length && items[next]?.disabled) {
            next = (next + 1) % items.length
          }
          return next >= items.length ? prev : next
        })
        break
      }
      case 'ArrowUp': {
        e.preventDefault()
        setFocusedIndex(prev => {
          let next = prev - 1
          // Skip disabled items
          while (next >= 0 && items[next]?.disabled) {
            next = next - 1
          }
          return next < 0 ? prev : next
        })
        break
      }
      case 'Enter':
      case ' ': {
        e.preventDefault()
        if (focusedIndex >= 0 && items[focusedIndex] && !items[focusedIndex]?.disabled) {
          const item = items[focusedIndex]
          if (item.children) {
            setExpandedIndex(prev => prev === focusedIndex ? null : focusedIndex)
          } else if (item.onClick) {
            item.onClick()
          } else if (onSelect) {
            onSelect(item, focusedIndex)
          }
        }
        break
      }
      case 'Escape': {
        e.preventDefault()
        if (expandedIndex !== null) {
          setExpandedIndex(null)
        } else if (onEscape) {
          onEscape()
        }
        break
      }
      case 'Tab': {
        if (onEscape) {
          onEscape()
        }
        break
      }
    }
  }, [items, focusedIndex, expandedIndex, onSelect, onEscape])

  return {
    focusedIndex,
    expandedIndex,
    setFocusedIndex,
    setExpandedIndex,
    handleKeyDown,
  }
}

/**
 * useGridNavigation - Hook for grid-based keyboard navigation
 */
interface UseGridNavigationOptions {
  columns: number
  rowCount: number
  onSelect?: (row: number, col: number) => void
  onActivate?: (row: number, col: number) => void
}

export function useGridNavigation(options: UseGridNavigationOptions) {
  const { columns, rowCount, onSelect, onActivate } = options
  const [focusedCell, setFocusedCell] = useState<{ row: number; col: number }>({ row: -1, col: -1 })

  const handleKeyDown = useCallback((e: ReactKeyboardEvent<HTMLElement>) => {
    const { key, shiftKey } = e
    const totalRows = rowCount

    switch (key) {
      case 'ArrowRight': {
        e.preventDefault()
        setFocusedCell(prev => {
          let nextCol = prev.col + 1
          let nextRow = prev.row

          if (nextCol >= columns) {
            if (!shiftKey && prev.row < totalRows - 1) {
              nextCol = 0
              nextRow = prev.row + 1
            } else {
              nextCol = columns - 1
            }
          }

          if (onSelect && nextRow >= 0 && nextCol >= 0) {
            onSelect(nextRow, nextCol)
          }

          return { row: nextRow, col: nextCol }
        })
        break
      }
      case 'ArrowLeft': {
        e.preventDefault()
        setFocusedCell(prev => {
          let nextCol = prev.col - 1
          let nextRow = prev.row

          if (nextCol < 0) {
            if (!shiftKey && prev.row > 0) {
              nextCol = columns - 1
              nextRow = prev.row - 1
            } else {
              nextCol = 0
            }
          }

          if (onSelect && nextRow >= 0 && nextCol >= 0) {
            onSelect(nextRow, nextCol)
          }

          return { row: nextRow, col: nextCol }
        })
        break
      }
      case 'ArrowDown': {
        e.preventDefault()
        setFocusedCell(prev => {
          const nextRow = Math.min(prev.row + 1, totalRows - 1)
          if (onSelect && nextRow >= 0) {
            onSelect(nextRow, prev.col)
          }
          return { row: nextRow, col: prev.col }
        })
        break
      }
      case 'ArrowUp': {
        e.preventDefault()
        setFocusedCell(prev => {
          const nextRow = Math.max(prev.row - 1, 0)
          if (onSelect && nextRow >= 0) {
            onSelect(nextRow, prev.col)
          }
          return { row: nextRow, col: prev.col }
        })
        break
      }
      case 'Enter':
      case ' ': {
        e.preventDefault()
        if (focusedCell.row >= 0 && focusedCell.col >= 0 && onActivate) {
          onActivate(focusedCell.row, focusedCell.col)
        }
        break
      }
      case 'Home': {
        e.preventDefault()
        setFocusedCell({ row: 0, col: 0 })
        if (onSelect) {
          onSelect(0, 0)
        }
        break
      }
      case 'End': {
        e.preventDefault()
        const lastRow = totalRows - 1
        const lastCol = columns - 1
        setFocusedCell({ row: lastRow, col: lastCol })
        if (onSelect) {
          onSelect(lastRow, lastCol)
        }
        break
      }
    }
  }, [columns, rowCount, focusedCell, onSelect, onActivate])

  return {
    focusedCell,
    handleKeyDown,
    setFocusedCell,
  }
}

export default useKeyboardNav