import React from 'react'

interface SuspenseFallbackProps {
  message?: string
  size?: 'sm' | 'md' | 'lg'
}

export const SuspenseFallback: React.FC<SuspenseFallbackProps> = ({
  message = 'Loading...',
  size = 'md'
}) => {
  const sizeClasses = {
    sm: 'w-6 h-6 border-2',
    md: 'w-10 h-10 border-[3px]',
    lg: 'w-14 h-14 border-4'
  }

  return (
    <div className="flex flex-col items-center justify-center p-8 bg-hades-dark/50 rounded-lg">
      <div className={`${sizeClasses[size]} border-hades-primary/30 border-t-hades-primary rounded-full animate-spin`} />
      <p className="mt-4 text-gray-400 text-sm">{message}</p>
    </div>
  )
}

export default SuspenseFallback
