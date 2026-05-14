import React from 'react'
import { Skeleton } from './Skeleton'
import { SkeletonCard } from './SkeletonCard'
import { SkeletonTable } from './SkeletonTable'
import { SkeletonChart } from './SkeletonChart'

type PageType = 'dashboard' | 'table' | 'form' | 'list'

interface PageSkeletonProps {
  type?: PageType
  className?: string
}

const DashboardSkeleton: React.FC<{ className?: string }> = ({ className }) => (
  <div className={`space-y-6 ${className}`}>
    <div className="flex items-center justify-between">
      <Skeleton width={200} height={32} />
      <Skeleton width={120} height={40} />
    </div>
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {Array.from({ length: 4 }).map((_, index) => (
        <SkeletonCard key={index} showHeader={false} bodyLines={2} />
      ))}
    </div>
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <SkeletonChart height={250} showAxes />
      <SkeletonChart height={250} showAxes />
    </div>
    <div className="grid grid-cols-1 gap-6">
      <SkeletonTable columns={4} rows={5} />
    </div>
  </div>
)

const TableSkeleton: React.FC<{ className?: string }> = ({ className }) => (
  <div className={`space-y-4 ${className}`}>
    <div className="flex items-center justify-between">
      <Skeleton width={150} height={32} />
      <div className="flex gap-2">
        <Skeleton width={100} height={40} />
        <Skeleton width={80} height={40} />
      </div>
    </div>
    <SkeletonTable columns={5} rows={8} />
  </div>
)

const FormSkeleton: React.FC<{ className?: string }> = ({ className }) => (
  <div className={`space-y-6 ${className}`}>
    <div className="flex items-center justify-between">
      <Skeleton width={180} height={32} />
    </div>
    <SkeletonCard />
  </div>
)

const ListSkeleton: React.FC<{ className?: string }> = ({ className }) => (
  <div className={`space-y-4 ${className}`}>
    <div className="flex items-center justify-between">
      <Skeleton width={150} height={32} />
      <Skeleton width={100} height={40} />
    </div>
    <div className="space-y-3">
      {Array.from({ length: 6 }).map((_, index) => (
        <SkeletonCard key={index} showHeader={false} bodyLines={2} />
      ))}
    </div>
  </div>
)

export const PageSkeleton: React.FC<PageSkeletonProps> = ({
  type = 'dashboard',
  className = '',
}) => {
  const skeletonComponents: Record<PageType, React.ReactNode> = {
    dashboard: <DashboardSkeleton className={className} />,
    table: <TableSkeleton className={className} />,
    form: <FormSkeleton className={className} />,
    list: <ListSkeleton className={className} />,
  }

  return <>{skeletonComponents[type]}</>
}

export default PageSkeleton
