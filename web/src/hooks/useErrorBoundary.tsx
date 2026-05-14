import { useState, useCallback, useRef } from 'react'

export interface ErrorBoundaryState {
  error: Error | null
  hasError: boolean
}

export interface UseErrorBoundaryReturn {
  showError: (error: Error) => void
  error: Error | null
  resetError: () => void
  ErrorBoundaryComponent: React.ComponentType<{
    children: React.ReactNode
    fallback?: React.ReactNode
    onError?: (error: Error) => void
  }>
}

/**
 * Hook version of ErrorBoundary for functional components
 * Returns error state and methods to show/reset errors
 */
export function useErrorBoundary(): UseErrorBoundaryReturn {
  const [error, setError] = useState<Error | null>(null)
  const errorCallbacksRef = useRef<Set<(error: Error) => void>>(new Set())

  const showError = useCallback((err: Error) => {
    setError(err)
    console.error('useErrorBoundary caught error:', err)
    errorCallbacksRef.current.forEach((callback) => callback(err))
  }, [])

  const resetError = useCallback(() => {
    setError(null)
  }, [])

  // Internal ErrorBoundary component that uses this hook's state
  const ErrorBoundaryComponent: React.ComponentType<{
    children: React.ReactNode
    fallback?: React.ReactNode
    onError?: (error: Error) => void
  }> = ({ children, fallback, onError: _onError }) => {
    const [_hasError, _setHasError] = useState(false)
    const [localError, _setLocalError] = useState<Error | null>(null)

    if (error || localError) {
      const currentError = error || localError
      if (fallback) {
        return typeof fallback === 'function'
          ? (fallback as (error: Error) => React.ReactNode)(currentError!)
          : fallback
      }

      return (
        <div className="flex flex-col items-center justify-center p-6 bg-gray-900 rounded-lg border border-red-500/30">
          <div className="w-12 h-12 rounded-full bg-red-500/10 flex items-center justify-center mb-4">
            <svg
              className="w-6 h-6 text-red-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-white mb-2">Error Occurred</h3>
          <p className="text-gray-400 mb-4 text-center max-w-md">
            {currentError?.message || 'An unexpected error happened'}
          </p>
          <button
            onClick={resetError}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium transition-colors"
          >
            Dismiss
          </button>
        </div>
      )
    }

    return <>{children}</>
  }

  return {
    showError,
    error,
    resetError,
    ErrorBoundaryComponent
  }
}

/**
 * Simple version that returns just state and handlers
 * For use with useState pattern
 */
export function useErrorBoundarySimple() {
  const [error, setError] = useState<Error | null>(null)

  const showError = useCallback((err: Error) => {
    setError(err)
    console.error('Error captured:', err)
  }, [])

  const resetError = useCallback(() => {
    setError(null)
  }, [])

  return { showError, error, resetError }
}

export default useErrorBoundary