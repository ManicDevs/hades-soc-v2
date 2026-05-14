import { useState, useEffect } from 'react'
import { Code, Users, Shield, Key, Database, Settings, Plus, Edit, Trash2, Eye } from 'lucide-react'

function Dev() {
  const [activeTab, setActiveTab] = useState('users')
  const [users, setUsers] = useState<any[]>([])
  const [roles, setRoles] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchDevData()
  }, [])

  const fetchDevData = async () => {
    try {
      // Import API_CONFIG locally to avoid circular dependencies
      const { default: API_CONFIG } = await import('../api/config')
      
      // Fetch users
      const usersData = await API_CONFIG.request('/users')
      setUsers(Array.isArray(usersData) ? usersData : [])

      // Fetch roles
      const rolesData = await API_CONFIG.request('/users/roles')
      setRoles(Array.isArray(rolesData) ? rolesData : [])
    } catch (error) {
      console.error('Failed to fetch dev data:', error)
    } finally {
      setLoading(false)
    }
  }

  const tabs = [
    { id: 'users', label: 'User Management', icon: Users },
    { id: 'roles', label: 'Role Management', icon: Shield },
    { id: 'auth', label: 'Authentication', icon: Key },
    { id: 'database', label: 'Database', icon: Database },
    { id: 'config', label: 'Configuration', icon: Settings },
  ]

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-hades-primary"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <div className="flex items-center space-x-3 mb-4">
          <Code className="w-8 h-8 text-hades-primary" />
          <div>
            <h1 className="text-2xl font-bold text-white">Dev Access</h1>
            <p className="text-gray-400">Development environment management tools</p>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex space-x-1 border-b border-gray-800">
          {tabs.map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center space-x-2 px-4 py-2 border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? 'border-hades-primary text-hades-primary'
                    : 'border-transparent text-gray-400 hover:text-white'
                }`}
              >
                <Icon className="w-4 h-4" />
                <span>{tab.label}</span>
              </button>
            )
          })}
        </div>
      </div>

      {/* Tab Content */}
      <div className="bg-hades-darker rounded-lg border border-gray-800">
        {activeTab === 'users' && <UserManagement users={users} onRefresh={fetchDevData} />}
        {activeTab === 'roles' && <RoleManagement roles={roles} onRefresh={fetchDevData} />}
        {activeTab === 'auth' && <AuthenticationManagement />}
        {activeTab === 'database' && <DatabaseManagement />}
        {activeTab === 'config' && <ConfigurationManagement />}
      </div>
    </div>
  )
}

// User Management Component
function UserManagement({ users, onRefresh }: { users?: any; onRefresh?: any }) {
  const [showCreateForm, setShowCreateForm] = useState(false)

  const handleCreateUser = async (userData: any) => {
    try {
      // Import API_CONFIG locally to avoid circular dependencies
      const { default: API_CONFIG } = await import('../api/config')
      
      await API_CONFIG.request('/users', {
        method: 'POST',
        body: JSON.stringify(userData)
      })
      
      setShowCreateForm(false)
      onRefresh()
    } catch (error) {
      console.error('Failed to create user:', error)
    }
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-semibold text-white">User Management</h2>
        <button
          onClick={() => setShowCreateForm(true)}
          className="flex items-center space-x-2 px-4 py-2 bg-hades-primary text-white rounded-lg hover:bg-hades-primary/90"
        >
          <Plus className="w-4 h-4" />
          <span>Create User</span>
        </button>
      </div>

      {showCreateForm && (
        <CreateUserForm
          onSubmit={handleCreateUser}
          onCancel={() => setShowCreateForm(false)}
        />
      )}

      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-800">
              <th className="text-left py-3 px-4 text-gray-400 font-medium">Username</th>
              <th className="text-left py-3 px-4 text-gray-400 font-medium">Email</th>
              <th className="text-left py-3 px-4 text-gray-400 font-medium">Role</th>
              <th className="text-left py-3 px-4 text-gray-400 font-medium">Status</th>
              <th className="text-left py-3 px-4 text-gray-400 font-medium">Last Login</th>
              <th className="text-left py-3 px-4 text-gray-400 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user: any) => (
              <tr key={user.id} className="border-b border-gray-800/50">
                <td className="py-3 px-4 text-white">{user.username}</td>
                <td className="py-3 px-4 text-gray-300">{user.email}</td>
                <td className="py-3 px-4">
                  <span className="px-2 py-1 bg-hades-primary/20 text-hades-primary rounded text-sm">
                    {user.role}
                  </span>
                </td>
                <td className="py-3 px-4">
                  <span className={`px-2 py-1 rounded text-sm ${
                    user.status === 'active' 
                      ? 'bg-green-500/20 text-green-400' 
                      : 'bg-gray-500/20 text-gray-400'
                  }`}>
                    {user.status}
                  </span>
                </td>
                <td className="py-3 px-4 text-gray-300">
                  {new Date(user.lastLogin).toLocaleString()}
                </td>
                <td className="py-3 px-4">
                  <div className="flex space-x-2">
                    <button className="p-1 text-gray-400 hover:text-white">
                      <Eye className="w-4 h-4" />
                    </button>
                    <button className="p-1 text-gray-400 hover:text-white">
                      <Edit className="w-4 h-4" />
                    </button>
                    <button className="p-1 text-gray-400 hover:text-red-400">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

// Role Management Component
function RoleManagement({ roles, onRefresh }: { roles?: any; onRefresh?: any }) {
  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold text-white mb-6">Role Management</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {roles.map((role: any) => (
          <div key={role.id} className="bg-hades-dark border border-gray-800 rounded-lg p-4">
            <h3 className="text-white font-medium mb-2">{role.name}</h3>
            <p className="text-gray-400 text-sm mb-3">{role.description}</p>
            <div className="flex flex-wrap gap-2">
              {role.permissions?.map((permission: any) => (
                <span
                  key={permission}
                  className="px-2 py-1 bg-hades-primary/20 text-hades-primary rounded text-xs"
                >
                  {permission}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// Authentication Management Component
function AuthenticationManagement() {
  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold text-white mb-6">Authentication Settings</h2>
      <div className="space-y-4">
        <div className="bg-hades-dark border border-gray-800 rounded-lg p-4">
          <h3 className="text-white font-medium mb-2">Development Credentials</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-400">Dev User:</span>
              <code className="text-hades-primary">dev / dev123</code>
            </div>
            <p className="text-gray-500 text-xs mt-2">
              Note: Default credentials are only available in development environment.
              Production requires proper user setup.
            </p>
          </div>
        </div>
        
        <div className="bg-hades-dark border border-gray-800 rounded-lg p-4">
          <h3 className="text-white font-medium mb-2">JWT Configuration</h3>
          <div className="space-y-2 text-sm text-gray-400">
            <div className="flex justify-between">
              <span>Token Expiry:</span>
              <span>24 hours</span>
            </div>
            <div className="flex justify-between">
              <span>Algorithm:</span>
              <span>HS256</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// Database Management Component
function DatabaseManagement() {
  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold text-white mb-6">Database Management</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-hades-dark border border-gray-800 rounded-lg p-4">
          <h3 className="text-white font-medium mb-2">Connection Status</h3>
          <div className="space-y-2 text-sm">
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              <span className="text-gray-400">SQLite Database</span>
            </div>
            <div className="text-gray-500 text-xs">./hades_toolkit.db</div>
          </div>
        </div>
        
        <div className="bg-hades-dark border border-gray-800 rounded-lg p-4">
          <h3 className="text-white font-medium mb-2">Quick Actions</h3>
          <div className="space-y-2">
            <button className="w-full px-3 py-2 bg-hades-primary text-white rounded text-sm hover:bg-hades-primary/90">
              Run Migrations
            </button>
            <button className="w-full px-3 py-2 bg-gray-700 text-white rounded text-sm hover:bg-gray-600">
              Backup Database
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// Configuration Management Component
function ConfigurationManagement() {
  return (
    <div className="p-6">
      <h2 className="text-xl font-semibold text-white mb-6">System Configuration</h2>
      <div className="space-y-4">
        <div className="bg-hades-dark border border-gray-800 rounded-lg p-4">
          <h3 className="text-white font-medium mb-2">Environment Settings</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-400">Environment:</span>
              <span className="text-hades-primary">Development</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-400">Debug Mode:</span>
              <span className="text-hades-primary">Enabled</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// Create User Form Component
function CreateUserForm({ onSubmit, onCancel }: { onSubmit?: any; onCancel?: any }) {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    role: 'Security Analyst'
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit(formData)
  }

  return (
    <div className="mb-6 p-4 bg-hades-dark border border-gray-800 rounded-lg">
      <h3 className="text-white font-medium mb-4">Create New User</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-gray-400 text-sm mb-1">Username</label>
            <input
              type="text"
              value={formData.username}
              onChange={(e) => setFormData({...formData, username: e.target.value})}
              className="w-full px-3 py-2 bg-hades-darker border border-gray-700 rounded text-white"
              required
            />
          </div>
          <div>
            <label className="block text-gray-400 text-sm mb-1">Email</label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({...formData, email: e.target.value})}
              className="w-full px-3 py-2 bg-hades-darker border border-gray-700 rounded text-white"
              required
            />
          </div>
          <div>
            <label className="block text-gray-400 text-sm mb-1">Password</label>
            <input
              type="password"
              value={formData.password}
              onChange={(e) => setFormData({...formData, password: e.target.value})}
              className="w-full px-3 py-2 bg-hades-darker border border-gray-700 rounded text-white"
              required
            />
          </div>
          <div>
            <label className="block text-gray-400 text-sm mb-1">Role</label>
            <select
              value={formData.role}
              onChange={(e) => setFormData({...formData, role: e.target.value})}
              className="w-full px-3 py-2 bg-hades-darker border border-gray-700 rounded text-white"
            >
              <option value="Security Analyst">Security Analyst</option>
              <option value="Security Engineer">Security Engineer</option>
              <option value="Auditor">Auditor</option>
            </select>
          </div>
        </div>
        <div className="flex space-x-4">
          <button
            type="submit"
            className="px-4 py-2 bg-hades-primary text-white rounded hover:bg-hades-primary/90"
          >
            Create User
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 bg-gray-700 text-white rounded hover:bg-gray-600"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}

export default Dev
