import React, { useState, useEffect } from 'react'
import { useHotSwap } from '../context/HotReloadContext'

const HotSwapTest: React.FC = () => {
  const { data, lastUpdate } = useHotSwap('HotSwapTest')
  const [counter, setCounter] = useState(0)

  useEffect(() => {
    if (data) {
      console.log('🔥 HotSwapTest component received hot swap:', data)
    }
  }, [data])

  return (
    <div className="p-6 bg-[var(--bg-secondary)] rounded-lg border border-[var(--border-color)]">
      <h3 className="text-[var(--text-primary)] text-lg font-medium mb-4">
        🔥 Hot Swap Test - UPDATED Component
      </h3>
      
      <div className="space-y-4">
        <div className="text-[var(--text-secondary)]">
          <p>Last hot swap: {lastUpdate ? new Date(lastUpdate).toLocaleTimeString() : 'Never'}</p>
          <p>Hot swap data: {data ? JSON.stringify(data) : 'None'}</p>
        </div>
        
        <div className="flex items-center space-x-4">
          <button
            onClick={() => setCounter(c => c + 1)}
            className="px-4 py-2 bg-[var(--accent-color)] hover:bg-[var(--accent-hover)] text-white rounded transition-colors"
          >
            Counter: {counter}
          </button>
          
          <div className="text-[var(--text-secondary)] text-sm">
            This counter should persist during hot swaps
          </div>
        </div>
      </div>
    </div>
  )
}

export default HotSwapTest
