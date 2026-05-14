import React from 'react'
import { Skeleton } from './Skeleton'
import { SkeletonText } from './SkeletonText'

interface SkeletonCardProps {
  showHeader?: boolean
  headerHeight?: string | number
  bodyLines?: number
  headerWidth?: string | number
  className?: string
}

export const SkeletonCard: React.FC<SkeletonCardProps> = ({
  showHeader = true,
  headerHeight = 48,
  bodyLines = 4,
  headerWidth = 200,
  className = '',
}) => {
  return (
    <div className={`bg-gray-800 rounded-lg border border-gray-700 ${className}`}>
      {showHeader && (
        <div className="px-4 py-3 border-b border-gray-700">
          <Skeleton height={headerHeight} width={headerWidth} />
        </div>
      )}
      <div className="p-4">
        <SkeletonText lines={bodyLines} />
      </div>
    </div>
  )
}

export default SkeletonCard
