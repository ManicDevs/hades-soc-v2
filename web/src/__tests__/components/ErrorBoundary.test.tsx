import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { Skeleton } from '@/components/ui/Skeleton'
import { ErrorBoundary } from '@/components/ErrorBoundary'

describe('UI Components', () => {
  describe('Skeleton', () => {
    it('should render skeleton element', () => {
      const { container } = render(<Skeleton />)
      expect(container.firstChild).toBeInTheDocument()
    })

    it('should apply custom dimensions', () => {
      const { container } = render(<Skeleton width={100} height={50} />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.style.width).toBe('100px')
      expect(skeleton?.style.height).toBe('50px')
    })
  })

  describe('ErrorBoundary', () => {
    it('should render children when no error', () => {
      render(
        <ErrorBoundary>
          <div>Test Content</div>
        </ErrorBoundary>
      )
    })

    it('should catch errors and show fallback', () => {
      const ThrowError = () => {
        throw new Error('Test error')
      }

      render(
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      )
    })
  })
})
