import React, { ReactNode } from 'react'
import { ErrorBoundary } from './ErrorBoundary'

interface ErrorBoundaryWithRetryProps {
  children: ReactNode
  onRetry?: () => void
  errorMessage?: string
  className?: string
}

export const ErrorBoundaryWithRetry: React.FC<ErrorBoundaryWithRetryProps> = ({
  children,
  onRetry,
  errorMessage = 'An error occurred',
  className = ''
}) => {
  const handleError = (): void => {
    // Error is already logged by ErrorBoundary
  }

  const fallbackUI = (
    <div className={`flex flex-col items-center justify-center p-6 ${className}`}>
      <div className="bg-gray-800/80 border border-red-500/30 rounded-lg p-6 max-w-md w-full text-center">
        {/* Warning Icon */}
        <div className="w-16 h-16 rounded-full bg-red-500/10 flex items-center justify-center mx-auto mb-4">
          <svg
            className="w-8 h-8 text-red-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>

        {/* Error Message */}
        <h3 className="text-lg font-semibold text-white mb-2">
          Operation Failed
        </h3>
        <p className="text-gray-400 mb-6">
          {errorMessage}
        </p>

        {/* Retry Button */}
        {onRetry && (
          <button
            onClick={onRetry}
            className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium transition-colors inline-flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Retry
          </button>
        )}
      </div>
    </div>
  )

  return (
    <ErrorBoundary fallbackUI={fallbackUI} onError={handleError}>
      {children}
    </ErrorBoundary>
  )
}

export default ErrorBoundaryWithRetry