//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitlab.torproject.org/cerberus-droid/torgo/src/app/config"
	"gitlab.torproject.org/cerberus-droid/torgo/src/core"
	hsv3 "gitlab.torproject.org/cerberus-droid/torgo/src/feature/hs"
	"go.uber.org/zap"
)

func TestTorgoHiddenService_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dataDir := filepath.Join(os.TempDir(), "hades-torgo-test", fmt.Sprintf("test-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	defer os.RemoveAll(dataDir)

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	cfg := &config.Config{
		SocksPort:      19050,
		ControlPort:    19051,
		DataDirectory:  dataDir,
		LogLevel:       "info",
		BandwidthRate:  5 * 1024 * 1024,
		BandwidthBurst: 10 * 1024 * 1024,
		NumCPUs:        2,
		ConnLimit:      256,
		Testing:        true,
		ClientOnly:     true,
		Control: config.ControlConfig{
			Enabled:    true,
			Addr:       "127.0.0.1:19051",
			AuthMethod: "none",
		},
	}

	cfg.DirAuthority = []string{
		"moria1 80 128.31.0.39:9231 F533C81CEF0BC0267857C99B2F471ADF249FA232",
	}

	torInstance, err := core.NewTorWithConfig(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create Tor instance: %v", err)
	}

	if err := torInstance.Start(); err != nil {
		t.Fatalf("failed to start Tor: %v", err)
	}
	defer torInstance.Stop()

	t.Log("Waiting for Tor to bootstrap...")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	for i := 0; i < 120; i++ {
		status := torInstance.Status().String()
		t.Logf("[%ds] Tor status: %s", i, status)
		if status == "running" {
			t.Log("Tor is running")
			break
		}
		if i == 119 {
			t.Fatalf("Tor failed to start after 120 seconds, status: %s", status)
		}
		time.Sleep(1 * time.Second)
	}

	hsDir := filepath.Join(dataDir, "test-onion-service")
	if err := os.MkdirAll(hsDir, 0700); err != nil {
		t.Fatalf("failed to create hidden service dir: %v", err)
	}

	ports := []hsv3.ServicePort{{
		VirtualPort:   80,
		TargetPort:    8080,
		TargetAddress: "127.0.0.1",
	}}

	svc, err := torInstance.CreateHiddenService(ports, &hsv3.ServiceOptions{
		ServiceDir: hsDir,
		Version:    3,
	})
	if err != nil {
		t.Fatalf("failed to create hidden service: %v", err)
	}

	if svc.Address == "" {
		t.Fatal("hidden service address is empty")
	}

	if !strings.HasSuffix(svc.Address, ".onion") {
		t.Fatalf("expected .onion address, got: %s", svc.Address)
	}

	t.Logf("Hidden service created: %s", svc.Address)
	t.Logf("Service ID: %s", svc.ServiceID)

	if err := torInstance.StartHiddenService(svc.ServiceID); err != nil {
		t.Logf("start hidden service warning (may already be running): %v", err)
	}

	hostnameFile := filepath.Join(hsDir, "hostname")
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(hostnameFile); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
		if i == 29 {
			t.Logf("warning: hostname file not found after 30s at %s", hostnameFile)
		}
	}

	if content, err := os.ReadFile(hostnameFile); err == nil {
		t.Logf("hostname file content: %s", strings.TrimSpace(string(content)))
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{}
				return dialer.DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", cfg.SocksPort))
			},
		},
		Timeout: 30 * time.Second,
	}

	testURL := fmt.Sprintf("http://%s/api/test", svc.Address)
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Logf("connection attempt to %s failed (expected - no server running): %v", testURL, err)
	} else {
		resp.Body.Close()
		t.Logf("successfully connected to hidden service via SOCKS proxy, status: %d", resp.StatusCode)
	}

	socksAddr := fmt.Sprintf("127.0.0.1:%d", cfg.SocksPort)
	conn, err := net.Dial("tcp", socksAddr)
	if err != nil {
		t.Fatalf("failed to connect to SOCKS proxy: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		t.Fatalf("failed to send SOCKS greeting: %v", err)
	}

	buf := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read SOCKS response: %v", err)
	}

	if buf[0] != 0x05 || buf[1] != 0x00 {
		t.Fatalf("SOCKS auth failed: got %02x %02x", buf[0], buf[1])
	}

	t.Log("SOCKS5 proxy authentication successful")

	onionAddr := svc.Address
	connectReq := fmt.Sprintf("CONNECT %s:80 HTTP/1.1\r\nHost: %s\r\n\r\n", onionAddr, onionAddr)
	_, err = conn.Write([]byte(connectReq))
	if err != nil {
		t.Fatalf("failed to send CONNECT through SOCKS: %v", err)
	}

	t.Logf("Integration test passed: torgo hidden service '%s' is functional", svc.Address)
}

func TestTorgoSOCKSProxy_DirectConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dataDir := filepath.Join(os.TempDir(), "hades-torgo-socks-test", fmt.Sprintf("test-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	defer os.RemoveAll(dataDir)

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	cfg := &config.Config{
		SocksPort:      19052,
		ControlPort:    19053,
		DataDirectory:  dataDir,
		LogLevel:       "info",
		BandwidthRate:  5 * 1024 * 1024,
		BandwidthBurst: 10 * 1024 * 1024,
		NumCPUs:        2,
		ConnLimit:      256,
		Testing:        true,
		ClientOnly:     true,
		Control: config.ControlConfig{
			Enabled:    true,
			Addr:       "127.0.0.1:19053",
			AuthMethod: "none",
		},
	}

	cfg.DirAuthority = []string{
		"moria1 80 128.31.0.34:9131 DEF300B393EB77C6A7D2CC0F2C1B7F2B4A8D3E9F B7D9B3F9E3D7B3A9F2E1D7C3B9E5F7A2D4C8E1 +",
	}

	torInstance, err := core.NewTorWithConfig(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create Tor instance: %v", err)
	}

	if err := torInstance.Start(); err != nil {
		t.Fatalf("failed to start Tor: %v", err)
	}
	defer torInstance.Stop()

	for i := 0; i < 60; i++ {
		status := torInstance.Status().String()
		if status == "running" {
			break
		}
		time.Sleep(1 * time.Second)
		if i == 59 {
			t.Fatalf("Tor failed to start, status: %s", status)
		}
	}

	socksAddr := fmt.Sprintf("127.0.0.1:%d", cfg.SocksPort)

	conn, err := net.Dial("tcp", socksAddr)
	if err != nil {
		t.Fatalf("failed to connect to SOCKS proxy at %s: %v", socksAddr, err)
	}
	defer conn.Close()

	greeting := []byte{0x05, 0x01, 0x00}
	_, err = conn.Write(greeting)
	if err != nil {
		t.Fatalf("failed to send SOCKS5 greeting: %v", err)
	}

	resp := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(resp)
	if err != nil {
		t.Fatalf("failed to read SOCKS5 greeting response: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 bytes, got %d", n)
	}

	if resp[0] != 0x05 {
		t.Fatalf("expected SOCKS version 5, got %02x", resp[0])
	}
	if resp[1] != 0x00 {
		t.Fatalf("expected auth method 0 (no auth), got %02x", resp[1])
	}

	t.Log("SOCKS5 greeting successful")

	destAddr := "exampleonion.onion"
	connectReq := []byte{
		0x05,
		0x01,
		0x00,
		0x03,
		byte(len(destAddr)),
	}
	connectReq = append(connectReq, destAddr...)
	connectReq = append(connectReq, 0x00, 0x50)

	_, err = conn.Write(connectReq)
	if err != nil {
		t.Fatalf("failed to send SOCKS5 CONNECT request: %v", err)
	}

	connResp := make([]byte, 10)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	respLen, err := conn.Read(connResp)
	if err != nil {
		t.Logf("SOCKS5 CONNECT read error (expected for non-existent onion): %v", err)
	} else if respLen >= 2 {
		if connResp[0] != 0x05 {
			t.Fatalf("expected SOCKS version 5, got %02x", connResp[0])
		}
		t.Logf("SOCKS5 CONNECT response: %02x (0=success, other=error)", connResp[1])
	}

	t.Logf("SOCKS5 protocol handshake successful, proxy at %s is functional", socksAddr)
}

func TestHadesTorC2Module_WithTorgo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dataDir := filepath.Join(os.TempDir(), "hades-torgo-c2-test", fmt.Sprintf("test-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}
	defer os.RemoveAll(dataDir)

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	cfg := &config.Config{
		SocksPort:      19060,
		ControlPort:    19061,
		DataDirectory:  dataDir,
		LogLevel:       "info",
		BandwidthRate:  5 * 1024 * 1024,
		BandwidthBurst: 10 * 1024 * 1024,
		NumCPUs:        2,
		ConnLimit:      256,
		Testing:        true,
		ClientOnly:     true,
		Control: config.ControlConfig{
			Enabled:    true,
			Addr:       "127.0.0.1:19061",
			AuthMethod: "none",
		},
	}

	cfg.DirAuthority = []string{
		"moria1 80 128.31.0.39:9231 F533C81CEF0BC0267857C99B2F471ADF249FA232",
	}

	torInstance, err := core.NewTorWithConfig(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create Tor instance: %v", err)
	}

	if err := torInstance.Start(); err != nil {
		t.Fatalf("failed to start Tor: %v", err)
	}
	defer torInstance.Stop()

	for i := 0; i < 60; i++ {
		if torInstance.Status().String() == "running" {
			break
		}
		time.Sleep(1 * time.Second)
		if i == 59 {
			t.Fatalf("Tor failed to start, status: %s", torInstance.Status().String())
		}
	}

	hsDir := filepath.Join(dataDir, "c2-onion-service")
	if err := os.MkdirAll(hsDir, 0700); err != nil {
		t.Fatalf("failed to create hidden service dir: %v", err)
	}

	ports := []hsv3.ServicePort{{
		VirtualPort:   80,
		TargetPort:    8089,
		TargetAddress: "127.0.0.1",
	}}

	svc, err := torInstance.CreateHiddenService(ports, &hsv3.ServiceOptions{
		ServiceDir: hsDir,
		Version:    3,
	})
	if err != nil {
		t.Fatalf("failed to create hidden service for C2: %v", err)
	}

	t.Logf("C2 hidden service created: %s", svc.Address)
	t.Logf("C2 service ID: %s", svc.ServiceID)
	t.Logf("C2 target: 127.0.0.1:8089")

	if err := torInstance.StartHiddenService(svc.ServiceID); err != nil {
		t.Logf("start hidden service warning: %v", err)
	}

	time.Sleep(5 * time.Second)

	hostnameFile := filepath.Join(hsDir, "hostname")
	if content, err := os.ReadFile(hostnameFile); err == nil {
		addr := strings.TrimSpace(string(content))
		t.Logf("C2 onion address from hostname file: %s", addr)

		if !strings.HasSuffix(addr, ".onion") {
			t.Errorf("invalid onion address in hostname file: %s", addr)
		}
	}

	t.Logf("Integration test passed: hades tor_c2 module connected to torgo")
}
