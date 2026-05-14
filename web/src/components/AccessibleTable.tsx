import React, { useState, useCallback, useMemo, ReactNode } from 'react'

/**
 * AccessibleTable - A fully accessible data table component
 * Features: Sortable columns, pagination, responsive design, proper ARIA attributes
 */

export interface TableColumn<T> {
  key: keyof T | string
  header: string
  sortable?: boolean
  render?: (value: unknown, row: T, index: number) => ReactNode
  width?: string
  align?: 'left' | 'center' | 'right'
}

interface AccessibleTableProps<T> {
  data: T[]
  columns: TableColumn<T>[]
  caption?: string
  sortKey?: string
  sortDirection?: 'asc' | 'desc'
  onSort?: (key: string) => void
  pageSize?: number
  currentPage?: number
  onPageChange?: (page: number) => void
  emptyMessage?: string
  loading?: boolean
  ariaLabel?: string
  className?: string
  id?: string
}

/**
 * SortIndicator - Visual indicator for sort direction
 */
const SortIndicator: React.FC<{ direction?: 'asc' | 'desc' | undefined }> = ({ direction }) => {
  if (!direction) {
    return (
      <span className="ml-1 inline-block w-4 h-4 text-gray-500" aria-hidden="true">
        ⬍
      </span>
    )
  }

  return (
    <span className="ml-1 inline-block w-4 h-4 text-hades-primary" aria-hidden="true">
      {direction === 'asc' ? '↑' : '↓'}
    </span>
  )
}

/**
 * AccessibleTable - Main component
 */
function AccessibleTable<T extends Record<string, unknown>>({
  data,
  columns,
  caption,
  sortKey,
  sortDirection = 'asc',
  onSort,
  pageSize = 10,
  currentPage = 1,
  onPageChange,
  emptyMessage = 'No data available',
  loading = false,
  ariaLabel,
  className = '',
  id,
}: AccessibleTableProps<T>) {
  const [internalPage, setInternalPage] = useState(1)

  // Use controlled or uncontrolled page
  const activePage = onPageChange ? currentPage : internalPage

  // Sort data if sort key is provided
  const sortedData = useMemo(() => {
    if (!sortKey) return data

    return [...data].sort((a, b) => {
      const aValue = a[sortKey as keyof T]
      const bValue = b[sortKey as keyof T]

      if (aValue === bValue) return 0
      if (aValue === undefined || aValue === null) return 1
      if (bValue === undefined || bValue === null) return -1

      const comparison = aValue < bValue ? -1 : 1
      return sortDirection === 'asc' ? comparison : -comparison
    })
  }, [data, sortKey, sortDirection])

  // Paginate data
  const paginatedData = useMemo(() => {
    if (pageSize <= 0) return sortedData
    const start = (activePage - 1) * pageSize
    return sortedData.slice(start, start + pageSize)
  }, [sortedData, activePage, pageSize])

  // Calculate total pages
  const totalPages = pageSize > 0 ? Math.ceil(sortedData.length / pageSize) : 1

  // Handle page change
  const handlePageChange = useCallback((newPage: number) => {
    if (onPageChange) {
      onPageChange(newPage)
    } else {
      setInternalPage(newPage)
    }
  }, [onPageChange])

  // Handle sort
  const handleSort = useCallback((key: string) => {
    if (onSort) {
      onSort(key)
    }
  }, [onSort])

  // Generate unique ID for ARIA
  const tableId = id || `table-${Math.random().toString(36).substr(2, 9)}`

  const alignClasses = {
    left: 'text-left',
    center: 'text-center',
    right: 'text-right',
  }

  return (
    <div className={`overflow-x-auto ${className}`}>
      <table
        id={tableId}
        className="min-w-full divide-y divide-gray-700"
        aria-label={ariaLabel || caption}
        role="grid"
      >
        {caption && (
          <caption className="sr-only">
            {caption}
          </caption>
        )}

        {/* Table Header */}
        <thead className="bg-gray-800">
          <tr role="row">
            {columns.map((column) => (
              <th
                key={String(column.key)}
                scope="col"
                aria-sort={sortKey === column.key ? (sortDirection === 'asc' ? 'ascending' : 'descending') : undefined}
                className={`px-4 py-3 text-xs font-medium text-gray-400 uppercase tracking-wider ${
                  column.align ? alignClasses[column.align] : 'text-left'
                } ${column.sortable ? 'cursor-pointer select-none hover:text-white' : ''}`}
                style={{ width: column.width }}
                tabIndex={column.sortable ? 0 : undefined}
                onClick={column.sortable ? () => handleSort(String(column.key)) : undefined}
                onKeyDown={
                  column.sortable
                    ? (e: React.KeyboardEvent) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault()
                          handleSort(String(column.key))
                        }
                      }
                    : undefined
                }
              >
                <span className="flex items-center">
                  {column.header}
                  {column.sortable && (
                    <SortIndicator
                      direction={sortKey === column.key ? sortDirection ?? undefined : undefined}
                    />
                  )}
                </span>
              </th>
            ))}
          </tr>
        </thead>

        {/* Table Body */}
        <tbody className="divide-y divide-gray-700 bg-hades-dark">
          {loading ? (
            <tr>
              <td
                colSpan={columns.length}
                className="px-4 py-8 text-center text-gray-400"
                aria-busy="true"
              >
                <div className="flex items-center justify-center">
                  <div className="w-6 h-6 border-2 border-hades-primary border-t-transparent rounded-full animate-spin mr-2" />
                  Loading...
                </div>
              </td>
            </tr>
          ) : paginatedData.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length}
                className="px-4 py-8 text-center text-gray-400"
                role="status"
                aria-live="polite"
              >
                {emptyMessage}
              </td>
            </tr>
          ) : (
            paginatedData.map((row, rowIndex) => (
              <tr
                key={rowIndex}
                role="row"
                className="hover:bg-gray-800/50 transition-colors"
              >
                {columns.map((column) => {
                  const value = row[column.key as keyof T]
                  return (
                    <td
                      key={String(column.key)}
                      className={`px-4 py-4 whitespace-nowrap text-sm text-gray-300 ${
                        column.align ? alignClasses[column.align] : ''
                      }`}
                    >
                      {column.render
                        ? column.render(value, row, rowIndex)
                        : value !== undefined && value !== null
                        ? String(value)
                        : ''}
                    </td>
                  )
                })}
              </tr>
            ))
          )}
        </tbody>
      </table>

      {/* Pagination */}
      {pageSize > 0 && totalPages > 1 && (
        <nav
          className="flex items-center justify-between px-4 py-3 border-t border-gray-700"
          aria-label="Table pagination"
        >
          <div className="text-sm text-gray-400">
            Showing {Math.min((activePage - 1) * pageSize + 1, sortedData.length)} to{' '}
            {Math.min(activePage * pageSize, sortedData.length)} of {sortedData.length} results
          </div>

          <div className="flex items-center gap-1">
            <button
              onClick={() => handlePageChange(activePage - 1)}
              disabled={activePage === 1}
              className="px-3 py-1 text-sm text-gray-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed rounded transition-colors focus:outline-none focus:ring-2 focus:ring-hades-primary"
              aria-label="Previous page"
            >
              Previous
            </button>

            {/* Page numbers */}
            {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
              let pageNum: number
              if (totalPages <= 5) {
                pageNum = i + 1
              } else if (activePage <= 3) {
                pageNum = i + 1
              } else if (activePage >= totalPages - 2) {
                pageNum = totalPages - 4 + i
              } else {
                pageNum = activePage - 2 + i
              }

              return (
                <button
                  key={pageNum}
                  onClick={() => handlePageChange(pageNum)}
                  className={`px-3 py-1 text-sm rounded transition-colors focus:outline-none focus:ring-2 focus:ring-hades-primary ${
                    activePage === pageNum
                      ? 'bg-hades-primary text-white'
                      : 'text-gray-400 hover:text-white'
                  }`}
                  aria-label={`Page ${pageNum}`}
                  aria-current={activePage === pageNum ? 'page' : undefined}
                >
                  {pageNum}
                </button>
              )
            })}

            <button
              onClick={() => handlePageChange(activePage + 1)}
              disabled={activePage === totalPages}
              className="px-3 py-1 text-sm text-gray-400 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed rounded transition-colors focus:outline-none focus:ring-2 focus:ring-hades-primary"
              aria-label="Next page"
            >
              Next
            </button>
          </div>
        </nav>
      )}
    </div>
  )
}

export default AccessibleTable