import React, { ReactNode, Suspense } from 'react'
import { ErrorBoundary } from './ErrorBoundary'

interface AsyncErrorBoundaryProps {
  children: ReactNode
  loadingFallback?: ReactNode
  errorFallback?: ReactNode
  suspenseOptions?: {
    fallback?: ReactNode
  }
}

const defaultLoadingFallback = (
  <div className="flex flex-col items-center justify-center p-8 space-y-4">
    {/* Skeleton Animation */}
    <div className="relative w-16 h-16">
      <div className="absolute inset-0 border-4 border-blue-500/20 rounded-full"></div>
      <div className="absolute inset-0 border-4 border-transparent border-t-blue-500 rounded-full animate-spin"></div>
    </div>

    {/* Skeleton Lines */}
    <div className="space-y-3 w-48">
      <div className="h-4 bg-gray-700/50 rounded animate-pulse"></div>
      <div className="h-4 bg-gray-700/30 rounded animate-pulse w-3/4"></div>
    </div>

    <p className="text-gray-400 text-sm">Loading...</p>
  </div>
)

const defaultErrorFallback = (
  <div className="flex flex-col items-center justify-center p-6">
    <div className="bg-gray-800/80 border border-red-500/30 rounded-lg p-6 max-w-md w-full text-center">
      <div className="w-14 h-14 rounded-full bg-red-500/10 flex items-center justify-center mx-auto mb-4">
        <svg
          className="w-7 h-7 text-red-400"
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
      <h3 className="text-lg font-semibold text-white mb-2">
        Failed to Load
      </h3>
      <p className="text-gray-400 text-sm mb-4">
        An error occurred while loading this content
      </p>
    </div>
  </div>
)

export const AsyncErrorBoundary: React.FC<AsyncErrorBoundaryProps> = ({
  children,
  loadingFallback = defaultLoadingFallback,
  errorFallback = defaultErrorFallback,
  suspenseOptions = {}
}) => {
  const fallback = suspenseOptions.fallback ?? loadingFallback

  return (
    <ErrorBoundary fallbackUI={errorFallback}>
      <Suspense fallback={fallback}>
        {children}
      </Suspense>
    </ErrorBoundary>
  )
}

export default AsyncErrorBoundary