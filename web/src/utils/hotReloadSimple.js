// Simple hot reload system using direct WebSocket connection
export class SimpleHotReload {
  constructor() {
    this.ws = null
    this.connected = false
    this.port = 3001
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    try {
      // Connect directly to hot reload server
      this.ws = new WebSocket(`ws://localhost:${this.port}`)
      
      this.ws.onopen = () => {
        console.log('🔥 Hot reload connected')
        this.connected = true
        this.reconnectAttempts = 0
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
        console.log('🔄 Hot reload disconnected')
        this.connected = false
        this.attemptReconnect()
      }

      this.ws.onerror = (error) => {
        console.error('Hot reload WebSocket error:', error)
        this.connected = false
      }
    } catch (error) {
      console.error('Failed to connect to hot reload server:', error)
      this.attemptReconnect()
    }
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 5000)
      console.log(`🔄 Attempting to reconnect in ${delay}ms...`)
      setTimeout(() => this.connect(), delay)
    } else {
      console.error('Max reconnection attempts reached')
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
