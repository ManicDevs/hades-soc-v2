package auxiliary

import (
	"context"
	"fmt"
	"time"

	"hades-v2/pkg/sdk"
)

// TrendAnalyzer provides trend analysis functionality
type TrendAnalyzer struct {
	*sdk.BaseModule
	dataPoints int
}

// NewTrendAnalyzer creates a new trend analyzer instance
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{
		BaseModule: sdk.NewBaseModule(
			"trend_analyzer",
			"Analyze trends in metrics and data",
			sdk.CategoryReporting,
		),
		dataPoints: 100,
	}
}

// Execute starts trend analysis
func (ta *TrendAnalyzer) Execute(ctx context.Context) error {
	ta.SetStatus(sdk.StatusRunning)
	defer ta.SetStatus(sdk.StatusIdle)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(220 * time.Millisecond):
		ta.SetStatus(sdk.StatusCompleted)
		return nil
	}
}

// SetDataPoints configures the number of data points
func (ta *TrendAnalyzer) SetDataPoints(points int) error {
	if points <= 0 {
		return fmt.Errorf("hades.auxiliary.trend_analyzer: data points must be positive")
	}
	ta.dataPoints = points
	return nil
}

// GetResult returns analysis status
func (ta *TrendAnalyzer) GetResult() string {
	return fmt.Sprintf("Trend analyzer initialized with %d data points", ta.dataPoints)
}
