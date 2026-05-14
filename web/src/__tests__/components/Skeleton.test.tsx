import { describe, it, expect } from 'vitest'
import { render } from '@testing-library/react'
import { Skeleton } from '@/components/ui/Skeleton'

describe('Skeleton', () => {
  describe('renders with default props', () => {
    it('should render skeleton element', () => {
      const { container } = render(<Skeleton />)
      expect(container.firstChild).toBeInTheDocument()
    })

    it('should apply default wave animation', () => {
      const { container } = render(<Skeleton />)
      const skeleton = container.querySelector('[class*="animate-shimmer"]')
      expect(skeleton).toBeInTheDocument()
    })
  })

  describe('renders with custom dimensions', () => {
    it('should apply custom width as number', () => {
      const { container } = render(<Skeleton width={200} />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.style.width).toBe('200px')
    })

    it('should apply custom width as string', () => {
      const { container } = render(<Skeleton width="100%" />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.style.width).toBe('100%')
    })

    it('should apply custom height as number', () => {
      const { container } = render(<Skeleton height={50} />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.style.height).toBe('50px')
    })

    it('should apply custom height as string', () => {
      const { container } = render(<Skeleton height="100px" />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.style.height).toBe('100px')
    })
  })

  describe('renders circle shape', () => {
    it('should apply rounded-full for circle', () => {
      const { container } = render(<Skeleton circle />)
      const skeleton = container.querySelector('[class*="rounded-full"]')
      expect(skeleton).toBeInTheDocument()
    })
  })

  describe('animation variants', () => {
    it('should apply pulse animation', () => {
      const { container } = render(<Skeleton animation="pulse" />)
      const skeleton = container.querySelector('[class*="animate-pulse"]')
      expect(skeleton).toBeInTheDocument()
    })

    it('should apply wave animation', () => {
      const { container } = render(<Skeleton animation="wave" />)
      const skeleton = container.querySelector('[class*="animate-shimmer"]')
      expect(skeleton).toBeInTheDocument()
    })

    it('should not apply animation when set to none', () => {
      const { container } = render(<Skeleton animation="none" />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.className).not.toContain('animate-')
    })
  })

  describe('applies custom className', () => {
    it('should apply custom className', () => {
      const { container } = render(<Skeleton className="custom-class" />)
      const skeleton = container.querySelector('.custom-class')
      expect(skeleton).toBeInTheDocument()
    })
  })

  describe('edge cases', () => {
    it('should handle zero dimensions', () => {
      const { container } = render(<Skeleton width={0} height={0} />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.style.width).toBe('0px')
      expect(skeleton?.style.height).toBe('0px')
    })

    it('should render with valid dimensions', () => {
      const { container } = render(<Skeleton width={100} height={50} />)
      const skeleton = container.firstChild as HTMLElement
      expect(skeleton?.style.width).toBe('100px')
      expect(skeleton?.style.height).toBe('50px')
    })
  })
})
