import { useState, useEffect } from 'react'
import { BarChart3, TrendingUp, Brain, Activity, PieChart, LineChart, Target, Zap } from 'lucide-react'

function Analytics() {
  const [analyticsData, setAnalyticsData] = useState<any>(null)
  const [mlInsights, setMlInsights] = useState<any[]>([])
  const [predictions, setPredictions] = useState<any[]>([])
  const [metrics, setMetrics] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedMetric, setSelectedMetric] = useState<any>(null)

  useEffect(() => {
    fetchAnalyticsData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchAnalyticsData()
    }, 15000)

    return () => clearInterval(interval)
  }, [])

  const fetchAnalyticsData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [analytics, mlData, predictionData, metricsData] = await Promise.all([
        API_CONFIG.request('/analytics/overview'),
        API_CONFIG.request('/analytics/ml-insights'),
        API_CONFIG.request('/analytics/predictions'),
        API_CONFIG.request('/analytics/metrics')
      ])
      
      setAnalyticsData(analytics)
      setMlInsights(mlData.insights || [])
      setPredictions(predictionData.predictions || [])
      setMetrics(metricsData.metrics || [])
    } catch (error) {
      setError('Failed to fetch analytics data')
    } finally {
      setLoading(false)
    }
  }

  const getTrendColor = (trend: any) => {
    switch (trend) {
      case 'up':
        return 'text-green-400'
      case 'down':
        return 'text-red-400'
      case 'stable':
        return 'text-yellow-400'
      default:
        return 'text-gray-400'
    }
  }

  const getAccuracyColor = (accuracy: any) => {
    if (accuracy >= 95) return 'text-green-400'
    if (accuracy >= 85) return 'text-yellow-400'
    return 'text-red-400'
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading analytics...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchAnalyticsData} className="hades-button-primary">
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
          <BarChart3 className="mr-3 text-hades-primary" />
          Advanced Analytics & Machine Learning
        </h1>
        <p className="text-gray-400">Predictive analytics and ML-powered security insights</p>
      </div>

      {/* Key Analytics Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">ML Models Active</p>
              <p className="text-2xl font-bold text-white">{analyticsData?.overview?.total_events || 0}</p>
            </div>
            <Brain className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Predictions Made</p>
              <p className="text-2xl font-bold text-white">{analyticsData?.overview?.critical_alerts || 0}</p>
            </div>
            <Target className="w-8 h-8 text-green-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Avg Accuracy</p>
              <p className={`text-2xl font-bold ${getAccuracyColor(analyticsData?.overview?.ml_accuracy || 0)}`}>
                {analyticsData?.overview?.ml_accuracy || 0}%
              </p>
            </div>
            <TrendingUp className="w-8 h-8 text-yellow-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Data Points</p>
              <p className="text-2xl font-bold text-white">{analyticsData?.overview?.prediction_confidence || 0}</p>
            </div>
            <Activity className="w-8 h-8 text-blue-400" />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* ML Insights */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Brain className="mr-2 text-hades-primary" />
              Machine Learning Insights
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {mlInsights.map((insight, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{insight.type}</h3>
                    <p className="text-gray-400 text-sm mt-1">{insight.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`text-xs ${getAccuracyColor(insight.confidence)}`}>
                        {insight.confidence}% confidence
                      </span>
                      <span className={`text-xs ${getTrendColor(insight.severity)}`}>
                        {insight.severity} severity
                      </span>
                    </div>
                  </div>
                  <Zap className="w-5 h-5 text-hades-primary" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Predictions */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Target className="mr-2 text-green-400" />
              Predictive Analytics
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
                      <span className={`text-xs ${getAccuracyColor(prediction.confidence)}`}>
                        {prediction.confidence}% confidence
                      </span>
                      <span className="text-xs text-gray-400">
                        {prediction.timeframe}
                      </span>
                    </div>
                  </div>
                  <TrendingUp className="w-5 h-5 text-green-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Performance Metrics */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <LineChart className="mr-2 text-blue-400" />
              Performance Metrics
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Array.isArray(metrics) ? metrics.map((metric, index) => (
              <div 
                key={index} 
                className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700 cursor-pointer hover:border-hades-primary"
                onClick={() => setSelectedMetric({ name: metric.metric_name || `Metric ${index}`, value: metric.value || 0, description: metric.description || `${metric.metric_name || 'Metric'} performance metric` })}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{metric.metric_name || `Metric ${index}`}</h3>
                    <p className="text-gray-400 text-sm">{metric.description || `${metric.metric_name || 'Metric'} performance metric`}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        Current: {typeof metric.value === 'number' ? metric.value : 'N/A'}%
                      </span>
                      <span className={`text-xs ${getAccuracyColor(typeof metric.value === 'number' ? metric.value : 0)}`}>
                        {typeof metric.value === 'number' ? metric.value : 'N/A'}% performance
                      </span>
                    </div>
                  </div>
                  <Activity className="w-5 h-5 text-blue-400" />
                </div>
              </div>
            )) : (
              <div className="text-gray-400 text-sm">No metrics available</div>
            )}
          </div>
        </div>

        {/* Metric Details Panel */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <PieChart className="mr-2 text-hades-primary" />
              Metric Analysis
            </h2>
          </div>
          <div className="p-4">
            {selectedMetric ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">{selectedMetric.name?.replace(/_/g, ' ').toUpperCase() || 'Metric'}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedMetric.description || 'Performance metric analysis'}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Current Value</p>
                    <p className="text-2xl font-bold text-hades-primary">{selectedMetric.value}%</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Performance Level</p>
                    <p className={`text-lg font-medium ${getAccuracyColor(selectedMetric.value)}`}>
                      {selectedMetric.value}% {selectedMetric.value >= 90 ? 'Excellent' : selectedMetric.value >= 75 ? 'Good' : 'Needs Improvement'}
                    </p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Historical Data</p>
                    <div className="mt-2 h-32 bg-gray-900 rounded border border-gray-700 flex items-center justify-center">
                      <p className="text-gray-400 text-sm">Chart visualization would go here</p>
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Recommendations</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedMetric.value >= 90 ? [
                      'Performance is excellent, maintain current configuration',
                      'Consider sharing best practices with other teams'
                    ] : selectedMetric.value >= 75 ? [
                      'Performance is good, monitor for trends',
                      'Consider optimization opportunities'
                    ] : [
                      'Performance needs improvement',
                      'Review configuration and consider optimization',
                      'Consult with technical team for assistance'
                    ].map((rec, index) => (
                        <li key={index}>{rec}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <PieChart className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select a metric to view detailed analysis</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default Analytics
