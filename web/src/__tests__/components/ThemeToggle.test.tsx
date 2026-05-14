import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import ThemeToggle from '@/components/ThemeToggle'

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  Sun: () => <svg data-testid="sun-icon" />,
  Moon: () => <svg data-testid="moon-icon" />
}))

// Mock ThemeContext
const mockToggleTheme = vi.fn()
const mockTheme = { theme: 'dark' as const, isDark: true }

vi.mock('@/context/ThemeContext', () => ({
  useTheme: () => ({
    theme: mockTheme.theme,
    isDark: mockTheme.isDark,
    toggleTheme: mockToggleTheme
  })
}))

describe('ThemeToggle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('renders theme toggle button', () => {
    it('should render theme toggle button', () => {
      render(<ThemeToggle />)
      expect(screen.getByRole('button')).toBeInTheDocument()
    })

    it('should have correct aria-label for dark mode', () => {
      render(<ThemeToggle />)
      expect(screen.getByRole('button')).toHaveAttribute(
        'aria-label',
        'Switch to light mode'
      )
    })

    it('should render both sun and moon icons', () => {
      render(<ThemeToggle />)
      expect(screen.getByTestId('sun-icon')).toBeInTheDocument()
      expect(screen.getByTestId('moon-icon')).toBeInTheDocument()
    })
  })

  describe('handles theme switching', () => {
    it('should call toggleTheme when clicked', () => {
      render(<ThemeToggle />)
      const button = screen.getByRole('button')

      fireEvent.click(button)

      expect(mockToggleTheme).toHaveBeenCalledTimes(1)
    })
  })

  describe('applies correct CSS classes for dark mode', () => {
    it('should apply dark mode specific classes', () => {
      const { container } = render(<ThemeToggle />)
      const button = container.querySelector('button')
      expect(button?.className).toContain('bg-transparent')
    })
  })
})
