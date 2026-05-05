package threat

import (
	"math"
	"sync"
	"time"
)

// AnomalyDetector detects anomalies in security events
type AnomalyDetector struct {
	baseline    map[string]BaselineMetrics
	mu          sync.RWMutex
	initialized bool
}

// BaselineMetrics represents baseline metrics for anomaly detection
type BaselineMetrics struct {
	Mean        float64   `json:"mean"`
	StdDev      float64   `json:"std_dev"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	SampleCount int       `json:"sample_count"`
	LastUpdate  time.Time `json:"last_update"`
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		baseline: make(map[string]BaselineMetrics),
	}
}

// UpdateBaseline updates the baseline metrics
func (ad *AnomalyDetector) UpdateBaseline(metric string, values []float64) {
	if len(values) == 0 {
		return
	}

	ad.mu.Lock()
	defer ad.mu.Unlock()

	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)
	min, max := calculateMinMax(values)

	ad.baseline[metric] = BaselineMetrics{
		Mean:        mean,
		StdDev:      stdDev,
		Min:         min,
		Max:         max,
		SampleCount: len(values),
		LastUpdate:  time.Now(),
	}

	if len(ad.baseline) > 10 {
		ad.initialized = true
	}
}

// IsAnomaly checks if a value is anomalous
func (ad *AnomalyDetector) IsAnomaly(metric string, value float64) bool {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	baseline, exists := ad.baseline[metric]
	if !exists || !ad.initialized {
		return false
	}

	// Use Z-score for anomaly detection
	zScore := math.Abs(value-baseline.Mean) / baseline.StdDev

	// Consider anomalous if Z-score > 3 (3 standard deviations)
	return zScore > 3.0
}

// GetAnomalyScore returns the anomaly score for a value
func (ad *AnomalyDetector) GetAnomalyScore(metric string, value float64) float64 {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	baseline, exists := ad.baseline[metric]
	if !exists || !ad.initialized {
		return 0.0
	}

	zScore := math.Abs(value-baseline.Mean) / baseline.StdDev

	// Normalize to 0-1 range
	score := math.Min(zScore/3.0, 1.0)
	return score
}

// calculateMean calculates the mean of a slice of floats
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

// calculateStdDev calculates the standard deviation
func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		diff := value - mean
		sum += diff * diff
	}

	variance := sum / float64(len(values)-1)
	return math.Sqrt(variance)
}

// calculateMinMax calculates min and max values
func calculateMinMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	min := values[0]
	max := values[0]

	for _, value := range values[1:] {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}

	return min, max
}

// GetBaseline returns the baseline metrics for a metric
func (ad *AnomalyDetector) GetBaseline(metric string) (BaselineMetrics, bool) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	baseline, exists := ad.baseline[metric]
	return baseline, exists
}

// GetAllBaselines returns all baseline metrics
func (ad *AnomalyDetector) GetAllBaselines() map[string]BaselineMetrics {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	baselines := make(map[string]BaselineMetrics)
	for k, v := range ad.baseline {
		baselines[k] = v
	}

	return baselines
}
