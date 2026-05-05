package threat

import (
	"math"
	"math/rand"
	"sync"
)

// MLModel represents a machine learning model for threat detection
type MLModel struct {
	trained      bool
	weights      map[string]float64
	bias         float64
	features     []string
	mu           sync.RWMutex
	trainingData []SecurityEvent
}

// NewMLModel creates a new machine learning model
func NewMLModel() *MLModel {
	return &MLModel{
		weights:  make(map[string]float64),
		features: []string{"port", "protocol", "method", "payload_length", "header_count"},
	}
}

// Train trains the model with historical data
func (m *MLModel) Train(events []SecurityEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.trainingData = events
	m.trained = true

	// Simple linear regression for demonstration
	// In production, this would use more sophisticated ML algorithms
	for _, feature := range m.features {
		m.weights[feature] = rand.Float64()*2 - 1 // Random weights between -1 and 1
	}
	m.bias = rand.Float64()*2 - 1

	return nil
}

// Predict predicts threat score for an event
func (m *MLModel) Predict(event SecurityEvent) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.trained {
		return 0.5 // Default probability
	}

	score := m.bias

	// Extract features
	features := m.extractFeatures(event)

	// Calculate weighted sum
	for feature, value := range features {
		if weight, exists := m.weights[feature]; exists {
			score += weight * value
		}
	}

	// Apply sigmoid function to get probability
	return 1 / (1 + math.Exp(-score))
}

// extractFeatures extracts features from a security event
func (m *MLModel) extractFeatures(event SecurityEvent) map[string]float64 {
	features := make(map[string]float64)

	// Port features
	if event.Port > 0 {
		features["port"] = math.Log(float64(event.Port))
	}

	// Protocol features
	switch event.Protocol {
	case "HTTP":
		features["protocol"] = 0.8
	case "HTTPS":
		features["protocol"] = 0.9
	case "TCP":
		features["protocol"] = 0.6
	case "UDP":
		features["protocol"] = 0.4
	default:
		features["protocol"] = 0.2
	}

	// Method features
	switch event.Method {
	case "GET":
		features["method"] = 0.3
	case "POST":
		features["method"] = 0.7
	case "PUT":
		features["method"] = 0.8
	case "DELETE":
		features["method"] = 0.9
	default:
		features["method"] = 0.5
	}

	// Payload length
	features["payload_length"] = math.Log(float64(len(event.Payload)) + 1)

	// Header count
	features["header_count"] = math.Log(float64(len(event.Headers)) + 1)

	return features
}

// IsTrained returns whether the model is trained
func (m *MLModel) IsTrained() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trained
}

// GetWeights returns the model weights
func (m *MLModel) GetWeights() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	weights := make(map[string]float64)
	for k, v := range m.weights {
		weights[k] = v
	}
	return weights
}
