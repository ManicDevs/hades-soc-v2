package auxiliary

import (
	"context"
	"fmt"
	"time"

	"hades-v2/pkg/sdk"
)

// Dashboard provides monitoring dashboard functionality
type Dashboard struct {
	*sdk.BaseModule
	refreshInterval time.Duration
}

// NewDashboard creates a new dashboard instance
func NewDashboard() *Dashboard {
	return &Dashboard{
		BaseModule: sdk.NewBaseModule(
			"dashboard",
			"Monitoring dashboard for security operations",
			sdk.CategoryReporting,
		),
		refreshInterval: 30 * time.Second,
	}
}

// Execute starts the dashboard
func (d *Dashboard) Execute(ctx context.Context) error {
	d.SetStatus(sdk.StatusRunning)
	defer d.SetStatus(sdk.StatusIdle)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(200 * time.Millisecond):
		d.SetStatus(sdk.StatusCompleted)
		return nil
	}
}

// SetRefreshInterval configures the refresh interval
func (d *Dashboard) SetRefreshInterval(interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("hades.auxiliary.dashboard: interval must be positive")
	}
	d.refreshInterval = interval
	return nil
}

// GetResult returns dashboard status
func (d *Dashboard) GetResult() string {
	return fmt.Sprintf("Dashboard initialized with refresh interval: %v", d.refreshInterval)
}
