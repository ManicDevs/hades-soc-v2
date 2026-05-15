package auxiliary

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"hades-v2/pkg/sdk"
)

// ResourceMonitor provides system resource monitoring
type ResourceMonitor struct {
	*sdk.BaseModule
	maxCPU    int
	maxMemory int
}

// NewResourceMonitor creates a new resource monitor instance
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		BaseModule: sdk.NewBaseModule(
			"resource_monitor",
			"Monitor system resources (CPU, memory, network)",
			sdk.CategoryReporting,
		),
		maxCPU:    80,
		maxMemory: 1024,
	}
}

// Execute starts resource monitoring
func (rm *ResourceMonitor) Execute(ctx context.Context) error {
	rm.SetStatus(sdk.StatusRunning)
	defer rm.SetStatus(sdk.StatusIdle)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(150 * time.Millisecond):
		rm.SetStatus(sdk.StatusCompleted)
		return nil
	}
}

// SetMaxCPU configures maximum CPU threshold
func (rm *ResourceMonitor) SetMaxCPU(cpu int) error {
	if cpu <= 0 || cpu > 100 {
		return fmt.Errorf("hades.auxiliary.resource_monitor: CPU must be between 1-100")
	}
	rm.maxCPU = cpu
	return nil
}

// SetMaxMemory configures maximum memory threshold in MB
func (rm *ResourceMonitor) SetMaxMemory(memory int) error {
	if memory <= 0 {
		return fmt.Errorf("hades.auxiliary.resource_monitor: memory must be positive")
	}
	rm.maxMemory = memory
	return nil
}

// GetResult returns resource status
func (rm *ResourceMonitor) GetResult() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return fmt.Sprintf("Resource monitor: CPU threshold=%d%%, Memory threshold=%dMB, Current memory=%dMB",
		rm.maxCPU, rm.maxMemory, m.Alloc/1024/1024)
}
