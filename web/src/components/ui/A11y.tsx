import React, { useEffect, useRef, useState, useCallback, ReactNode, CSSProperties } from 'react'

/**
 * VisuallyHidden - Component that hides content visually but keeps it accessible to screen readers
 * Use for: labels, instructions, or content that needs to be announced but not visually visible
 */
interface VisuallyHiddenProps {
  children: ReactNode
  className?: string
  style?: CSSProperties
}

export const VisuallyHidden: React.FC<VisuallyHiddenProps> = ({ children, className = '', style }) => {
  return (
    <span
      className={`sr-only ${className}`}
      style={{
        position: 'absolute',
        width: '1px',
        height: '1px',
        padding: 0,
        margin: '-1px',
        overflow: 'hidden',
        clip: 'rect(0, 0, 0, 0)',
        whiteSpace: 'nowrap',
        border: 0,
        ...style,
      }}
    >
      {children}
    </span>
  )
}

/**
 * SkipLink - Component that allows keyboard users to skip navigation and jump to main content
 * Placement: Should be the first focusable element in the document body
 */
interface SkipLinkProps {
  targetId?: string
  children?: ReactNode
}

export const SkipLink: React.FC<SkipLinkProps> = ({ targetId = 'main-content', children = 'Skip to main content' }) => {
  const handleClick = useCallback((_e: React.MouseEvent<HTMLAnchorElement>) => {
    const target = document.getElementById(targetId)
    if (target) {
      target.focus()
      target.scrollIntoView({ behavior: 'smooth' })
    }
  }, [targetId])

  return (
    <a
      href={`#${targetId}`}
      onClick={handleClick}
      className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-50 focus:px-4 focus:py-2 focus:bg-hades-primary focus:text-white focus:rounded focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-hades-primary"
      style={{
        position: 'absolute',
        clip: 'rect(0, 0, 0, 0)',
      }}
    >
      {children}
    </a>
  )
}

/**
 * FocusTrap - Component that traps focus within a container
 * Use for: Modals, dialogs, popovers where focus should not escape
 */
interface FocusTrapProps {
  isActive: boolean
  children: ReactNode
  className?: string
  returnFocusRef?: React.RefObject<HTMLElement>
}

export const FocusTrap: React.FC<FocusTrapProps> = ({ isActive, children, className = '', returnFocusRef }) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const previousActiveElement = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (!isActive) return

    // Store the currently focused element
    previousActiveElement.current = document.activeElement as HTMLElement

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return

      const container = containerRef.current
      if (!container) return

      const focusableElements = container.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      )
      const firstElement = focusableElements[0]
      const lastElement = focusableElements[focusableElements.length - 1]

      if (e.shiftKey && document.activeElement === firstElement) {
        e.preventDefault()
        lastElement?.focus()
      } else if (!e.shiftKey && document.activeElement === lastElement) {
        e.preventDefault()
        firstElement?.focus()
      }
    }

    // Focus the first focusable element
    setTimeout(() => {
      const container = containerRef.current
      if (container) {
        const focusableElements = container.querySelectorAll<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        )
        focusableElements[0]?.focus()
      }
    }, 0)

    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      // Return focus to the previously focused element
      if (returnFocusRef?.current) {
        returnFocusRef.current.focus()
      } else if (previousActiveElement.current) {
        previousActiveElement.current.focus()
      }
    }
  }, [isActive, returnFocusRef])

  return (
    <div ref={containerRef} className={className} data-focus-trap="true">
      {children}
    </div>
  )
}

/**
 * LiveRegion - Component for announcing dynamic content to screen readers
 * Use for: Status updates, notifications, form errors, loading states
 */
type LiveRegionPoliteness = 'polite' | 'assertive'

interface LiveRegionProps {
  message: string
  politeness?: LiveRegionPoliteness
  atomic?: boolean
  className?: string
}

export const LiveRegion: React.FC<LiveRegionProps> = ({
  message,
  politeness = 'polite',
  atomic = true,
  className = '',
}) => {
  const [currentMessage, setCurrentMessage] = useState('')
  const previousMessageRef = useRef('')

  useEffect(() => {
    // Only announce if the message has changed
    if (message !== previousMessageRef.current) {
      previousMessageRef.current = message
      // Small delay to ensure screen reader picks up the change
      setTimeout(() => setCurrentMessage(message), 100)
    }
  }, [message])

  return (
    <div
      role="status"
      aria-live={politeness}
      aria-atomic={atomic}
      className={className}
      style={{
        position: 'absolute',
        width: '1px',
        height: '1px',
        padding: 0,
        margin: '-1px',
        overflow: 'hidden',
        clip: 'rect(0, 0, 0, 0)',
        whiteSpace: 'nowrap',
        border: 0,
      }}
    >
      {currentMessage}
    </div>
  )
}

/**
 * Announcement - Hook-based live region for programmatic announcements
 */
interface UseAnnouncementOptions {
  politeness?: LiveRegionPoliteness
  atomic?: boolean
}

export function useAnnouncement(options: UseAnnouncementOptions = {}) {
  const [announcement, setAnnouncement] = useState('')
  const previousRef = useRef('')

  const announce = useCallback((message: string) => {
    if (message !== previousRef.current) {
      previousRef.current = message
      setTimeout(() => setAnnouncement(message), 50)
    }
  }, [])

  const Component: React.FC<{ className?: string }> = ({ className = '' }) => (
    <LiveRegion
      message={announcement}
      politeness={options.politeness ?? 'polite'}
      atomic={options.atomic ?? true}
      className={className}
    />
  )

  return { announce, AnnouncementComponent: Component }
}

/**
 * Modal - Accessible modal component with focus trap and escape key handling
 */
interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  children: ReactNode
  className?: string
  closeOnOverlayClick?: boolean
  closeOnEscape?: boolean
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  children,
  className = '',
  closeOnOverlayClick = true,
  closeOnEscape = true,
}) => {
  const modalRef = useRef<HTMLDivElement>(null)
  const previousFocusRef = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (!isOpen) return

    // Store the currently focused element
    previousFocusRef.current = document.activeElement as HTMLElement

    // Handle escape key
    const handleKeyDown = (e: KeyboardEvent) => {
      if (closeOnEscape && e.key === 'Escape') {
        onClose()
      }
    }

    // Prevent body scroll
    document.body.style.overflow = 'hidden'

    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      document.body.style.overflow = ''
      // Return focus to the previously focused element
      previousFocusRef.current?.focus()
    }
  }, [isOpen, onClose, closeOnEscape])

  // Focus trap logic
  useEffect(() => {
    if (!isOpen) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return

      const modal = modalRef.current
      if (!modal) return

      const focusableElements = modal.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      )
      const firstElement = focusableElements[0]
      const lastElement = focusableElements[focusableElements.length - 1]

      if (e.shiftKey && document.activeElement === firstElement) {
        e.preventDefault()
        lastElement?.focus()
      } else if (!e.shiftKey && document.activeElement === lastElement) {
        e.preventDefault()
        firstElement?.focus()
      }
    }

    // Focus the first element after mount
    setTimeout(() => {
      const modal = modalRef.current
      if (modal) {
        const focusableElements = modal.querySelectorAll<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        )
        focusableElements[0]?.focus()
      }
    }, 0)

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen])

  if (!isOpen) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
    >
      {/* Overlay */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={closeOnOverlayClick ? onClose : undefined}
        aria-hidden="true"
      />

      {/* Modal content */}
      <div
        ref={modalRef}
        className={`relative bg-hades-dark border border-gray-700 rounded-lg shadow-xl max-w-lg w-full mx-4 max-h-[90vh] overflow-y-auto ${className}`}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-700">
          <h2 id="modal-title" className="text-lg font-semibold text-white">
            {title}
          </h2>
          <button
            onClick={onClose}
            className="p-1 text-gray-400 hover:text-white rounded transition-colors focus:outline-none focus:ring-2 focus:ring-hades-primary"
            aria-label="Close modal"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Body */}
        <div className="p-4">{children}</div>
      </div>
    </div>
  )
}

/**
 * Tooltip - Accessible tooltip component
 */
interface TooltipProps {
  content: string
  children: ReactNode
  position?: 'top' | 'bottom' | 'left' | 'right'
}

export const Tooltip: React.FC<TooltipProps> = ({ content, children, position = 'top' }) => {
  const [isVisible, setIsVisible] = useState(false)
  const triggerRef = useRef<HTMLButtonElement>(null)

  const positionClasses = {
    top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
    bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
    left: 'right-full top-1/2 -translate-y-1/2 mr-2',
    right: 'left-full top-1/2 -translate-y-1/2 ml-2',
  }

  return (
    <div className="relative inline-block">
      <button
        ref={triggerRef}
        onMouseEnter={() => setIsVisible(true)}
        onMouseLeave={() => setIsVisible(false)}
        onFocus={() => setIsVisible(true)}
        onBlur={() => setIsVisible(false)}
        aria-describedby="tooltip"
        className="focus:outline-none focus:ring-2 focus:ring-hades-primary focus:ring-offset-2 focus:ring-offset-hades-dark rounded"
      >
        {children}
      </button>
      {isVisible && (
        <div
          id="tooltip"
          role="tooltip"
          className={`absolute z-50 px-2 py-1 text-sm bg-gray-900 text-white rounded shadow-lg whitespace-nowrap ${positionClasses[position]}`}
        >
          {content}
        </div>
      )}
    </div>
  )
}

/**
 * Button with accessible label support
 */
interface AccessibleButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  label: string
  icon?: ReactNode
  iconPosition?: 'left' | 'right'
}

export const AccessibleButton: React.FC<AccessibleButtonProps> = ({
  label,
  icon,
  iconPosition = 'left',
  children,
  className = '',
  ...props
}) => {
  return (
    <button
      className={`focus:outline-none focus:ring-2 focus:ring-hades-primary focus:ring-offset-2 focus:ring-offset-hades-dark ${className}`}
      aria-label={label}
      {...props}
    >
      {icon && iconPosition === 'left' && <span className="mr-2">{icon}</span>}
      {children || label}
      {icon && iconPosition === 'right' && <span className="ml-2">{icon}</span>}
    </button>
  )
}

/**
 * IconButton - Accessible icon-only button
 */
interface IconButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  icon: ReactNode
  label: string
  size?: 'sm' | 'md' | 'lg'
}

export const IconButton: React.FC<IconButtonProps> = ({
  icon,
  label,
  size = 'md',
  className = '',
  ...props
}) => {
  const sizeClasses = {
    sm: 'p-1',
    md: 'p-2',
    lg: 'p-3',
  }

  return (
    <button
      className={`inline-flex items-center justify-center rounded text-gray-400 hover:text-white hover:bg-gray-700 transition-colors focus:outline-none focus:ring-2 focus:ring-hades-primary ${sizeClasses[size]} ${className}`}
      aria-label={label}
      {...props}
    >
      {icon}
    </button>
  )
}