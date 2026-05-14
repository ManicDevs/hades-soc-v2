import React, { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { ErrorBoundary } from './ErrorBoundary'

interface PageErrorBoundaryProps {
  children: ReactNode
  pageName?: string
}

const PageErrorFallback: React.FC<{ pageName?: string | undefined }> = ({ pageName }) => {
  const navigate = useNavigate()

  const handleGoToDashboard = (): void => {
    navigate('/dashboard')
  }

  const handleReloadPage = (): void => {
    window.location.reload()
  }

  return (
    <div className="min-h-screen bg-hades-dark flex items-center justify-center p-4">
      <div className="bg-gray-800 border border-red-500/30 rounded-xl max-w-lg w-full p-8 text-center shadow-2xl">
        {/* Error Icon */}
        <div className="w-20 h-20 rounded-full bg-red-500/10 flex items-center justify-center mx-auto mb-6">
          <svg
            className="w-10 h-10 text-red-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>

        {/* Title and Description */}
        <h1 className="text-2xl font-bold text-white mb-3">
          Page Failed to Load
        </h1>
        <p className="text-gray-400 mb-2">
          {pageName ? `"${pageName}" encountered an error` : 'This page encountered an error'}
        </p>
        <p className="text-gray-500 text-sm mb-8">
          Please try again or return to the dashboard
        </p>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          <button
            onClick={handleGoToDashboard}
            className="px-5 py-2.5 bg-gray-700 hover:bg-gray-600 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
              />
            </svg>
            Go to Dashboard
          </button>
          <button
            onClick={handleReloadPage}
            className="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            Reload Page
          </button>
        </div>

        {/* Decorative Element */}
        <div className="mt-8 pt-6 border-t border-gray-700/50">
          <p className="text-xs text-gray-600">
            If this problem persists, please contact support
          </p>
        </div>
      </div>
    </div>
  )
}

export const PageErrorBoundary: React.FC<PageErrorBoundaryProps> = ({
  children,
  pageName
}) => {
  return (
    <ErrorBoundary fallbackUI={<PageErrorFallback pageName={pageName} />}>
      {children}
    </ErrorBoundary>
  )
}

export default PageErrorBoundary