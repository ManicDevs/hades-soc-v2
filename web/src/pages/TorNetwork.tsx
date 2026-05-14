import { useState, useEffect } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import { Activity, Globe, AlertCircle } from 'lucide-react'

const TorDashboard = () => {
  const [status, setStatus] = useState<any>(null)
  const [circuits, setCircuits] = useState<any[]>([])
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<any>(null)

  useEffect(() => {
    fetchStatus()
    const interval = setInterval(fetchStatus, 30000)
    return () => clearInterval(interval)
  }, [])

  const fetchStatus = async () => {
    try {
      setLoading(true)
      const response = await fetch('/api/v2/tor/status')
      if (!response.ok) {
        const text = await response.text()
        throw new Error(`HTTP ${response.status}: ${text}`)
      }
      const data = await response.json()
      setStatus(data)
      
      const circuitsRes = await fetch('/api/v2/tor/circuit/status')
      if (!circuitsRes.ok) {
        const text = await circuitsRes.text()
        throw new Error(`HTTP ${circuitsRes.status}: ${text}`)
      }
      const circuitsData = await circuitsRes.json()
      setCircuits(circuitsData.circuits || [])
      
      const statsRes = await fetch('/api/v2/tor/stats')
      if (!statsRes.ok) {
        const text = await statsRes.text()
        throw new Error(`HTTP ${statsRes.status}: ${text}`)
      }
      const statsData = await statsRes.json()
      setStats(statsData)
      
      setError(null)
    } catch (err: any) {
      console.error('Failed to fetch Tor status:', err)
      setError(err.message || 'Failed to fetch Tor status')
      setStatus(null)
      setStats(null)
      setCircuits([])
    } finally {
      setLoading(false)
    }
  }

  const createOnionService = async () => {
    const port = prompt('Enter local port to expose:', '8080')
    if (!port) return
    
    try {
      const response = await fetch('/api/v2/tor/onion/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ port: parseInt(port), target_port: parseInt(port) })
      })
      const data = await response.json()
      alert(`Onion service created: ${data.onion_address}`)
      fetchStatus()
    } catch (err: any) {
      alert('Failed to create onion service: ' + err.message)
    }
  }

  const testConnection = async () => {
    try {
      const response = await fetch('/api/v2/tor/test-connection', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: 'http://check.torproject.org/api/ip' })
      })
      const data = await response.json()
      alert('Connection test result: ' + JSON.stringify(data.response, null, 2))
    } catch (err: any) {
      alert('Connection test failed: ' + err.message)
    }
  }

  if (loading) {
    return <div className="p-4">Loading Tor status...</div>
  }

  return (
    <div className="p-4">
      <h2 className="mb-4">
        <span className="me-2">🧅</span>
        Tor Network (TorGo) Dashboard
      </h2>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 p-4 rounded-lg mb-4 flex items-center gap-2">
          <AlertCircle className="w-5 h-5" />
          Error: {error}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className={`p-4 rounded-lg ${status?.running ? 'bg-green-500/10 border border-green-500' : 'bg-red-500/10 border border-red-500'}`}>
          <div className="text-sm text-gray-400">Tor Status</div>
          <div className="text-2xl font-bold">{status?.running ? 'Running' : 'Stopped'}</div>
        </div>
        
        <div className="bg-gray-800 p-4 rounded-lg">
          <div className="text-sm text-gray-400">SOCKS Port</div>
          <div className="text-2xl font-bold">{status?.socks_port || 19050}</div>
        </div>
        
        <div className="bg-gray-800 p-4 rounded-lg">
          <div className="text-sm text-gray-400">Control Port</div>
          <div className="text-2xl font-bold">{status?.control_port || 19051}</div>
        </div>
        
        <div className="bg-gray-800 p-4 rounded-lg">
          <div className="text-sm text-gray-400">Version</div>
          <div className="text-sm font-mono">{status?.version || '0.4.9.3-alpha'}</div>
        </div>
      </div>

      <div className="mb-6">
        <Card>
          <CardHeader className="flex justify-between items-center">
            <CardTitle>Actions</CardTitle>
          </CardHeader>
          <CardContent className="flex gap-2">
            <Button onClick={createOnionService} variant="default">
              ➕ Create Onion Service
            </Button>
            <Button onClick={testConnection} variant="default">
              🔗 Test Connection
            </Button>
            <Button onClick={fetchStatus} variant="default">
              🔄 Refresh
            </Button>
          </CardContent>
        </Card>
      </div>

      {stats && (
        <div className="mb-6">
          <Card>
            <CardHeader>
              <CardTitle>Statistics</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
                <div>
                  <div className="text-sm text-gray-400">Bytes Read</div>
                  <div className="font-mono">{(stats.bytes_read / 1024 / 1024).toFixed(2)} MB</div>
                </div>
                <div>
                  <div className="text-sm text-gray-400">Bytes Written</div>
                  <div className="font-mono">{(stats.bytes_written / 1024 / 1024).toFixed(2)} MB</div>
                </div>
                <div>
                  <div className="text-sm text-gray-400">Circuits</div>
                  <div className="font-mono">{stats.circuits}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-400">Streams</div>
                  <div className="font-mono">{stats.streams}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-400">Onion Services</div>
                  <div className="font-mono">{stats.onion_services}</div>
                </div>
                <div>
                  <div className="text-sm text-gray-400">Uptime</div>
                  <div className="font-mono">{stats.uptime}</div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      <div className="mb-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="w-5 h-5" />
              Active Circuits
            </CardTitle>
          </CardHeader>
          <CardContent>
            {circuits.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-gray-400 border-b border-gray-700">
                      <th className="pb-2">ID</th>
                      <th className="pb-2">State</th>
                      <th className="pb-2">Path</th>
                      <th className="pb-2">Created</th>
                      <th className="pb-2">Flags</th>
                    </tr>
                  </thead>
                  <tbody>
                    {circuits.map((circuit) => (
                      <tr key={circuit.id} className="border-b border-gray-700/50">
                        <td className="py-2 font-mono">{circuit.id}</td>
                        <td className="py-2">
                          <Badge variant={circuit.state === 'BUILT' ? 'success' : 'warning'}>
                            {circuit.state}
                          </Badge>
                        </td>
                        <td className="py-2 text-gray-400">{circuit.path?.join(' → ')}</td>
                        <td className="py-2 text-gray-400">{circuit.created}</td>
                        <td className="py-2">
                          {circuit.flags?.map((flag) => (
                            <Badge key={flag} variant="warning" className="mr-1">{flag}</Badge>
                          ))}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-gray-400">No active circuits</p>
            )}
          </CardContent>
        </Card>
      </div>

      {status?.onion_services?.length > 0 && (
        <div>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Globe className="w-5 h-5" />
                Onion Services
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-gray-400 border-b border-gray-700">
                      <th className="pb-2">Onion Address</th>
                      <th className="pb-2">Port</th>
                      <th className="pb-2">Target Port</th>
                      <th className="pb-2">Created</th>
                      <th className="pb-2">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {status.onion_services.map((service) => (
                      <tr key={service.onion_address} className="border-b border-gray-700/50">
                        <td className="py-2 font-mono text-sm">{service.onion_address}</td>
                        <td className="py-2">{service.port}</td>
                        <td className="py-2">{service.target_port}</td>
                        <td className="py-2 text-gray-400">{new Date(service.created_at).toLocaleString()}</td>
                        <td className="py-2">
                          <Button 
                            size="sm" 
                            variant="default"
                            onClick={() => {
                              fetch(`/api/v2/tor/onion/delete/${service.onion_address}`, { method: 'DELETE' })
                                .then(() => fetchStatus());
                            }}
                          >
                            Delete
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

export default TorDashboard