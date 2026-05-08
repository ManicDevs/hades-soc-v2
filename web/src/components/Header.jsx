import React, { useState, useRef, useEffect } from 'react'
import { Bell, Search, User, LogOut, Menu, Shield, Settings, ChevronDown } from 'lucide-react'

function Header({ user, onLogout, onToggleSidebar }) {
  const [showNotifications, setShowNotifications] = useState(false)
  const [showUserMenu, setShowUserMenu] = useState(false)
  const notificationsRef = useRef(null)
  const userMenuRef = useRef(null)

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (notificationsRef.current && !notificationsRef.current.contains(event.target)) {
        setShowNotifications(false)
      }
      if (userMenuRef.current && !userMenuRef.current.contains(event.target)) {
        setShowUserMenu(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])
  return (
    <header className="bg-gray-800 border-b border-gray-700 px-4 py-3">
      <div className="flex items-center justify-between">
        {/* Left side */}
        <div className="flex items-center space-x-4">
          <button
            onClick={onToggleSidebar}
            className="text-gray-400 hover:text-white transition-colors"
          >
            <Menu className="w-5 h-5" />
          </button>
          
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search..."
              className="pl-10 pr-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-hades-primary focus:border-transparent w-64"
            />
          </div>
        </div>

        {/* Right side */}
        <div className="flex items-center space-x-4">
          {/* Notifications */}
          <div className="relative" ref={notificationsRef}>
            <button 
              onClick={() => setShowNotifications(!showNotifications)}
              className="relative text-gray-400 hover:text-white transition-colors"
            >
              <Bell className="w-5 h-5" />
              <span className="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full"></span>
            </button>
            
            {showNotifications && (
              <div className="absolute right-0 mt-2 w-80 bg-gray-800 border border-gray-700 rounded-lg shadow-lg z-50">
                <div className="p-4 border-b border-gray-700">
                  <h3 className="text-white font-medium">Notifications</h3>
                </div>
                <div className="max-h-96 overflow-y-auto">
                  <div className="p-4 text-gray-400 text-sm">
                    <p>No new notifications</p>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* User menu */}
          <div className="relative" ref={userMenuRef}>
            <button 
              onClick={() => setShowUserMenu(!showUserMenu)}
              className="flex items-center space-x-3 text-gray-400 hover:text-white transition-colors"
            >
              <div className="text-right">
                <p className="text-white font-medium text-sm">{user?.username || 'Admin'}</p>
                <p className="text-gray-400 text-xs">{user?.role || 'Administrator'}</p>
              </div>
              
              <div className="w-8 h-8 bg-hades-primary rounded-full flex items-center justify-center">
                <User className="w-4 h-4 text-white" />
              </div>
              
              <ChevronDown className="w-4 h-4" />
            </button>
            
            {showUserMenu && (
              <div className="absolute right-0 mt-2 w-48 bg-gray-800 border border-gray-700 rounded-lg shadow-lg z-50">
                <div className="py-2">
                  <a href="/settings" className="flex items-center space-x-2 px-4 py-2 text-gray-300 hover:bg-gray-700 hover:text-white">
                    <Settings className="w-4 h-4" />
                    <span>Settings</span>
                  </a>
                  <button 
                    onClick={onLogout}
                    className="flex items-center space-x-2 w-full px-4 py-2 text-gray-300 hover:bg-gray-700 hover:text-red-400"
                  >
                    <LogOut className="w-4 h-4" />
                    <span>Logout</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}

export default Header
