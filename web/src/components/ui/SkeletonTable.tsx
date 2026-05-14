import React from 'react'
import { Skeleton } from './Skeleton'

interface SkeletonTableProps {
  columns?: number
  rows?: number
  colWidths?: (string | number)[]
  headerHeight?: number
  rowHeight?: number
  className?: string
}

export const SkeletonTable: React.FC<SkeletonTableProps> = ({
  columns = 5,
  rows = 5,
  colWidths,
  headerHeight = 44,
  rowHeight = 48,
  className = '',
}) => {
  const defaultColumnWidths = Array.from({ length: columns }, (_, i) => 
    colWidths?.[i] || `${100 / columns}%`
  )

  return (
    <div className={`bg-gray-800 rounded-lg border border-gray-700 overflow-hidden ${className}`}>
      <table className="w-full">
        <thead>
          <tr className="bg-gray-700/30 border-b border-gray-700">
            {defaultColumnWidths.map((width, index) => (
              <th key={`header-${index}`} style={{ width: typeof width === 'number' ? `${width}px` : width }}>
                <div className="px-4 py-3">
                  <Skeleton height={headerHeight} />
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {Array.from({ length: rows }).map((_, rowIndex) => (
            <tr 
              key={`row-${rowIndex}`} 
              className="border-b border-gray-700/50 last:border-b-0"
            >
              {defaultColumnWidths.map((width, colIndex) => (
                <td key={`cell-${rowIndex}-${colIndex}`} style={{ width: typeof width === 'number' ? `${width}px` : width }}>
                  <div className="px-4 py-3">
                    <Skeleton height={rowHeight - 24} />
                  </div>
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default SkeletonTable
