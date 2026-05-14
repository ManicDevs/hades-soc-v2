package auxiliary

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"hades-v2/pkg/sdk"

	"gitlab.torproject.org/cerberus-droid/torgo/src/app/config"
	"gitlab.torproject.org/cerberus-droid/torgo/src/core"
	hsv3 "gitlab.torproject.org/cerberus-droid/torgo/src/feature/hs"
	"go.uber.org/zap"
)

type TorManager struct {
	*sdk.BaseModule

	mu           sync.RWMutex
	torInstance  *core.Tor
	socksPort    int
	controlPort  int
	dataDir      string
	httpClient   *http.Client
	running      bool
	onionAddress string
	logger       *zap.Logger
}

func NewTorManager() *TorManager {
	return &TorManager{
		BaseModule: sdk.NewBaseModule(
			"tor_manager",
			"Tor network integration via torgo for anonymous operations",
			sdk.CategoryAuxiliary,
		),
		socksPort:   19050,
		controlPort: 19051,
		dataDir:     "/tmp/hades-torgo-data",
	}
}

func (tm *TorManager) Execute(ctx context.Context) error {
	tm.SetStatus(sdk.StatusRunning)
	defer tm.SetStatus(sdk.StatusIdle)

	if err := tm.startTorgo(ctx); err != nil {
		return fmt.Errorf("hades.tor_manager: failed to start torgo: %w", err)
	}

	<-ctx.Done()

	tm.stopTorgo()

	return nil
}

func (tm *TorManager) startTorgo(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if err := tm.ensureDataDir(); err != nil {
		return err
	}

	cfg := tm.createConfig()
	torgoCfg, err := config.LoadFromFile("/tmp/torgo/hades-torrc")
	if err != nil {
		torgoCfg = cfg
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("hades.tor_manager: failed to create logger: %w", err)
	}
	tm.logger = logger

	tor, err := core.NewTorWithConfig(torgoCfg, logger)
	if err != nil {
		return fmt.Errorf("hades.tor_manager: failed to create Tor instance: %w", err)
	}

	if err := tor.Start(); err != nil {
		return fmt.Errorf("hades.tor_manager: failed to start Tor: %w", err)
	}

	tm.torInstance = tor
	tm.running = true

	time.Sleep(3 * time.Second)

	tm.httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{}
				return dialer.DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", tm.socksPort))
			},
		},
	}

	fmt.Printf("tor_manager: started torgo on SOCKS %d, Control %d\n", tm.socksPort, tm.controlPort)

	return nil
}

func (tm *TorManager) createConfig() *config.Config {
	return &config.Config{
		SocksPort:      tm.socksPort,
		ControlPort:    tm.controlPort,
		DataDirectory:  tm.dataDir,
		LogLevel:       "info",
		LogFile:        "",
		BandwidthRate:  5 * 1024 * 1024,
		BandwidthBurst: 10 * 1024 * 1024,
		NumCPUs:        1,
		ConnLimit:      512,
		Control: config.ControlConfig{
			Enabled:    true,
			Addr:       fmt.Sprintf("127.0.0.1:%d", tm.controlPort),
			AuthMethod: "none",
		},
	}
}

func (tm *TorManager) ensureDataDir() error {
	if _, err := tm.ensureDir(tm.dataDir); err != nil {
		return fmt.Errorf("hades.tor_manager: failed to create data dir: %w", err)
	}
	return nil
}

func (tm *TorManager) ensureDir(path string) (string, error) {
	return path, nil
}

func (tm *TorManager) stopTorgo() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.torInstance != nil {
		tm.torInstance.Stop()
		tm.torInstance = nil
		tm.running = false
		fmt.Println("tor_manager: stopped torgo")
	}

	if tm.logger != nil {
		tm.logger.Sync()
	}
}

func (tm *TorManager) GetSocksAddr() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return fmt.Sprintf("127.0.0.1:%d", tm.socksPort)
}

func (tm *TorManager) GetControlAddr() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return fmt.Sprintf("127.0.0.1:%d", tm.controlPort)
}

func (tm *TorManager) GetHTTPClient() *http.Client {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.httpClient
}

func (tm *TorManager) IsRunning() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.running
}

func (tm *TorManager) CreateOnionService(ctx context.Context, port int) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running || tm.torInstance == nil {
		return "", fmt.Errorf("hades.tor_manager: tor not running")
	}

	// Use the hidden service manager from Tor instance
	hsMgr := tm.torInstance.GetHiddenServiceManager()
	if hsMgr == nil {
		return "", fmt.Errorf("hades.tor_manager: hidden service manager not available")
	}

	serviceDir := tm.dataDir + "/hidden_services/hades-onion"
	if err := os.MkdirAll(serviceDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create service dir: %w", err)
	}

	hsPort := hsv3.ServicePort{
		VirtualPort:   80,
		TargetPort:    uint16(port),
		TargetAddress: "127.0.0.1",
	}

	hsService, err := hsMgr.CreateService("", []hsv3.ServicePort{hsPort}, &hsv3.ServiceOptions{
		ServiceDir: serviceDir,
		Version:    3,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create hidden service: %w", err)
	}

	// Start the service
	if err := hsMgr.StartService(hsService.ServiceID); err != nil {
		tm.logger.Warn("failed to start hidden service immediately", zap.Error(err))
	}

	tm.onionAddress = hsService.Address
	fmt.Printf("tor_manager: created hidden service %s -> localhost:%d\n", hsService.Address, port)

	return hsService.Address, nil
}

func (tm *TorManager) GetOnionAddress() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.onionAddress
}

func (tm *TorManager) NewRequestWithTor(ctx context.Context, url string) (*http.Request, error) {
	client := tm.GetHTTPClient()
	if client == nil {
		return nil, fmt.Errorf("hades.tor_manager: tor client not initialized")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func generateOnionKey() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz234567"[i%32]
	}
	return string(b)
}
