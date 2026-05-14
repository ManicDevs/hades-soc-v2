import { useState } from 'react'
import { Minimize2, Maximize2, Bot, Activity } from 'lucide-react'

interface AgentActivityToggleProps {
  children: React.ReactNode
  [key: string]: any
}

const AgentActivityToggle = ({ children, ...props }: AgentActivityToggleProps) => {
  const [isMinimized, setIsMinimized] = useState(true)
  const [showPopup, setShowPopup] = useState(false)

  const toggleMinimize = () => {
    setIsMinimized(!isMinimized)
  }

  const togglePopup = () => {
    setShowPopup(!showPopup)
  }

  if (isMinimized) {
    return (
      <div className="fixed bottom-4 right-4 z-50">
        {/* Minimized floating button */}
        <button
          onClick={togglePopup}
          className="flex items-center gap-2 bg-hades-primary/90 backdrop-blur-sm text-white px-3 py-2 rounded-lg shadow-lg border border-hades-primary/50 hover:bg-hades-primary transition-all duration-200"
        >
          <Activity className="w-4 h-4" />
          <Bot className="w-4 h-4" />
          <span className="text-xs font-medium">Agent Activity</span>
          <Maximize2 className="w-3 h-3" />
        </button>

        {/* Popup modal */}
        {showPopup && (
          <div className="absolute bottom-16 right-0 w-96 max-h-96 bg-hades-darker border border-hades-primary/50 rounded-lg shadow-2xl overflow-hidden">
            {/* Popup header */}
            <div className="flex items-center justify-between p-3 bg-hades-primary/20 border-b border-hades-primary/30">
              <div className="flex items-center gap-2">
                <Bot className="w-4 h-4 text-hades-primary" />
                <span className="text-sm font-medium text-white">Agent Activity</span>
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={toggleMinimize}
                  className="p-1 text-gray-400 hover:text-white transition-colors"
                  title="Restore to panel"
                >
                  <Maximize2 className="w-3 h-3" />
                </button>
                <button
                  onClick={togglePopup}
                  className="p-1 text-gray-400 hover:text-white transition-colors"
                  title="Close popup"
                >
                  ×
                </button>
              </div>
            </div>
            
            {/* Popup content */}
            <div className="max-h-80 overflow-y-auto">
              {children}
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="relative" {...props}>
      {/* Full panel with minimize button */}
      <button
        onClick={toggleMinimize}
        className="absolute top-2 right-2 z-10 p-1 text-gray-400 hover:text-white transition-colors bg-hades-darker/80 rounded"
        title="Minimize to floating button"
      >
        <Minimize2 className="w-3 h-3" />
      </button>
      {children}
    </div>
  )
}

export default AgentActivityToggle
