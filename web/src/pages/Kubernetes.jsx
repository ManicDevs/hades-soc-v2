import React, { useState, useEffect } from 'react'
import { Server, Activity, Shield, Settings, Plus, Minus, Eye, Zap, Database } from 'lucide-react'

function Kubernetes({ user }) {
  const [clusters, setClusters] = useState({})
  const [deployments, setDeployments] = useState({})
  const [services, setServices] = useState({})
  const [autoscalers, setAutoscalers] = useState({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [selectedCluster, setSelectedCluster] = useState(null)
  const [selectedDeployment, setSelectedDeployment] = useState(null)

  useEffect(() => {
    fetchKubernetesData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchKubernetesData()
    }, 8000)

    return () => clearInterval(interval)
  }, [])

  const fetchKubernetesData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [clustersData, deploymentsData, servicesData, autoscalersData] = await Promise.all([
        API_CONFIG.request('/kubernetes/clusters'),
        API_CONFIG.request('/kubernetes/deployments'),
        API_CONFIG.request('/kubernetes/services'),
        API_CONFIG.request('/kubernetes/autoscalers')
      ])
      
      setClusters(clustersData.clusters || {})
      setDeployments(deploymentsData.deployments || {})
      setServices(servicesData.services || {})
      setAutoscalers(autoscalersData.autoscalers || {})
    } catch (error) {
      setError('Failed to fetch Kubernetes data')
      console.error('Kubernetes fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const scaleDeployment = async (deploymentId, replicas) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/kubernetes/scale', {
        method: 'POST',
        body: JSON.stringify({ deployment_id: deploymentId, replicas })
      })
      
      // Refresh data
      fetchKubernetesData()
    } catch (error) {
      console.error('Failed to scale deployment:', error)
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'ready':
      case 'available':
      case 'active':
        return 'text-green-400 bg-green-900/20 border-green-500/20'
      case 'pending':
      case 'deploying':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      case 'failed':
      case 'unhealthy':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getResourceColor = (percentage) => {
    if (percentage >= 80) return 'text-red-400'
    if (percentage >= 60) return 'text-yellow-400'
    return 'text-green-400'
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading Kubernetes data...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchKubernetesData} className="hades-button-primary">
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-white mb-2 flex items-center">
          <Server className="mr-3 text-hades-primary" />
          Cloud-Native Kubernetes Deployment
        </h1>
        <p className="text-gray-400">Enterprise-grade container orchestration and cluster management</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Clusters</p>
              <p className="text-2xl font-bold text-white">{Object.keys(clusters).length}</p>
            </div>
            <Server className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Deployments</p>
              <p className="text-2xl font-bold text-white">{Object.keys(deployments).length}</p>
            </div>
            <Activity className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Services</p>
              <p className="text-2xl font-bold text-white">{Object.keys(services).length}</p>
            </div>
            <Shield className="w-8 h-8 text-blue-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Autoscalers</p>
              <p className="text-2xl font-bold text-white">{Object.keys(autoscalers).length}</p>
            </div>
            <Settings className="w-8 h-8 text-yellow-400" />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Clusters */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Server className="mr-2 text-hades-primary" />
              Kubernetes Clusters
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(clusters).map((cluster, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{cluster.name}</h3>
                    <p className="text-gray-400 text-sm">{cluster.version} • {cluster.provider}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(cluster.status)}`}>
                        {cluster.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {cluster.nodes?.length || 0} nodes
                      </span>
                      <span className="text-xs text-gray-400">
                        {cluster.region}
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">Resources:</p>
                      <div className="flex space-x-4 mt-1">
                        <span className="text-xs text-gray-300">
                          CPU: {cluster.resources?.used_cpu || '0'}/{cluster.resources?.total_cpu || '0'}
                        </span>
                        <span className="text-xs text-gray-300">
                          Memory: {cluster.resources?.used_memory || '0'}/{cluster.resources?.total_memory || '0'}
                        </span>
                      </div>
                    </div>
                  </div>
                  <button 
                    onClick={() => setSelectedCluster(cluster)}
                    className="ml-3 p-2 text-gray-400 hover:text-white"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Deployments */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Activity className="mr-2 text-green-400" />
              Deployments
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(deployments).map((deployment, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{deployment.name}</h3>
                    <p className="text-gray-400 text-sm">{deployment.image}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(deployment.status)}`}>
                        {deployment.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {deployment.ready_replicas}/{deployment.replicas} ready
                      </span>
                      <span className="text-xs text-gray-400">
                        Port: {deployment.port}
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">Resources:</p>
                      <div className="flex space-x-4 mt-1">
                        <span className="text-xs text-gray-300">
                          CPU: {deployment.resources?.requests?.cpu || 'N/A'}
                        </span>
                        <span className="text-xs text-gray-300">
                          Memory: {deployment.resources?.requests?.memory || 'N/A'}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button 
                      onClick={() => scaleDeployment(deployment.id, deployment.replicas + 1)}
                      className="p-2 text-green-400 hover:text-green-300"
                      title="Scale Up"
                    >
                      <Plus className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={() => scaleDeployment(deployment.id, Math.max(1, deployment.replicas - 1))}
                      className="p-2 text-red-400 hover:text-red-300"
                      title="Scale Down"
                    >
                      <Minus className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={() => setSelectedDeployment(deployment)}
                      className="p-2 text-gray-400 hover:text-white"
                    >
                      <Eye className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Services */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-blue-400" />
              Services
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(services).map((service, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{service.name}</h3>
                    <p className="text-gray-400 text-sm">{service.type}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(service.status)}`}>
                        {service.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {service.cluster_ip}
                      </span>
                      <span className="text-xs text-gray-400">
                        {service.ports?.length || 0} ports
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">Ports:</p>
                      <div className="flex space-x-2 mt-1">
                        {service.ports?.map((port, i) => (
                          <span key={i} className="text-xs text-gray-300">
                            {port.port}:{port.target_port}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                  <Shield className="w-5 h-5 text-blue-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Autoscalers */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Settings className="mr-2 text-yellow-400" />
              Horizontal Pod Autoscalers
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(autoscalers).map((autoscaler, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{autoscaler.name}</h3>
                    <p className="text-gray-400 text-sm">Target: {autoscaler.target_deployment}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(autoscaler.status)}`}>
                        {autoscaler.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {autoscaler.current_replicas}/{autoscaler.max_replicas}
                      </span>
                      <span className="text-xs text-gray-400">
                        Min: {autoscaler.min_replicas}
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">Metrics:</p>
                      <div className="flex space-x-4 mt-1">
                        <span className="text-xs text-gray-300">
                          CPU: {autoscaler.target_cpu_utilization}%
                        </span>
                        <span className="text-xs text-gray-300">
                          Memory: {autoscaler.target_memory_utilization}%
                        </span>
                      </div>
                    </div>
                  </div>
                  <Settings className="w-5 h-5 text-yellow-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Cluster Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Eye className="mr-2 text-hades-primary" />
              Cluster Details
            </h2>
          </div>
          <div className="p-4">
            {selectedCluster ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedCluster.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Version</p>
                    <p className="text-white">{selectedCluster.version}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Provider</p>
                    <p className="text-white">{selectedCluster.provider}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Region</p>
                    <p className="text-white">{selectedCluster.region}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Status</p>
                    <span className={`px-2 py-1 rounded text-sm ${getStatusColor(selectedCluster.status)}`}>
                      {selectedCluster.status}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Nodes ({selectedCluster.nodes?.length || 0})</p>
                    <div className="space-y-2 mt-2">
                      {selectedCluster.nodes?.map((node, index) => (
                        <div key={index} className="p-2 bg-gray-900 rounded border border-gray-700">
                          <p className="text-white">{node.name}</p>
                          <p className="text-gray-400 text-sm">{node.status} • {node.ip}</p>
                          <div className="flex space-x-4 mt-1">
                            <span className="text-xs text-gray-300">
                              CPU: {node.resources?.cpu?.used || '0'}/{node.resources?.cpu?.capacity || '0'}
                            </span>
                            <span className="text-xs text-gray-300">
                              Memory: {node.resources?.memory?.used || '0'}/{node.resources?.memory?.capacity || '0'}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Network Configuration</p>
                    <div className="space-y-1 mt-2">
                      <p className="text-white text-sm">Pod CIDR: {selectedCluster.networking?.pod_cidr}</p>
                      <p className="text-white text-sm">Service CIDR: {selectedCluster.networking?.service_cidr}</p>
                      <p className="text-white text-sm">CNI: {selectedCluster.networking?.cni}</p>
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Security Features</p>
                    <div className="flex space-x-2 mt-2">
                      <span className={`px-2 py-1 rounded text-xs ${selectedCluster.security?.rbac_enabled ? 'text-green-400 bg-green-900/20 border-green-500/20' : 'text-gray-400 bg-gray-900/20 border-gray-500/20'}`}>
                        RBAC
                      </span>
                      <span className={`px-2 py-1 rounded text-xs ${selectedCluster.security?.network_policy ? 'text-green-400 bg-green-900/20 border-green-500/20' : 'text-gray-400 bg-gray-900/20 border-gray-500/20'}`}>
                        Network Policy
                      </span>
                      <span className={`px-2 py-1 rounded text-xs ${selectedCluster.security?.encryption_at_rest ? 'text-green-400 bg-green-900/20 border-green-500/20' : 'text-gray-400 bg-gray-900/20 border-gray-500/20'}`}>
                        Encryption
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            ) : selectedDeployment ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">Deployment: {selectedDeployment.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Image</p>
                    <p className="text-white">{selectedDeployment.image}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Namespace</p>
                    <p className="text-white">{selectedDeployment.namespace}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Replicas</p>
                    <p className="text-white">{selectedDeployment.ready_replicas}/{selectedDeployment.replicas} ready</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Status</p>
                    <span className={`px-2 py-1 rounded text-sm ${getStatusColor(selectedDeployment.status)}`}>
                      {selectedDeployment.status}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Resource Requests</p>
                    <div className="space-y-1 mt-2">
                      <p className="text-white text-sm">CPU: {selectedDeployment.resources?.requests?.cpu || 'N/A'}</p>
                      <p className="text-white text-sm">Memory: {selectedDeployment.resources?.requests?.memory || 'N/A'}</p>
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Resource Limits</p>
                    <div className="space-y-1 mt-2">
                      <p className="text-white text-sm">CPU: {selectedDeployment.resources?.limits?.cpu || 'N/A'}</p>
                      <p className="text-white text-sm">Memory: {selectedDeployment.resources?.limits?.memory || 'N/A'}</p>
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Deployment Strategy</p>
                    <p className="text-white">{selectedDeployment.strategy?.type}</p>
                    {selectedDeployment.strategy?.rolling_update && (
                      <div className="mt-2">
                        <p className="text-white text-sm">Max Unavailable: {selectedDeployment.strategy.rolling_update.max_unavailable}</p>
                        <p className="text-white text-sm">Max Surge: {selectedDeployment.strategy.rolling_update.max_surge}</p>
                      </div>
                    )}
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Environment Variables</p>
                    <div className="mt-2">
                      <pre className="text-xs text-gray-300 bg-gray-900 p-2 rounded border border-gray-700">
                        {JSON.stringify(selectedDeployment.environment, null, 2)}
                      </pre>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Server className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select a cluster or deployment to view details</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Kubernetes
