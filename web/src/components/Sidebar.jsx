import React from 'react'
import { useLocation } from 'react-router-dom'
import { Shield, Activity, AlertTriangle, Users, Settings, Home, Lock, Database, Globe, Code, Brain, BarChart3, Target, Hash, Atom, Server, FileText, ShieldCheck, CheckCircle } from 'lucide-react'

function Sidebar({ isOpen, onToggle, user }) {
  const location = useLocation()
  const menuItems = [
    { icon: Home, label: 'Dashboard', path: '/dashboard' },
    { icon: Shield, label: 'Security', path: '/security' },
    { icon: CheckCircle, label: 'Approval Queue', path: '/approval-queue' },
    { icon: Lock, label: 'Threats', path: '/threats' },
    { icon: Users, label: 'Users', path: '/users' },
    { icon: Database, label: 'Database', path: '/database' },
    { icon: Globe, label: 'Network', path: '/network' },
    
    // Strategic Enhancement Menu Items
    { icon: Brain, label: 'AI Threat Intelligence', path: '/threat-intelligence' },
    { icon: BarChart3, label: 'Advanced Analytics', path: '/analytics' },
    { icon: Target, label: 'Threat Hunting', path: '/threat-hunting' },
    { icon: Hash, label: 'Blockchain Audit', path: '/blockchain' },
    { icon: ShieldCheck, label: 'Zero Trust', path: '/zero-trust' },
    { icon: Atom, label: 'Quantum Crypto', path: '/quantum' },
    { icon: Server, label: 'SIEM', path: '/siem' },
    { icon: FileText, label: 'Incident Response', path: '/incident-response' },
    { icon: Target, label: 'Threat Modeling', path: '/threat-modeling' },
    { icon: Server, label: 'Kubernetes', path: '/kubernetes' },
    
    // Dev access menu items - only show for dev users
    ...(user?.role === 'Developer' || user?.permissions?.includes('dev') ? [
      { icon: Code, label: 'Dev Access', path: '/dev' }
    ] : []),
    { icon: Settings, label: 'Settings', path: '/settings' },
  ]

  return (
    <div className={`${isOpen ? 'w-64' : 'w-20'} bg-hades-darker border-r border-gray-800 transition-all duration-300 flex flex-col`}>
      {/* Logo */}
      <div className="p-4 border-b border-gray-800">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-hades-primary rounded-lg flex items-center justify-center">
            <Shield className="w-6 h-6 text-white" />
          </div>
          {isOpen && (
            <div>
              <h1 className="text-white font-bold text-lg">Hades Toolkit</h1>
              <p className="text-gray-400 text-xs">Enterprise Security</p>
            </div>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4">
        <ul className="space-y-2">
          {menuItems.map((item, index) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <li key={index}>
                <a
                  href={item.path}
                  className={`flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors ${
                    isActive
                      ? 'bg-hades-primary text-white'
                      : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                  }`}
                >
                  <Icon className="w-5 h-5 flex-shrink-0" />
                  {isOpen && <span className="text-sm font-medium">{item.label}</span>}
                  {item.label === 'Dev Access' && (
                    <span className="ml-auto px-2 py-1 bg-hades-primary/20 text-hades-primary rounded text-xs">
                      DEV
                    </span>
                  )}
                </a>
              </li>
            )
          })}
        </ul>
      </nav>

      {/* Toggle Button */}
      <div className="p-4 border-t border-gray-800">
        <button
          onClick={onToggle}
          className="w-full flex items-center justify-center space-x-2 px-3 py-2 text-gray-400 hover:bg-gray-800 hover:text-white rounded-lg transition-colors"
        >
          <Settings className="w-5 h-5" />
          {isOpen && <span className="text-sm">Collapse</span>}
        </button>
      </div>
    </div>
  )
}

export default Sidebar
