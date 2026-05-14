import { useState, useCallback, useMemo } from 'react'
import { Skeleton } from '../components/ui/Skeleton'
import { SkeletonText } from '../components/ui/SkeletonText'
import { SkeletonCard } from '../components/ui/SkeletonCard'
import { SkeletonTable } from '../components/ui/SkeletonTable'
import { SkeletonChart } from '../components/ui/SkeletonChart'
import { PageSkeleton } from '../components/ui/PageSkeleton'

interface UseSkeletonOptions {
  initialLoading?: boolean
  debounceMs?: number
}

interface SkeletonComponentProps {
  width?: string | number
  height?: string | number
  className?: string
  circle?: boolean
  animation?: 'pulse' | 'wave' | 'none'
  style?: React.CSSProperties
}

interface SkeletonTextProps {
  lines?: number
  lastLineWidth?: string | number
  className?: string
  lineHeight?: string | number
  lineSpacing?: string | number
}

interface SkeletonTableProps {
  columns?: number
  rows?: number
  colWidths?: (string | number)[]
  headerHeight?: number
  rowHeight?: number
  className?: string
}

interface SkeletonChartProps {
  height?: number
  showAxes?: boolean
  className?: string
  chartWidth?: string
}

interface SkeletonCardProps {
  showHeader?: boolean
  headerHeight?: string | number
  bodyLines?: number
  headerWidth?: string | number
  className?: string
}

interface PageSkeletonOptions {
  type?: 'dashboard' | 'table' | 'form' | 'list'
  className?: string
}

interface UseSkeletonReturn {
  isLoading: boolean
  Skeleton: React.FC<SkeletonComponentProps>
  SkeletonText: React.FC<SkeletonTextProps>
  SkeletonCard: React.FC<SkeletonCardProps>
  SkeletonTable: React.FC<SkeletonTableProps>
  SkeletonChart: React.FC<SkeletonChartProps>
  PageSkeleton: React.FC<PageSkeletonOptions>
  showSkeleton: () => void
  hideSkeleton: () => void
  toggleSkeleton: () => void
  setLoadingState: (loading: boolean) => void
}

export const useSkeleton = (options: UseSkeletonOptions = {}): UseSkeletonReturn => {
  const { initialLoading = true, debounceMs = 0 } = options
  
  const [isLoading, setIsLoading] = useState(initialLoading)

  const debounceTimerRef = useMemo(() => ({ current: null as NodeJS.Timeout | null }), [])

  const setLoadingState = useCallback((loading: boolean) => {
    if (debounceMs > 0) {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
      debounceTimerRef.current = setTimeout(() => {
        setIsLoading(loading)
      }, debounceMs)
    } else {
      setIsLoading(loading)
    }
  }, [debounceMs, debounceTimerRef])

  const showSkeleton = useCallback(() => {
    setLoadingState(true)
  }, [setLoadingState])

  const hideSkeleton = useCallback(() => {
    setLoadingState(false)
  }, [setLoadingState])

  const toggleSkeleton = useCallback(() => {
    setIsLoading(prev => !prev)
  }, [])

  return {
    isLoading,
    Skeleton,
    SkeletonText,
    SkeletonCard,
    SkeletonTable,
    SkeletonChart,
    PageSkeleton,
    showSkeleton,
    hideSkeleton,
    toggleSkeleton,
    setLoadingState,
  }
}

export default useSkeleton
