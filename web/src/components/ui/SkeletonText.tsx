import React from 'react'
import { Skeleton } from './Skeleton'

interface SkeletonTextProps {
  lines?: number
  lastLineWidth?: string | number
  className?: string
  lineHeight?: string | number
  lineSpacing?: string | number
}

export const SkeletonText: React.FC<SkeletonTextProps> = ({
  lines = 4,
  lastLineWidth = '60%',
  className = '',
  lineHeight = 16,
}) => {
  const lineWidths = Array.from({ length: lines }, (_, index) => {
    if (index === lines - 1) {
      return lastLineWidth
    }
    return `${100 - index * 15}%`
  })

  return (
    <div className={`space-y-2 ${className}`}>
      {lineWidths.map((width, index) => (
        <Skeleton
          key={index}
          height={lineHeight}
          width={width}
        />
      ))}
    </div>
  )
}

export default SkeletonText
