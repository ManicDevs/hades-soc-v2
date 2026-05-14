import { useState, useEffect } from 'react'
import { Brain, AlertTriangle, TrendingUp, Shield, Activity, Eye, Target, Zap } from 'lucide-react'

function ThreatIntelligence() {
  const [threatData, setThreatData] = useState<any>(null)
  const [anomalyData, setAnomalyData] = useState<any[]>([])
  const [predictions, setPredictions] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedThreat, setSelectedThreat] = useState<any>(null)

  useEffect(() => {
    fetchThreatIntelligenceData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchThreatIntelligenceData()
    }, 10000)

    return () => clearInterval(interval)
  }, [])

  const fetchThreatIntelligenceData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [threats, anomalies, mlPredictions] = await Promise.all([
        API_CONFIG.request('/ai/threats'),
        API_CONFIG.request('/ai/anomalies'),
        API_CONFIG.request('/ai/predictions')
      ])
      
      setThreatData(threats)
      setAnomalyData(anomalies.anomalies || [])
      setPredictions(mlPredictions.predictions || [])
    } catch (error) {
      setError('Failed to fetch threat intelligence data')
      console.error('Threat intelligence fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'critical':
        return 'text-red-400 bg-red-900/20 border-red-500/20'
      case 'high':
        return 'text-orange-400 bg-orange-900/20 border-orange-500/20'
      case 'medium':
        return 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'
      case 'low':
        return 'text-blue-400 bg-blue-900/20 border-blue-500/20'
      default:
        return 'text-gray-400 bg-gray-900/20 border-gray-500/20'
    }
  }

  const getConfidenceColor = (confidence) => {
    if (confidence >= 90) return 'text-green-400'
    if (confidence >= 70) return 'text-yellow-400'
    return 'text-red-400'
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading threat intelligence...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchThreatIntelligenceData} className="hades-button-primary">
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
          <Brain className="mr-3 text-hades-primary" />
          AI-Powered Threat Intelligence
        </h1>
        <p className="text-gray-400">Machine learning-driven threat analysis and anomaly detection</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">AI Threats Detected</p>
              <p className="text-2xl font-bold text-white">{threatData?.threats?.length || 0}</p>
            </div>
            <AlertTriangle className="w-8 h-8 text-red-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Anomalies Found</p>
              <p className="text-2xl font-bold text-white">{anomalyData.length}</p>
            </div>
            <Activity className="w-8 h-8 text-yellow-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">ML Predictions</p>
              <p className="text-2xl font-bold text-white">{predictions.length}</p>
            </div>
            <TrendingUp className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">AI Accuracy</p>
              <p className="text-2xl font-bold text-white">{threatData?.accuracy || 0}%</p>
            </div>
            <Target className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* AI Detected Threats */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Brain className="mr-2 text-hades-primary" />
              AI Detected Threats
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {threatData?.threats?.map((threat, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{threat.name || 'Unknown Threat'}</h3>
                    <p className="text-gray-400 text-sm mt-1">{threat.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getSeverityColor(threat.severity)}`}>
                        {threat.severity}
                      </span>
                      <span className={`text-xs ${getConfidenceColor(threat.confidence)}`}>
                        {threat.confidence}% confidence
                      </span>
                    </div>
                  </div>
                  <button 
                    onClick={() => setSelectedThreat(threat)}
                    className="ml-3 p-2 text-gray-400 hover:text-white"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Anomaly Detection */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Activity className="mr-2 text-yellow-400" />
              Anomaly Detection
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {anomalyData.map((anomaly, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{anomaly.type}</h3>
                    <p className="text-gray-400 text-sm">{anomaly.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        Score: {anomaly.score}
                      </span>
                      <span className={`text-xs ${getConfidenceColor(anomaly.confidence)}`}>
                        {anomaly.confidence}% confidence
                      </span>
                    </div>
                  </div>
                  <Zap className="w-5 h-5 text-yellow-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* ML Predictions */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <TrendingUp className="mr-2 text-green-400" />
              ML Predictions
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {predictions.map((prediction, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{prediction.type}</h3>
                    <p className="text-gray-400 text-sm">{prediction.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`text-xs ${getConfidenceColor(prediction.confidence)}`}>
                        {prediction.confidence}% confidence
                      </span>
                      <span className="text-xs text-gray-400">
                        Timeframe: {prediction.timeframe}
                      </span>
                    </div>
                  </div>
                  <Target className="w-5 h-5 text-green-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Threat Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-hades-primary" />
              Threat Analysis
            </h2>
          </div>
          <div className="p-4">
            {selectedThreat ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedThreat.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedThreat.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Severity</p>
                    <span className={`px-2 py-1 rounded text-sm ${getSeverityColor(selectedThreat.severity)}`}>
                      {selectedThreat.severity}
                    </span>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">AI Confidence</p>
                    <p className={`text-lg font-medium ${getConfidenceColor(selectedThreat.confidence)}`}>
                      {selectedThreat.confidence}%
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Recommended Actions</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedThreat.recommendations?.map((rec, index) => (
                        <li key={index}>{rec}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Eye className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select a threat to view detailed analysis</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default ThreatIntelligence
