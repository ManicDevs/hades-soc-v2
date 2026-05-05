import React, { useEffect, useRef, useState } from 'react'

export const AgentThoughtStream = ({ maxItems = 100, className = '' }) => {
  const [events, setEvents] = useState([])
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState(null)
  const wsRef = useRef(null)
  const scrollRef = useRef(null)
  const [autoScroll, setAutoScroll] = useState(true)

  useEffect(() => {
    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws/agent-stream`
    
    const connect = () => {
      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        setIsConnected(true)
        setError(null)
        console.log('Agent Stream WebSocket connected')
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          setEvents((prev) => {
            const newEvents = [data, ...prev]
            return newEvents.slice(0, maxItems)
          })
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }

      ws.onclose = () => {
        setIsConnected(false)
        console.log('Agent Stream WebSocket disconnected')
        setTimeout(connect, 3000)
      }

      ws.onerror = (err) => {
        setError('WebSocket error occurred')
        setIsConnected(false)
        console.error('Agent Stream WebSocket error:', err)
      }
    }

    connect()

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [maxItems])

  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = 0
    }
  }, [events, autoScroll])

  const clearEvents = () => {
    setEvents([])
  }

  const getCategoryColor = (category) => {
    switch (category) {
      case 'recon':
        return 'bg-blue-500/20 border-blue-500/50 text-blue-300'
      case 'thought':
        return 'bg-yellow-500/20 border-yellow-500/50 text-yellow-300'
      case 'action':
        return 'bg-green-500/20 border-green-500/50 text-green-300'
      case 'quantum':
        return 'bg-purple-500/20 border-purple-500/50 text-purple-300'
      case 'critical':
        return 'bg-red-500/20 border-red-500/50 text-red-300'
      case 'remediation':
        return 'bg-orange-500/20 border-orange-500/50 text-orange-300'
      case 'hot_swap':
        return 'bg-pink-500/20 border-pink-500/50 text-pink-300'
      case 'network_containment':
        return 'bg-red-600/30 border-red-600/70 text-red-200 animate-pulse'
      case 'honey_token_trap':
        return 'bg-red-700/40 border-red-700/80 text-red-100 animate-pulse shadow-lg shadow-red-900/50'
      case 'honey_file_accessed':
        return 'bg-amber-600/40 border-amber-600/90 text-amber-100 animate-pulse shadow-lg shadow-amber-900/50'
      case 'honey_file_modified':
        return 'bg-orange-600/40 border-orange-600/90 text-orange-100 animate-pulse shadow-lg shadow-orange-900/50'
      case 'honey_file_rotation':
        return 'bg-green-600/40 border-green-600/90 text-green-100 animate-pulse shadow-lg shadow-green-900/50'
      default:
        return 'bg-gray-500/20 border-gray-500/50 text-gray-300'
    }
  }

  const getCategoryIcon = (category) => {
    switch (category) {
      case 'recon':
        return '🔍'
      case 'thought':
        return '💭'
      case 'action':
        return '⚡'
      case 'quantum':
        return '🔐'
      case 'critical':
        return '🚨'
      case 'remediation':
        return '🔧'
      case 'hot_swap':
        return '🔄'
      case 'network_containment':
        return '🚫'
      case 'honey_token_trap':
        return '🍯'
      case 'honey_file_accessed':
        return '📄'
      case 'honey_file_rotation':
        return '🔄'
      default:
        return '📋'
    }
  }

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp)
    return date.toLocaleTimeString('en-US', { 
      hour12: false, 
      hour: '2-digit', 
      minute: '2-digit', 
      second: '2-digit' 
    })
  }

  return (
    <div className={`hades-card flex flex-col ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
          <h3 className="text-lg font-semibold text-white">Agent Thought Stream</h3>
          <span className="text-xs text-slate-500">
            {isConnected ? 'Live' : 'Disconnected'}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={clearEvents}
            className="text-xs px-2 py-1 rounded bg-slate-700 text-slate-400 hover:bg-slate-600"
          >
            Clear
          </button>
          <span className="text-xs text-slate-500">{events.length} events</span>
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 px-4 py-2 bg-slate-800/40 border-b border-slate-700/50 text-xs flex-wrap">
        <span className="text-slate-500">Categories:</span>
        <span className="flex items-center gap-1 text-blue-400"><span className="w-2 h-2 rounded-full bg-blue-500"></span> Recon</span>
        <span className="flex items-center gap-1 text-yellow-400"><span className="w-2 h-2 rounded-full bg-yellow-500"></span> Logic</span>
        <span className="flex items-center gap-1 text-green-400"><span className="w-2 h-2 rounded-full bg-green-500"></span> Action</span>
        <span className="flex items-center gap-1 text-purple-400"><span className="w-2 h-2 rounded-full bg-purple-500"></span> Quantum</span>
        <span className="flex items-center gap-1 text-red-400"><span className="w-2 h-2 rounded-full bg-red-500"></span> Critical</span>
        <span className="flex items-center gap-1 text-orange-400"><span className="w-2 h-2 rounded-full bg-orange-500"></span> Remediation</span>
        <span className="flex items-center gap-1 text-pink-400"><span className="w-2 h-2 rounded-full bg-pink-500"></span> Hot-Swap</span>
        <span className="flex items-center gap-1 text-red-500 font-bold"><span className="w-2 h-2 rounded-full bg-red-600 animate-pulse"></span> Network Containment</span>
        <span className="flex items-center gap-1 text-red-600 font-bold"><span className="w-2 h-2 rounded-full bg-red-700 animate-pulse shadow-red-500"></span> 🍯 Honey-Token Trap</span>
        <span className="flex items-center gap-1 text-amber-500 font-bold"><span className="w-2 h-2 rounded-full bg-amber-600 animate-pulse shadow-amber-500"></span> 📄 Honey-File Accessed</span>
      <span className="flex items-center gap-1 text-green-500 font-bold"><span className="w-2 h-2 rounded-full bg-green-600 animate-pulse shadow-green-500"></span> 🔄 Trap Rotation</span>
      </div>

      {/* Event Stream */}
      <div 
        ref={scrollRef}
        className="flex-1 overflow-y-auto p-4 space-y-3 max-h-[500px] min-h-[300px]"
        onScroll={(e) => {
          const { scrollTop } = e.currentTarget
          if (scrollTop > 10) {
            setAutoScroll(false)
          }
        }}
      >
        {events.length === 0 && (
          <div className="flex items-center justify-center h-full text-slate-500">
            <div className="text-center">
              <div className="text-4xl mb-2">🤖</div>
              <p className="text-sm">Waiting for agent events...</p>
              <p className="text-xs text-slate-600 mt-1">
                Events will appear here when agents process data
              </p>
            </div>
          </div>
        )}

        {events.map((event, index) => (
          <div
            key={event.id || index}
            className={`p-3 rounded-lg border ${getCategoryColor(event.category)} transition-all hover:opacity-80`}
          >
            <div className="flex items-start gap-3">
              <span className="text-lg">{getCategoryIcon(event.category)}</span>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-mono text-slate-400">
                    {formatTimestamp(event.timestamp)}
                  </span>
                  <span className="text-xs px-2 py-0.5 rounded bg-slate-700/50 text-slate-300">
                    {event.source}
                  </span>
                  {event.severity && (
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      event.severity === 'critical' ? 'bg-red-500/20 text-red-300' :
                      event.severity === 'high' ? 'bg-orange-500/20 text-orange-300' :
                      'bg-blue-500/20 text-blue-300'
                    }`}>
                      {event.severity}
                    </span>
                  )}
                </div>
                
                <p className="text-sm text-white font-medium mb-1">
                  {event.message || event.agent_name}
                </p>
                
                {event.internal_reasoning && (
                  <p className="text-xs text-slate-400 mt-1 italic">
                    💭 {event.internal_reasoning}
                  </p>
                )}

                {event.target && event.target !== 'dashboard' && (
                  <p className="text-xs text-slate-500 mt-1">
                    → Target: {event.target}
                  </p>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {error && (
        <div className="px-4 py-2 bg-red-900/20 border-t border-red-500/50 text-red-400 text-sm">
          Error: {error}
        </div>
      )}
    </div>
  )
}

export default AgentThoughtStream
