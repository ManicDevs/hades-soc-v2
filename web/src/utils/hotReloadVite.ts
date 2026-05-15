export class ViteHotReload {
  constructor() {
    this.connected = false
    // this.listeners = new Map() // No longer needed
  }

  connect(): boolean {
    // Vite HMR is automatically available in development
    if (import.meta.hot) {
      this.connected = true
      console.log('🔥 Vite HMR connected')
      
      // Listen for HMR updates
      import.meta.hot.on('vite:beforeUpdate', (payload: { updates: Array<{ type: string; path: string }> }) => {
        console.log('🔄 Vite HMR update:', payload)
        this.handleHMRUpdate(payload)
      })

      // Listen for HMR errors
      import.meta.hot.on('vite:error', (err: any) => {
        console.error('🔥 Vite HMR error:', err)
      })

      // Listen for HMR disconnects
      import.meta.hot.on('vite:invalidate', () => {
        console.log('🔄 Vite HMR invalidated')
      })

      return true
    }
    
    return false
  }

  handleHMRUpdate(payload: { updates: Array<{ type: string; path: string }> }): void {
    // Broadcast custom hot swap events for components
    payload.updates?.forEach(update => {
      const componentId = this.getComponentId(update.path)
      if (componentId) {
        this.broadcastComponentUpdate(componentId, {
          type: 'hmr',
          path: update.path,
          timestamp: Date.now()
        })
      }
    })
  }

  getComponentId(filePath: string): string | null {
    // Extract component name from file path
    const parts = filePath.split('/')
    const fileName = parts[parts.length - 1]
    return fileName?.replace(/\.(tsx|jsx|ts|js)$/, '') || null
  }

  broadcastComponentUpdate(componentId: string, data: any): void {
    const event = new CustomEvent('component-hot-swap', {
      detail: { componentId, data }
    })
    window.dispatchEvent(event)
  }

  // Add custom event listener (no longer needed)
  // addEventListener(eventType: string, callback: EventListener): void {
  //   this.listeners.set(eventType, callback)
  //   window.addEventListener(eventType, callback)
  // }

  // Remove custom event listener (no longer needed)
  // removeEventListener(eventType: string): void {
  //   const callback = this.listeners.get(eventType)
  //   if (callback) {
  //     window.removeEventListener(eventType, callback)
  //     this.listeners.delete(eventType)
  //   }
  // }

  disconnect(): void {
    this.connected = false
    // this.listeners.clear() // No longer needed
  }

  isConnected(): boolean {
    return this.connected
  }
}
