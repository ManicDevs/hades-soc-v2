import React from 'react'

type SpinnerSize = 'sm' | 'md' | 'lg'

interface LoadingSpinnerProps {
  size?: SpinnerSize
  message?: string
  className?: string
  fullScreen?: boolean
}

const sizeMap: Record<SpinnerSize, { container: string; dot: string }> = {
  sm: {
    container: 'w-4 h-4',
    dot: 'w-1 h-1',
  },
  md: {
    container: 'w-8 h-8',
    dot: 'w-2 h-2',
  },
  lg: {
    container: 'w-12 h-12',
    dot: 'w-3 h-3',
  },
}

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'md',
  message,
  className = '',
  fullScreen = false,
}) => {
  const { container, dot } = sizeMap[size]

  const spinner = (
    <div className={`${className}`}>
      <div className={`${container} relative flex items-center justify-center`}>
        {Array.from({ length: 12 }).map((_, index) => (
          <div
            key={index}
            className={`absolute ${dot} bg-hades-primary rounded-full animate-pulse`}
            style={{
              animationDelay: `${index * 75}ms`,
              transform: `rotate(${index * 30}deg) translateY(-50%)`,
            }}
          />
        ))}
      </div>
      {message && (
        <p className="mt-3 text-sm text-gray-400 text-center">{message}</p>
      )}
    </div>
  )

  if (fullScreen) {
    return (
      <div className="fixed inset-0 bg-hades-darker/80 flex items-center justify-center z-50">
        <div className="flex flex-col items-center">
          {spinner}
        </div>
      </div>
    )
  }

  return spinner
}

export default LoadingSpinner
