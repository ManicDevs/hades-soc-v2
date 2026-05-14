import { useEffect, useRef, useCallback, useState } from 'react'

interface UseFocusManagementOptions {
  isOpen?: boolean
  closeOnEscape?: boolean
  onClose?: () => void
  returnFocusRef?: React.RefObject<HTMLElement>
}

interface FocusableElement {
  element: HTMLElement
  index: number
}

function getFocusableElements(container: HTMLElement): FocusableElement[] {
  const selectors = [
    'button:not([disabled])',
    'a[href]',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
  ]

  const elements = container.querySelectorAll<HTMLElement>(selectors.join(', '))
  return Array.from(elements).map((element, index) => ({ element, index }))
}

export function trapFocus(
  containerRef: React.RefObject<HTMLElement | null>,
  onEscape?: () => void
): () => void {
  const previousActiveElement = document.activeElement as HTMLElement | null

  const handleKeyDown = (e: KeyboardEvent): void => {
    if (e.key !== 'Tab') {
      if (e.key === 'Escape' && onEscape) {
        onEscape()
      }
      return
    }

    const container = containerRef.current
    if (!container) return

    const focusableElements = getFocusableElements(container)
    if (focusableElements.length === 0) return

    const firstElement = focusableElements[0]?.element
    const lastElement = focusableElements[focusableElements.length - 1]?.element

    if (e.shiftKey && document.activeElement === firstElement) {
      e.preventDefault()
      lastElement?.focus()
    } else if (!e.shiftKey && document.activeElement === lastElement) {
      e.preventDefault()
      firstElement?.focus()
    }
  }

  document.addEventListener('keydown', handleKeyDown)

  setTimeout(() => {
    const container = containerRef.current
    if (container) {
      const focusableElements = getFocusableElements(container)
      if (focusableElements.length > 0) {
        focusableElements[0]?.element?.focus()
      }
    }
  }, 0)

  return () => {
    document.removeEventListener('keydown', handleKeyDown)
    previousActiveElement?.focus()
  }
}

export function useFocusManagement(options: UseFocusManagementOptions = {}) {
  const { isOpen = false, closeOnEscape = true, onClose, returnFocusRef } = options

  const containerRef = useRef<HTMLDivElement | null>(null)
  const previousActiveElement = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (isOpen) {
      previousActiveElement.current = document.activeElement as HTMLElement
    }
  }, [isOpen])

  useEffect(() => {
    if (!isOpen || !containerRef.current) return

    const handleKeyDown = (e: KeyboardEvent): void => {
      if (closeOnEscape && e.key === 'Escape' && onClose) {
        onClose()
        return
      }

      if (e.key !== 'Tab') return

      const container = containerRef.current
      if (!container) return

      const focusableElements = getFocusableElements(container)
      if (focusableElements.length === 0) return

      const firstElement = focusableElements[0]?.element
      const lastElement = focusableElements[focusableElements.length - 1]?.element

      if (e.shiftKey && document.activeElement === firstElement) {
        e.preventDefault()
        lastElement?.focus()
      } else if (!e.shiftKey && document.activeElement === lastElement) {
        e.preventDefault()
        firstElement?.focus()
      }
    }

    setTimeout(() => {
      const container = containerRef.current
      if (container) {
        const focusableElements = getFocusableElements(container)
        if (focusableElements.length > 0) {
          focusableElements[0]?.element?.focus()
        }
      }
    }, 0)

    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [isOpen, closeOnEscape, onClose])

  useEffect(() => {
    if (!isOpen && previousActiveElement.current) {
      const focusTarget = returnFocusRef?.current || previousActiveElement.current
      if (typeof focusTarget.focus === 'function') {
        focusTarget.focus()
      }
    }
  }, [isOpen, returnFocusRef])

  const focusFirstElement = useCallback((): void => {
    const container = containerRef.current
    if (container) {
      const focusableElements = getFocusableElements(container)
      if (focusableElements.length > 0) {
        focusableElements[0]?.element?.focus()
      }
    }
  }, [])

  const focusLastElement = useCallback((): void => {
    const container = containerRef.current
    if (container) {
      const focusableElements = getFocusableElements(container)
      if (focusableElements.length > 0) {
        focusableElements[focusableElements.length - 1]?.element?.focus()
      }
    }
  }, [])

  return {
    containerRef,
    focusFirstElement,
    focusLastElement,
  }
}

export function useFocusTrap(isActive: boolean) {
  const previousActiveElement = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (!isActive) return

    previousActiveElement.current = document.activeElement as HTMLElement

    return () => {
      if (previousActiveElement.current) {
        previousActiveElement.current.focus()
      }
    }
  }, [isActive])

  return { previousActiveElement }
}

export function useFocusVisible() {
  const [isFocused, setIsFocused] = useState(false)

  useEffect(() => {
    const handleFocus = () => setIsFocused(true)
    const handleBlur = () => setIsFocused(false)

    window.addEventListener('focus', handleFocus)
    window.addEventListener('blur', handleBlur)

    return () => {
      window.removeEventListener('focus', handleFocus)
      window.removeEventListener('blur', handleBlur)
    }
  }, [])

  return isFocused
}

export default useFocusManagement