import { useState } from 'react'
import { useAgentEvents } from '../context/AgentEventContext'

const AgentActivityPanel = () => {
  const { 
    events, 
    isConnected, 
    clearEvents
  } = useAgentEvents()
  
  const [activeTab, setActiveTab] = useState('activity')
  const [filter, setFilter] = useState('all')

  const eventsArray = (events as any[]) || []
  const recentEvents = eventsArray.slice(-20).reverse()
  const decisions = eventsArray.filter((e: any) => e.type.includes('decision'))
  const monologueEvents = eventsArray.filter((e: any) => e.type.includes('log.event'))

  const getEventIcon = (type: string) => {
    if (type.includes('log.event')) return '💭'
    if (type.includes('decision')) return '🧠'
    if (type.includes('module.launched')) return '🚀'
    if (type.includes('module.completed')) return '✅'
    if (type.includes('threat.critical')) return '🚨'
    if (type.includes('threat.detected')) return '⚠️'
    if (type.includes('port.discovered')) return '🔍'
    if (type.includes('vulnerability')) return '💉'
    if (type.includes('domain.found')) return '🌐'
    if (type.includes('node.isolated')) return '🔒'
    if (type.includes('recon.complete')) return '📡'
    if (type.includes('exploitation')) return '💥'
    return '📋'
  }

  const getEventColor = (type: string) => {
    if (type.includes('log.event')) return 'border-cyan-500 bg-cyan-500/10'
    if (type.includes('critical')) return 'border-red-500 bg-red-500/10'
    if (type.includes('decision')) return 'border-purple-500 bg-purple-500/10'
    if (type.includes('module.launched')) return 'border-green-500 bg-green-500/10'
    if (type.includes('threat')) return 'border-orange-500 bg-orange-500/10'
    if (type.includes('vulnerability')) return 'border-yellow-500 bg-yellow-500/10'
    if (type.includes('isolated')) return 'border-red-400 bg-red-400/10'
    return 'border-gray-600 bg-gray-600/10'
  }

  const formatPayload = (payload: any) => {
    if (!payload) return ''
    if (typeof payload === 'string') return payload
    
    const entries = Object.entries(payload).slice(0, 4)
    return entries.map(([k, v]) => `${k}: ${typeof v === 'object' ? JSON.stringify(v).slice(0, 30) : v}`).join(', ')
  }

  const filteredEvents = filter === 'all' 
    ? recentEvents 
    : eventsArray.filter((e: any) => e.type.includes(filter)).slice(-20).reverse()

  return (
    <div className="w-full h-full bg-hades-dark/95 backdrop-blur border border-hades-primary/30 rounded-lg shadow-2xl overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 bg-hades-primary/20 border-b border-hades-primary/30">
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}></span>
          <span className="text-sm font-medium text-white">Agent Activity</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-gray-400">{eventsArray.length} events</span>
          <button 
            onClick={() => {
              const parent = document.querySelector('[data-agent-toggle="true"]')
              if (parent) {
                const toggleBtn = parent.querySelector('button[title="Minimize to floating button"]') as HTMLButtonElement
                if (toggleBtn) toggleBtn.click()
              }
            }}
            className="p-1 text-gray-400 hover:text-white transition-colors"
            title="Minimize to floating button"
          >
            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 3H5a2 2 0 0 0-2 2v3m0 0V5a2 2 0 0 1 2-2h3m0 0V5a2 2 0 0 1 2 2h3m0 0V5a2 2 0 0 1 2-2h3m0 0V5a2 2 0 0 1 2 2h3m0 0V5a2 2 0 0 1 2-2h3m0 0V5a2 2 0 0 1 2 2h3" />
            </svg>
          </button>
          <button 
            onClick={clearEvents}
            className="text-xs text-gray-400 hover:text-white transition-colors"
          >
            Clear
          </button>
        </div>
      </div>

      <div className="flex border-b border-hades-primary/20">
        <button
          onClick={() => setActiveTab('activity')}
          className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
            activeTab === 'activity' 
              ? 'text-hades-primary border-b-2 border-hades-primary' 
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Live Feed
        </button>
        <button
          onClick={() => setActiveTab('monologue')}
          className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
            activeTab === 'monologue' 
              ? 'text-cyan-400 border-b-2 border-cyan-400' 
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Monologue ({monologueEvents.length})
        </button>
        <button
          onClick={() => setActiveTab('decisions')}
          className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
            activeTab === 'decisions' 
              ? 'text-hades-primary border-b-2 border-hades-primary' 
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Decisions ({decisions.length})
        </button>
      </div>

      <div className="p-2 overflow-y-auto max-h-[350px]">
        {activeTab === 'activity' && (
          <div className="space-y-1">
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className="w-full mb-2 px-2 py-1 text-xs bg-hades-dark border border-hades-primary/30 rounded text-white"
            >
              <option value="all">All Events</option>
              <option value="log.event">Monologue</option>
              <option value="agent.decision">Decisions</option>
              <option value="module">Modules</option>
              <option value="threat">Threats</option>
              <option value="port">Discovery</option>
            </select>
            
            {filteredEvents.length === 0 ? (
              <div className="text-center py-8 text-gray-500 text-sm">
                {isConnected ? 'Waiting for events...' : 'Connecting to agent...'}
              </div>
            ) : (
              filteredEvents.map((event) => (
                <div
                  key={event.id}
                  className={`p-2 rounded border-l-2 ${getEventColor(event.type)}`}
                >
                  <div className="flex items-start gap-2">
                    <span className="text-sm">{getEventIcon(event.type)}</span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-medium text-white truncate">
                          {event.type.replace('agent.', '').replace('recon.', '').replace('exploitation.', '')}
                        </span>
                        <span className="text-xs text-gray-500">{event.displayTime}</span>
                      </div>
                      {event.target && (
                        <div className="text-xs text-gray-400 truncate">
                          Target: {event.target}
                        </div>
                      )}
                      {event.payload && (
                        <div className="text-xs text-gray-500 truncate mt-1">
                          {formatPayload(event.payload)}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {activeTab === 'decisions' && (
          <div className="space-y-2">
            {decisions.length === 0 ? (
              <div className="text-center py-8 text-gray-500 text-sm">
                No agent decisions yet
              </div>
            ) : (
              decisions.slice(0, 10).map((decision) => (
                <div
                  key={decision.id}
                  className="p-3 rounded border border-purple-500/30 bg-purple-500/10"
                >
                  <div className="flex items-center gap-2 mb-1">
                    <span>🧠</span>
                    <span className="text-xs font-medium text-purple-400">Agent Decision</span>
                  </div>
                  <div className="text-xs text-white mb-1">
                    {decision.type.replace('agent.', '')}
                  </div>
                  {decision.target && (
                    <div className="text-xs text-gray-400">
                      Target: {decision.target}
                    </div>
                  )}
                  {decision.payload?.action && (
                    <div className="text-xs text-green-400 mt-1">
                      Action: {decision.payload.action}
                    </div>
                  )}
                  {decision.payload?.reason && (
                    <div className="text-xs text-gray-500 mt-1">
                      Reason: {decision.payload.reason}
                    </div>
                  )}
                  <div className="text-xs text-gray-600 mt-1">
                    {decision.displayTime}
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {activeTab === 'monologue' && (
          <div className="space-y-2">
            {monologueEvents.length === 0 ? (
              <div className="text-center py-8 text-gray-500 text-sm">
                Agent&apos;s internal monologue will appear here...
              </div>
            ) : (
              monologueEvents.slice(0, 15).reverse().map((event) => (
                <div
                  key={event.id}
                  className="p-3 rounded border border-cyan-500/30 bg-cyan-500/10"
                >
                  <div className="flex items-center gap-2 mb-2">
                    <span>💭</span>
                    <span className="text-xs font-medium text-cyan-400">Agent Thinking</span>
                    <span className="text-xs text-gray-500 ml-auto">{event.displayTime}</span>
                  </div>
                  {event.payload?.reasoning && (
                    <div className="text-xs text-cyan-100 leading-relaxed">
                      {event.payload.reasoning}
                    </div>
                  )}
                  {event.payload?.rule_name && (
                    <div className="text-xs text-gray-500 mt-2">
                      Rule: {event.payload.rule_name}
                    </div>
                  )}
                  {event.payload?.target && (
                    <div className="text-xs text-gray-400">
                      Target: {event.payload.target}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export default AgentActivityPanel
