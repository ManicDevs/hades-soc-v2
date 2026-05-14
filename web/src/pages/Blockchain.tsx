import { useState, useEffect } from 'react'
import { Shield, Hash, CheckCircle, AlertTriangle, Database, Link, Eye } from 'lucide-react'

function Blockchain() {
  const [auditLogs, setAuditLogs] = useState<any[]>([])
  const [blocks, setBlocks] = useState<any[]>([])
  const [transactions, setTransactions] = useState<any[]>([])
  const [integrity, setIntegrity] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedLog, setSelectedLog] = useState<any>(null)
  const [selectedBlock, setSelectedBlock] = useState<any>(null)

  useEffect(() => {
    fetchBlockchainData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchBlockchainData()
    }, 12000)

    return () => clearInterval(interval)
  }, [])

  const fetchBlockchainData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [logsData, blocksData, transactionsData, integrityData] = await Promise.all([
        API_CONFIG.request('/blockchain/audit-logs'),
        API_CONFIG.request('/blockchain/blocks'),
        API_CONFIG.request('/blockchain/transactions'),
        API_CONFIG.request('/blockchain/integrity')
      ])
      
      setAuditLogs(logsData.logs || [])
      setBlocks(blocksData.blocks || [])
      setTransactions(transactionsData.transactions || [])
      setIntegrity(integrityData)
    } catch (error) {
      setError('Failed to fetch blockchain data')
      console.error('Blockchain fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const verifyIntegrity = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      const result = await API_CONFIG.request('/blockchain/verify', {
        method: 'POST'
      })
      setIntegrity(result)
    } catch (error) {
      console.error('Failed to verify integrity:', error)
    }
  }

  const getStatusColor = (status: any) => {
    switch (status) {
      case 'verified':
        return 'text-green-400 bg-green-900/20 border-green-500/20'
      case 'pending':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      case 'corrupted':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getHashColor = (hash: any) => {
    return hash ? 'text-hades-primary' : 'text-gray-400'
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading blockchain data...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchBlockchainData} className="hades-button-primary">
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
          <Hash className="mr-3 text-hades-primary" />
          Blockchain Audit Logging & Immutability
        </h1>
        <p className="text-gray-400">Cryptographic audit trails with blockchain-based integrity verification</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Total Blocks</p>
              <p className="text-2xl font-bold text-white">{blocks.length}</p>
            </div>
            <Database className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Audit Logs</p>
              <p className="text-2xl font-bold text-white">{auditLogs.length}</p>
            </div>
            <Shield className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Transactions</p>
              <p className="text-2xl font-bold text-white">{transactions.length}</p>
            </div>
            <Link className="w-8 h-8 text-blue-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Integrity</p>
              <p className={`text-2xl font-bold ${integrity?.valid ? 'text-green-400' : 'text-red-400'}`}>
                {integrity?.valid ? 'Valid' : 'Invalid'}
              </p>
            </div>
            <CheckCircle className={`w-8 h-8 ${integrity?.valid ? 'text-green-400' : 'text-red-400'}`} />
          </div>
        </div>
      </div>

      {/* Integrity Verification */}
      <div className="bg-gray-800 rounded-lg border border-gray-700 p-4 mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-green-400" />
              Chain Integrity
            </h2>
            <p className="text-gray-400 text-sm mt-1">
              Last verified: {integrity?.last_verified ? new Date(integrity.last_verified).toLocaleString() : 'Never'}
            </p>
          </div>
          <button 
            onClick={verifyIntegrity}
            className="hades-button-primary"
          >
            Verify Integrity
          </button>
        </div>
        {integrity && (
          <div className="mt-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-white font-medium">
                  Status: <span className={integrity.valid ? 'text-green-400' : 'text-red-400'}>
                    {integrity.valid ? 'Chain Intact' : 'Chain Compromised'}
                  </span>
                </p>
                <p className="text-gray-400 text-sm">
                  Blocks verified: {integrity.verified_blocks || 0}/{blocks.length}
                </p>
              </div>
              {integrity.valid ? (
                <CheckCircle className="w-6 h-6 text-green-400" />
              ) : (
                <AlertTriangle className="w-6 h-6 text-red-400" />
              )}
            </div>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Audit Logs */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-green-400" />
              Audit Logs
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {auditLogs.map((log, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{log.action}</h3>
                    <p className="text-gray-400 text-sm mt-1">{log.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getStatusColor(log.status)}`}>
                        {log.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        {new Date(log.timestamp).toLocaleString()}
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">Hash:</p>
                      <p className={`text-xs font-mono ${getHashColor(log.hash)}`}>
                        {log.hash || 'Pending...'}
                      </p>
                    </div>
                  </div>
                  <button 
                    onClick={() => setSelectedLog(log)}
                    className="ml-3 p-2 text-gray-400 hover:text-white"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Blockchain Blocks */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Database className="mr-2 text-hades-primary" />
              Blockchain Blocks
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {blocks.slice().reverse().map((block, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">Block #{block.index}</h3>
                    <p className="text-gray-400 text-sm mt-1">
                      {block.transactions?.length || 0} transactions
                    </p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        {new Date(block.timestamp).toLocaleString()}
                      </span>
                      <span className={`text-xs ${getStatusColor(block.status)}`}>
                        {block.status}
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">Hash:</p>
                      <p className={`text-xs font-mono ${getHashColor(block.hash)}`}>
                        {block.hash || 'Mining...'}
                      </p>
                    </div>
                  </div>
                  <button 
                    onClick={() => setSelectedBlock(block)}
                    className="ml-3 p-2 text-gray-400 hover:text-white"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Transactions */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Link className="mr-2 text-blue-400" />
              Transactions
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {transactions.slice().reverse().map((tx, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{tx.type}</h3>
                    <p className="text-gray-400 text-sm">{tx.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`text-xs ${getStatusColor(tx.status)}`}>
                        {tx.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        Block: #{tx.block_index}
                      </span>
                    </div>
                  </div>
                  <Hash className="w-5 h-5 text-blue-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Eye className="mr-2 text-hades-primary" />
              Details
            </h2>
          </div>
          <div className="p-4">
            {selectedLog ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">Audit Log Details</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Action</p>
                    <p className="text-white">{selectedLog.action}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedLog.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">User</p>
                    <p className="text-white">{selectedLog.user}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Timestamp</p>
                    <p className="text-white">{new Date(selectedLog.timestamp).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Hash</p>
                    <p className={`text-sm font-mono ${getHashColor(selectedLog.hash)}`}>
                      {selectedLog.hash || 'Pending...'}
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Status</p>
                    <span className={`px-2 py-1 rounded text-sm ${getStatusColor(selectedLog.status)}`}>
                      {selectedLog.status}
                    </span>
                  </div>
                </div>
              </div>
            ) : selectedBlock ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">Block Details</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Block Index</p>
                    <p className="text-white">#{selectedBlock.index}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Transactions</p>
                    <p className="text-white">{selectedBlock.transactions?.length || 0}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Timestamp</p>
                    <p className="text-white">{new Date(selectedBlock.timestamp).toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Previous Hash</p>
                    <p className={`text-sm font-mono ${getHashColor(selectedBlock.previous_hash)}`}>
                      {selectedBlock.previous_hash}
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Hash</p>
                    <p className={`text-sm font-mono ${getHashColor(selectedBlock.hash)}`}>
                      {selectedBlock.hash || 'Mining...'}
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Nonce</p>
                    <p className="text-white">{selectedBlock.nonce || 'Mining...'}</p>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Eye className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select an audit log or block to view details</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Blockchain
