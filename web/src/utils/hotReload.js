// Simple hot reload system for development
export class SimpleHotReload {
  constructor() {
    this.ws = null
    this.connected = false
    this.port = 3001
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    try {
      this.ws = new WebSocket(`ws://localhost:${this.port}`)
      
      this.ws.onopen = () => {
        console.log('🔥 Hot reload connected')
        this.connected = true
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.handleMessage(data)
        } catch (error) {
          console.error('Failed to parse hot reload message:', error)
        }
      }

      this.ws.onclose = () => {
        console.log('🔥 Hot reload disconnected')
        this.connected = false
      }

      this.ws.onerror = (error) => {
        console.error('Hot reload WebSocket error:', error)
        this.connected = false
      }
    } catch (error) {
      console.error('Failed to connect to hot reload server:', error)
    }
  }

  handleMessage(data) {
    console.log('🔄 Hot reload message:', data)
    
    switch (data.type) {
      case 'component-update':
        if (data.componentId) {
          this.broadcastComponentUpdate(data.componentId, data.data)
        }
        break
        
      case 'style-update':
        this.broadcastStyleUpdate(data.data)
        break
        
      case 'config-update':
        this.broadcastConfigUpdate(data.data)
        break
        
      case 'force-reload':
        window.location.reload()
        break
    }
  }

  broadcastComponentUpdate(componentId, data) {
    const event = new CustomEvent('component-hot-swap', {
      detail: { componentId, data }
    })
    window.dispatchEvent(event)
  }

  broadcastStyleUpdate(data) {
    const event = new CustomEvent('style-hot-update', {
      detail: data
    })
    window.dispatchEvent(event)
  }

  broadcastConfigUpdate(data) {
    const event = new CustomEvent('config-hot-update', {
      detail: data
    })
    window.dispatchEvent(event)
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.connected = false
  }

  isConnected() {
    return this.connected
  }
}
