package websocket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"hades-v2/internal/bus"
	"hades-v2/internal/database"

	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections and real-time updates
type WebSocketManager struct {
	connections map[*websocket.Conn]bool
	broadcast   chan []byte
	mu          sync.RWMutex
	database    database.Database
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    *int        `json:"user_id,omitempty"`
}

// ActionRequest represents an action request with Manual ACK support
type ActionRequest struct {
	ActionName       string                 `json:"action_name"`
	Target           string                 `json:"target"`
	Reasoning        string                 `json:"reasoning"`
	RequiresApproval bool                   `json:"requires_approval"`
	Requester        string                 `json:"requester"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RealTimeUpdate represents real-time data updates
type RealTimeUpdate struct {
	Type      string      `json:"type"`
	Entity    string      `json:"entity"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(db database.Database) *WebSocketManager {
	wsm := &WebSocketManager{
		connections: make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte, 256),
		mu:          sync.RWMutex{},
		database:    db,
	}

	go wsm.subscribeToEventBus()

	return wsm
}

// subscribeToEventBus subscribes to agent events and forwards them to WebSocket clients
func (wsm *WebSocketManager) subscribeToEventBus() {
	bus.Default().Subscribe(bus.EventTypeAgentDecision, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	bus.Default().Subscribe(bus.EventTypeModuleLaunched, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	bus.Default().Subscribe(bus.EventTypeModuleCompleted, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	bus.Default().Subscribe(bus.EventTypeCriticalThreat, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	bus.Default().Subscribe(bus.EventTypeNodeIsolated, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	bus.Default().Subscribe(bus.EventTypePortDiscovered, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	bus.Default().Subscribe(bus.EventTypeVulnerabilityDetected, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	bus.Default().Subscribe(bus.EventTypeActionRequest, func(event bus.Event) error {
		// Check if this is a governor block event
		if governorBlock, ok := event.Payload["governor_block"].(bool); ok && governorBlock {
			// Send as GOVERNOR_INTERCEPT message
			wsm.broadcastGovernorIntercept(event)
		} else {
			// Send as regular event
			wsm.broadcastEvent(event)
		}
		return nil
	})

	bus.Default().Subscribe(bus.EventTypeLogEvent, func(event bus.Event) error {
		wsm.broadcastEvent(event)
		return nil
	})

	log.Println("WebSocketManager: Subscribed to agent events")
}

// broadcastGovernorIntercept sends a governor intercept message to all connected WebSocket clients
func (wsm *WebSocketManager) broadcastGovernorIntercept(event bus.Event) {
	// Create GOVERNOR_INTERCEPT message
	interceptMessage := WebSocketMessage{
		Type: "GOVERNOR_INTERCEPT",
		Data: map[string]interface{}{
			"action_name":       event.Payload["action_name"],
			"target":            event.Payload["target"],
			"reasoning":         event.Payload["reasoning"],
			"requires_approval": event.Payload["requires_approval"],
			"requester":         event.Payload["requester"],
			"timestamp":         event.Payload["timestamp"],
			"block_reason":      event.Payload["block_reason"],
			"source":            event.Source,
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(interceptMessage)
	if err != nil {
		log.Printf("Failed to marshal governor intercept message for WebSocket broadcast: %v", err)
		return
	}

	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	for conn := range wsm.connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Failed to send governor intercept to WebSocket client: %v", err)
			conn.Close()
			delete(wsm.connections, conn)
		}
	}
}

// broadcastEvent sends an event to all connected WebSocket clients
func (wsm *WebSocketManager) broadcastEvent(event bus.Event) {
	eventData := map[string]interface{}{
		"id":        event.ID,
		"type":      string(event.Type),
		"source":    event.Source,
		"target":    event.Target,
		"payload":   event.Payload,
		"timestamp": event.Timestamp,
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Failed to marshal event for WebSocket broadcast: %v", err)
		return
	}

	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	for conn := range wsm.connections {
		select {
		case wsm.broadcast <- data:
		default:
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to send event to WebSocket client: %v", err)
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	log.Printf("New WebSocket connection established from %s", r.RemoteAddr)

	// Add connection to manager
	wsm.mu.Lock()
	wsm.connections[conn] = true
	wsm.mu.Unlock()

	// Start message handler for this connection
	go wsm.handleConnection(conn)
}

// handleConnection handles messages from a WebSocket connection
func (wsm *WebSocketManager) handleConnection(conn *websocket.Conn) {
	defer func() {
		// Remove connection when done
		wsm.mu.Lock()
		delete(wsm.connections, conn)
		wsm.mu.Unlock()
		if err := conn.Close(); err != nil {
			log.Printf("Warning: failed to close connection: %v", err)
		}
	}()

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Printf("Warning: failed to set read deadline: %v", err)
	}

	for {
		// Read message from WebSocket
		var message WebSocketMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err) {
				log.Printf("WebSocket connection closed normally")
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Process message based on type
		wsm.processMessage(conn, message)
	}
}

// processMessage processes incoming WebSocket messages
func (wsm *WebSocketManager) processMessage(conn *websocket.Conn, message WebSocketMessage) {
	switch message.Type {
	case "subscribe":
		wsm.handleSubscription(conn, message)
	case "unsubscribe":
		wsm.handleUnsubscription(conn, message)
	case "ping":
		wsm.sendToConnection(conn, WebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		})
	default:
		log.Printf("Unknown WebSocket message type: %s", message.Type)
	}
}

// handleSubscription handles subscription requests
func (wsm *WebSocketManager) handleSubscription(conn *websocket.Conn, message WebSocketMessage) {
	subscriptionData, ok := message.Data.(map[string]interface{})
	if !ok {
		wsm.sendError(conn, "Invalid subscription data")
		return
	}

	entity, ok := subscriptionData["entity"].(string)
	if !ok {
		wsm.sendError(conn, "Entity is required for subscription")
		return
	}

	log.Printf("Client subscribed to: %s", entity)

	// Send initial data for the subscribed entity
	switch entity {
	case "threats":
		wsm.sendThreats(conn)
	case "users":
		wsm.sendUsers(conn)
	case "metrics":
		wsm.sendMetrics(conn)
	case "notifications":
		wsm.sendNotifications(conn)
	default:
		log.Printf("Unknown subscription entity: %s", entity)
	}
}

// handleUnsubscription handles unsubscription requests
func (wsm *WebSocketManager) handleUnsubscription(conn *websocket.Conn, message WebSocketMessage) {
	subscriptionData, ok := message.Data.(map[string]interface{})
	if !ok {
		wsm.sendError(conn, "Invalid unsubscription data")
		return
	}

	entity, ok := subscriptionData["entity"].(string)
	if !ok {
		wsm.sendError(conn, "Entity is required for unsubscription")
		return
	}

	log.Printf("Client unsubscribed from: %s", entity)
}

// sendThreats sends current threat data to WebSocket connection
func (wsm *WebSocketManager) sendThreats(conn *websocket.Conn) {
	threats, err := wsm.getRecentThreats()
	if err != nil {
		wsm.sendError(conn, "Failed to fetch threats")
		return
	}

	update := RealTimeUpdate{
		Type:      "threats_update",
		Entity:    "threats",
		Data:      threats,
		Timestamp: time.Now(),
	}

	wsm.sendToConnection(conn, WebSocketMessage{
		Type:      "update",
		Data:      update,
		Timestamp: time.Now(),
	})
}

// sendUsers sends current user data to WebSocket connection
func (wsm *WebSocketManager) sendUsers(conn *websocket.Conn) {
	users, err := wsm.getActiveUsers()
	if err != nil {
		wsm.sendError(conn, "Failed to fetch users")
		return
	}

	update := RealTimeUpdate{
		Type:      "users_update",
		Entity:    "users",
		Data:      users,
		Timestamp: time.Now(),
	}

	wsm.sendToConnection(conn, WebSocketMessage{
		Type:      "update",
		Data:      update,
		Timestamp: time.Now(),
	})
}

// sendMetrics sends current system metrics to WebSocket connection
func (wsm *WebSocketManager) sendMetrics(conn *websocket.Conn) {
	metrics, err := wsm.getSystemMetrics()
	if err != nil {
		wsm.sendError(conn, "Failed to fetch metrics")
		return
	}

	update := RealTimeUpdate{
		Type:      "metrics_update",
		Entity:    "metrics",
		Data:      metrics,
		Timestamp: time.Now(),
	}

	wsm.sendToConnection(conn, WebSocketMessage{
		Type:      "update",
		Data:      update,
		Timestamp: time.Now(),
	})
}

// sendNotifications sends current notifications to WebSocket connection
func (wsm *WebSocketManager) sendNotifications(conn *websocket.Conn) {
	notifications, err := wsm.getNotifications()
	if err != nil {
		wsm.sendError(conn, "Failed to fetch notifications")
		return
	}

	update := RealTimeUpdate{
		Type:      "notifications_update",
		Entity:    "notifications",
		Data:      notifications,
		Timestamp: time.Now(),
	}

	wsm.sendToConnection(conn, WebSocketMessage{
		Type:      "update",
		Data:      update,
		Timestamp: time.Now(),
	})
}

// BroadcastUpdate sends updates to all connected clients
func (wsm *WebSocketManager) BroadcastUpdate(updateType, entity string, data interface{}) {
	update := RealTimeUpdate{
		Type:      updateType,
		Entity:    entity,
		Data:      data,
		Timestamp: time.Now(),
	}

	message := WebSocketMessage{
		Type:      "broadcast",
		Data:      update,
		Timestamp: time.Now(),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	// Send to all connections
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	for conn := range wsm.connections {
		err := conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Printf("Failed to send message to WebSocket client: %v", err)
		}
	}
}

// BroadcastThreatUpdate sends threat updates to all connected clients
func (wsm *WebSocketManager) BroadcastThreatUpdate(threat database.Threat) {
	wsm.BroadcastUpdate("threat_created", "threats", threat)
}

// BroadcastUserUpdate sends user updates to all connected clients
func (wsm *WebSocketManager) BroadcastUserUpdate(user database.User) {
	wsm.BroadcastUpdate("user_updated", "users", user)
}

// BroadcastNotification sends notification to all connected clients
func (wsm *WebSocketManager) BroadcastNotification(notification database.Notification) {
	wsm.BroadcastUpdate("notification_created", "notifications", notification)
}

// GetConnectionCount returns the number of active connections
func (wsm *WebSocketManager) GetConnectionCount() int {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()
	return len(wsm.connections)
}

// Helper methods for data fetching
func (wsm *WebSocketManager) getRecentThreats() ([]database.Threat, error) {
	sqlDB, ok := wsm.database.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT id, title, description, severity, status, source, target, detected_at, created_at
		FROM threats 
		ORDER BY detected_at DESC 
		LIMIT 50
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch threats: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var threats []database.Threat
	for rows.Next() {
		var threat database.Threat
		err := rows.Scan(&threat.ID, &threat.Title, &threat.Description, &threat.Severity,
			&threat.Status, &threat.Source, &threat.Target, &threat.DetectedAt, &threat.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan threat: %w", err)
		}
		threats = append(threats, threat)
	}

	return threats, nil
}

func (wsm *WebSocketManager) getActiveUsers() ([]database.User, error) {
	sqlDB, ok := wsm.database.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT id, username, email, role, status, last_login, created_at
		FROM users 
		WHERE status = 'active'
		ORDER BY last_login DESC
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var users []database.User
	for rows.Next() {
		var user database.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role,
			&user.Status, &user.LastLogin, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (wsm *WebSocketManager) getSystemMetrics() (database.SystemMetrics, error) {
	sqlDB, ok := wsm.database.GetConnection().(*sql.DB)
	if !ok {
		return database.SystemMetrics{}, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT cpu_usage, memory_usage, disk_usage, network_in, network_out, 
			   active_users, total_requests, error_rate, timestamp
		FROM system_metrics 
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var metrics database.SystemMetrics
	err := sqlDB.QueryRow(query).Scan(&metrics.CPUUsage, &metrics.MemoryUsage, &metrics.DiskUsage,
		&metrics.NetworkIn, &metrics.NetworkOut, &metrics.ActiveUsers, &metrics.TotalRequests,
		&metrics.ErrorRate, &metrics.Timestamp)

	return metrics, err
}

func (wsm *WebSocketManager) getNotifications() ([]database.Notification, error) {
	sqlDB, ok := wsm.database.GetConnection().(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("database is not an SQL database")
	}

	query := `
		SELECT id, user_id, title, message, type, severity, read, created_at
		FROM notifications 
		WHERE read = false 
		ORDER BY created_at DESC 
		LIMIT 20
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: failed to close rows: %v", err)
		}
	}()

	var notifications []database.Notification
	for rows.Next() {
		var notification database.Notification
		err := rows.Scan(&notification.ID, &notification.UserID, &notification.Title,
			&notification.Message, &notification.Type, &notification.Severity,
			&notification.Read, &notification.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// Helper methods for sending messages
func (wsm *WebSocketManager) sendToConnection(conn *websocket.Conn, message WebSocketMessage) {
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Failed to send WebSocket message: %v", err)
	}
}

func (wsm *WebSocketManager) sendError(conn *websocket.Conn, errorMsg string) {
	errorMessage := WebSocketMessage{
		Type:      "error",
		Data:      errorMsg,
		Timestamp: time.Now(),
	}
	wsm.sendToConnection(conn, errorMessage)
}

// AgentStreamWebSocketManager is a specialized WebSocket manager for agent thought stream
// It subscribes to LogEvent and ActionRequest events and broadcasts them to connected clients
type AgentStreamWebSocketManager struct {
	*WebSocketManager
}

// NewAgentStreamWebSocketManager creates a new WebSocket manager for agent thought stream
func NewAgentStreamWebSocketManager(db database.Database) *AgentStreamWebSocketManager {
	wsm := NewWebSocketManager(db)
	asm := &AgentStreamWebSocketManager{WebSocketManager: wsm}
	asm.subscribeToAgentStreamEvents()
	return asm
}

// subscribeToAgentStreamEvents subscribes to LogEvent and ActionRequest events
func (asm *AgentStreamWebSocketManager) subscribeToAgentStreamEvents() {
	// Subscribe to LogEvent (Thoughts)
	bus.Default().Subscribe(bus.EventTypeLogEvent, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "thought")
		return nil
	})

	// Subscribe to ActionRequest (Actions)
	bus.Default().Subscribe(bus.EventTypeActionRequest, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "action")
		return nil
	})

	// Subscribe to SecurityUpgradeRequest (Quantum Shield)
	bus.Default().Subscribe(bus.EventTypeSecurityUpgradeRequest, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "quantum")
		return nil
	})

	// Subscribe to NewAssetEvent (Recon)
	bus.Default().Subscribe(bus.EventTypeNewAsset, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "recon")
		return nil
	})

	// Subscribe to Threat (Critical)
	bus.Default().Subscribe(bus.EventTypeThreat, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "critical")
		return nil
	})

	// Subscribe to VulnerabilityFound (Remediation)
	bus.Default().Subscribe(bus.EventTypeVulnerabilityFound, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "remediation")
		return nil
	})

	// Subscribe to ModuleHotSwap (Hot-swap)
	bus.Default().Subscribe(bus.EventTypeModuleHotSwap, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "hot_swap")
		return nil
	})

	// Subscribe to LateralMovement (Network Containment)
	bus.Default().Subscribe(bus.EventTypeLateralMovement, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "network_containment")
		return nil
	})

	// Subscribe to HoneyTokenTriggered (Identity Deception)
	bus.Default().Subscribe(bus.EventTypeHoneyTokenTriggered, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "honey_token_trap")
		return nil
	})

	// Subscribe to HoneyFileAccessed (File-Based Deception)
	bus.Default().Subscribe(bus.EventTypeHoneyFileAccessed, func(event bus.Event) error {
		asm.broadcastAgentStreamEvent(event, "honey_file_accessed")
		return nil
	})

	// Subscribe to ActionRequest for honey file rotation
	bus.Default().Subscribe(bus.EventTypeActionRequest, func(event bus.Event) error {
		if actionName, ok := event.Payload["action_name"].(string); ok && actionName == "RotateHoneyTraps" {
			asm.broadcastAgentStreamEvent(event, "honey_file_rotation")
		}
		return nil
	})

	log.Println("AgentStreamWebSocketManager: Subscribed to agent stream events (LogEvent, ActionRequest, SecurityUpgradeRequest, NewAsset, Threat, VulnerabilityFound, ModuleHotSwap, LateralMovement, HoneyTokenTriggered, HoneyFileAccessed, TrapRotation)")
}

// broadcastAgentStreamEvent sends an agent stream event to all connected clients with enhanced metadata
func (asm *AgentStreamWebSocketManager) broadcastAgentStreamEvent(event bus.Event, eventCategory string) {
	// Extract internal reasoning from payload if available
	var internalReasoning string
	var agentName string
	var severity string
	var message string

	if payload, ok := event.Payload["data"]; ok {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			if ir, ok := payloadMap["internal_reasoning"].(string); ok {
				internalReasoning = ir
			}
			if an, ok := payloadMap["agent_name"].(string); ok {
				agentName = an
			}
			if msg, ok := payloadMap["message"].(string); ok {
				message = msg
			}
		}
	}

	// Fallback to direct payload fields
	if internalReasoning == "" {
		if ir, ok := event.Payload["internal_reasoning"].(string); ok {
			internalReasoning = ir
		}
	}
	if agentName == "" {
		if an, ok := event.Payload["agent_name"].(string); ok {
			agentName = an
		}
	}
	if message == "" {
		if msg, ok := event.Payload["message"].(string); ok {
			message = msg
		}
	}
	if severity == "" {
		if sev, ok := event.Payload["severity"].(string); ok {
			severity = sev
		} else {
			severity = eventCategory
		}
	}

	// Build enriched event data
	eventData := map[string]interface{}{
		"id":                 event.ID,
		"type":               string(event.Type),
		"category":           eventCategory,
		"source":             event.Source,
		"target":             event.Target,
		"timestamp":          event.Timestamp,
		"agent_name":         agentName,
		"message":            message,
		"internal_reasoning": internalReasoning,
		"severity":           severity,
		"payload":            event.Payload,
	}

	data, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Failed to marshal agent stream event: %v", err)
		return
	}

	asm.mu.Lock()
	defer asm.mu.Unlock()

	for conn := range asm.connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Failed to send agent stream event to client: %v", err)
		}
	}
}
