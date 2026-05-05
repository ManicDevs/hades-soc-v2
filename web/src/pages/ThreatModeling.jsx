import React, { useState, useEffect } from 'react'
import { Shield, Target, AlertTriangle, Play, Eye, Settings, BarChart3, Lock } from 'lucide-react'

function ThreatModeling({ user }) {
  const [threatModels, setThreatModels] = useState([])
  const [attackScenarios, setAttackScenarios] = useState([])
  const [vulnerabilities, setVulnerabilities] = useState([])
  const [mitigations, setMitigations] = useState([])
  const [simulations, setSimulations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [selectedModel, setSelectedModel] = useState(null)
  const [selectedScenario, setSelectedScenario] = useState(null)

  useEffect(() => {
    fetchThreatModelingData()
    
    // Set up real-time updates
    const interval = setInterval(() => {
      fetchThreatModelingData()
    }, 10000)

    return () => clearInterval(interval)
  }, [])

  const fetchThreatModelingData = async () => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      
      const [modelsData, scenariosData, vulnsData, mitigationsData, simulationsData] = await Promise.all([
        API_CONFIG.request('/threat/models'),
        API_CONFIG.request('/threat/scenarios'),
        API_CONFIG.request('/threat/vulnerabilities'),
        API_CONFIG.request('/threat/mitigations'),
        API_CONFIG.request('/threat/simulations')
      ])
      
      setThreatModels(modelsData.models || {})
      setAttackScenarios(scenariosData.scenarios || {})
      setVulnerabilities(vulnsData.vulnerabilities || {})
      setMitigations(mitigationsData.mitigations || {})
      setSimulations(simulationsData.simulations || {})
    } catch (error) {
      setError('Failed to fetch threat modeling data')
      console.error('Threat modeling fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  const runSimulation = async (scenarioId) => {
    try {
      const { default: API_CONFIG } = await import('../api/config')
      const result = await API_CONFIG.request('/threat/simulations', {
        method: 'POST',
        body: JSON.stringify({ scenario_id: scenarioId })
      })
      
      // Refresh data
      fetchThreatModelingData()
      return result
    } catch (error) {
      console.error('Failed to run simulation:', error)
    }
  }

  const getRiskColor = (riskLevel) => {
    switch (riskLevel) {
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

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'critical':
        return 'text-red-400'
      case 'high':
        return 'text-orange-400'
      case 'medium':
        return 'text-yellow-400'
      case 'low':
        return 'text-blue-400'
      default:
        return 'text-gray-400'
    }
  }

  if (loading) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-hades-primary mx-auto mb-4"></div>
          <p className="text-gray-400">Loading threat modeling...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <button onClick={fetchThreatModelingData} className="hades-button-primary">
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
          <Shield className="mr-3 text-hades-primary" />
          Advanced Threat Modeling & Attack Simulation
        </h1>
        <p className="text-gray-400">MITRE ATT&CK-based threat modeling with attack scenario simulation</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Threat Models</p>
              <p className="text-2xl font-bold text-white">{Object.keys(threatModels).length}</p>
            </div>
            <Shield className="w-8 h-8 text-hades-primary" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Attack Scenarios</p>
              <p className="text-2xl font-bold text-white">{Object.keys(attackScenarios).length}</p>
            </div>
            <Target className="w-8 h-8 text-red-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Vulnerabilities</p>
              <p className="text-2xl font-bold text-white">{Object.keys(vulnerabilities).length}</p>
            </div>
            <AlertTriangle className="w-8 h-8 text-orange-400" />
          </div>
        </div>
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Simulations Run</p>
              <p className="text-2xl font-bold text-white">{Object.keys(simulations).length}</p>
            </div>
            <Play className="w-8 h-8 text-green-400" />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Threat Models */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Shield className="mr-2 text-hades-primary" />
              Threat Models
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(threatModels).map((model, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{model.name}</h3>
                    <p className="text-gray-400 text-sm mt-1">{model.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        {model.assets?.length || 0} assets
                      </span>
                      <span className="text-xs text-gray-400">
                        {model.threats?.length || 0} threats
                      </span>
                      <span className="text-xs text-gray-400">
                        {model.risks?.length || 0} risks
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">Risk Levels:</p>
                      <div className="flex space-x-2 mt-1">
                        {model.risks?.slice(0, 3).map((risk, i) => (
                          <span key={i} className={`px-2 py-1 rounded text-xs ${getRiskColor(risk.risk_level)}`}>
                            {risk.risk_level}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                  <button 
                    onClick={() => setSelectedModel(model)}
                    className="ml-3 p-2 text-gray-400 hover:text-white"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Attack Scenarios */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Target className="mr-2 text-red-400" />
              Attack Scenarios
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(attackScenarios).map((scenario, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-white font-medium">{scenario.name}</h3>
                    <p className="text-gray-400 text-sm mt-1">{scenario.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        {scenario.phase}
                      </span>
                      <span className="text-xs text-gray-400">
                        {scenario.techniques?.length || 0} techniques
                      </span>
                    </div>
                    <div className="mt-2">
                      <p className="text-xs text-gray-400">MITRE Techniques:</p>
                      <div className="flex flex-wrap gap-1 mt-1">
                        {scenario.techniques?.slice(0, 3).map((tech, i) => (
                          <span key={i} className="px-2 py-1 rounded text-xs bg-gray-800 text-gray-300">
                            {tech.technique_id}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button 
                      onClick={() => runSimulation(scenario.id)}
                      className="p-2 text-green-400 hover:text-green-300"
                      title="Run Simulation"
                    >
                      <Play className="w-4 h-4" />
                    </button>
                    <button 
                      onClick={() => setSelectedScenario(scenario)}
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

        {/* Vulnerabilities */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <AlertTriangle className="mr-2 text-orange-400" />
              Vulnerabilities
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(vulnerabilities).map((vuln, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{vuln.name}</h3>
                    <p className="text-gray-400 text-sm">{vuln.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${getRiskColor(vuln.severity)}`}>
                        {vuln.severity}
                      </span>
                      <span className="text-xs text-gray-400">
                        CVSS: {vuln.cvss}
                      </span>
                      <span className="text-xs text-gray-400">
                        {vuln.cve}
                      </span>
                    </div>
                  </div>
                  <AlertTriangle className={`w-5 h-5 ${getSeverityColor(vuln.severity)}`} />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Mitigations */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <Lock className="mr-2 text-green-400" />
              Mitigations
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(mitigations).map((mitigation, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{mitigation.name}</h3>
                    <p className="text-gray-400 text-sm">{mitigation.description}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className="text-xs text-gray-400">
                        {mitigation.type}
                      </span>
                      <span className="text-xs text-gray-400">
                        Priority: {mitigation.priority}
                      </span>
                      <span className="text-xs text-gray-400">
                        {mitigation.effectiveness}% effective
                      </span>
                    </div>
                  </div>
                  <Lock className="w-5 h-5 text-green-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Simulation Results */}
        <div className="bg-gray-800 rounded-lg border border-gray-700">
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center">
              <BarChart3 className="mr-2 text-blue-400" />
              Simulation Results
            </h2>
          </div>
          <div className="p-4 max-h-96 overflow-y-auto">
            {Object.values(simulations).slice().reverse().map((simulation, index) => (
              <div key={index} className="mb-4 p-3 bg-gray-900 rounded-lg border border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-white font-medium">{simulation.name}</h3>
                    <p className="text-gray-400 text-sm">{simulation.scenario_id}</p>
                    <div className="flex items-center mt-2 space-x-4">
                      <span className={`px-2 py-1 rounded text-xs ${simulation.status === 'completed' ? 'text-green-400 bg-green-900/20 border-green-500/20' : 'text-yellow-400 bg-yellow-900/20 border-yellow-500/20'}`}>
                        {simulation.status}
                      </span>
                      <span className="text-xs text-gray-400">
                        Success Rate: {Math.round(simulation.metrics?.success_rate * 100) || 0}%
                      </span>
                      <span className="text-xs text-gray-400">
                        Duration: {simulation.metrics?.duration || 0}s
                      </span>
                    </div>
                    <div className="mt-2">
                      <div className="w-full bg-gray-700 rounded-full h-2">
                        <div 
                          className="bg-blue-400 h-2 rounded-full" 
                          style={{ width: `${(simulation.metrics?.success_rate || 0) * 100}%` }}
                        ></div>
                      </div>
                    </div>
                  </div>
                  <BarChart3 className="w-5 h-5 text-blue-400" />
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
            {selectedModel ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">Threat Model: {selectedModel.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedModel.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Assets</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedModel.assets?.map((asset, index) => (
                        <li key={index}>{asset.name} - Value: ${asset.value}</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Threats</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedModel.threats?.map((threat, index) => (
                        <li key={index}>{threat.name} - {threat.likelihood} likelihood</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Risk Assessment</p>
                    <div className="space-y-2">
                      {selectedModel.risks?.map((risk, index) => (
                        <div key={index} className="p-2 bg-gray-900 rounded border border-gray-700">
                          <p className="text-white">{risk.asset}</p>
                          <p className="text-gray-400 text-sm">{risk.threat}</p>
                          <span className={`px-2 py-1 rounded text-xs ${getRiskColor(risk.risk_level)}`}>
                            {risk.risk_level} (Score: {risk.risk_score})
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            ) : selectedScenario ? (
              <div>
                <h3 className="text-lg font-medium text-white mb-3">Attack Scenario: {selectedScenario.name}</h3>
                <div className="space-y-3">
                  <div>
                    <p className="text-gray-400 text-sm">Description</p>
                    <p className="text-white">{selectedScenario.description}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Phase</p>
                    <p className="text-white">{selectedScenario.phase}</p>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">MITRE ATT&CK Techniques</p>
                    <div className="space-y-2">
                      {selectedScenario.techniques?.map((technique, index) => (
                        <div key={index} className="p-2 bg-gray-900 rounded border border-gray-700">
                          <p className="text-white">{technique.name}</p>
                          <p className="text-gray-400 text-sm">{technique.technique_id} - {technique.tactic}</p>
                          <p className="text-gray-400 text-xs">{technique.description}</p>
                        </div>
                      ))}
                    </div>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Prerequisites</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedScenario.prerequisites?.map((prereq, index) => (
                        <li key={index}>{prereq}</li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <p className="text-gray-400 text-sm">Detection Methods</p>
                    <ul className="text-white text-sm list-disc list-inside">
                      {selectedScenario.detection?.map((method, index) => (
                        <li key={index}>{method}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400 py-8">
                <Eye className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>Select a threat model or attack scenario to view details</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default ThreatModeling
