import React from 'react'

interface SkeletonProps {
  width?: string | number
  height?: string | number
  circle?: boolean
  className?: string
  animation?: 'pulse' | 'wave' | 'none'
}

export const Skeleton: React.FC<SkeletonProps> = ({
  width,
  height,
  circle = false,
  className = '',
  animation = 'wave',
}) => {
  const baseClasses = 'bg-gray-600/50'
  
  const animationClasses = {
    pulse: 'animate-pulse',
    wave: 'animate-shimmer',
    none: '',
  }

  const shapeClasses = circle
    ? 'rounded-full'
    : 'rounded-md'

  const style: React.CSSProperties = {
    width: typeof width === 'number' ? `${width}px` : width,
    height: typeof height === 'number' ? `${height}px` : height,
  }

  return (
    <div
      className={`${baseClasses} ${animationClasses[animation]} ${shapeClasses} ${className}`}
      style={style}
    />
  )
}

export default Skeleton
