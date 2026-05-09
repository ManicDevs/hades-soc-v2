package analytics

import (
	"testing"
	"time"
)

func TestAnalyticsEngineCreation(t *testing.T) {
	engine, err := NewAnalyticsEngine(nil)
	if err != nil {
		t.Fatalf("NewAnalyticsEngine returned error: %v", err)
	}
	if engine == nil {
		t.Fatal("NewAnalyticsEngine returned nil")
	}
}

func TestMLModelCreation(t *testing.T) {
	model := MLModel{
		ModelPath:   "/models/test.pb",
		Version:     "1.0",
		Accuracy:    0.95,
		IsLoaded:    true,
		LastTrained: time.Now(),
		InputSize:   64,
		OutputSize:  10,
	}

	if model.ModelPath != "/models/test.pb" {
		t.Errorf("Expected ModelPath '/models/test.pb', got '%s'", model.ModelPath)
	}
	if !model.IsLoaded {
		t.Error("IsLoaded should be true")
	}
	if model.Accuracy != 0.95 {
		t.Errorf("Expected Accuracy 0.95, got %f", model.Accuracy)
	}
}

func TestPredictionModel(t *testing.T) {
	model := PredictionModel{
		Name:        "threat_predictor",
		Type:        "classification",
		Accuracy:    0.87,
		Features:    []string{"ip", "port", "protocol"},
		LastUpdated: time.Now(),
		IsEnabled:   true,
	}

	if model.Name != "threat_predictor" {
		t.Errorf("Expected Name 'threat_predictor', got '%s'", model.Name)
	}
	if len(model.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(model.Features))
	}
	if !model.IsEnabled {
		t.Error("IsEnabled should be true")
	}
}

func TestDataAggregator(t *testing.T) {
	agg := DataAggregator{
		WindowSizes: []time.Duration{
			1 * time.Minute,
			5 * time.Minute,
			1 * time.Hour,
		},
		Metrics:      make(map[string]*MetricDefinition),
		Aggregations: make(map[string]*AggregationResult),
	}

	if len(agg.WindowSizes) != 3 {
		t.Errorf("Expected 3 window sizes, got %d", len(agg.WindowSizes))
	}
}