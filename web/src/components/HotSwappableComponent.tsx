import React, { useEffect, useState, ReactNode } from 'react'
import { useHotSwap } from '../context/HotReloadContext'

interface HotSwappableComponentProps {
  componentId: string
  children: ReactNode
  fallback?: ReactNode
  onHotSwap?: (data: any) => void
}

const HotSwappableComponent: React.FC<HotSwappableComponentProps> = ({ 
  componentId, 
  children, 
  fallback,
  onHotSwap 
}) => {
  const { data, lastUpdate } = useHotSwap(componentId)
  const [isHotSwapping, setIsHotSwapping] = useState(false)

  useEffect(() => {
    if (data && onHotSwap) {
      console.log(`🔥 Hot swapping component ${componentId}`)
      setIsHotSwapping(true)
      
      try {
        onHotSwap(data)
      } catch (error) {
        console.error('Error during hot swap:', error)
      } finally {
        // Small delay to ensure smooth transition
        setTimeout(() => {
          setIsHotSwapping(false)
        }, 100)
      }
    }
  }, [data, componentId, onHotSwap])

  if (isHotSwapping) {
    return (
      <div className="relative">
        <div className="absolute inset-0 bg-[var(--bg-secondary)]/80 flex items-center justify-center z-50">
          <div className="text-center">
            <div className="w-8 h-8 border-[3px] border-[var(--accent-color)] border-t-transparent rounded-full animate-spin mx-auto mb-2" />
            <p className="text-[var(--text-primary)] text-sm">Hot Swapping...</p>
          </div>
        </div>
        {fallback || children}
      </div>
    )
  }

  return (
    <div 
      className="hot-swappable-component" 
      data-component-id={componentId}
      data-last-update={lastUpdate}
    >
      {children}
    </div>
  )
}

export default HotSwappableComponent
