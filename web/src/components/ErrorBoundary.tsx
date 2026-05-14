import React, { Component, ReactNode } from 'react'

interface ErrorBoundaryProps {
  children: ReactNode
  fallbackUI?: ReactNode
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void
}

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
  errorInfo: React.ErrorInfo | null
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null
    }
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    this.setState({ errorInfo })
    console.error('ErrorBoundary caught an error:', error, errorInfo)
    this.props.onError?.(error, errorInfo)
  }

  handleReset = (): void => {
    this.setState({ hasError: false, error: null, errorInfo: null })
  }

  handleReportIssue = (): void => {
    const { error, errorInfo } = this.state
    const subject = encodeURIComponent('HADES Dashboard Error Report')
    const body = encodeURIComponent(
      `Error: ${error?.message || 'Unknown error'}\n\nStack: ${error?.stack || 'No stack trace'}\n\nComponent Stack: ${errorInfo?.componentStack || 'No component stack'}`
    )
    window.open(`mailto:support@hades.local?subject=${subject}&body=${body}`)
  }

  renderFallback(): ReactNode {
    const { error, errorInfo } = this.state

    return (
      <div className="min-h-screen bg-hades-dark flex items-center justify-center p-4">
        <div className="bg-gray-800 border border-red-500/50 rounded-lg max-w-2xl w-full p-6 shadow-2xl">
          {/* Error Header */}
          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 rounded-full bg-red-500/20 flex items-center justify-center">
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
            <div>
              <h2 className="text-xl font-bold text-white">Something went wrong</h2>
              <p className="text-gray-400 text-sm">An unexpected error occurred</p>
            </div>
          </div>

          {/* Error Details - Collapsible */}
          <details className="bg-gray-900/50 rounded-lg mb-4">
            <summary className="p-3 cursor-pointer text-gray-300 hover:text-white transition-colors">
              Error Details
            </summary>
            <div className="p-3 pt-0 space-y-3">
              {error && (
                <div>
                  <p className="text-red-400 font-mono text-sm break-all">
                    {error.message}
                  </p>
                </div>
              )}
              {error?.stack && (
                <div className="bg-gray-950 rounded p-3 overflow-auto max-h-48">
                  <pre className="text-xs text-gray-400 whitespace-pre-wrap font-mono">
                    {error.stack}
                  </pre>
                </div>
              )}
              {errorInfo?.componentStack && (
                <div className="bg-gray-950 rounded p-3 overflow-auto max-h-48">
                  <p className="text-xs text-gray-500 uppercase mb-1">Component Stack</p>
                  <pre className="text-xs text-gray-400 whitespace-pre-wrap font-mono">
                    {errorInfo.componentStack}
                  </pre>
                </div>
              )}
            </div>
          </details>

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-3 justify-end">
            <button
              onClick={this.handleReportIssue}
              className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md font-medium transition-colors flex items-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              Report Issue
            </button>
            <button
              onClick={this.handleReset}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium transition-colors flex items-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
              Try Again
            </button>
          </div>
        </div>
      </div>
    )
  }

  render(): ReactNode {
    const { fallbackUI, children } = this.props

    if (this.state.hasError) {
      return fallbackUI ?? this.renderFallback()
    }

    return children
  }
}

export default ErrorBoundary