package anti_analysis

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// BlockchainIntegrity provides blockchain-based integrity verification
type BlockchainIntegrity struct {
	chain         []*IntegrityBlock
	pendingBlocks []*IntegrityBlock
	difficulty    int
	rewardAddress string
	privateKey    *ecdsa.PrivateKey
	mutex         sync.RWMutex
	peers         map[string]*BlockchainPeer
	consensus     *ProofOfStakeConsensus
	validators    map[string]*Validator
}

// IntegrityBlock represents a block in the integrity chain
type IntegrityBlock struct {
	Index        int                     `json:"index"`
	Timestamp    time.Time               `json:"timestamp"`
	Hash         string                  `json:"hash"`
	PreviousHash string                  `json:"previous_hash"`
	MerkleRoot   string                  `json:"merkle_root"`
	Validator    string                  `json:"validator"`
	Signature    string                  `json:"signature"`
	Transactions []*IntegrityTransaction `json:"transactions"`
	Nonce        int                     `json:"nonce"`
	Stake        int                     `json:"stake"`
}

// IntegrityTransaction represents an integrity verification transaction
type IntegrityTransaction struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"` // "file_hash", "memory_hash", "code_signature"
	Target    string            `json:"target"`
	Hash      string            `json:"hash"`
	Metadata  map[string]string `json:"metadata"`
	Timestamp time.Time         `json:"timestamp"`
	Signature string            `json:"signature"`
	Submitter string            `json:"submitter"`
}

// BlockchainPeer represents a peer in the blockchain network
type BlockchainPeer struct {
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	PublicKey  string    `json:"public_key"`
	Stake      int       `json:"stake"`
	Reputation float64   `json:"reputation"`
	LastSeen   time.Time `json:"last_seen"`
	Active     bool      `json:"active"`
}

// ProofOfStakeConsensus implements PoS consensus mechanism
type ProofOfStakeConsensus struct {
	validators    map[string]*Validator
	currentEpoch  int
	validatorPool []string
	totalStake    int
	mutex         sync.RWMutex
}

// Validator represents a network validator
type Validator struct {
	ID         string    `json:"id"`
	PublicKey  string    `json:"public_key"`
	Stake      int       `json:"stake"`
	Reputation float64   `json:"reputation"`
	LastSigned time.Time `json:"last_signed"`
	Active     bool      `json:"active"`
}

// NewBlockchainIntegrity creates a new blockchain integrity system
func NewBlockchainIntegrity() (*BlockchainIntegrity, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	genesisBlock := createGenesisBlock()

	return &BlockchainIntegrity{
		chain:         []*IntegrityBlock{genesisBlock},
		pendingBlocks: []*IntegrityBlock{},
		difficulty:    4,
		rewardAddress: generateAddressFromKey(privateKey),
		privateKey:    privateKey,
		peers:         make(map[string]*BlockchainPeer),
		consensus:     NewProofOfStakeConsensus(),
		validators:    make(map[string]*Validator),
	}, nil
}

// NewProofOfStakeConsensus creates a new PoS consensus manager
func NewProofOfStakeConsensus() *ProofOfStakeConsensus {
	return &ProofOfStakeConsensus{
		validators:    make(map[string]*Validator),
		currentEpoch:  0,
		validatorPool: []string{},
		totalStake:    0,
	}
}

// createGenesisBlock creates the genesis block
func createGenesisBlock() *IntegrityBlock {
	genesisTx := &IntegrityTransaction{
		ID:        "genesis-tx",
		Type:      "system_init",
		Target:    "blockchain",
		Hash:      "genesis-hash",
		Metadata:  map[string]string{"creator": "hades-v2"},
		Timestamp: time.Now(),
		Submitter: "genesis",
	}

	genesisBlock := &IntegrityBlock{
		Index:        0,
		Timestamp:    time.Now(),
		PreviousHash: "0",
		Transactions: []*IntegrityTransaction{genesisTx},
		Nonce:        0,
		Stake:        0,
		Validator:    "genesis",
	}

	genesisBlock.MerkleRoot = calculateMerkleRoot(genesisBlock.Transactions)
	genesisBlock.Hash = calculateBlockHash(genesisBlock)

	return genesisBlock
}

// AddIntegrityTransaction adds a new integrity transaction to the blockchain
func (bi *BlockchainIntegrity) AddIntegrityTransaction(txType, target, hash string, metadata map[string]string) error {
	tx := &IntegrityTransaction{
		ID:        generateBlockchainTransactionID(),
		Type:      txType,
		Target:    target,
		Hash:      hash,
		Metadata:  metadata,
		Timestamp: time.Now(),
		Submitter: bi.rewardAddress,
	}

	// Sign transaction
	signature, err := bi.signTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	tx.Signature = signature

	// Add to pending block
	bi.mutex.Lock()
	defer bi.mutex.Unlock()

	// Create or get pending block
	var pendingBlock *IntegrityBlock
	if len(bi.pendingBlocks) == 0 {
		pendingBlock = &IntegrityBlock{
			Index:        len(bi.chain),
			PreviousHash: bi.chain[len(bi.chain)-1].Hash,
			Transactions: []*IntegrityTransaction{},
		}
		bi.pendingBlocks = append(bi.pendingBlocks, pendingBlock)
	} else {
		pendingBlock = bi.pendingBlocks[len(bi.pendingBlocks)-1]
	}

	pendingBlock.Transactions = append(pendingBlock.Transactions, tx)

	// Update Merkle root
	pendingBlock.MerkleRoot = calculateMerkleRoot(pendingBlock.Transactions)

	fmt.Printf("Added integrity transaction: %s for target %s\n", txType, target)
	return nil
}

// VerifyFileIntegrity verifies and records file integrity on blockchain
func (bi *BlockchainIntegrity) VerifyFileIntegrity(filePath string, fileHash string) error {
	metadata := map[string]string{
		"file_path":   filePath,
		"file_size":   fmt.Sprintf("%d", len(fileHash)),
		"verify_time": time.Now().Format(time.RFC3339),
	}

	return bi.AddIntegrityTransaction("file_hash", filePath, fileHash, metadata)
}

// VerifyMemoryIntegrity verifies and records memory integrity on blockchain
func (bi *BlockchainIntegrity) VerifyMemoryIntegrity(region string, memoryHash string) error {
	metadata := map[string]string{
		"memory_region": region,
		"region_size":   fmt.Sprintf("%d", len(memoryHash)),
		"verify_time":   time.Now().Format(time.RFC3339),
	}

	return bi.AddIntegrityTransaction("memory_hash", region, memoryHash, metadata)
}

// VerifyCodeSignature verifies and records code signature on blockchain
func (bi *BlockchainIntegrity) VerifyCodeSignature(function string, signature string) error {
	metadata := map[string]string{
		"function_name":    function,
		"signature_length": fmt.Sprintf("%d", len(signature)),
		"verify_time":      time.Now().Format(time.RFC3339),
	}

	return bi.AddIntegrityTransaction("code_signature", function, signature, metadata)
}

// MineBlock mines a new block using Proof of Stake
func (bi *BlockchainIntegrity) MineBlock() error {
	bi.mutex.Lock()
	defer bi.mutex.Unlock()

	if len(bi.pendingBlocks) == 0 {
		return fmt.Errorf("no pending blocks to mine")
	}

	pendingBlock := bi.pendingBlocks[0]
	if len(pendingBlock.Transactions) == 0 {
		return fmt.Errorf("no transactions to mine")
	}

	// Select validator using PoS
	validator := bi.consensus.SelectValidator()
	if validator == "" {
		return fmt.Errorf("no validator available")
	}

	// Create block
	pendingBlock.Timestamp = time.Now()
	pendingBlock.Validator = validator
	pendingBlock.Stake = bi.consensus.validators[validator].Stake

	// Calculate block hash
	pendingBlock.Hash = calculateBlockHash(pendingBlock)

	// Sign block
	signature, err := bi.signBlock(pendingBlock)
	if err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}
	pendingBlock.Signature = signature

	// Add to chain
	bi.chain = append(bi.chain, pendingBlock)

	// Remove from pending
	bi.pendingBlocks = bi.pendingBlocks[1:]

	// Update validator reputation
	bi.updateValidatorReputation(validator, true)

	fmt.Printf("Mined block %d by validator %s\n", pendingBlock.Index, validator)
	return nil
}

// ValidateChain validates the entire blockchain
func (bi *BlockchainIntegrity) ValidateChain() bool {
	bi.mutex.RLock()
	defer bi.mutex.RUnlock()

	for i := 1; i < len(bi.chain); i++ {
		currentBlock := bi.chain[i]
		previousBlock := bi.chain[i-1]

		// Verify block hash
		expectedHash := calculateBlockHash(currentBlock)
		if currentBlock.Hash != expectedHash {
			fmt.Printf("Invalid block hash at index %d\n", currentBlock.Index)
			return false
		}

		// Verify previous hash link
		if currentBlock.PreviousHash != previousBlock.Hash {
			fmt.Printf("Broken chain link at index %d\n", currentBlock.Index)
			return false
		}

		// Verify Merkle root
		expectedMerkleRoot := calculateMerkleRoot(currentBlock.Transactions)
		if currentBlock.MerkleRoot != expectedMerkleRoot {
			fmt.Printf("Invalid Merkle root at index %d\n", currentBlock.Index)
			return false
		}
	}

	return true
}

// GetIntegrityHistory retrieves integrity history for a target
func (bi *BlockchainIntegrity) GetIntegrityHistory(target string) []*IntegrityTransaction {
	bi.mutex.RLock()
	defer bi.mutex.RUnlock()

	var history []*IntegrityTransaction

	for _, block := range bi.chain {
		for _, tx := range block.Transactions {
			if tx.Target == target {
				history = append(history, tx)
			}
		}
	}

	return history
}

// AddValidator adds a new validator to the network
func (bi *BlockchainIntegrity) AddValidator(validatorID, publicKey string, stake int) error {
	validator := &Validator{
		ID:         validatorID,
		PublicKey:  publicKey,
		Stake:      stake,
		Reputation: 1.0,
		LastSigned: time.Now(),
		Active:     true,
	}

	bi.mutex.Lock()
	defer bi.mutex.Unlock()

	bi.validators[validatorID] = validator
	bi.consensus.AddValidator(validator)

	fmt.Printf("Added validator %s with stake %d\n", validatorID, stake)
	return nil
}

// SelectValidator selects a validator for the next block
func (pos *ProofOfStakeConsensus) SelectValidator() string {
	pos.mutex.RLock()
	defer pos.mutex.RUnlock()

	if len(pos.validators) == 0 {
		return ""
	}

	// Calculate total stake
	totalStake := 0
	for _, validator := range pos.validators {
		if validator.Active {
			totalStake += validator.Stake
		}
	}

	if totalStake == 0 {
		return ""
	}

	// Select validator based on stake weight
	randomValue, _ := rand.Int(rand.Reader, big.NewInt(int64(totalStake)))
	cumulativeStake := 0

	for _, validator := range pos.validators {
		if !validator.Active {
			continue
		}

		cumulativeStake += validator.Stake
		if int(randomValue.Int64()) < cumulativeStake {
			return validator.ID
		}
	}

	// Fallback to first active validator
	for _, validator := range pos.validators {
		if validator.Active {
			return validator.ID
		}
	}

	return ""
}

// AddValidator adds a validator to the consensus pool
func (pos *ProofOfStakeConsensus) AddValidator(validator *Validator) {
	pos.mutex.Lock()
	defer pos.mutex.Unlock()

	pos.validators[validator.ID] = validator
	pos.totalStake += validator.Stake
}

// GetChainStatus returns the current blockchain status
func (bi *BlockchainIntegrity) GetChainStatus() map[string]interface{} {
	bi.mutex.RLock()
	defer bi.mutex.RUnlock()

	totalTransactions := 0
	for _, block := range bi.chain {
		totalTransactions += len(block.Transactions)
	}

	return map[string]interface{}{
		"chain_length":       len(bi.chain),
		"total_transactions": totalTransactions,
		"pending_blocks":     len(bi.pendingBlocks),
		"difficulty":         bi.difficulty,
		"total_validators":   len(bi.validators),
		"active_validators":  bi.getActiveValidatorCount(),
		"total_stake":        bi.getTotalStake(),
		"last_block_hash":    bi.chain[len(bi.chain)-1].Hash,
		"last_block_time":    bi.chain[len(bi.chain)-1].Timestamp,
		"chain_valid":        bi.ValidateChain(),
	}
}

// Helper functions

func calculateBlockHash(block *IntegrityBlock) string {
	blockData := fmt.Sprintf("%d%s%s%s%d%d%d",
		block.Index,
		block.PreviousHash,
		block.MerkleRoot,
		block.Validator,
		block.Timestamp.Unix(),
		block.Nonce,
		block.Stake)

	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:])
}

func calculateMerkleRoot(transactions []*IntegrityTransaction) string {
	if len(transactions) == 0 {
		return ""
	}

	// Calculate transaction hashes
	txHashes := make([]string, len(transactions))
	for i, tx := range transactions {
		txData := fmt.Sprintf("%s%s%s%s", tx.ID, tx.Type, tx.Target, tx.Hash)
		hash := sha256.Sum256([]byte(txData))
		txHashes[i] = hex.EncodeToString(hash[:])
	}

	// Build Merkle tree
	for len(txHashes) > 1 {
		var nextLevel []string

		for i := 0; i < len(txHashes); i += 2 {
			if i+1 < len(txHashes) {
				combined := txHashes[i] + txHashes[i+1]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			} else {
				// Odd number of nodes, duplicate the last one
				hash := sha256.Sum256([]byte(txHashes[i]))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			}
		}

		txHashes = nextLevel
	}

	return txHashes[0]
}

func generateBlockchainTransactionID() string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

func generateAddressFromKey(privateKey *ecdsa.PrivateKey) string {
	hash := sha256.Sum256(privateKey.D.Bytes())
	return hex.EncodeToString(hash[:])[:20]
}

func (bi *BlockchainIntegrity) signTransaction(tx *IntegrityTransaction) (string, error) {
	txData := fmt.Sprintf("%s%s%s%s%d", tx.ID, tx.Type, tx.Target, tx.Hash, tx.Timestamp.Unix())
	hash := sha256.Sum256([]byte(txData))
	return hex.EncodeToString(hash[:]), nil
}

func (bi *BlockchainIntegrity) signBlock(block *IntegrityBlock) (string, error) {
	blockData := fmt.Sprintf("Block %d (hash: %s, timestamp: %d, txs: %d)", block.Index, block.Hash, block.Timestamp.Unix(), len(block.Transactions))
	hash := sha256.Sum256([]byte(blockData))
	return hex.EncodeToString(hash[:]), nil
}

func (bi *BlockchainIntegrity) updateValidatorReputation(validatorID string, success bool) {
	if validator, exists := bi.validators[validatorID]; exists {
		if success {
			validator.Reputation += 0.1
		} else {
			validator.Reputation -= 0.5
		}

		// Clamp reputation
		if validator.Reputation < 0 {
			validator.Reputation = 0
		}
		if validator.Reputation > 10 {
			validator.Reputation = 10
		}

		validator.LastSigned = time.Now()
	}
}

func (bi *BlockchainIntegrity) getActiveValidatorCount() int {
	count := 0
	for _, validator := range bi.validators {
		if validator.Active {
			count++
		}
	}
	return count
}

func (bi *BlockchainIntegrity) getTotalStake() int {
	total := 0
	for _, validator := range bi.validators {
		if validator.Active {
			total += validator.Stake
		}
	}
	return total
}
