import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react'
import { io, Socket } from 'socket.io-client'

// Hot reload event types
interface HotReloadEvent {
  type: 'component-update' | 'style-update' | 'config-update'
  componentId?: string
  data?: any
  timestamp: number
}

interface HotReloadContextType {
  isEnabled: boolean
  connected: boolean
  lastUpdate: number
  forceReload: () => void
  updateComponent: (componentId: string, data: any) => void
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
  const [socket, setSocket] = useState<Socket | null>(null)
  const [connected, setConnected] = useState(false)
  const [lastUpdate, setLastUpdate] = useState(Date.now())
  const [isEnabled] = useState(enabled)

  const connectSocket = useCallback(() => {
    if (!enabled || socket) return

    const newSocket = io('ws://localhost:3001', {
      transports: ['websocket'],
      upgrade: false,
      rememberUpgrade: false,
    })

    newSocket.on('connect', () => {
      console.log('🔥 Hot reload connected')
      setConnected(true)
    })

    newSocket.on('disconnect', () => {
      console.log('🔥 Hot reload disconnected')
      setConnected(false)
    })

    newSocket.on('hot-reload', (event: HotReloadEvent) => {
      console.log('🔄 Hot reload event:', event)
      setLastUpdate(event.timestamp)
      
      switch (event.type) {
        case 'component-update':
          if (event.componentId) {
            // Trigger component hot swap without page reload
            const customEvent = new CustomEvent('component-hot-swap', {
              detail: { componentId: event.componentId, data: event.data }
            })
            window.dispatchEvent(customEvent)
          }
          break
          
        case 'style-update':
          // Hot swap CSS without page reload
          const styleElements = document.querySelectorAll('style[data-hot-reload]')
          styleElements.forEach(el => el.remove())
          
          if (event.data) {
            const style = document.createElement('style')
            style.textContent = event.data
            style.setAttribute('data-hot-reload', 'true')
            document.head.appendChild(style)
          }
          break
          
        case 'config-update':
          // Update configuration without page reload
          const configEvent = new CustomEvent('config-hot-update', {
            detail: event.data
          })
          window.dispatchEvent(configEvent)
          break
      }
    })

    newSocket.on('force-reload', () => {
      console.log('🔄 Force reload requested')
      window.location.reload()
    })

    setSocket(newSocket)
  }, [enabled, socket])

  const disconnectSocket = useCallback(() => {
    if (socket) {
      socket.disconnect()
      setSocket(null)
      setConnected(false)
    }
  }, [socket])

  const forceReload = useCallback(() => {
    window.location.reload()
  }, [])

  const updateComponent = useCallback((componentId: string, data: any) => {
    if (socket && connected) {
      socket.emit('component-update', { componentId, data })
    }
  }, [socket, connected])

  // Auto-connect on mount
  React.useEffect(() => {
    if (enabled) {
      connectSocket()
    }

    return () => {
      disconnectSocket()
    }
  }, [enabled, connectSocket, disconnectSocket])

  const contextValue: HotReloadContextType = {
    isEnabled,
    connected,
    lastUpdate,
    forceReload,
    updateComponent
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

  React.useEffect(() => {
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

  React.useEffect(() => {
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
