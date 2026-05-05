import React, { useState } from 'react'
import { Settings as SettingsIcon, Shield, Bell, Database, Globe, Lock, User, Palette } from 'lucide-react'

function Settings({ user }) {
  const [activeTab, setActiveTab] = useState('general')
  const [settings, setSettings] = useState({
    general: {
      companyName: 'Hades Toolkit',
      timezone: 'UTC',
      language: 'en',
      theme: 'dark'
    },
    security: {
      sessionTimeout: '30',
      passwordPolicy: 'strong',
      twoFactorAuth: 'enabled',
      ipWhitelist: 'disabled'
    },
    notifications: {
      emailAlerts: 'enabled',
      pushNotifications: 'enabled',
      threatAlerts: 'immediate',
      systemAlerts: 'daily'
    },
    database: {
      backupFrequency: 'daily',
      retentionPeriod: '90',
      encryption: 'enabled',
      compression: 'enabled'
    }
  })

  const tabs = [
    { id: 'general', label: 'General', icon: SettingsIcon },
    { id: 'security', label: 'Security', icon: Shield },
    { id: 'notifications', label: 'Notifications', icon: Bell },
    { id: 'database', label: 'Database', icon: Database },
  ]

  const handleSettingChange = (category, key, value) => {
    setSettings(prev => ({
      ...prev,
      [category]: {
        ...prev[category],
        [key]: value
      }
    }))
  }

  const renderTabContent = () => {
    switch (activeTab) {
      case 'general':
        return (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Company Name</label>
              <input
                type="text"
                value={settings.general.companyName}
                onChange={(e) => handleSettingChange('general', 'companyName', e.target.value)}
                className="hades-input w-full"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Timezone</label>
              <select
                value={settings.general.timezone}
                onChange={(e) => handleSettingChange('general', 'timezone', e.target.value)}
                className="hades-input w-full"
              >
                <option value="UTC">UTC</option>
                <option value="EST">EST</option>
                <option value="PST">PST</option>
                <option value="GMT">GMT</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Language</label>
              <select
                value={settings.general.language}
                onChange={(e) => handleSettingChange('general', 'language', e.target.value)}
                className="hades-input w-full"
              >
                <option value="en">English</option>
                <option value="es">Spanish</option>
                <option value="fr">French</option>
                <option value="de">German</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Theme</label>
              <select
                value={settings.general.theme}
                onChange={(e) => handleSettingChange('general', 'theme', e.target.value)}
                className="hades-input w-full"
              >
                <option value="dark">Dark</option>
                <option value="light">Light</option>
                <option value="auto">Auto</option>
              </select>
            </div>
          </div>
        )
      
      case 'security':
        return (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Session Timeout (minutes)</label>
              <input
                type="number"
                value={settings.security.sessionTimeout}
                onChange={(e) => handleSettingChange('security', 'sessionTimeout', e.target.value)}
                className="hades-input w-full"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Password Policy</label>
              <select
                value={settings.security.passwordPolicy}
                onChange={(e) => handleSettingChange('security', 'passwordPolicy', e.target.value)}
                className="hades-input w-full"
              >
                <option value="weak">Weak</option>
                <option value="medium">Medium</option>
                <option value="strong">Strong</option>
                <option value="very-strong">Very Strong</option>
              </select>
            </div>
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.security.twoFactorAuth === 'enabled'}
                  onChange={(e) => handleSettingChange('security', 'twoFactorAuth', e.target.checked ? 'enabled' : 'disabled')}
                  className="w-4 h-4 text-hades-primary bg-gray-700 border-gray-600 rounded focus:ring-hades-primary"
                />
                <span className="text-gray-300">Enable Two-Factor Authentication</span>
              </label>
            </div>
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.security.ipWhitelist === 'enabled'}
                  onChange={(e) => handleSettingChange('security', 'ipWhitelist', e.target.checked ? 'enabled' : 'disabled')}
                  className="w-4 h-4 text-hades-primary bg-gray-700 border-gray-600 rounded focus:ring-hades-primary"
                />
                <span className="text-gray-300">Enable IP Whitelist</span>
              </label>
            </div>
          </div>
        )
      
      case 'notifications':
        return (
          <div className="space-y-6">
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.notifications.emailAlerts === 'enabled'}
                  onChange={(e) => handleSettingChange('notifications', 'emailAlerts', e.target.checked ? 'enabled' : 'disabled')}
                  className="w-4 h-4 text-hades-primary bg-gray-700 border-gray-600 rounded focus:ring-hades-primary"
                />
                <span className="text-gray-300">Email Alerts</span>
              </label>
            </div>
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.notifications.pushNotifications === 'enabled'}
                  onChange={(e) => handleSettingChange('notifications', 'pushNotifications', e.target.checked ? 'enabled' : 'disabled')}
                  className="w-4 h-4 text-hades-primary bg-gray-700 border-gray-600 rounded focus:ring-hades-primary"
                />
                <span className="text-gray-300">Push Notifications</span>
              </label>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Threat Alerts</label>
              <select
                value={settings.notifications.threatAlerts}
                onChange={(e) => handleSettingChange('notifications', 'threatAlerts', e.target.value)}
                className="hades-input w-full"
              >
                <option value="immediate">Immediate</option>
                <option value="hourly">Hourly</option>
                <option value="daily">Daily</option>
                <option value="weekly">Weekly</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">System Alerts</label>
              <select
                value={settings.notifications.systemAlerts}
                onChange={(e) => handleSettingChange('notifications', 'systemAlerts', e.target.value)}
                className="hades-input w-full"
              >
                <option value="immediate">Immediate</option>
                <option value="hourly">Hourly</option>
                <option value="daily">Daily</option>
                <option value="weekly">Weekly</option>
              </select>
            </div>
          </div>
        )
      
      case 'database':
        return (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Backup Frequency</label>
              <select
                value={settings.database.backupFrequency}
                onChange={(e) => handleSettingChange('database', 'backupFrequency', e.target.value)}
                className="hades-input w-full"
              >
                <option value="hourly">Hourly</option>
                <option value="daily">Daily</option>
                <option value="weekly">Weekly</option>
                <option value="monthly">Monthly</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Retention Period (days)</label>
              <input
                type="number"
                value={settings.database.retentionPeriod}
                onChange={(e) => handleSettingChange('database', 'retentionPeriod', e.target.value)}
                className="hades-input w-full"
              />
            </div>
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.database.encryption === 'enabled'}
                  onChange={(e) => handleSettingChange('database', 'encryption', e.target.checked ? 'enabled' : 'disabled')}
                  className="w-4 h-4 text-hades-primary bg-gray-700 border-gray-600 rounded focus:ring-hades-primary"
                />
                <span className="text-gray-300">Enable Encryption</span>
              </label>
            </div>
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={settings.database.compression === 'enabled'}
                  onChange={(e) => handleSettingChange('database', 'compression', e.target.checked ? 'enabled' : 'disabled')}
                  className="w-4 h-4 text-hades-primary bg-gray-700 border-gray-600 rounded focus:ring-hades-primary"
                />
                <span className="text-gray-300">Enable Compression</span>
              </label>
            </div>
          </div>
        )
      
      default:
        return null
    }
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Settings</h1>
        <p className="text-gray-400">Configure system settings and preferences</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar */}
        <div className="lg:col-span-1">
          <div className="hades-card p-4">
            <nav className="space-y-2">
              {tabs.map((tab) => {
                const Icon = tab.icon
                return (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`w-full flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors ${
                      activeTab === tab.id
                        ? 'bg-hades-primary text-white'
                        : 'text-gray-400 hover:bg-gray-700 hover:text-white'
                    }`}
                  >
                    <Icon className="w-4 h-4" />
                    <span className="text-sm font-medium">{tab.label}</span>
                  </button>
                )
              })}
            </nav>
          </div>
        </div>

        {/* Content */}
        <div className="lg:col-span-3">
          <div className="hades-card p-6">
            <h2 className="text-xl font-semibold text-white mb-6 capitalize">
              {activeTab} Settings
            </h2>
            {renderTabContent()}
            
            {/* Save Button */}
            <div className="mt-8 flex justify-end space-x-4">
              <button className="hades-button-secondary">
                Cancel
              </button>
              <button className="hades-button-primary">
                Save Changes
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Settings
