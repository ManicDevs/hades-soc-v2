import React, { createContext, useContext, useState, useCallback, ReactNode, useEffect } from 'react'
import { ViteHotReload } from '../utils/hotReloadVite' // Import ViteHotReload
// Hot reload event types
interface HotReloadEvent {
  type: 'component-update' | 'style-update' | 'config-update' | 'force-reload'
  componentId?: string
  data?: any
  timestamp: number
}

interface HotReloadContextType {
  isEnabled: boolean
  connected: boolean
  lastUpdate: number
  forceReload: () => void
  // updateComponent: (componentId: string, data: any) => void // Not needed with Vite HMR
}

const HotReloadContext = createContext<HotReloadContextType | null>(null)

export const useHotReload = () => {
  const context = useContext(HotReloadContext)
  if (!context) {
    throw new Error('useHotReload must be used within HotReloadProvider')
  }
  return context
}

interface HotReloadProviderProps {
  children: ReactNode
  enabled?: boolean
}

export const HotReloadProvider: React.FC<HotReloadProviderProps> = ({ 
  children, 
  enabled = process.env.NODE_ENV === 'development' 
}) => {
  const [viteHotReload, setViteHotReload] = useState<ViteHotReload | null>(null)
  const [connected, setConnected] = useState(false)
  const [lastUpdate, setLastUpdate] = useState(Date.now())
  const [isEnabled] = useState(enabled)

  const handleComponentHotSwap = useCallback((event: Event) => {
    const customEvent = event as CustomEvent<{ componentId: string; data: any }>
    if (customEvent.detail.componentId) {
        setLastUpdate(Date.now())
      // You can add more specific handling here if needed
      console.log(`🔥 Hot swapped component: ${customEvent.detail.componentId}`, customEvent.detail.data)
      }
  }, [])

  const handleConfigHotUpdate = useCallback((event: Event) => {
    const customEvent = event as CustomEvent<any>
    setLastUpdate(Date.now())
    console.log('🔥 Hot config update:', customEvent.detail)
  }, [])

  useEffect(() => {
    if (!enabled) return

    const hmr = new ViteHotReload()
    if (hmr.connect()) {
      setViteHotReload(hmr)
      setConnected(true)

      // Listen to custom events dispatched by ViteHotReload
      window.addEventListener('component-hot-swap', handleComponentHotSwap as EventListener)
      window.addEventListener('config-hot-update', handleConfigHotUpdate as EventListener)
    } else {
      console.warn('Vite HMR is not available. Hot reload disabled.')
}

    return () => {
      if (hmr) {
        hmr.disconnect()
      }
      window.removeEventListener('component-hot-swap', handleComponentHotSwap as EventListener)
      window.removeEventListener('config-hot-update', handleConfigHotUpdate as EventListener)
    }
  }, [enabled, handleComponentHotSwap, handleConfigHotUpdate])

  const forceReload = useCallback(() => {
    window.location.reload()
  }, [])

  const contextValue: HotReloadContextType = {
    isEnabled,
    connected,
    lastUpdate,
    forceReload,
  }

  return (
    <HotReloadContext.Provider value={contextValue}>
      {children}
    </HotReloadContext.Provider>
  )
}

// Hook for components to listen for hot swaps
export const useHotSwap = (componentId: string) => {
  const [data, setData] = useState<any>(null)
  const [lastUpdate, setLastUpdate] = useState(Date.now())

  useEffect(() => {
    const handleHotSwap = (event: CustomEvent) => {
      if (event.detail.componentId === componentId) {
        console.log(`🔥 Hot swapping component: ${componentId}`)
        setData(event.detail.data)
        setLastUpdate(Date.now())
      }
    }

    window.addEventListener('component-hot-swap', handleHotSwap as EventListener)
    
    return () => {
      window.removeEventListener('component-hot-swap', handleHotSwap as EventListener)
    }
  }, [componentId])

  return { data, lastUpdate }
}

// Hook for configuration hot updates
export const useHotConfig = () => {
  const [config, setConfig] = useState<any>(null)
  const [lastUpdate, setLastUpdate] = useState(Date.now())

  useEffect(() => {
    const handleConfigUpdate = (event: CustomEvent) => {
      console.log('🔥 Hot config update:', event.detail)
      setConfig(event.detail)
      setLastUpdate(Date.now())
    }

    window.addEventListener('config-hot-update', handleConfigUpdate as EventListener)
    
    return () => {
      window.removeEventListener('config-hot-update', handleConfigUpdate as EventListener)
    }
  }, [])

  return { config, lastUpdate }
}

