import React, { useState, lazy, Suspense, Component, ReactNode, useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Sidebar from './components/Sidebar'
import Header from './components/Header'
import { useAuth } from './hooks/useAuth'
import { AuthProvider } from './hooks/useAuth'
import { AgentEventProvider } from './context/AgentEventContext'
import { ThemeProvider } from './context/ThemeContext'
import { HotReloadProvider } from './context/HotReloadContext'
import PageLoader from './components/PageLoader'
import { SuspenseFallback } from './components/SuspenseFallback'
import ViteHotReload from './utils/hotReloadVite.js'

// Lazy loaded page components
const Dashboard = lazy(() => import('./pages/Dashboard'))
const Login = lazy(() => import('./pages/Login'))
const Security = lazy(() => import('./pages/Security'))
const Threats = lazy(() => import('./pages/Threats'))
const Users = lazy(() => import('./pages/Users'))
const Settings = lazy(() => import('./pages/Settings'))
const Dev = lazy(() => import('./pages/Dev'))
const ThreatIntelligence = lazy(() => import('./pages/ThreatIntelligence'))
const Analytics = lazy(() => import('./pages/Analytics'))
const ThreatHunting = lazy(() => import('./pages/ThreatHunting'))
const Blockchain = lazy(() => import('./pages/Blockchain'))
const ZeroTrust = lazy(() => import('./pages/ZeroTrust'))
const Quantum = lazy(() => import('./pages/Quantum'))
const SIEM = lazy(() => import('./pages/SIEM'))
const IncidentResponse = lazy(() => import('./pages/IncidentResponse'))
const ThreatModeling = lazy(() => import('./pages/ThreatModeling'))
const Kubernetes = lazy(() => import('./pages/Kubernetes'))
const PeerNetwork = lazy(() => import('./pages/PeerNetwork'))
const TorNetwork = lazy(() => import('./pages/TorNetwork'))

// Lazy loaded components
const ApprovalQueue = lazy(() => import('./components/ApprovalQueue'))
const AgentActivityPanel = lazy(() => import('./components/AgentActivityPanel'))
const AgentActivityToggle = lazy(() => import('./components/AgentActivityToggle'))

// Error Boundary Component
interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
}

class AppErrorBoundary extends Component<{ children: ReactNode }, ErrorBoundaryState> {
  constructor(props: { children: ReactNode }) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error }
  }

  resetError = (): void => {
    this.setState({ hasError: false, error: null })
  }

  render(): ReactNode {
    if (this.state.hasError) {
      return (
        <div className="flex flex-col items-center justify-center min-h-[200px] p-6 bg-[var(--bg-secondary)] rounded-lg border border-[var(--border-color)]">
          <div className="text-red-400 text-4xl mb-4">⚠️</div>
          <h3 className="text-[var(--text-primary)] text-lg font-medium mb-2">Something went wrong</h3>
          <p className="text-[var(--text-secondary)] text-sm mb-4">
            {this.state.error?.message || 'An unexpected error occurred'}
          </p>
          <button
            onClick={this.resetError}
            className="px-4 py-2 bg-[var(--accent-color)] hover:bg-[var(--accent-hover)] text-white text-sm rounded transition-colors"
          >
            Try Again
          </button>
        </div>
      )
    }

    return this.props.children
  }
}

// Page wrapper with error boundary and suspense
interface PageWrapperProps {
  children: ReactNode
  fallback?: ReactNode
}

const PageWrapper: React.FC<PageWrapperProps> = ({ children, fallback }) => (
  <AppErrorBoundary>
    <Suspense fallback={fallback || <PageLoader />}>
      {children}
    </Suspense>
  </AppErrorBoundary>
)

function AppContent() {
  const { user, isAuthenticated, loading, login, logout } = useAuth()
  const [sidebarOpen, setSidebarOpen] = useState(true)

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--bg-primary)] flex items-center justify-center">
        <div className="text-center">
          <div className="w-10 h-10 border-[3px] border-[var(--accent-color)] border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-[var(--text-secondary)]">Loading...</p>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <Suspense fallback={<PageLoader message="Loading authentication..." />}>
        <Login onLogin={login} />
      </Suspense>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--bg-primary)] flex theme-transition">
      <Sidebar isOpen={sidebarOpen} onToggle={() => setSidebarOpen(!sidebarOpen)} user={user} />

      <div className="flex-1 flex flex-col overflow-hidden">
        <Header
          user={user}
          onLogout={logout}
          onToggleSidebar={() => setSidebarOpen(!sidebarOpen)}
        />

        <main className="flex-1 overflow-y-auto theme-transition">
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={
              <PageWrapper>
                <Dashboard />
              </PageWrapper>
            } />
            <Route path="/security" element={
              <PageWrapper>
                <Security />
              </PageWrapper>
            } />
            <Route path="/approval-queue" element={
              <PageWrapper>
                <ApprovalQueue />
              </PageWrapper>
            } />
            <Route path="/threats" element={
              <PageWrapper>
                <Threats />
              </PageWrapper>
            } />
            <Route path="/users" element={
              <PageWrapper>
                <Users />
              </PageWrapper>
            } />
            <Route path="/dev" element={
              <PageWrapper>
                <Dev />
              </PageWrapper>
            } />
            <Route path="/settings" element={
              <PageWrapper>
                <Settings />
              </PageWrapper>
            } />

            {/* Strategic Enhancement Pages */}
            <Route path="/threat-intelligence" element={
              <PageWrapper>
                <ThreatIntelligence />
              </PageWrapper>
            } />
            <Route path="/analytics" element={
              <PageWrapper>
                <Analytics />
              </PageWrapper>
            } />
            <Route path="/threat-hunting" element={
              <PageWrapper>
                <ThreatHunting />
              </PageWrapper>
            } />
            <Route path="/blockchain" element={
              <PageWrapper>
                <Blockchain />
              </PageWrapper>
            } />
            <Route path="/zero-trust" element={
              <PageWrapper>
                <ZeroTrust />
              </PageWrapper>
            } />
            <Route path="/quantum" element={
              <PageWrapper>
                <Quantum />
              </PageWrapper>
            } />
            <Route path="/siem" element={
              <PageWrapper>
                <SIEM />
              </PageWrapper>
            } />
            <Route path="/incident-response" element={
              <PageWrapper>
                <IncidentResponse />
              </PageWrapper>
            } />
            <Route path="/threat-modeling" element={
              <PageWrapper>
                <ThreatModeling />
              </PageWrapper>
            } />
            <Route path="/kubernetes" element={
              <PageWrapper>
                <Kubernetes />
              </PageWrapper>
            } />
            <Route path="/peer-network" element={
              <PageWrapper>
                <PeerNetwork />
              </PageWrapper>
            } />
            <Route path="/tor-network" element={
              <PageWrapper>
                <TorNetwork />
              </PageWrapper>
            } />

            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </main>
      </div>

      <Suspense fallback={<SuspenseFallback message="Loading agent panel..." size="sm" />}>
        <AgentActivityToggle data-agent-toggle="true">
          <AgentActivityPanel />
        </AgentActivityToggle>
      </Suspense>
    </div>
  )
}

function App() {
  // Initialize hot reload in development
  const hotReloadEnabled = process.env.NODE_ENV === 'development'
  
  // The useEffect block for ViteHotReload is moved to HotReloadProvider
  // useEffect(() => {
  //   let hotReload: ViteHotReload | null = null
    
  //   if (hotReloadEnabled) {
  //     hotReload = new ViteHotReload()
  //     hotReload.connect()
  //   }
    
  //   return () => {
  //     if (hotReload) {
  //       hotReload.disconnect()
  //     }
  //   }
  // }, [hotReloadEnabled])

  return (
    <HotReloadProvider enabled={hotReloadEnabled}>
      <ThemeProvider>
        <AuthProvider>
          <AgentEventProvider>
            <AppContent />
          </AgentEventProvider>
        </AuthProvider>
      </ThemeProvider>
    </HotReloadProvider>
  )
}

export default App