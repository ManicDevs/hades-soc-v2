import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react'

const AgentEventContext = createContext(null)

export function useAgentEvents() {
  const context = useContext(AgentEventContext)
  if (!context) {
    throw new Error('useAgentEvents must be used within an AgentEventProvider')
  }
  return context
}

export function AgentEventProvider({ children }) {
  const [events, setEvents] = useState([])
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef(null)
  const eventQueueRef = useRef([])
  const reconnectAttemptsRef = useRef(0)
  const maxEvents = 100

  const connect = useCallback(() => {
    // ENFORCE WSS:// protocol for security compliance
    const protocol = 'wss:' // Always use secure WebSocket per AGENTS.md directives
    const wsUrl = `${protocol}//${window.location.host}/api/v2/ws/events`
    
    const websocket = new WebSocket(wsUrl)
    
    websocket.onopen = () => {
      console.log('AgentEvent WebSocket connected')
      setIsConnected(true)
      reconnectAttemptsRef.current = 0 // Reset reconnection attempts on successful connect
    }
    
    websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        
        const agentEvent = {
          id: data.id || Date.now(),
          type: data.type,
          source: data.source,
          target: data.target,
          payload: data.payload,
          timestamp: data.timestamp || new Date().toISOString(),
          displayTime: new Date().toLocaleTimeString()
        }
        
        eventQueueRef.current = [...eventQueueRef.current, agentEvent].slice(-maxEvents)
        setEvents([...eventQueueRef.current])
      } catch (err) {
        console.error('Failed to parse agent event:', err)
      }
    }
    
    websocket.onclose = () => {
      console.log('AgentEvent WebSocket disconnected')
      setIsConnected(false)
      wsRef.current = null
      
      // EXPONENTIAL BACKOFF for reconnection per AGENTS.md directives
      reconnectAttemptsRef.current++
      const baseDelay = 1000 // 1 second base
      const maxDelay = 30000 // 30 second max
      const exponentialDelay = Math.min(baseDelay * Math.pow(2, reconnectAttemptsRef.current - 1), maxDelay)
      
      console.log(`AgentEvent WebSocket reconnecting in ${exponentialDelay}ms (attempt ${reconnectAttemptsRef.current})`)
      
      setTimeout(() => {
        const newWs = connect()
        if (newWs) wsRef.current = newWs
      }, exponentialDelay)
    }
    
    websocket.onerror = (error) => {
      console.error('AgentEvent WebSocket error:', error)
    }
    
    wsRef.current = websocket
    return websocket
  }, [])

  useEffect(() => {
    const websocket = connect()
    return () => {
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        websocket.close()
      }
    }
  }, [connect])

  const clearEvents = useCallback(() => {
    eventQueueRef.current = []
    setEvents([])
  }, [])

  const getEventsByType = useCallback((type) => {
    return events.filter(e => e.type && e.type.includes(type))
  }, [events])

  const getRecentEvents = useCallback((count = 10) => {
    return events.slice(-count).reverse()
  }, [events])

  const getAgentDecisions = useCallback(() => {
    return events.filter(e => 
      e.type && (
        e.type.includes('agent.decision') || 
        e.type.includes('module.launched') ||
        e.type.includes('threat')
      )
    )
  }, [events])

  const eventTypes = {
    RECON_COMPLETE: 'recon.complete',
    EXPLOITATION_COMPLETE: 'exploitation.complete',
    THREAT_DETECTED: 'threat.detected',
    CRITICAL_THREAT: 'agent.threat.critical',
    AGENT_DECISION: 'agent.decision',
    MODULE_LAUNCHED: 'agent.module.launched',
    MODULE_COMPLETED: 'agent.module.completed',
    NODE_ISOLATED: 'agent.node.isolated',
    PORT_DISCOVERED: 'agent.port.discovered',
    VULNERABILITY_DETECTED: 'agent.vulnerability.detected',
    DOMAIN_FOUND: 'agent.domain.found'
  }

  const value = {
    events,
    isConnected,
    clearEvents,
    getEventsByType,
    getRecentEvents,
    getAgentDecisions,
    eventTypes
  }

  return React.createElement(AgentEventContext.Provider, { value }, children)
}
