import { useState, useEffect } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Badge } from '../components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs'
import { Activity, Users, Globe, Shield, Zap } from 'lucide-react'

const PeerNetwork = () => {
  const [peers, setPeers] = useState<any[]>([])
  const [networkStatus, setNetworkStatus] = useState<any>({})
  const [discoveredPeers, setDiscoveredPeers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [nodeId, setNodeId] = useState('')

  useEffect(() => {
    // Simulate peer network data
    const mockPeers = [
      {
        id: 'e3cc56349b8a1895a29e83da1e34085d',
        address: 'peer1.hades.local:8080',
        lastSeen: new Date(),
        reputation: 1.0,
        capabilities: ['key_sharing', 'consensus', 'blockchain_sync'],
        active: true
      },
      {
        id: '871f7268b42b0a148966d1a0c82bd451',
        address: 'peer2.hades.local:8080',
        lastSeen: new Date(),
        reputation: 1.0,
        capabilities: ['key_sharing', 'consensus', 'blockchain_sync'],
        active: true
      },
      {
        id: 'ce8cd7b7aac5f46fb2602fd8f3786b9a',
        address: 'peer3.hades.local:8080',
        lastSeen: new Date(),
        reputation: 1.0,
        capabilities: ['key_sharing', 'consensus', 'blockchain_sync'],
        active: true
      },
      {
        id: '589ee1c57f49bca4eb101983ca3269f0',
        address: 'peer4.hades.local:8080',
        lastSeen: new Date(),
        reputation: 1.0,
        capabilities: ['key_sharing', 'consensus', 'blockchain_sync'],
        active: true
      },
      {
        id: 'c3dfbc1c16fed9907e771f9b1a5c0485',
        address: 'peer5.hades.local:8080',
        lastSeen: new Date(),
        reputation: 1.0,
        capabilities: ['key_sharing', 'consensus', 'blockchain_sync'],
        active: true
      }
    ]

    const mockDiscoveredPeers = [
      'peer1.hades.local:8080',
      'peer2.hades.local:8080',
      'peer3.hades.local:8080',
      'peer4.hades.local:8080',
      'peer5.hades.local:8080'
    ]

    const mockNetworkStatus = {
      nodeId: 'hades-node-1847',
      totalPeers: 5,
      activePeers: 5,
      networkUptime: '2.5 hours',
      capabilities: [
        'Key Distribution: Shamir\'s Secret Sharing',
        'Consensus: PBFT (Practical Byzantine Fault Tolerance)',
        'Blockchain: Integrity Verification',
        'Multi-Chain: Cross-chain Interoperability'
      ]
    }

    setPeers(mockPeers)
    setDiscoveredPeers(mockDiscoveredPeers)
    setNetworkStatus(mockNetworkStatus)
    setNodeId(mockNetworkStatus.nodeId)
    setLoading(false)
  }, [])

  const handleDiscoverPeers = () => {
    // Simulate peer discovery
    setLoading(true)
    setTimeout(() => {
      setLoading(false)
    }, 2000)
  }

  const handleConnectPeer = (address: any) => {
    console.log(`Connecting to peer: ${address}`)
    // In real implementation, this would call API to connect to peer
  }

  const formatPeerId = (id: any) => {
    return id.length > 18 ? id.substring(0, 18) + '...' : id
  }

  const formatLastSeen = (date: any) => {
    return new Date(date).toLocaleString()
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-hades-primary"></div>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Peer Network</h1>
          <p className="text-gray-400">HADES P2P Network Management</p>
        </div>
        <div className="flex items-center space-x-2">
          <Badge variant="default" className="text-green-400 border-green-400">
            <Globe className="w-4 h-4 mr-1" />
            Network Active
          </Badge>
          <Badge variant="default" className="text-blue-400 border-blue-400">
            <Users className="w-4 h-4 mr-1" />
            {networkStatus.totalPeers} Peers
          </Badge>
        </div>
      </div>

      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="peers">Connected Peers</TabsTrigger>
          <TabsTrigger value="discovery">Discovery</TabsTrigger>
          <TabsTrigger value="capabilities">Capabilities</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Node ID</CardTitle>
                <Shield className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{nodeId}</div>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Peers</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{networkStatus.totalPeers}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Active Peers</CardTitle>
                <Activity className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-400">{networkStatus.activePeers}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Network Uptime</CardTitle>
                <Zap className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{networkStatus.networkUptime}</div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Network Capabilities</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {networkStatus.capabilities?.map((capability, index) => (
                  <div key={index} className="flex items-center space-x-2">
                    <Shield className="w-4 h-4 text-hades-primary" />
                    <span className="text-sm text-gray-300">{capability}</span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="peers" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Connected Peers</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-gray-700">
                      <th className="text-left p-2 text-gray-400">Peer ID</th>
                      <th className="text-left p-2 text-gray-400">Address</th>
                      <th className="text-left p-2 text-gray-400">Last Seen</th>
                      <th className="text-left p-2 text-gray-400">Status</th>
                      <th className="text-left p-2 text-gray-400">Capabilities</th>
                    </tr>
                  </thead>
                  <tbody>
                    {peers.map((peer, index) => (
                      <tr key={index} className="border-b border-gray-800 hover:bg-gray-800/50">
                        <td className="p-2 text-gray-300 font-mono text-sm">
                          {formatPeerId(peer.id)}
                        </td>
                        <td className="p-2 text-gray-300">{peer.address}</td>
                        <td className="p-2 text-gray-300">
                          {formatLastSeen(peer.lastSeen)}
                        </td>
                        <td className="p-2">
                          <Badge variant={peer.active ? "default" : "warning"}>
                            {peer.active ? 'Active' : 'Inactive'}
                          </Badge>
                        </td>
                        <td className="p-2">
                          <div className="flex flex-wrap gap-1">
                            {peer.capabilities.map((cap, capIndex) => (
                              <Badge key={capIndex} variant="default" className="text-xs">
                                {cap}
                              </Badge>
                            ))}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="discovery" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Peer Discovery</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between">
                <p className="text-gray-400">
                  Discover HADES peers on the network
                </p>
                <Button onClick={handleDiscoverPeers} disabled={loading}>
                  <Globe className="w-4 h-4 mr-2" />
                  {loading ? 'Discovering...' : 'Discover Peers'}
                </Button>
              </div>
              
              {discoveredPeers.length > 0 && (
                <div className="space-y-2">
                  <h3 className="text-lg font-semibold text-white">Discovered Peers</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                    {discoveredPeers.map((peer, index) => (
                      <div key={index} className="flex items-center justify-between p-3 bg-gray-800 rounded-lg">
                        <span className="text-gray-300 font-mono text-sm">{peer}</span>
                        <Button 
                          size="sm" 
                          onClick={() => handleConnectPeer(peer)}
                          className="ml-2"
                        >
                          Connect
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="capabilities" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Network Capabilities</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-white">Security Features</h3>
                  <div className="space-y-2">
                    <div className="flex items-center space-x-2">
                      <Shield className="w-4 h-4 text-green-400" />
                      <span className="text-gray-300">Key Distribution via Shamir&apos;s Secret Sharing</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Shield className="w-4 h-4 text-green-400" />
                      <span className="text-gray-300">PBFT Consensus Mechanism</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Shield className="w-4 h-4 text-green-400" />
                      <span className="text-gray-300">Blockchain Integrity Verification</span>
                    </div>
                  </div>
                </div>
                
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold text-white">Network Features</h3>
                  <div className="space-y-2">
                    <div className="flex items-center space-x-2">
                      <Zap className="w-4 h-4 text-blue-400" />
                      <span className="text-gray-300">Multi-Chain Interoperability</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Zap className="w-4 h-4 text-blue-400" />
                      <span className="text-gray-300">Real-time Peer Discovery</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Zap className="w-4 h-4 text-blue-400" />
                      <span className="text-gray-300">Dynamic Connection Management</span>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

export default PeerNetwork
