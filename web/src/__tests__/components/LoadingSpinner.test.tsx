import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { LoadingSpinner } from '@/components/ui/LoadingSpinner'

describe('LoadingSpinner', () => {
  describe('renders with different sizes', () => {
    it('should render small spinner', () => {
      const { container } = render(<LoadingSpinner size="sm" />)
      const spinner = container.querySelector('.w-4')
      expect(spinner).toBeInTheDocument()
    })

    it('should render medium spinner (default)', () => {
      const { container } = render(<LoadingSpinner />)
      const spinner = container.querySelector('.w-8')
      expect(spinner).toBeInTheDocument()
    })

    it('should render large spinner', () => {
      const { container } = render(<LoadingSpinner size="lg" />)
      const spinner = container.querySelector('.w-12')
      expect(spinner).toBeInTheDocument()
    })
  })

  describe('renders spinner dots', () => {
    it('should render 12 dots for animation', () => {
      const { container } = render(<LoadingSpinner />)
      const dots = container.querySelectorAll('.absolute.bg-hades-primary')
      expect(dots).toHaveLength(12)
    })

    it('should apply animation delays to dots', () => {
      const { container } = render(<LoadingSpinner />)
      const firstDot = container.querySelector('.animate-pulse')
      expect(firstDot).toBeInTheDocument()
    })
  })

  describe('renders with message', () => {
    it('should display message when provided', () => {
      render(<LoadingSpinner message="Loading data..." />)
      expect(screen.getByText('Loading data...')).toBeInTheDocument()
    })

    it('should not display message when not provided', () => {
      const { container } = render(<LoadingSpinner />)
      expect(container.querySelector('p')).not.toBeInTheDocument()
    })
  })

  describe('renders fullscreen variant', () => {
    it('should render fullscreen overlay when fullScreen is true', () => {
      const { container } = render(<LoadingSpinner fullScreen />)
      const overlay = container.querySelector('.fixed.inset-0')
      expect(overlay).toBeInTheDocument()
    })

    it('should render centered content in fullscreen', () => {
      const { container } = render(<LoadingSpinner fullScreen message="Loading..." />)
      const centered = container.querySelector('.flex.flex-col.items-center')
      expect(centered).toBeInTheDocument()
    })
  })

  describe('applies custom className', () => {
    it('should apply custom className', () => {
      const { container } = render(<LoadingSpinner className="custom-class" />)
      const customClass = container.querySelector('.custom-class')
      expect(customClass).toBeInTheDocument()
    })
  })

  describe('animation transforms', () => {
    it('should apply rotation transforms to dots', () => {
      const { container } = render(<LoadingSpinner />)
      const dots = container.querySelectorAll('[style*="transform"]')
      expect(dots.length).toBeGreaterThan(0)
    })
  })
})
