import React from 'react'
import { Skeleton } from './Skeleton'

interface SkeletonChartProps {
  height?: number
  showAxes?: boolean
  className?: string
  chartWidth?: string
}

export const SkeletonChart: React.FC<SkeletonChartProps> = ({
  height = 300,
  showAxes = true,
  className = '',
}) => {
  return (
    <div className={`bg-gray-800 rounded-lg border border-gray-700 p-4 ${className}`}>
      <div className="flex" style={{ height }}>
        {showAxes && (
          <div className="flex flex-col justify-between pr-4 border-r border-gray-700">
            <Skeleton width={30} height={8} />
            <Skeleton width={30} height={8} />
            <Skeleton width={30} height={8} />
            <Skeleton width={30} height={8} />
            <Skeleton width={30} height={8} />
          </div>
        )}
        
        <div className="flex-1 flex items-end justify-around pl-4">
          {Array.from({ length: 12 }).map((_, index) => {
            const barHeight = 20 + Math.random() * 70
            return (
              <Skeleton
                key={index}
                width={16}
                height={`${barHeight}%`}
                className="mb-4"
              />
            )
          })}
        </div>
      </div>

      {showAxes && (
        <div className="flex justify-around pt-4 border-t border-gray-700 mt-4">
          {Array.from({ length: 12 }).map((_, index) => (
            <Skeleton key={`label-${index}`} width={24} height={8} />
          ))}
        </div>
      )}
    </div>
  )
}

export default SkeletonChart
