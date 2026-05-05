import React, { useState, useEffect } from 'react'
import { Atom, Shield, Key, Lock, Zap, CheckCircle, AlertTriangle, RefreshCw, Eye } from 'lucide-react'

function Quantum({ user }) {
  const [algorithms, setAlgorithms] = useState([])
  const [keys, setKeys] = useState([])
  const [certificates, setCertificates] = useState([])
  const [securityMetrics, setSecurityMetrics] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [selectedAlgorithm, setSelectedAlgorithm] = useState(null)

  useEffect(() => {
    fetchQuantumData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchQuantumData()
    }, 15000)

    return () => clearInterval(interval)
  }, [])

  const fetchQuantumData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [algorithmsData, keysData, certificatesData, metricsData] = await Promise.all([
        API_CONFIG.request('/quantum/algorithms'),
        API_CONFIG.request('/quantum/keys'),
        API_CONFIG.request('/quantum/certificates'),
        API_CONFIG.request('/quantum/metrics')
      ])
      
      setAlgorithms(algorithmsData.algorithms || [])
      setKeys(keysData.keys || [])
      setCertificates(certificatesData.certificates || [])
      setSecurityMetrics(metricsData)
    } catch (error) {
      setError('Failed to fetch quantum cryptography data')
      console.error('Quantum fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const generateKey = async (algorithmId) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/quantum/generate-key', {
        method: 'POST',
        body: JSON.stringify({ algorithm_id: algorithmId })
      })
      
      // Refresh data
      fetchQuantumData()
    } catch (error) {
      console.error('Failed to generate key:', error)
    }
  }

  const rotateKeys = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      await API_CONFIG.request('/quantum/rotate-keys', {
        method: 'POST'
      })
      
      // Refresh data
      fetchQuantumData()
    } catch (error) {
      console.error('Failed to rotate keys:', error)
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case 'active':
        return 'text-green-400 bg-green-900/20 border-green-500/20'
      case 'pending':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      case 'expired':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      case 'inactive':
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getSecurityColor = (level) => {
    switch (level) {
      case 'quantum-safe':
        return 'text-green-400'
      case 'quantum-resistant':
        return 'text-blue-400'
      case 'quantum-vulnerable':
        return 'text-red-400'
      default:
        return 'text-gray-400'
    }
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading quantum cryptography...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchQuantumData} className="hades-button-primary">
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
          <Atom className="mr-3 text-hades-primary" />
          Quantum-Resistant Cryptography
        </h1>
        <p className="text-gray-400">Post-quantum encryption and quantum-safe cryptographic operations</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Active Algorithms</p>
              <p className="text-2xl font-bold text-white">{algorithms.filter(a => a.status === 'active').length}</p>
            </div>
            <Atom className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Quantum Keys</p>
              <p className="text-2xl font-bold text-white">{keys.filter(k => k.status === 'active').length}</p>
            </div>
            <Key className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Certificates</p>
              <p className="text-2xl font-bold text-white">{certificates.filter(c => c.status === 'active').length}</p>
            </div>
            <Shield className="w-8 h-8 text-blue-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Security Level</p>
              <p className={`text-2xl font-bold ${getSecurityColor(securityMetrics?.security_level)}`}>
                {securityMetrics?.security_level || 'Unknown'}
              </p>
            </div>
            <Lock className={`w-8 h-8 ${getSecurityColor(securityMetrics?.security_level)}`} />
          </div>
        </div>
      </div>

      {/* Quantum Security Status */}
      <div className="bg-gray-800 rounded-lg border border-gray-700 p-4 mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-green-400" />
              Quantum Security Status
            </h2>
            <p className="text-gray-400 text-sm mt-1">
              Last assessment: {securityMetrics?.last_assessment ? new Date(securityMetrics.last_assessment).toLocaleString() : 'Never'}
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <button 
              onClick={rotateKeys}
              className="hades-button-primary"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Rotate Keys
            </button>
          </div>
        </div>
        {securityMetrics && (
          <div className="mt-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <p className="text-white font-medium">Security Level</p>
                <p className={`text-lg ${getSecurityColor(securityMetrics.security_level)}`}>
                  {securityMetrics.security_level}
                </p>
              </div>
              <div>
                <p className="text-white font-medium">Threat Assessment</p>
                <p className={`text-lg ${securityMetrics.quantum_threat === 'low' ? 'text-green-400' : securityMetrics.quantum_threat === 'medium' ? 'text-yellow-400' : 'text-red-400'}`}>
                  {securityMetrics.quantum_threat}
                </p>
              </div>
              <div>
                <p className="text-white font-medium">Compliance</p>
                <p className={`text-lg ${securityMetrics.compliant ? 'text-green-400' : 'text-red-400'}`}>
                  {securityMetrics.compliant ? 'Compliant' : 'Non-Compliant'}
                </p>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Quantum Algorithms */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Atom className="mr-2 text-hades-primary" />
              Quantum Algorithms
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {algorithms.map((algorithm, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{algorithm.name}</h3>
                    <p className="text-gray-400 text-sm mt-1">{algorithm.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(algorithm.status)}`}>
                        {algorithm.status}
                      </span>
                      <span className={`text-xs ${getSecurityColor(algorithm.security_level)}`}>
                        {algorithm.security_level}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button 
                      onClick={() => generateKey(algorithm.id)}
                      className="p-2 text-gray-400 hover:text-white"
                      title="Generate Key"
                    >
                      <Key className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={() => setSelectedAlgorithm(algorithm)}
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

        {/* Quantum Keys */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Key className="mr-2 text-green-400" />
              Quantum Keys
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {keys.map((key, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{key.algorithm}</h3>
                    <p className="text-gray-400 text-sm">Key ID: {key.id}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(key.status)}`}>
                        {key.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {key.key_size} bits
                      </span>
                      <span className="text-xs text-gray-400">
                        Expires: {new Date(key.expires_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <Zap className="w-5 h-5 text-green-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Quantum Certificates */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-blue-400" />
              Quantum Certificates
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {certificates.map((cert, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{cert.subject}</h3>
                    <p className="text-gray-400 text-sm">Issuer: {cert.issuer}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(cert.status)}`}>
                        {cert.status}
                      </span>
                      <span className={`text-xs ${getSecurityColor(cert.security_level)}`}>
                        {cert.security_level}
                      </span>
                      <span className="text-xs text-gray-400">
                        Expires: {new Date(cert.expires_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <Lock className="w-5 h-5 text-blue-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Algorithm Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Eye className="mr-2 text-hades-primary" />
              Algorithm Details
            </h2>
          </div>
          <div className="p-4">
            {selectedAlgorithm ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedAlgorithm.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedAlgorithm.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Type</p>
                    <p className="text-white">{selectedAlgorithm.type}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Security Level</p>
                    <p className={`text-lg font-medium ${getSecurityColor(selectedAlgorithm.security_level)}`}>
                      {selectedAlgorithm.security_level}
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Key Size</p>
                    <p className="text-white">{selectedAlgorithm.key_size} bits</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Status</p>
                    <span className={`px-2 py-1 rounded text-sm ${getStatusColor(selectedAlgorithm.status)}`}>
                      {selectedAlgorithm.status}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Implementation</p>
                    <p className="text-white">{selectedAlgorithm.implementation}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Performance</p>
                    <div className="grid grid-cols-2 gap-4 mt-2">
                      <div>
                        <p className="text-xs text-gray-400">Key Generation</p>
                        <p className="text-white">{selectedAlgorithm.performance?.key_generation || 'N/A'}</p>
                      </div>
                      <div>
                        <p className="text-xs text-gray-400">Encryption</p>
                        <p className="text-white">{selectedAlgorithm.performance?.encryption || 'N/A'}</p>
                      </div>
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Quantum Resistance</p>
                    <p className="text-white">{selectedAlgorithm.quantum_resistance}</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Atom className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select an algorithm to view detailed information</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Quantum
