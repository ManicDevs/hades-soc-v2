import React from 'react'

interface PageLoaderProps {
  message?: string
}

export const PageLoader: React.FC<PageLoaderProps> = ({ message = 'Loading page...' }) => {
  return (
    <div className="min-h-[400px] flex flex-col items-center justify-center bg-hades-dark p-8">
      <div className="relative">
        <div className="w-16 h-16 border-4 border-hades-primary/20 rounded-full" />
        <div className="absolute top-0 left-0 w-16 h-16 border-4 border-hades-primary border-t-transparent rounded-full animate-spin" />
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
          <svg
            className="w-6 h-6 text-hades-primary animate-pulse"
            fill="none"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              className="opacity-75"
              fill="currentColor"
              d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"
            />
          </svg>
        </div>
      </div>
      <p className="mt-6 text-gray-400 text-sm font-medium tracking-wide">{message}</p>
      <div className="mt-4 flex gap-1">
        <div className="w-2 h-2 bg-hades-primary rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
        <div className="w-2 h-2 bg-hades-primary rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
        <div className="w-2 h-2 bg-hades-primary rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
      </div>
    </div>
  )
}

export default PageLoader
