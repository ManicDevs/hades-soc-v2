import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react'

interface AgentEvent {
  events: any[]
  isConnected: boolean
  clearEvents: () => void
  getEventsByType: (type: string) => any[]
  getRecentEvents: (count?: number) => any[]
  getAgentDecisions: () => any[]
  eventTypes: Record<string, string>
}

const AgentEventContext = createContext<AgentEvent | null>(null)

export function useAgentEvents(): AgentEvent {
  const context = useContext(AgentEventContext)
  if (context === null) {
    throw new Error('useAgentEvents must be used within an AgentEventProvider')
  }
  return context
}

export function AgentEventProvider({ children }: { children: React.ReactNode }) {
  const [events, setEvents] = useState<any[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const eventQueueRef = useRef<any[]>([])
  const reconnectAttemptsRef = useRef(0)
  const maxEvents = 100

  const connect = useCallback(() => {
    // Use backend port 8080 for WebSocket connection
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//192.168.0.2:8080/api/v2/ws/events`
    
    const websocket = new WebSocket(wsUrl)
    
    websocket.onopen = () => {
        setIsConnected(true)
        reconnectAttemptsRef.current = 0
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
      setIsConnected(false)
      wsRef.current = null
      
      // Exponential backoff for reconnection
      reconnectAttemptsRef.current++
      const baseDelay = 1000
      const maxDelay = 30000
      const exponentialDelay = Math.min(baseDelay * Math.pow(2, reconnectAttemptsRef.current - 1), maxDelay)
      
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
    
    // Generate mock events for development if no real events are received
    const mockEventInterval = setInterval(() => {
      if (events.length === 0 || Math.random() > 0.7) {
        const mockEvents = [
          {
            id: Date.now(),
            type: 'agent.decision',
            source: 'hades-agent',
            target: '192.168.1.100',
            payload: { action: 'scan', reason: 'Suspicious activity detected' },
            timestamp: new Date().toISOString(),
            displayTime: new Date().toLocaleTimeString()
          },
          {
            id: Date.now() + 1,
            type: 'log.event',
            source: 'hades-agent',
            payload: { reasoning: 'Analyzing network traffic patterns for anomalies', rule_name: 'traffic_analysis' },
            timestamp: new Date().toISOString(),
            displayTime: new Date().toLocaleTimeString()
          },
          {
            id: Date.now() + 2,
            type: 'module.launched',
            source: 'hades-agent',
            target: 'network-scanner',
            payload: { module: 'port_scanner', status: 'active' },
            timestamp: new Date().toISOString(),
            displayTime: new Date().toLocaleTimeString()
          },
          {
            id: Date.now() + 3,
            type: 'threat.detected',
            source: 'hades-agent',
            target: 'malware-sample.exe',
            payload: { severity: 'medium', signature: 'Trojan.Generic' },
            timestamp: new Date().toISOString(),
            displayTime: new Date().toLocaleTimeString()
          }
        ]
        
        const randomEvent = mockEvents[Math.floor(Math.random() * mockEvents.length)]
        eventQueueRef.current = [...eventQueueRef.current, randomEvent].slice(-maxEvents)
        setEvents([...eventQueueRef.current])
      }
    }, 3000) // Generate event every 3 seconds
    
    return () => {
      if (websocket && websocket.readyState === WebSocket.OPEN) {
        websocket.close()
      }
      clearInterval(mockEventInterval)
    }
  }, [connect, events.length])

  const clearEvents = useCallback(() => {
    eventQueueRef.current = []
    setEvents([])
  }, [])

  const getEventsByType = useCallback((type: string) => {
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
