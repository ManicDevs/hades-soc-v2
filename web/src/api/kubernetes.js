// Cloud-Native Kubernetes Deployment API
import API_CONFIG from './config'

export const kubernetesAPI = {
  // Get clusters
  getClusters: async () => {
    return await API_CONFIG.request('/kubernetes/clusters')
  },

  // Get deployments
  getDeployments: async () => {
    return await API_CONFIG.request('/kubernetes/deployments')
  },

  // Get services
  getServices: async () => {
    return await API_CONFIG.request('/kubernetes/services')
  },

  // Get autoscalers
  getAutoscalers: async () => {
    return await API_CONFIG.request('/kubernetes/autoscalers')
  },

  // Deploy application
  deployApplication: async (deploymentData) => {
    return await API_CONFIG.request('/kubernetes/deployments', {
      method: 'POST',
      body: JSON.stringify(deploymentData)
    })
  },

  // Scale deployment
  scaleDeployment: async (deploymentId, replicas) => {
    return await API_CONFIG.request('/kubernetes/scale', {
      method: 'POST',
      body: JSON.stringify({ deployment_id: deploymentId, replicas })
    })
  },

  // Get cluster status
  getClusterStatus: async () => {
    return await API_CONFIG.request('/kubernetes/status')
  },

  // Get deployment logs
  getDeploymentLogs: async (deploymentId) => {
    return await API_CONFIG.request(`/kubernetes/deployments/${deploymentId}/logs`)
  },

  // Get node details
  getNodeDetails: async (clusterId, nodeId) => {
    return await API_CONFIG.request(`/kubernetes/clusters/${clusterId}/nodes/${nodeId}`)
  },

  // Update deployment
  updateDeployment: async (deploymentId, deploymentData) => {
    return await API_CONFIG.request(`/kubernetes/deployments/${deploymentId}`, {
      method: 'PUT',
      body: JSON.stringify(deploymentData)
    })
  }
}
