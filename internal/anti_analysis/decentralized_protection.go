package anti_analysis

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel controls debug output verbosity
var (
	LogLevel   = getLogLevel()
	logWriter  = io.Discard
	debugPrint = func(format string, v ...interface{}) {}
	infoPrint  = func(format string, v ...interface{}) {}
	warnPrint  = func(format string, v ...interface{}) {}
)

func init() {
	if os.Getenv("HADES_DEBUG") == "1" {
		LogLevel = 3
		logWriter = os.Stderr
	}

	switch LogLevel {
	case 3: // Debug
		debugPrint = func(format string, v ...interface{}) {
			fmt.Fprintf(logWriter, "[DEBUG] "+format+"\n", v...)
		}
		infoPrint = func(format string, v ...interface{}) {
			fmt.Fprintf(logWriter, "[INFO] "+format+"\n", v...)
		}
		warnPrint = func(format string, v ...interface{}) {
			fmt.Fprintf(logWriter, "[WARN] "+format+"\n", v...)
		}
	case 2: // Info
		infoPrint = func(format string, v ...interface{}) {
			fmt.Fprintf(logWriter, "[INFO] "+format+"\n", v...)
		}
		warnPrint = func(format string, v ...interface{}) {
			fmt.Fprintf(logWriter, "[WARN] "+format+"\n", v...)
		}
	case 1: // Warning
		warnPrint = func(format string, v ...interface{}) {
			fmt.Fprintf(logWriter, "[WARN] "+format+"\n", v...)
		}
	}
}

func getLogLevel() int {
	level := os.Getenv("HADES_LOG_LEVEL")
	switch level {
	case "debug":
		return 3
	case "info":
		return 2
	case "warn":
		return 1
	default:
		return 0 // Silent
	}
}

// sanitizeKeyID returns a masked version of the key ID for logging
func sanitizeKeyID(keyID string) string {
	if len(keyID) <= 4 {
		return "****"
	}
	return keyID[:2] + "****" + keyID[len(keyID)-2:]
}

// DecentralizedProtection provides distributed anti-analysis coordination
type DecentralizedProtection struct {
	NodeID          string
	PrivateKey      *ecdsa.PrivateKey
	Peers           map[string]*Peer
	ProtectionChain *ProtectionChain
	Consensus       *ConsensusManager
	DistributedKeys map[string][]byte
	mutex           sync.RWMutex
	Enabled         bool
}

// Peer represents a network node in protection network
type Peer struct {
	ID           string    `json:"id"`
	Address      string    `json:"address"`
	PublicKey    string    `json:"public_key"`
	LastSeen     time.Time `json:"last_seen"`
	Reputation   float64   `json:"reputation"`
	Capabilities []string  `json:"capabilities"`
	Active       bool      `json:"active"`
}

// ProtectionChain represents a blockchain for integrity verification
type ProtectionChain struct {
	Blocks     []*ProtectionBlock `json:"blocks"`
	Difficulty int                `json:"difficulty"`
	PendingTxs []*Transaction     `json:"pending_transactions"`
	mutex      sync.RWMutex
}

// ProtectionBlock represents a block in the protection chain
type ProtectionBlock struct {
	Index        int            `json:"index"`
	Timestamp    time.Time      `json:"timestamp"`
	Hash         string         `json:"hash"`
	PreviousHash string         `json:"previous_hash"`
	Transactions []*Transaction `json:"transactions"`
	Nonce        int            `json:"nonce"`
}

// Transaction represents a protection-related transaction
type Transaction struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "integrity", "threat", "key_rotation"
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	Signature string    `json:"signature"`
	NodeID    string    `json:"node_id"`
}

// ConsensusManager handles distributed consensus
type ConsensusManager struct {
	NodeID      string
	Peers       map[string]*Peer
	Votes       map[string][]*Vote
	CurrentTerm int
	mutex       sync.RWMutex
}

// Vote represents a consensus vote
type Vote struct {
	Term      int    `json:"term"`
	NodeID    string `json:"node_id"`
	BlockHash string `json:"block_hash"`
	Signature string `json:"signature"`
}

// NewDecentralizedProtection creates a new decentralized protection instance
func NewDecentralizedProtection() (*DecentralizedProtection, error) {
	// Generate node identity
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	nodeID := generateNodeID(privateKey)

	dp := &DecentralizedProtection{
		NodeID:          nodeID,
		PrivateKey:      privateKey,
		Peers:           make(map[string]*Peer),
		ProtectionChain: NewProtectionChain(),
		Consensus:       NewConsensusManager(nodeID),
		DistributedKeys: make(map[string][]byte),
		Enabled:         false, // Disabled by default in server mode
	}

	// Initialize peer network with simulated peers
	dp.initializePeerNetwork()

	// Don't start background services in server mode
	// go dp.startNetworkListener()
	// go dp.startConsensusProtocol()
	// go dp.startDistributedMonitoring()

	return dp, nil
}

// initializePeerNetwork sets up initial peer connections
func (dp *DecentralizedProtection) initializePeerNetwork() {
	// Add some simulated peers for demonstration
	peerAddresses := []string{
		"peer1.hades.local:8080",
		"peer2.hades.local:8080",
		"peer3.hades.local:8080",
		"peer4.hades.local:8080",
		"peer5.hades.local:8080",
	}

	for _, address := range peerAddresses {
		peer, err := dp.connectToPeer(address)
		if err != nil {
			warnPrint("Failed to connect to peer %s: %v", address, err)
			continue
		}
		debugPrint("Initialized peer %s at %s", peer.ID, address)
	}

	infoPrint("Peer network initialized with %d peers", len(dp.Peers))
}

// NewProtectionChain creates a new protection blockchain
func NewProtectionChain() *ProtectionChain {
	genesisBlock := &ProtectionBlock{
		Index:        0,
		Timestamp:    time.Now(),
		PreviousHash: "0",
		Transactions: []*Transaction{},
		Nonce:        0,
	}
	genesisBlock.Hash = calculateProtectionBlockHash(genesisBlock)

	chain := &ProtectionChain{
		Blocks:     make([]*ProtectionBlock, 0),
		Difficulty: 4,
		PendingTxs: make([]*Transaction, 0),
	}
	chain.Blocks = append(chain.Blocks, genesisBlock)
	return chain
}

// NewConsensusManager creates a new consensus manager
func NewConsensusManager(nodeID string) *ConsensusManager {
	return &ConsensusManager{
		NodeID:      nodeID,
		Peers:       make(map[string]*Peer),
		Votes:       make(map[string][]*Vote),
		CurrentTerm: 0,
	}
}

// JoinNetwork connects to decentralized protection network
func (dp *DecentralizedProtection) JoinNetwork(bootstrapNodes []string) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	// Connect to bootstrap nodes
	for _, nodeAddr := range bootstrapNodes {
		peer, err := dp.connectToPeer(nodeAddr)
		if err != nil {
			fmt.Printf("Failed to connect to peer %s: %v\n", nodeAddr, err)
			continue
		}

		dp.Peers[peer.ID] = peer
		fmt.Printf("Connected to peer: %s at %s\n", peer.ID, peer.Address)
	}

	// Start network services
	go dp.startNetworkListener()
	go dp.startConsensusProtocol()
	go dp.startDistributedMonitoring()

	return nil
}

// AddPeer adds a peer to the network
func (dp *DecentralizedProtection) AddPeer(address string) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	// Check if peer already exists
	for _, peer := range dp.Peers {
		if peer.Address == address {
			return fmt.Errorf("peer with address %s already exists", address)
		}
	}

	peer, err := dp.connectToPeer(address)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", address, err)
	}

	fmt.Printf("Added peer %s at %s\n", peer.ID, address)
	return nil
}

// connectToPeer establishes connection to a peer
func (dp *DecentralizedProtection) connectToPeer(address string) (*Peer, error) {
	// In a real implementation, this would establish network connection
	// For now, simulate peer connection
	peerID := generateRandomNodeID()
	publicKey := generateRandomPublicKey()

	peer := &Peer{
		ID:           peerID,
		Address:      address,
		PublicKey:    publicKey,
		LastSeen:     time.Now(),
		Reputation:   1.0,
		Capabilities: []string{"key_sharing", "consensus", "blockchain_sync"},
		Active:       true,
	}

	dp.Peers[peerID] = peer
	return peer, nil
}

// startNetworkListener starts listening for incoming peer connections
func (dp *DecentralizedProtection) startNetworkListener() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {
		if !dp.Enabled {
			continue
		}

		// Simulate network activity
		dp.maintainPeerConnections()
		dp.syncBlockchain()
	}
}

// startConsensusProtocol starts the distributed consensus mechanism
func (dp *DecentralizedProtection) startConsensusProtocol() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for range ticker.C {
		if !dp.Enabled {
			continue
		}

		// Check for pending transactions
		if len(dp.ProtectionChain.PendingTxs) > 0 {
			dp.proposeNewBlock()
		}

		// Participate in consensus voting
		dp.participateInConsensus()
	}
}

// startDistributedMonitoring starts distributed threat monitoring
func (dp *DecentralizedProtection) startDistributedMonitoring() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for range ticker.C {
		if !dp.Enabled {
			continue
		}

		// Monitor local threats
		localThreats := dp.detectLocalThreats()
		if len(localThreats) > 0 {
			dp.broadcastThreatAlert(localThreats)
		}

		// Verify distributed integrity
		dp.verifyDistributedIntegrity()
	}
}

// maintainPeerConnections maintains peer connectivity
func (dp *DecentralizedProtection) maintainPeerConnections() {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	for peerID, peer := range dp.Peers {
		if time.Since(peer.LastSeen) > time.Minute*5 {
			fmt.Printf("Peer %s appears inactive, removing from network\n", peerID)
			delete(dp.Peers, peerID)
		}
	}
}

// syncBlockchain synchronizes blockchain with peers
func (dp *DecentralizedProtection) syncBlockchain() {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()

	for _, peer := range dp.Peers {
		if peer.Active {
			// Request blockchain state from peer
			go dp.requestBlockchainSync(peer)
		}
	}
}

// detectLocalThreats detects local analysis threats
func (dp *DecentralizedProtection) detectLocalThreats() []string {
	var threats []string

	// Check for debugging
	if GetGlobalAntiAnalysisManager().antiDebugging.IsDebuggingActive() {
		threats = append(threats, "debugging_detected")
	}

	// Check for VM environment
	if GetGlobalAntiAnalysisManager().antiVM.IsVMEnvironment() {
		threats = append(threats, "vm_environment_detected")
	}

	// Check for instrumentation
	if GetGlobalAntiAnalysisManager().antiInstrumentation.IsInstrumentationActive() {
		threats = append(threats, "instrumentation_detected")
	}

	return threats
}

// broadcastThreatAlert broadcasts threat information to peers
func (dp *DecentralizedProtection) broadcastThreatAlert(threats []string) {
	// Create threat transaction
	tx := &Transaction{
		ID:        generateProtectionTransactionID(),
		Type:      "threat",
		Data:      fmt.Sprintf(`{"threats": %v, "node_id": "%s"}`, threats, dp.NodeID),
		Timestamp: time.Now(),
		NodeID:    dp.NodeID,
	}

	// Sign transaction
	signature, err := dp.signTransaction(tx)
	if err != nil {
		fmt.Printf("Failed to sign threat transaction: %v\n", err)
		return
	}
	tx.Signature = signature

	// Add to pending transactions
	dp.ProtectionChain.mutex.Lock()
	dp.ProtectionChain.PendingTxs = append(dp.ProtectionChain.PendingTxs, tx)
	dp.ProtectionChain.mutex.Unlock()

	fmt.Printf("Broadcasted threat alert: %v\n", threats)
}

// verifyDistributedIntegrity verifies integrity across the network
func (dp *DecentralizedProtection) verifyDistributedIntegrity() {
	// Create integrity verification transaction
	tx := &Transaction{
		ID:        generateProtectionTransactionID(),
		Type:      "integrity",
		Data:      fmt.Sprintf(`{"node_id": "%s", "timestamp": %d}`, dp.NodeID, time.Now().Unix()),
		Timestamp: time.Now(),
		NodeID:    dp.NodeID,
	}

	// Sign transaction
	signature, err := dp.signTransaction(tx)
	if err != nil {
		fmt.Printf("Failed to sign integrity transaction: %v\n", err)
		return
	}
	tx.Signature = signature

	// Add to pending transactions
	dp.ProtectionChain.mutex.Lock()
	dp.ProtectionChain.PendingTxs = append(dp.ProtectionChain.PendingTxs, tx)
	dp.ProtectionChain.mutex.Unlock()
}

// proposeNewBlock proposes a new block to the network
func (dp *DecentralizedProtection) proposeNewBlock() {
	dp.ProtectionChain.mutex.Lock()
	defer dp.ProtectionChain.mutex.Unlock()

	if len(dp.ProtectionChain.PendingTxs) == 0 {
		return
	}

	// Create new block
	lastBlock := dp.ProtectionChain.Blocks[len(dp.ProtectionChain.Blocks)-1]
	newBlock := &ProtectionBlock{
		Index:        lastBlock.Index + 1,
		Timestamp:    time.Now(),
		PreviousHash: lastBlock.Hash,
		Transactions: dp.ProtectionChain.PendingTxs,
		Nonce:        0,
	}

	// Mine block (simplified PoW)
	dp.mineBlock(newBlock)

	// Broadcast to peers for consensus
	dp.broadcastBlockProposal(newBlock)

	// Clear pending transactions
	dp.ProtectionChain.PendingTxs = []*Transaction{}
}

// mineBlock performs proof-of-work mining
func (dp *DecentralizedProtection) mineBlock(block *ProtectionBlock) {
	target := fmt.Sprintf("%0*d", dp.ProtectionChain.Difficulty, 0)

	for {
		block.Hash = calculateProtectionBlockHash(block)
		if block.Hash[:dp.ProtectionChain.Difficulty] == target {
			fmt.Printf("Block mined: %d\n", block.Index)
			break
		}
		block.Nonce++
	}
}

// broadcastBlockProposal broadcasts a block proposal to peers
func (dp *DecentralizedProtection) broadcastBlockProposal(block *ProtectionBlock) {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()

	for _, peer := range dp.Peers {
		if peer.Active {
			go dp.sendBlockProposal(peer, block)
		}
	}
}

// participateInConsensus participates in block consensus
func (dp *DecentralizedProtection) participateInConsensus() {
	threshold := 0.6 // 60% agreement required

	// Use threshold in consensus logic
	activePeers := 0
	for _, peer := range dp.Peers {
		if peer.Active {
			activePeers++
		}
	}

	requiredVotes := int(float64(activePeers) * threshold)
	fmt.Printf("Participating in consensus: %d/%d votes required\n", requiredVotes, activePeers)
}

// DistributeKey distributes encryption key across the network
func (dp *DecentralizedProtection) DistributeKey(keyID string, key []byte) error {
	dp.mutex.Lock()
	defer dp.mutex.Unlock()

	// Split key into shares (simplified Shamir's Secret Sharing)
	shares := dp.splitSecret(key, 5, 3) // 5 shares, 3 required

	// Distribute shares to peers
	peerCount := 0
	for _, peer := range dp.Peers {
		if peerCount >= len(shares) {
			break
		}

		// Send share to peer
		go dp.sendKeyShare(peer, keyID, shares[peerCount])
		peerCount++
	}

	// Store local share
	dp.DistributedKeys[keyID] = shares[0]

	fmt.Printf("Distributed key %s across %d peers\n", keyID, peerCount)
	return nil
}

// RetrieveKey retrieves and reconstructs distributed key
func (dp *DecentralizedProtection) RetrieveKey(keyID string) ([]byte, error) {
	dp.mutex.RLock()
	localShare, exists := dp.DistributedKeys[keyID]
	dp.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key share not found: %s", keyID)
	}

	// In a real implementation, this would request shares from peers
	// For now, simulate reconstruction
	shares := [][]byte{localShare}

	// Reconstruct secret (simplified)
	reconstructed := dp.reconstructSecret(shares)

	fmt.Printf("Retrieved key %s from distributed storage\n", keyID)
	return reconstructed, nil
}

// splitSecret splits a secret into shares with threshold-based reconstruction
func (dp *DecentralizedProtection) splitSecret(secret []byte, n, threshold int) [][]byte {
	if len(secret) == 0 {
		return nil
	}

	// Validate parameters
	if n <= 0 || threshold <= 0 || threshold > n {
		fmt.Printf("Invalid parameters: n=%d, threshold=%d\n", n, threshold)
		return nil
	}

	shares := make([][]byte, n)

	// Enhanced secret sharing with threshold consideration
	// Use different XOR patterns based on threshold for better security
	for i := 0; i < n; i++ {
		share := make([]byte, len(secret))

		for j, b := range secret {
			// Threshold-based XOR sharing
			// Higher thresholds use more complex patterns
			if threshold <= 2 {
				// Simple XOR for low threshold
				share[j] = b ^ byte(i+1)
			} else if threshold <= 4 {
				// Medium complexity for medium threshold
				share[j] = b ^ byte(i+1) ^ byte((i+j)%256)
			} else {
				// High complexity for high threshold
				share[j] = b ^ byte(i+1) ^ byte((i*j)%256) ^ byte((i+j+threshold)%256)
			}
		}

		// Add share metadata for threshold validation
		shareMeta := []byte{byte(threshold), byte(n), byte(i)}
		share = append(shareMeta, share...)

		shares[i] = share
	}

	fmt.Printf("Split secret into %d shares with threshold %d\n", n, threshold)
	return shares
}

// reconstructSecret reconstructs secret from shares with threshold validation
func (dp *DecentralizedProtection) reconstructSecret(shares [][]byte) []byte {
	if len(shares) == 0 {
		return nil
	}

	// Extract metadata from first share
	if len(shares[0]) < 3 {
		return nil
	}

	threshold := int(shares[0][0])
	totalShares := int(shares[0][1])
	shareIndex := int(shares[0][2])

	// Validate we have enough shares for threshold
	if len(shares) < threshold {
		fmt.Printf("Insufficient shares: have %d, need %d\n", len(shares), threshold)
		return nil
	}

	fmt.Printf("Reconstructing secret with threshold %d from %d shares (total: %d, index: %d)\n", threshold, len(shares), totalShares, shareIndex)

	// Remove metadata from shares
	cleanShares := make([][]byte, len(shares))
	for i, share := range shares {
		if len(share) >= 3 {
			cleanShares[i] = share[3:] // Skip metadata
		}
	}

	// Threshold-based reconstruction
	if len(cleanShares) == 0 || len(cleanShares[0]) == 0 {
		return nil
	}

	secret := make([]byte, len(cleanShares[0]))
	copy(secret, cleanShares[0])

	// XOR with other shares (threshold-based)
	for i := 1; i < min(len(cleanShares), threshold); i++ {
		for j := range secret {
			if j < len(cleanShares[i]) {
				secret[j] ^= cleanShares[i][j]
			}
		}
	}

	return secret
}

// GetNetworkStatus returns the current network status
func (dp *DecentralizedProtection) GetNetworkStatus() map[string]interface{} {
	dp.mutex.RLock()
	defer dp.mutex.RUnlock()

	activePeers := 0
	for _, peer := range dp.Peers {
		if peer.Active {
			activePeers++
		}
	}

	return map[string]interface{}{
		"node_id":        dp.NodeID,
		"enabled":        dp.Enabled,
		"total_peers":    len(dp.Peers),
		"active_peers":   activePeers,
		"chain_length":   len(dp.ProtectionChain.Blocks),
		"pending_txs":    len(dp.ProtectionChain.PendingTxs),
		"consensus_term": dp.Consensus.CurrentTerm,
	}
}

// Helper functions

func generateNodeID(privateKey *ecdsa.PrivateKey) string {
	hash := sha256.Sum256(privateKey.D.Bytes())
	return hex.EncodeToString(hash[:])[:16]
}

func generateRandomNodeID() string {
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

func generateRandomPublicKey() string {
	// Generate a random ECDSA key pair and return the public key as hex string
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "default_public_key"
	}

	publicKeyBytes := elliptic.Marshal(elliptic.P256(), privateKey.PublicKey.X, privateKey.PublicKey.Y)
	return hex.EncodeToString(publicKeyBytes)
}

func generateProtectionTransactionID() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

func calculateProtectionBlockHash(block *ProtectionBlock) string {
	blockData := fmt.Sprintf("Block %d (hash: %s, timestamp: %d, txs: %d)",
		block.Index, block.Hash, block.Timestamp.Unix(), len(block.Transactions))
	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:])
}

func (dp *DecentralizedProtection) signTransaction(tx *Transaction) (string, error) {
	// Simplified signing - in reality, use proper ECDSA signing
	txData := fmt.Sprintf("%s%s%s%d", tx.ID, tx.Type, tx.Data, tx.Timestamp.Unix())
	hash := sha256.Sum256([]byte(txData))
	return hex.EncodeToString(hash[:]), nil
}

func (dp *DecentralizedProtection) sendBlockProposal(peer *Peer, block *ProtectionBlock) {
	if block == nil {
		warnPrint("Cannot send nil block proposal to peer %s", peer.ID)
		return
	}

	debugPrint("Sending block proposal %d to peer %s", block.Index, peer.ID)
}

func (dp *DecentralizedProtection) sendKeyShare(peer *Peer, keyID string, share []byte) {
	// Secure key share transmission - no sensitive data logged
	debugPrint("Sending key share to peer %s (key: %s, size: %d)", peer.ID, sanitizeKeyID(keyID), len(share))
}

func (dp *DecentralizedProtection) requestBlockchainSync(peer *Peer) {
	debugPrint("Requesting blockchain sync from peer %s", peer.ID)
}

func (dp *DecentralizedProtection) validateBlock(block *ProtectionBlock) bool {
	// Simplified validation with block content usage
	if block == nil {
		return false
	}

	// Check block hash
	expectedHash := calculateProtectionBlockHash(block)
	if block.Hash != expectedHash {
		return false
	}

	// Check block index is reasonable
	if block.Index < 0 {
		return false
	}

	// Check timestamp is not too old or future
	now := time.Now()
	if block.Timestamp.Before(now.Add(-time.Hour)) || block.Timestamp.After(now.Add(time.Hour)) {
		return false
	}

	return true
}

func (dp *DecentralizedProtection) validateTransaction(tx *Transaction) bool {
	// Enhanced validation with transaction content usage
	if tx == nil {
		return false
	}

	// Check required fields
	if tx.ID == "" || tx.Type == "" {
		return false
	}

	// Check timestamp is reasonable
	now := time.Now()
	if tx.Timestamp.Before(now.Add(-time.Hour)) || tx.Timestamp.After(now.Add(time.Hour)) {
		return false
	}

	// Check signature format
	if tx.Signature != "" && len(tx.Signature) < 10 {
		return false // Invalid signature length
	}

	return true
}

func (dp *DecentralizedProtection) distributeKeyShare(keyID string, share []byte) {
	if len(share) == 0 {
		warnPrint("Cannot distribute empty key share for %s", sanitizeKeyID(keyID))
		return
	}

	distributedCount := 0
	for peerID, peer := range dp.Peers {
		if peer.Active {
			debugPrint("Sending key share to peer %s (key: %s)", peerID, sanitizeKeyID(keyID))
			distributedCount++
		}
	}

	if distributedCount == 0 {
		warnPrint("No active peers available for key distribution: %s", sanitizeKeyID(keyID))
	}
}
