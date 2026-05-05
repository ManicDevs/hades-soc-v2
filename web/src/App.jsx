import React, { useState } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Sidebar from './components/Sidebar'
import Header from './components/Header'
import ApprovalQueue from './components/ApprovalQueue'
import Dashboard from './pages/Dashboard'
import Login from './pages/Login'
import Security from './pages/Security'
import Threats from './pages/Threats'
import Users from './pages/Users'
import Settings from './pages/Settings'
import Dev from './pages/Dev'
import ThreatIntelligence from './pages/ThreatIntelligence'
import Analytics from './pages/Analytics'
import ThreatHunting from './pages/ThreatHunting'
import Blockchain from './pages/Blockchain'
import ZeroTrust from './pages/ZeroTrust'
import Quantum from './pages/Quantum'
import SIEM from './pages/SIEM'
import IncidentResponse from './pages/IncidentResponse'
import ThreatModeling from './pages/ThreatModeling'
import Kubernetes from './pages/Kubernetes'
import { AuthProvider, useAuth } from './hooks/useAuth'
import { AgentEventProvider } from './context/AgentEventContext'
import AgentActivityPanel from './components/AgentActivityPanel'

function AppContent() {
  const { user, isAuthenticated, loading, login, logout } = useAuth()
  const [sidebarOpen, setSidebarOpen] = useState(true)

  // Check if we should show login page or dashboard
  const isDevelopmentEnvironment = window.location.hostname === 'localhost' ||
                                  window.location.hostname === '127.0.0.1' ||
                                  window.location.hostname.includes('dev') ||
                                  window.location.hostname.includes('test') ||
                                  window.location.hostname.includes('qa') ||
                                  window.location.hostname.includes('staging')

  // For development environments, we still use authentication but allow easier access
  const currentUser = user
  const isAuth = isAuthenticated

  // Show loading spinner while checking authentication
  if (loading) {
    return (
      <div className="min-h-screen bg-hades-dark flex items-center justify-center">
        <div className="text-white text-center">
          <div className="w-8 h-8 border-2 border-hades-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p>Loading...</p>
        </div>
      </div>
    )
  }

  // Show login page if not authenticated
  if (!isAuth) {
    return <Login onLogin={login} />
  }

  // Show dashboard if authenticated
  return (
    <div className="min-h-screen bg-hades-dark flex">
      <Sidebar isOpen={sidebarOpen} onToggle={() => setSidebarOpen(!sidebarOpen)} user={currentUser} />
      
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header 
          user={currentUser} 
          onLogout={logout} 
          onToggleSidebar={() => setSidebarOpen(!sidebarOpen)}
        />
        
        <main className="flex-1 overflow-y-auto">
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<Dashboard user={currentUser} />} />
            <Route path="/security" element={<Security user={currentUser} />} />
            <Route path="/approval-queue" element={<ApprovalQueue user={currentUser} />} />
            <Route path="/threats" element={<Threats user={currentUser} />} />
            <Route path="/users" element={<Users user={currentUser} />} />
            <Route path="/dev" element={<Dev user={currentUser} />} />
            <Route path="/settings" element={<Settings user={currentUser} />} />
            
            {/* Strategic Enhancement Pages */}
            <Route path="/threat-intelligence" element={<ThreatIntelligence user={currentUser} />} />
            <Route path="/analytics" element={<Analytics user={currentUser} />} />
            <Route path="/threat-hunting" element={<ThreatHunting user={currentUser} />} />
            <Route path="/blockchain" element={<Blockchain user={currentUser} />} />
            <Route path="/zero-trust" element={<ZeroTrust user={currentUser} />} />
            <Route path="/quantum" element={<Quantum user={currentUser} />} />
            <Route path="/siem" element={<SIEM user={currentUser} />} />
            <Route path="/incident-response" element={<IncidentResponse user={currentUser} />} />
            <Route path="/threat-modeling" element={<ThreatModeling user={currentUser} />} />
            <Route path="/kubernetes" element={<Kubernetes user={currentUser} />} />
            
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </main>
      </div>
    </div>
  )
}

function App() {
  return (
    <AuthProvider>
      <AgentEventProvider>
        <AppContent />
        <AgentActivityPanel />
      </AgentEventProvider>
    </AuthProvider>
  )
}

export default App
