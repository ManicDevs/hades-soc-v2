import React, { useState } from 'react'
import { Shield, Activity, AlertTriangle, Users, TrendingUp, Server, Database, Lock, FileText } from 'lucide-react'
import { useDashboard } from '../hooks/useDashboard'
import AgentThoughtStream from '../components/AgentThoughtStream'
import ReportsViewer from '../components/ReportsViewer'

function Dashboard({ user }) {
  const [activeTab, setActiveTab] = useState('overview')
  const { metrics, activity, systemStatus, securityOverview, loading, error, refreshData } = useDashboard()

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={refreshData} className="hades-button-primary">
            Retry
          </button>
        </div>
      </div>
    )
  }

  const metricCards = [
    {
      title: 'Security Score',
      value: metrics?.securityScore + '%' || '98%',
      icon: Shield,
      color: 'text-green-400',
      bgColor: 'bg-green-900/20',
      borderColor: 'border-green-500/20'
    },
    {
      title: 'Active Threats',
      value: metrics?.activeThreats || 3,
      icon: AlertTriangle,
      color: 'text-red-400',
      bgColor: 'bg-red-900/20',
      borderColor: 'border-red-500/20'
    },
    {
      title: 'Blocked Attacks',
      value: (metrics?.blockedAttacks || 1247).toLocaleString(),
      icon: Lock,
      color: 'text-blue-400',
      bgColor: 'bg-blue-900/20',
      borderColor: 'border-blue-500/20'
    },
    {
      title: 'System Health',
      value: metrics?.systemHealth + '%' || '99%',
      icon: Activity,
      color: 'text-green-400',
      bgColor: 'bg-green-900/20',
      borderColor: 'border-green-500/20'
    },
  ]

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Security Dashboard</h1>
        <p className="text-gray-400">Real-time monitoring and threat detection</p>
      </div>

      {/* Tab Navigation */}
      <div className="flex gap-4 mb-6 border-b border-slate-700">
        <button
          onClick={() => setActiveTab('overview')}
          className={`pb-3 px-4 text-sm font-medium transition-colors ${
            activeTab === 'overview'
              ? 'text-hades-primary border-b-2 border-hades-primary'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          <span className="flex items-center gap-2">
            <Activity className="w-4 h-4" />
            Overview
          </span>
        </button>
        <button
          onClick={() => setActiveTab('thoughts')}
          className={`pb-3 px-4 text-sm font-medium transition-colors ${
            activeTab === 'thoughts'
              ? 'text-hades-primary border-b-2 border-hades-primary'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          <span className="flex items-center gap-2">
            <Shield className="w-4 h-4" />
            Agent Thoughts
          </span>
        </button>
        <button
          onClick={() => setActiveTab('reports')}
          className={`pb-3 px-4 text-sm font-medium transition-colors ${
            activeTab === 'reports'
              ? 'text-hades-primary border-b-2 border-hades-primary'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          <span className="flex items-center gap-2">
            <FileText className="w-4 h-4" />
            Historical Reports
          </span>
        </button>
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <>
          {/* Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {metricCards.map((metric, index) => {
          const Icon = metric.icon
          return (
            <div key={index} className={`hades-card p-6 ${metric.bgColor} ${metric.borderColor}`}>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-400 text-sm font-medium">{metric.title}</p>
                  <p className={`text-2xl font-bold ${metric.color} mt-2`}>{metric.value}</p>
                </div>
                <div className={`p-3 rounded-lg ${metric.bgColor}`}>
                  <Icon className={`w-6 h-6 ${metric.color}`} />
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Charts and Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Security Overview Chart */}
        <div className="lg:col-span-2 hades-card p-6">
          <h2 className="text-xl font-semibold text-white mb-4">Security Overview</h2>
          <div className="h-64 flex items-center justify-center bg-gray-700/30 rounded-lg border border-gray-600">
            <div className="text-center">
              <TrendingUp className="w-12 h-12 text-hades-primary mx-auto mb-2" />
              <p className="text-gray-400">Security metrics visualization</p>
              <p className="text-gray-500 text-sm">Chart component would be rendered here</p>
            </div>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="hades-card p-6">
          <h2 className="text-xl font-semibold text-white mb-4">Recent Activity</h2>
          <div className="space-y-3">
            {activity?.map((activity) => (
              <div key={activity.id} className="flex items-start space-x-3 p-3 rounded-lg bg-gray-700/30 border border-gray-600">
                <div className={`w-2 h-2 rounded-full mt-2 ${
                  activity.severity === 'high' ? 'bg-red-400' :
                  activity.severity === 'medium' ? 'bg-yellow-400' : 'bg-green-400'
                }`}></div>
                <div className="flex-1">
                  <p className="text-white text-sm">{activity.message}</p>
                  <p className="text-gray-500 text-xs">{activity.time}</p>
                </div>
              </div>
            )) || [
              <div key="1" className="flex items-start space-x-3 p-3 rounded-lg bg-gray-700/30 border border-gray-600">
                <div className="w-2 h-2 rounded-full mt-2 bg-yellow-400"></div>
                <div className="flex-1">
                  <p className="text-white text-sm">Security scan completed</p>
                  <p className="text-gray-500 text-xs">2 min ago</p>
                </div>
              </div>
            ]}
          </div>
        </div>
      </div>

      {/* System Status */}
      <div className="mt-6 grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="hades-card p-6">
          <h2 className="text-xl font-semibold text-white mb-4">System Status</h2>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-gray-300">Database</span>
              <span className="text-green-400 text-sm">Operational</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-300">API Server</span>
              <span className="text-green-400 text-sm">Running</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-300">Security Engine</span>
              <span className="text-green-400 text-sm">Active</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-300">Backup Service</span>
              <span className="text-yellow-400 text-sm">Scheduled</span>
            </div>
          </div>
        </div>

        <div className="hades-card p-6">
          <h2 className="text-xl font-semibold text-white mb-4">Quick Actions</h2>
          <div className="grid grid-cols-2 gap-3">
            <button className="hades-button-secondary text-sm">Run Security Scan</button>
            <button className="hades-button-secondary text-sm">View Logs</button>
            <button className="hades-button-secondary text-sm">Update Rules</button>
            <button className="hades-button-secondary text-sm">Export Report</button>
          </div>
        </div>
      </div>

      {/* Agent Thought Stream - Real-time Autonomous Agent Monitoring */}
          <div className="mt-6">
            <AgentThoughtStream className="h-[500px]" />
          </div>
        </>
      )}

      {activeTab === 'thoughts' && (
        <div className="mt-6">
          <AgentThoughtStream className="h-[600px]" />
        </div>
      )}

      {activeTab === 'reports' && (
        <div className="mt-6">
          <ReportsViewer />
        </div>
      )}
    </div>
  )
}

export default Dashboard
