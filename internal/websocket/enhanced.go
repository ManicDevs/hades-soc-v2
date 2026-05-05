package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// EnhancedWebSocketManager manages enhanced WebSocket connections
type EnhancedWebSocketManager struct {
	clients        map[string]*EnhancedClient
	rooms          map[string]*Room
	broadcastChan  chan BroadcastMessage
	registerChan   chan *EnhancedClient
	unregisterChan chan *EnhancedClient
	mu             sync.RWMutex
	upgrader       websocket.Upgrader
	messageQueue   *MessageQueue
	authenticator  *WebSocketAuthenticator
}

// EnhancedClient represents an enhanced WebSocket client
type EnhancedClient struct {
	ID            string
	Connection    *websocket.Conn
	SendChan      chan []byte
	Authenticated bool
	UserID        string
	Roles         []string
	Permissions   []string
	LastActivity  time.Time
	Rooms         map[string]bool
	Metadata      map[string]interface{}
	mu            sync.RWMutex
}

// Room represents a WebSocket room for group messaging
type Room struct {
	ID       string
	Name     string
	Clients  map[string]*EnhancedClient
	Created  time.Time
	Metadata map[string]interface{}
	mu       sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	Type      string                 `json:"type"`
	Room      string                 `json:"room,omitempty"`
	Target    string                 `json:"target,omitempty"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Sender    string                 `json:"sender"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageQueue handles message queuing and delivery
type MessageQueue struct {
	messages   chan QueuedMessage
	workers    int
	maxSize    int
	mu         sync.RWMutex
	processing bool
}

// QueuedMessage represents a queued message
type QueuedMessage struct {
	Message   BroadcastMessage
	Timestamp time.Time
	Retries   int
	Target    string
}

// WebSocketAuthenticator handles WebSocket authentication
type WebSocketAuthenticator struct {
	tokens map[string]*AuthSession
	mu     sync.RWMutex
}

// AuthSession represents an authenticated session
type AuthSession struct {
	Token       string
	UserID      string
	Roles       []string
	Permissions []string
	Created     time.Time
	Expires     time.Time
}

// NewEnhancedWebSocketManager creates a new enhanced WebSocket manager
func NewEnhancedWebSocketManager() *EnhancedWebSocketManager {
	wsm := &EnhancedWebSocketManager{
		clients:        make(map[string]*EnhancedClient),
		rooms:          make(map[string]*Room),
		broadcastChan:  make(chan BroadcastMessage, 1000),
		registerChan:   make(chan *EnhancedClient),
		unregisterChan: make(chan *EnhancedClient),
		messageQueue:   NewMessageQueue(10, 10000),
		authenticator:  NewWebSocketAuthenticator(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	return wsm
}

// Start starts the enhanced WebSocket manager
func (wsm *EnhancedWebSocketManager) Start(ctx context.Context) {
	go wsm.runBroadcastLoop(ctx)
	go wsm.runClientManager(ctx)
	go wsm.messageQueue.Start(ctx)
}

// runBroadcastLoop runs the broadcast loop
func (wsm *EnhancedWebSocketManager) runBroadcastLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-wsm.broadcastChan:
			wsm.handleBroadcast(message)
		}
	}
}

// runClientManager runs the client manager
func (wsm *EnhancedWebSocketManager) runClientManager(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-wsm.registerChan:
			wsm.registerClient(client)
		case client := <-wsm.unregisterChan:
			wsm.unregisterClient(client)
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (wsm *EnhancedWebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &EnhancedClient{
		ID:           fmt.Sprintf("client_%d", time.Now().UnixNano()),
		Connection:   conn,
		SendChan:     make(chan []byte, 256),
		LastActivity: time.Now(),
		Rooms:        make(map[string]bool),
		Metadata:     make(map[string]interface{}),
	}

	// Register client
	wsm.registerChan <- client

	// Start client goroutines
	go wsm.writePump(client)
	go wsm.readPump(client)
}

// registerClient registers a new client
func (wsm *EnhancedWebSocketManager) registerClient(client *EnhancedClient) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	wsm.clients[client.ID] = client
	log.Printf("Client %s connected", client.ID)

	// Send welcome message
	welcome := BroadcastMessage{
		Type:      "welcome",
		Data:      map[string]string{"client_id": client.ID},
		Timestamp: time.Now(),
		Sender:    "system",
	}

	wsm.sendToClient(client, welcome)
}

// unregisterClient unregisters a client
func (wsm *EnhancedWebSocketManager) unregisterClient(client *EnhancedClient) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	if _, exists := wsm.clients[client.ID]; exists {
		delete(wsm.clients, client.ID)
		close(client.SendChan)
		client.Connection.Close()

		// Remove from all rooms
		for roomID := range client.Rooms {
			if room, exists := wsm.rooms[roomID]; exists {
				room.mu.Lock()
				delete(room.Clients, client.ID)
				room.mu.Unlock()
			}
		}

		log.Printf("Client %s disconnected", client.ID)
	}
}

// writePump handles writing messages to client
func (wsm *EnhancedWebSocketManager) writePump(client *EnhancedClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-client.SendChan:
			client.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error for client %s: %v", client.ID, err)
				return
			}

		case <-ticker.C:
			client.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump handles reading messages from client
func (wsm *EnhancedWebSocketManager) readPump(client *EnhancedClient) {
	defer func() {
		wsm.unregisterChan <- client
	}()

	client.Connection.SetReadLimit(512)
	client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Connection.SetPongHandler(func(string) error {
		client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := client.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error for client %s: %v", client.ID, err)
			}
			break
		}

		client.LastActivity = time.Now()
		wsm.handleClientMessage(client, message)
	}
}

// handleClientMessage handles messages from clients
func (wsm *EnhancedWebSocketManager) handleClientMessage(client *EnhancedClient, message []byte) {
	var msg struct {
		Type     string                 `json:"type"`
		Room     string                 `json:"room,omitempty"`
		Data     interface{}            `json:"data"`
		Target   string                 `json:"target,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message from client %s: %v", client.ID, err)
		return
	}

	switch msg.Type {
	case "authenticate":
		wsm.handleAuthentication(client, msg.Data)
	case "join_room":
		wsm.handleJoinRoom(client, msg.Room)
	case "leave_room":
		wsm.handleLeaveRoom(client, msg.Room)
	case "message":
		wsm.handleUserMessage(client, msg)
	case "ping":
		wsm.handlePing(client)
	default:
		log.Printf("Unknown message type from client %s: %s", client.ID, msg.Type)
	}
}

// handleAuthentication handles client authentication
func (wsm *EnhancedWebSocketManager) handleAuthentication(client *EnhancedClient, data interface{}) {
	authData, ok := data.(map[string]interface{})
	if !ok {
		wsm.sendError(client, "invalid_auth_data")
		return
	}

	token, ok := authData["token"].(string)
	if !ok {
		wsm.sendError(client, "missing_token")
		return
	}

	session, valid := wsm.authenticator.ValidateToken(token)
	if !valid {
		wsm.sendError(client, "invalid_token")
		return
	}

	client.mu.Lock()
	client.Authenticated = true
	client.UserID = session.UserID
	client.Roles = session.Roles
	client.Permissions = session.Permissions
	client.mu.Unlock()

	response := BroadcastMessage{
		Type:      "authenticated",
		Data:      map[string]string{"user_id": session.UserID},
		Timestamp: time.Now(),
		Sender:    "system",
	}

	wsm.sendToClient(client, response)
}

// handleJoinRoom handles joining a room
func (wsm *EnhancedWebSocketManager) handleJoinRoom(client *EnhancedClient, roomID string) {
	if !client.Authenticated {
		wsm.sendError(client, "not_authenticated")
		return
	}

	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	room, exists := wsm.rooms[roomID]
	if !exists {
		room = &Room{
			ID:       roomID,
			Name:     roomID,
			Clients:  make(map[string]*EnhancedClient),
			Created:  time.Now(),
			Metadata: make(map[string]interface{}),
		}
		wsm.rooms[roomID] = room
	}

	room.mu.Lock()
	room.Clients[client.ID] = client
	room.mu.Unlock()

	client.mu.Lock()
	client.Rooms[roomID] = true
	client.mu.Unlock()

	// Notify room
	notification := BroadcastMessage{
		Type:      "user_joined",
		Room:      roomID,
		Data:      map[string]string{"user_id": client.UserID, "client_id": client.ID},
		Timestamp: time.Now(),
		Sender:    "system",
	}

	wsm.broadcastToRoom(roomID, notification, client.ID)
}

// handleLeaveRoom handles leaving a room
func (wsm *EnhancedWebSocketManager) handleLeaveRoom(client *EnhancedClient, roomID string) {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	if room, exists := wsm.rooms[roomID]; exists {
		room.mu.Lock()
		delete(room.Clients, client.ID)
		room.mu.Unlock()

		client.mu.Lock()
		delete(client.Rooms, roomID)
		client.mu.Unlock()

		// Notify room
		notification := BroadcastMessage{
			Type:      "user_left",
			Room:      roomID,
			Data:      map[string]string{"user_id": client.UserID, "client_id": client.ID},
			Timestamp: time.Now(),
			Sender:    "system",
		}

		wsm.broadcastToRoom(roomID, notification, client.ID)
	}
}

// handleUserMessage handles user messages
func (wsm *EnhancedWebSocketManager) handleUserMessage(client *EnhancedClient, msg struct {
	Type     string                 `json:"type"`
	Room     string                 `json:"room,omitempty"`
	Data     interface{}            `json:"data"`
	Target   string                 `json:"target,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}) {
	if !client.Authenticated {
		wsm.sendError(client, "not_authenticated")
		return
	}

	broadcast := BroadcastMessage{
		Type:      "message",
		Room:      msg.Room,
		Target:    msg.Target,
		Data:      msg.Data,
		Timestamp: time.Now(),
		Sender:    client.UserID,
		Metadata:  msg.Metadata,
	}

	wsm.broadcastChan <- broadcast
}

// handlePing handles ping messages
func (wsm *EnhancedWebSocketManager) handlePing(client *EnhancedClient) {
	response := BroadcastMessage{
		Type:      "pong",
		Timestamp: time.Now(),
		Sender:    "system",
	}

	wsm.sendToClient(client, response)
}

// handleBroadcast handles broadcast messages
func (wsm *EnhancedWebSocketManager) handleBroadcast(message BroadcastMessage) {
	if message.Room != "" {
		wsm.broadcastToRoom(message.Room, message, "")
	} else if message.Target != "" {
		wsm.sendToTarget(message.Target, message)
	} else {
		wsm.broadcastToAll(message)
	}
}

// broadcastToRoom broadcasts a message to a room
func (wsm *EnhancedWebSocketManager) broadcastToRoom(roomID string, message BroadcastMessage, excludeClientID string) {
	wsm.mu.RLock()
	room, exists := wsm.rooms[roomID]
	wsm.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for clientID, client := range room.Clients {
		if clientID != excludeClientID {
			wsm.sendToClient(client, message)
		}
	}
}

// broadcastToAll broadcasts a message to all clients
func (wsm *EnhancedWebSocketManager) broadcastToAll(message BroadcastMessage) {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	for _, client := range wsm.clients {
		if client.Authenticated {
			wsm.sendToClient(client, message)
		}
	}
}

// sendToTarget sends a message to a specific target
func (wsm *EnhancedWebSocketManager) sendToTarget(targetID string, message BroadcastMessage) {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	for _, client := range wsm.clients {
		if client.UserID == targetID || client.ID == targetID {
			wsm.sendToClient(client, message)
			break
		}
	}
}

// sendToClient sends a message to a specific client
func (wsm *EnhancedWebSocketManager) sendToClient(client *EnhancedClient, message BroadcastMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.SendChan <- data:
	default:
		log.Printf("Client %s send channel full", client.ID)
	}
}

// sendError sends an error message to a client
func (wsm *EnhancedWebSocketManager) sendError(client *EnhancedClient, error string) {
	message := BroadcastMessage{
		Type:      "error",
		Data:      map[string]string{"error": error},
		Timestamp: time.Now(),
		Sender:    "system",
	}

	wsm.sendToClient(client, message)
}

// GetStats returns WebSocket statistics
func (wsm *EnhancedWebSocketManager) GetStats() map[string]interface{} {
	wsm.mu.RLock()
	defer wsm.mu.RUnlock()

	authenticatedClients := 0
	totalRooms := len(wsm.rooms)

	for _, client := range wsm.clients {
		if client.Authenticated {
			authenticatedClients++
		}
	}

	return map[string]interface{}{
		"total_clients":         len(wsm.clients),
		"authenticated_clients": authenticatedClients,
		"total_rooms":           totalRooms,
		"message_queue_size":    wsm.messageQueue.Size(),
		"active_sessions":       wsm.authenticator.ActiveSessions(),
	}
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(workers, maxSize int) *MessageQueue {
	return &MessageQueue{
		messages:   make(chan QueuedMessage, maxSize),
		workers:    workers,
		maxSize:    maxSize,
		processing: false,
	}
}

// Start starts the message queue
func (mq *MessageQueue) Start(ctx context.Context) {
	mq.mu.Lock()
	if mq.processing {
		mq.mu.Unlock()
		return
	}
	mq.processing = true
	mq.mu.Unlock()

	for i := 0; i < mq.workers; i++ {
		go mq.worker(ctx)
	}
}

// worker processes messages from the queue
func (mq *MessageQueue) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-mq.messages:
			// Process message
			time.Sleep(1 * time.Millisecond) // Simulate processing time
		}
	}
}

// Size returns the current queue size
func (mq *MessageQueue) Size() int {
	return len(mq.messages)
}

// NewWebSocketAuthenticator creates a new WebSocket authenticator
func NewWebSocketAuthenticator() *WebSocketAuthenticator {
	return &WebSocketAuthenticator{
		tokens: make(map[string]*AuthSession),
	}
}

// GenerateToken generates a new authentication token
func (wsa *WebSocketAuthenticator) GenerateToken(userID string, roles, permissions []string) string {
	token := fmt.Sprintf("token_%d_%s", time.Now().UnixNano(), userID)

	wsa.mu.Lock()
	defer wsa.mu.Unlock()

	wsa.tokens[token] = &AuthSession{
		Token:       token,
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
		Created:     time.Now(),
		Expires:     time.Now().Add(24 * time.Hour),
	}

	return token
}

// ValidateToken validates a token
func (wsa *WebSocketAuthenticator) ValidateToken(token string) (*AuthSession, bool) {
	wsa.mu.RLock()
	defer wsa.mu.RUnlock()

	session, exists := wsa.tokens[token]
	if !exists || time.Now().After(session.Expires) {
		return nil, false
	}

	return session, true
}

// ActiveSessions returns the number of active sessions
func (wsa *WebSocketAuthenticator) ActiveSessions() int {
	wsa.mu.RLock()
	defer wsa.mu.RUnlock()

	count := 0
	for _, session := range wsa.tokens {
		if time.Now().Before(session.Expires) {
			count++
		}
	}

	return count
}
