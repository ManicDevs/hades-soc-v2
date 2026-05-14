package p2p

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// LogLevel controls debug output verbosity
var (
	P2PLogLevel = getP2PLogLevel()
	p2pWriter   = io.Discard
	p2pDebug    = func(format string, v ...interface{}) {}
	p2pInfo     = func(format string, v ...interface{}) {}
	p2pWarn     = func(format string, v ...interface{}) {}
)

func initP2P() {
	if os.Getenv("HADES_DEBUG") == "1" {
		P2PLogLevel = 3
		p2pWriter = os.Stderr
	}

	switch P2PLogLevel {
	case 3:
		p2pDebug = func(format string, v ...interface{}) {
			fmt.Fprintf(p2pWriter, "[P2P-DEBUG] "+format+"\n", v...)
		}
		p2pInfo = func(format string, v ...interface{}) {
			fmt.Fprintf(p2pWriter, "[P2P-INFO] "+format+"\n", v...)
		}
		p2pWarn = func(format string, v ...interface{}) {
			fmt.Fprintf(p2pWriter, "[P2P-WARN] "+format+"\n", v...)
		}
	case 2:
		p2pInfo = func(format string, v ...interface{}) {
			fmt.Fprintf(p2pWriter, "[P2P-INFO] "+format+"\n", v...)
		}
		p2pWarn = func(format string, v ...interface{}) {
			fmt.Fprintf(p2pWriter, "[P2P-WARN] "+format+"\n", v...)
		}
	case 1:
		p2pWarn = func(format string, v ...interface{}) {
			fmt.Fprintf(p2pWriter, "[P2P-WARN] "+format+"\n", v...)
		}
	}
}

func getP2PLogLevel() int {
	level := os.Getenv("HADES_LOG_LEVEL")
	switch level {
	case "debug":
		return 3
	case "info":
		return 2
	case "warn":
		return 1
	default:
		return 0
	}
}

func init() {
	initP2P()
}

type MessageType byte

const (
	MsgHandshake MessageType = iota
	MsgKeyShare
	MsgBlockProposal
	MsgConsensusVote
	MsgHeartbeat
	MsgDisconnect
)

type Message struct {
	Type      MessageType
	SenderID  string
	Timestamp time.Time
	Payload   []byte
}

type Peer struct {
	ID        string
	Conn      net.Conn
	Connected bool
	LastSeen  time.Time
	mu        sync.RWMutex
}

type P2PNetwork struct {
	NodeID      string
	ListenAddr  string
	Peers       map[string]*Peer
	Listener    net.Listener
	MessageChan chan *Message
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

func NewP2PNetwork(nodeID, listenAddr string) *P2PNetwork {
	ctx, cancel := context.WithCancel(context.Background())
	return &P2PNetwork{
		NodeID:      nodeID,
		ListenAddr:  listenAddr,
		Peers:       make(map[string]*Peer),
		MessageChan: make(chan *Message, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (p2p *P2PNetwork) Start() error {
	ln, err := net.Listen("tcp", p2p.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", p2p.ListenAddr, err)
	}
	p2p.Listener = ln

	go p2p.acceptLoop()
	go p2p.heartbeatLoop()

	p2pInfo("Listening on %s", p2p.ListenAddr)
	return nil
}

func (p2p *P2PNetwork) acceptLoop() {
	for {
		conn, err := p2p.Listener.Accept()
		if err != nil {
			select {
			case <-p2p.ctx.Done():
				return
			default:
				p2pWarn("Accept error: %v", err)
				continue
			}
		}
		go p2p.handleConnection(conn)
	}
}

func (p2p *P2PNetwork) handleConnection(conn net.Conn) {
	msg, err := p2p.readMessage(conn)
	if err != nil {
		p2pWarn("Failed to read handshake from %s: %v", conn.RemoteAddr(), err)
		return
	}

	if msg.Type != MsgHandshake {
		p2pWarn("Expected handshake, got %v", msg.Type)
		return
	}

	peerID := msg.SenderID
	peer := &Peer{
		ID:        peerID,
		Conn:      conn,
		Connected: true,
		LastSeen:  time.Now(),
	}

	p2p.mu.Lock()
	p2p.Peers[peerID] = peer
	p2p.mu.Unlock()

	p2pDebug("Connected to peer %s at %s", peerID, conn.RemoteAddr())

	// Send handshake response
	resp := &Message{
		Type:      MsgHandshake,
		SenderID:  p2p.NodeID,
		Timestamp: time.Now(),
		Payload:   []byte("accepted"),
	}
	p2p.writeMessage(conn, resp)

	go p2p.readLoop(peer)
}

func (p2p *P2PNetwork) readLoop(peer *Peer) {
	for {
		msg, err := p2p.readMessage(peer.Conn)
		if err != nil {
			p2p.mu.Lock()
			peer.Connected = false
			delete(p2p.Peers, peer.ID)
			p2p.mu.Unlock()

			p2pDebug("Peer %s disconnected: %v", peer.ID, err)
			return
		}

		peer.LastSeen = time.Now()
		p2p.MessageChan <- msg
	}
}

func (p2p *P2PNetwork) ConnectToPeer(address string) error {
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	handshake := &Message{
		Type:      MsgHandshake,
		SenderID:  p2p.NodeID,
		Timestamp: time.Now(),
		Payload:   []byte(p2p.ListenAddr),
	}

	if err := p2p.writeMessage(conn, handshake); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	msg, err := p2p.readMessage(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Small delay to ensure server has started reading
	time.Sleep(100 * time.Millisecond)

	peerID := msg.SenderID
	peer := &Peer{
		ID:        peerID,
		Conn:      conn,
		Connected: true,
		LastSeen:  time.Now(),
	}

	p2p.mu.Lock()
	p2p.Peers[peerID] = peer
	p2p.mu.Unlock()

	go p2p.readLoop(peer)

	p2pDebug("Connected to peer %s at %s", peerID, address)
	return nil
}

func (p2p *P2PNetwork) SendToPeer(peerID string, msg *Message) error {
	p2p.mu.RLock()
	peer, exists := p2p.Peers[peerID]
	p2p.mu.RUnlock()

	if !exists || !peer.Connected {
		return fmt.Errorf("peer %s not connected", peerID)
	}

	msg.SenderID = p2p.NodeID
	msg.Timestamp = time.Now()

	return p2p.writeMessage(peer.Conn, msg)
}

func (p2p *P2PNetwork) Broadcast(msg *Message) {
	p2p.mu.RLock()
	defer p2p.mu.RUnlock()

	for peerID, peer := range p2p.Peers {
		if peer.Connected {
			msg.SenderID = p2p.NodeID
			msg.Timestamp = time.Now()
			if err := p2p.writeMessage(peer.Conn, msg); err != nil {
				p2pWarn("Failed to broadcast to %s: %v", peerID, err)
			}
		}
	}
}

func (p2p *P2PNetwork) readMessage(conn net.Conn) (*Message, error) {
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	var msg Message
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (p2p *P2PNetwork) writeMessage(conn net.Conn, msg *Message) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(msg); err != nil {
		return err
	}

	data := buf.Bytes()
	if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	_, err := conn.Write(data)
	return err
}

func (p2p *P2PNetwork) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p2p.ctx.Done():
			return
		case <-ticker.C:
			p2p.sendHeartbeats()
		}
	}
}

func (p2p *P2PNetwork) sendHeartbeats() {
	p2p.mu.RLock()
	defer p2p.mu.RUnlock()

	msg := &Message{
		Type:      MsgHeartbeat,
		SenderID:  p2p.NodeID,
		Timestamp: time.Now(),
	}

	for peerID, peer := range p2p.Peers {
		if peer.Connected {
			if err := p2p.writeMessage(peer.Conn, msg); err != nil {
				p2pWarn("Heartbeat failed to %s: %v", peerID, err)
				peer.Connected = false
			}
		}
	}
}

func (p2p *P2PNetwork) GetPeers() map[string]*Peer {
	p2p.mu.RLock()
	defer p2p.mu.RUnlock()

	result := make(map[string]*Peer)
	for k, v := range p2p.Peers {
		result[k] = v
	}
	return result
}

func (p2p *P2PNetwork) Close() error {
	p2p.cancel()

	p2p.mu.RLock()
	for _, peer := range p2p.Peers {
		peer.Conn.Close()
	}
	p2p.mu.RUnlock()

	if p2p.Listener != nil {
		return p2p.Listener.Close()
	}
	return nil
}
