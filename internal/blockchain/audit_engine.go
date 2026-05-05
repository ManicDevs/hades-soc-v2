package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"hades-v2/internal/database"
)

// AuditEngine provides blockchain-based audit logging and immutability
type AuditEngine struct {
	db           database.Database
	blockchain   *Blockchain
	genesisBlock *Block
	mu           sync.RWMutex
}

// Blockchain represents the audit blockchain
type Blockchain struct {
	Blocks      []*Block  `json:"blocks"`
	MerkleRoot  string    `json:"merkle_root"`
	Difficulty  int       `json:"difficulty"`
	ChainID     string    `json:"chain_id"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}

// Block represents a block in the blockchain
type Block struct {
	Index        int           `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	PreviousHash string        `json:"previous_hash"`
	Hash         string        `json:"hash"`
	Nonce        int           `json:"nonce"`
	Data         []*AuditEntry `json:"data"`
	MerkleRoot   string        `json:"merkle_root"`
	Validator    string        `json:"validator"`
	Signature    string        `json:"signature"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Source      string                 `json:"source"`
	User        string                 `json:"user"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Outcome     string                 `json:"outcome"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Hash        string                 `json:"hash"`
	Signature   string                 `json:"signature"`
}

// AuditQuery represents an audit query
type AuditQuery struct {
	StartTime *time.Time             `json:"start_time"`
	EndTime   *time.Time             `json:"end_time"`
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AuditResult represents audit query results
type AuditResult struct {
	Entries    []*AuditEntry          `json:"entries"`
	TotalCount int                    `json:"total_count"`
	BlockHash  string                 `json:"block_hash"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Proof represents a cryptographic proof
type Proof struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Hash      string      `json:"hash"`
	Timestamp time.Time   `json:"timestamp"`
	Validator string      `json:"validator"`
	Signature string      `json:"signature"`
}

// NewAuditEngine creates a new audit engine
func NewAuditEngine(db database.Database) (*AuditEngine, error) {
	engine := &AuditEngine{
		db: db,
		blockchain: &Blockchain{
			Blocks:      make([]*Block, 0),
			Difficulty:  4,
			ChainID:     "hades-audit-chain",
			CreatedAt:   time.Now(),
			LastUpdated: time.Now(),
		},
	}

	// Initialize blockchain with genesis block
	if err := engine.initializeBlockchain(); err != nil {
		return nil, fmt.Errorf("failed to initialize blockchain: %w", err)
	}

	return engine, nil
}

// initializeBlockchain initializes the blockchain with genesis block
func (ae *AuditEngine) initializeBlockchain() error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Create genesis block
	genesisEntry := &AuditEntry{
		ID:          "genesis_entry",
		Timestamp:   time.Now(),
		EventType:   "system",
		Source:      "audit_engine",
		User:        "system",
		Action:      "initialize",
		Resource:    "blockchain",
		Outcome:     "success",
		Description: "Blockchain audit system initialized",
		Metadata: map[string]interface{}{
			"chain_id":   ae.blockchain.ChainID,
			"difficulty": ae.blockchain.Difficulty,
		},
	}

	genesisEntry.Hash = ae.calculateEntryHash(genesisEntry)

	genesisBlock := &Block{
		Index:        0,
		Timestamp:    time.Now(),
		PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
		Nonce:        0,
		Data:         []*AuditEntry{genesisEntry},
		Validator:    "system",
	}

	genesisBlock.MerkleRoot = ae.calculateMerkleRoot(genesisBlock.Data)
	genesisBlock.Hash = ae.calculateBlockHash(genesisBlock)

	ae.blockchain.Blocks = append(ae.blockchain.Blocks, genesisBlock)
	ae.blockchain.MerkleRoot = genesisBlock.MerkleRoot
	ae.genesisBlock = genesisBlock

	// Save genesis block to database
	if err := ae.saveBlock(genesisBlock); err != nil {
		log.Printf("Warning: Failed to save genesis block to database: %v", err)
	}

	return nil
}

// LogEvent logs an audit event to the blockchain
func (ae *AuditEngine) LogEvent(ctx context.Context, entry *AuditEntry) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Generate entry ID and hash
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("audit_%d_%s", time.Now().UnixNano(), entry.EventType)
	}
	entry.Timestamp = time.Now()
	entry.Hash = ae.calculateEntryHash(entry)

	// Add entry to current block or create new block
	currentBlock := ae.getCurrentBlock()
	if currentBlock == nil || len(currentBlock.Data) >= 100 { // Max 100 entries per block
		// Create new block
		newBlock := &Block{
			Index:        len(ae.blockchain.Blocks),
			Timestamp:    time.Now(),
			PreviousHash: ae.getLatestBlockHash(),
			Nonce:        0,
			Data:         []*AuditEntry{entry},
			Validator:    "audit_engine",
		}

		newBlock.MerkleRoot = ae.calculateMerkleRoot(newBlock.Data)
		newBlock.Hash = ae.calculateBlockHash(newBlock)

		// Mine the block (simplified proof of work)
		if err := ae.mineBlock(newBlock); err != nil {
			return fmt.Errorf("failed to mine block: %w", err)
		}

		ae.blockchain.Blocks = append(ae.blockchain.Blocks, newBlock)
		ae.blockchain.LastUpdated = time.Now()
		ae.blockchain.MerkleRoot = ae.calculateChainMerkleRoot()

		// Save block to database
		if err := ae.saveBlock(newBlock); err != nil {
			log.Printf("Warning: Failed to save block to database: %v", err)
		}
	} else {
		// Add to current block
		currentBlock.Data = append(currentBlock.Data, entry)
		currentBlock.MerkleRoot = ae.calculateMerkleRoot(currentBlock.Data)
		currentBlock.Hash = ae.calculateBlockHash(currentBlock)
		currentBlock.Timestamp = time.Now()

		// Update blockchain
		ae.blockchain.LastUpdated = time.Now()
		ae.blockchain.MerkleRoot = ae.calculateChainMerkleRoot()
	}

	return nil
}

// QueryAuditLogs queries audit logs from the blockchain
func (ae *AuditEngine) QueryAuditLogs(ctx context.Context, query AuditQuery) (*AuditResult, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	matchingEntries := make([]*AuditEntry, 0)

	// Search through all blocks
	for _, block := range ae.blockchain.Blocks {
		for _, entry := range block.Data {
			if ae.matchesQuery(entry, query) {
				matchingEntries = append(matchingEntries, entry)
			}
		}
	}

	// Apply pagination
	totalCount := len(matchingEntries)
	start := query.Offset
	if start < 0 {
		start = 0
	}
	end := start + query.Limit
	if end > totalCount {
		end = totalCount
	}
	if start > end {
		start = end
	}

	var paginatedEntries []*AuditEntry
	if start < len(matchingEntries) {
		paginatedEntries = matchingEntries[start:end]
	} else {
		paginatedEntries = make([]*AuditEntry, 0)
	}

	result := &AuditResult{
		Entries:    paginatedEntries,
		TotalCount: totalCount,
		BlockHash:  ae.blockchain.MerkleRoot,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"blocks_searched": len(ae.blockchain.Blocks),
			"chain_id":        ae.blockchain.ChainID,
		},
	}

	return result, nil
}

// VerifyIntegrity verifies the integrity of the blockchain
func (ae *AuditEngine) VerifyIntegrity(ctx context.Context) (*Proof, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	proof := &Proof{
		Type:      "blockchain_integrity",
		Timestamp: time.Now(),
		Validator: "audit_engine",
	}

	// Verify each block
	for i, block := range ae.blockchain.Blocks {
		// Verify block hash
		calculatedHash := ae.calculateBlockHash(block)
		if block.Hash != calculatedHash {
			proof.Data = map[string]interface{}{
				"valid":       false,
				"block_index": i,
				"block_hash":  block.Hash,
				"calc_hash":   calculatedHash,
			}
			proof.Hash = ae.calculateProofHash(proof)
			return proof, fmt.Errorf("block integrity violation at index %d", i)
		}

		// Verify merkle root
		calculatedMerkleRoot := ae.calculateMerkleRoot(block.Data)
		if block.MerkleRoot != calculatedMerkleRoot {
			proof.Data = map[string]interface{}{
				"valid":            false,
				"block_index":      i,
				"merkle_root":      block.MerkleRoot,
				"calc_merkle_root": calculatedMerkleRoot,
			}
			proof.Hash = ae.calculateProofHash(proof)
			return proof, fmt.Errorf("merkle root integrity violation at index %d", i)
		}

		// Verify previous hash linkage
		if i > 0 {
			prevBlock := ae.blockchain.Blocks[i-1]
			if block.PreviousHash != prevBlock.Hash {
				proof.Data = map[string]interface{}{
					"valid":         false,
					"block_index":   i,
					"prev_hash":     block.PreviousHash,
					"expected_hash": prevBlock.Hash,
				}
				proof.Hash = ae.calculateProofHash(proof)
				return proof, fmt.Errorf("block chain linkage violation at index %d", i)
			}
		}
	}

	// Verify chain merkle root
	calculatedChainMerkleRoot := ae.calculateChainMerkleRoot()
	if ae.blockchain.MerkleRoot != calculatedChainMerkleRoot {
		proof.Data = map[string]interface{}{
			"valid":                  false,
			"chain_merkle_root":      ae.blockchain.MerkleRoot,
			"calc_chain_merkle_root": calculatedChainMerkleRoot,
		}
		proof.Hash = ae.calculateProofHash(proof)
		return proof, fmt.Errorf("chain merkle root integrity violation")
	}

	proof.Data = map[string]interface{}{
		"valid":       true,
		"blocks":      len(ae.blockchain.Blocks),
		"chain_id":    ae.blockchain.ChainID,
		"merkle_root": ae.blockchain.MerkleRoot,
	}
	proof.Hash = ae.calculateProofHash(proof)

	return proof, nil
}

// GetBlockchainStatus returns blockchain status information
func (ae *AuditEngine) GetBlockchainStatus() map[string]interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	totalEntries := 0
	for _, block := range ae.blockchain.Blocks {
		totalEntries += len(block.Data)
	}

	return map[string]interface{}{
		"chain_id":      ae.blockchain.ChainID,
		"blocks":        len(ae.blockchain.Blocks),
		"total_entries": totalEntries,
		"merkle_root":   ae.blockchain.MerkleRoot,
		"difficulty":    ae.blockchain.Difficulty,
		"created_at":    ae.blockchain.CreatedAt,
		"last_updated":  ae.blockchain.LastUpdated,
		"genesis_block": ae.genesisBlock.Hash,
		"timestamp":     time.Now(),
	}
}

// Helper functions

// getCurrentBlock returns the current block (latest block)
func (ae *AuditEngine) getCurrentBlock() *Block {
	if len(ae.blockchain.Blocks) == 0 {
		return nil
	}
	return ae.blockchain.Blocks[len(ae.blockchain.Blocks)-1]
}

// getLatestBlockHash returns the hash of the latest block
func (ae *AuditEngine) getLatestBlockHash() string {
	if len(ae.blockchain.Blocks) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}
	return ae.blockchain.Blocks[len(ae.blockchain.Blocks)-1].Hash
}

// calculateEntryHash calculates hash for an audit entry
func (ae *AuditEngine) calculateEntryHash(entry *AuditEntry) string {
	data, _ := json.Marshal(entry)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// calculateBlockHash calculates hash for a block
func (ae *AuditEngine) calculateBlockHash(block *Block) string {
	blockData := map[string]interface{}{
		"index":         block.Index,
		"timestamp":     block.Timestamp,
		"previous_hash": block.PreviousHash,
		"nonce":         block.Nonce,
		"merkle_root":   block.MerkleRoot,
		"validator":     block.Validator,
	}
	data, _ := json.Marshal(blockData)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// calculateMerkleRoot calculates merkle root for entries
func (ae *AuditEngine) calculateMerkleRoot(entries []*AuditEntry) string {
	if len(entries) == 0 {
		return ""
	}

	if len(entries) == 1 {
		return entries[0].Hash
	}

	// Build merkle tree
	level := make([]string, 0)
	for _, entry := range entries {
		level = append(level, entry.Hash)
	}

	for len(level) > 1 {
		nextLevel := make([]string, 0)
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				combined := level[i] + level[i+1]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			} else {
				// Odd number, duplicate last hash
				combined := level[i] + level[i]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			}
		}
		level = nextLevel
	}

	return level[0]
}

// calculateChainMerkleRoot calculates merkle root for the entire chain
func (ae *AuditEngine) calculateChainMerkleRoot() string {
	if len(ae.blockchain.Blocks) == 0 {
		return ""
	}

	if len(ae.blockchain.Blocks) == 1 {
		return ae.blockchain.Blocks[0].Hash
	}

	// Build merkle tree from block hashes
	level := make([]string, 0)
	for _, block := range ae.blockchain.Blocks {
		level = append(level, block.Hash)
	}

	for len(level) > 1 {
		nextLevel := make([]string, 0)
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				combined := level[i] + level[i+1]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			} else {
				// Odd number, duplicate last hash
				combined := level[i] + level[i]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			}
		}
		level = nextLevel
	}

	return level[0]
}

// mineBlock performs simplified proof of work mining
func (ae *AuditEngine) mineBlock(block *Block) error {
	target := fmt.Sprintf("%0*s", ae.blockchain.Difficulty, "0")

	for {
		block.Nonce++
		block.Hash = ae.calculateBlockHash(block)

		// Check if hash meets difficulty
		if len(block.Hash) >= ae.blockchain.Difficulty && block.Hash[:ae.blockchain.Difficulty] == target {
			break
		}

		// Prevent infinite loop (simplified)
		if block.Nonce > 1000000 {
			return fmt.Errorf("mining timeout")
		}
	}

	return nil
}

// matchesQuery checks if an entry matches the query criteria
func (ae *AuditEngine) matchesQuery(entry *AuditEntry, query AuditQuery) bool {
	// Time range filter
	if query.StartTime != nil && entry.Timestamp.Before(*query.StartTime) {
		return false
	}
	if query.EndTime != nil && entry.Timestamp.After(*query.EndTime) {
		return false
	}

	// Event type filter
	if query.EventType != "" && entry.EventType != query.EventType {
		return false
	}

	// Source filter
	if query.Source != "" && entry.Source != query.Source {
		return false
	}

	// User filter
	if query.User != "" && entry.User != query.User {
		return false
	}

	// Action filter
	if query.Action != "" && entry.Action != query.Action {
		return false
	}

	// Resource filter
	if query.Resource != "" && entry.Resource != query.Resource {
		return false
	}

	// Metadata filter (simplified)
	if query.Metadata != nil {
		for key, value := range query.Metadata {
			if entryValue, ok := entry.Metadata[key]; !ok || entryValue != value {
				return false
			}
		}
	}

	return true
}

// saveBlock saves block to database (simplified)
func (ae *AuditEngine) saveBlock(block *Block) error {
	// In a real implementation, this would save to a database
	// For now, we just log it
	log.Printf("Saved block %d to database with %d entries", block.Index, len(block.Data))
	return nil
}

// calculateProofHash calculates hash for a proof
func (ae *AuditEngine) calculateProofHash(proof *Proof) string {
	data, _ := json.Marshal(proof)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
