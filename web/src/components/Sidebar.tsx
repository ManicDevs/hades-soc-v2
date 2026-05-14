import React from 'react'
import { useLocation } from 'react-router-dom'
import {
  Shield,
  Users,
  Settings,
  Home,
  Lock,
  Database,
  Globe,
  Code,
  Brain,
  BarChart3,
  Target,
  Hash,
  Atom,
  Server,
  FileText,
  ShieldCheck,
  CheckCircle,
  Network,
  Wifi,
} from 'lucide-react'

interface MenuItem {
  icon: any
  label: string
  path: string
  badge?: string
}

function Sidebar({ isOpen, onToggle, user }: { isOpen: boolean; onToggle: () => void; user: any }) {
  const location = useLocation()

  const menuItems: MenuItem[] = [
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
    { icon: Network, label: 'Peer Network', path: '/peer-network' },
    { icon: Wifi, label: 'Tor Network', path: '/tor-network' },

    // Dev access menu items - only show for dev users
    ...(user?.role === 'Developer' || user?.permissions?.includes('dev')
      ? [{ icon: Code, label: 'Dev Access', path: '/dev', badge: 'DEV' }]
      : []),
    { icon: Settings, label: 'Settings', path: '/settings' },
  ]

  const handleKeyDown = (e: React.KeyboardEvent, _item: MenuItem, index: number) => {
    const items = menuItems
    let nextIndex: number

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        nextIndex = index < items.length - 1 ? index + 1 : 0
        document.getElementById(`sidebar-link-${nextIndex}`)?.focus()
        break
      case 'ArrowUp':
        e.preventDefault()
        nextIndex = index > 0 ? index - 1 : items.length - 1
        document.getElementById(`sidebar-link-${nextIndex}`)?.focus()
        break
      case 'Home':
        e.preventDefault()
        document.getElementById('sidebar-link-0')?.focus()
        break
      case 'End':
        e.preventDefault()
        document.getElementById(`sidebar-link-${items.length - 1}`)?.focus()
        break
    }
  }

  return (
    <aside
      className={`${isOpen ? 'w-64' : 'w-20'} bg-hades-darker border-r border-gray-800 transition-all duration-300 flex flex-col`}
      role="navigation"
      aria-label="Main navigation"
    >
      {/* Logo */}
      <div className="p-4 border-b border-gray-800">
        <div className="flex items-center space-x-3">
          <div
            className="w-10 h-10 bg-hades-primary rounded-lg flex items-center justify-center"
            role="img"
            aria-label="Hades Toolkit logo"
          >
            <Shield className="w-6 h-6 text-white" aria-hidden="true" />
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
      <nav className="flex-1 p-4" role="menubar" aria-label="Sidebar navigation">
        <ul className="space-y-2" role="none">
          {menuItems.map((item, index) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <li key={index} role="none">
                <a
                  id={`sidebar-link-${index}`}
                  href={item.path}
                  role="menuitem"
                  aria-current={isActive ? 'page' : undefined}
                  aria-label={item.label}
                  className={`flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-hades-primary focus:ring-inset ${
                    isActive
                      ? 'bg-hades-primary text-white'
                      : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                  }`}
                  tabIndex={isActive ? 0 : -1}
                  onKeyDown={(e) => handleKeyDown(e, item, index)}
                >
                  <Icon
                    className="w-5 h-5 flex-shrink-0"
                    aria-hidden="true"
                  />
                  {isOpen && (
                    <span className="text-sm font-medium">{item.label}</span>
                  )}
                  {item.badge && (
                    <span
                      className="ml-auto px-2 py-1 bg-hades-primary/20 text-hades-primary rounded text-xs"
                      aria-label={`${item.label} - Development access`}
                    >
                      {item.badge}
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
          className="w-full flex items-center justify-center space-x-2 px-3 py-2 text-gray-400 hover:bg-gray-800 hover:text-white rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-hades-primary"
          aria-label={isOpen ? 'Collapse sidebar' : 'Expand sidebar'}
          aria-expanded={isOpen}
        >
          <Settings className="w-5 h-5" aria-hidden="true" />
          {isOpen && <span className="text-sm">Collapse</span>}
        </button>
      </div>
    </aside>
  )
}

export default Sidebar