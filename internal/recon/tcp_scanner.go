package recon

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/pkg/sdk"
)

// TCPScanner performs TCP port scanning using connect() method
type TCPScanner struct {
	*sdk.BaseModule
	target    string
	ports     []int
	timeout   time.Duration
	threads   int
	openPorts []int
	mu        sync.Mutex
}

// NewTCPScanner creates a new TCP port scanner instance
func NewTCPScanner() *TCPScanner {
	return &TCPScanner{
		BaseModule: sdk.NewBaseModule(
			"tcp_scanner",
			"TCP port scanner using connect() method",
			sdk.CategoryScanning,
		),
		timeout: 5 * time.Second,
		threads: 50,
	}
}

// Execute runs the TCP port scanner
func (ts *TCPScanner) Execute(ctx context.Context) error {
	ts.SetStatus(sdk.StatusRunning)
	defer ts.SetStatus(sdk.StatusIdle)

	if err := ts.validateConfig(); err != nil {
		return fmt.Errorf("hades.recon.tcp_scanner: %w", err)
	}

	ts.openPorts = make([]int, 0)

	semaphore := make(chan struct{}, ts.threads)
	var wg sync.WaitGroup

	for _, port := range ts.ports {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				defer func() { <-semaphore }()

				if ts.scanPort(ctx, p) {
					ts.mu.Lock()
					ts.openPorts = append(ts.openPorts, p)
					ts.mu.Unlock()
				}
			}(port)
		}
	}

	wg.Wait()

	bus.Default().Publish(bus.Event{
		Type:   bus.EventTypeReconComplete,
		Source: "tcp_scanner",
		Target: ts.target,
		Payload: map[string]interface{}{
			"open_ports":  ts.GetOpenPorts(),
			"total_ports": len(ts.ports),
			"scanned_at":  time.Now().Unix(),
		},
	})

	ts.SetStatus(sdk.StatusCompleted)
	return nil
}

// SetTarget configures the scan target
func (ts *TCPScanner) SetTarget(target string) error {
	if target == "" {
		return fmt.Errorf("hades.recon.tcp_scanner: target cannot be empty")
	}
	ts.target = target
	return nil
}

// SetPorts configures the ports to scan
func (ts *TCPScanner) SetPorts(ports []int) error {
	if len(ports) == 0 {
		return fmt.Errorf("hades.recon.tcp_scanner: ports cannot be empty")
	}
	ts.ports = ports
	return nil
}

// SetPortsFromRange parses port range string (e.g., "1-1000,8080,9000-9100")
func (ts *TCPScanner) SetPortsFromRange(portRange string) error {
	if portRange == "" {
		return fmt.Errorf("hades.recon.tcp_scanner: port range cannot be empty")
	}

	var ports []int
	ranges := strings.Split(portRange, ",")

	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				return fmt.Errorf("hades.recon.tcp_scanner: invalid port range format: %s", r)
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return fmt.Errorf("hades.recon.tcp_scanner: invalid start port: %v", err)
			}

			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return fmt.Errorf("hades.recon.tcp_scanner: invalid end port: %v", err)
			}

			if start < 1 || end > 65535 || start > end {
				return fmt.Errorf("hades.recon.tcp_scanner: invalid port range: %d-%d", start, end)
			}

			for p := start; p <= end; p++ {
				ports = append(ports, p)
			}
		} else {
			port, err := strconv.Atoi(r)
			if err != nil {
				return fmt.Errorf("hades.recon.tcp_scanner: invalid port: %v", err)
			}
			if port < 1 || port > 65535 {
				return fmt.Errorf("hades.recon.tcp_scanner: port out of range: %d", port)
			}
			ports = append(ports, port)
		}
	}

	ts.ports = ports
	return nil
}

// SetTimeout configures connection timeout
func (ts *TCPScanner) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("hades.recon.tcp_scanner: timeout must be positive")
	}
	ts.timeout = timeout
	return nil
}

// SetThreads configures concurrent thread count
func (ts *TCPScanner) SetThreads(threads int) error {
	if threads <= 0 {
		return fmt.Errorf("hades.recon.tcp_scanner: threads must be positive")
	}
	ts.threads = threads
	return nil
}

// GetOpenPorts returns the list of discovered open ports
func (ts *TCPScanner) GetOpenPorts() []int {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	result := make([]int, len(ts.openPorts))
	copy(result, ts.openPorts)
	return result
}

// GetResult returns scan results as formatted string
func (ts *TCPScanner) GetResult() string {
	openPorts := ts.GetOpenPorts()
	if len(openPorts) == 0 {
		return fmt.Sprintf("No open ports found on %s", ts.target)
	}

	return fmt.Sprintf("Open ports on %s: %v", ts.target, openPorts)
}

// validateConfig ensures scanner configuration is valid
func (ts *TCPScanner) validateConfig() error {
	if ts.target == "" {
		return fmt.Errorf("hades.recon.tcp_scanner: target not configured")
	}
	if len(ts.ports) == 0 {
		return fmt.Errorf("hades.recon.tcp_scanner: ports not configured")
	}
	return nil
}

// scanPort checks if a single port is open
func (ts *TCPScanner) scanPort(ctx context.Context, port int) bool {
	address := fmt.Sprintf("%s:%d", ts.target, port)

	dialer := &net.Dialer{
		Timeout: ts.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
