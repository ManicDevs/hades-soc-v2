package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileOperation represents a file operation request
type FileOperation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // read, write, edit, create, delete, list
	Path        string                 `json:"path"`
	Content     string                 `json:"content,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	ManualACK   bool                   `json:"manual_ack"` // Required for destructive operations
	Status      string                 `json:"status"`
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
}

// FileEditRequest represents an AI-assisted file edit request
type FileEditRequest struct {
	Path        string `json:"path"`
	Instruction string `json:"instruction"` // Natural language instruction for the edit
	OldContent  string `json:"old_content,omitempty"`
	ManualACK   bool   `json:"manual_ack"`
}

// FileInfo represents file metadata
type FileInfo struct {
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	Size         int64       `json:"size"`
	Mode         fs.FileMode `json:"mode"`
	ModTime      time.Time   `json:"mod_time"`
	IsDir        bool        `json:"is_dir"`
	IsExecutable bool        `json:"is_executable"`
}

// FileOperationManager manages file operations with AI assistance
type FileOperationManager struct {
	operations map[string]*FileOperation
	mu         sync.RWMutex
	llm        *LLMService
	baseDir    string // Base directory for safe operations
	// Safety limits
	maxFileSize      int64     // Max file size in bytes
	blockedPaths     []string  // Paths that are blocked
	destructiveCount int       // Counter for destructive operations
	destructiveLimit int       // Max destructive operations per hour
	lastDestructive  time.Time // Last destructive operation timestamp
}

// NewFileOperationManager creates a new file operation manager
func NewFileOperationManager(baseDir string) *FileOperationManager {
	return &FileOperationManager{
		operations:  make(map[string]*FileOperation),
		llm:         GetLLMService(),
		baseDir:     baseDir,
		maxFileSize: 10 * 1024 * 1024, // 10MB default
		blockedPaths: []string{
			"/etc/passwd",
			"/etc/shadow",
			"/etc/sudoers",
			"/bin/",
			"/sbin/",
			"/usr/bin/",
			"/usr/sbin/",
			"/boot/",
			"/sys/",
			"/proc/",
		},
		destructiveLimit: 5, // Per safety governor requirements
	}
}

// ValidatePath checks if a path is safe to operate on
func (fom *FileOperationManager) ValidatePath(path string) error {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if within base directory
	if fom.baseDir != "" {
		relPath, err := filepath.Rel(fom.baseDir, absPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("path outside allowed directory: %s", path)
		}
	}

	// Check blocked paths
	for _, blocked := range fom.blockedPaths {
		if strings.HasPrefix(absPath, blocked) {
			return fmt.Errorf("path is blocked for security: %s", path)
		}
	}

	return nil
}

// CheckDestructiveLimit checks if destructive operation limit has been reached
func (fom *FileOperationManager) CheckDestructiveLimit() error {
	fom.mu.Lock()
	defer fom.mu.Unlock()

	now := time.Now()
	// Reset counter if more than an hour has passed
	if now.Sub(fom.lastDestructive) > time.Hour {
		fom.destructiveCount = 0
	}

	if fom.destructiveCount >= fom.destructiveLimit {
		return fmt.Errorf("destructive operation limit reached (%d/hour). Manual intervention required", fom.destructiveLimit)
	}

	return nil
}

// IncrementDestructiveCount increments the destructive operation counter
func (fom *FileOperationManager) IncrementDestructiveCount() {
	fom.mu.Lock()
	defer fom.mu.Unlock()

	fom.destructiveCount++
	fom.lastDestructive = time.Now()
}

// ReadFile reads a file's contents
func (fom *FileOperationManager) ReadFile(path string) (*FileOperation, error) {
	op := &FileOperation{
		ID:        fmt.Sprintf("op-%d", time.Now().UnixNano()),
		Type:      "read",
		Path:      path,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	// Validate path
	if err := fom.ValidatePath(path); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Check file size
	if int64(len(content)) > fom.maxFileSize {
		op.Status = "failed"
		op.Error = fmt.Sprintf("file too large: %d bytes (max %d)", len(content), fom.maxFileSize)
		op.CompletedAt = time.Now()
		return op, errors.New(op.Error)
	}

	op.Status = "completed"
	op.Result = string(content)
	op.CompletedAt = time.Now()

	fom.mu.Lock()
	fom.operations[op.ID] = op
	fom.mu.Unlock()

	return op, nil
}

// WriteFile writes content to a file
func (fom *FileOperationManager) WriteFile(path, content string, manualACK bool) (*FileOperation, error) {
	op := &FileOperation{
		ID:        fmt.Sprintf("op-%d", time.Now().UnixNano()),
		Type:      "write",
		Path:      path,
		Content:   content,
		ManualACK: manualACK,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	// Require manual ACK for write operations
	if !manualACK {
		op.Status = "failed"
		op.Error = "manual ACK required for write operations"
		op.CompletedAt = time.Now()
		return op, errors.New(op.Error)
	}

	// Check destructive limit
	if err := fom.CheckDestructiveLimit(); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Validate path
	if err := fom.ValidatePath(path); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	fom.IncrementDestructiveCount()

	op.Status = "completed"
	op.Result = fmt.Sprintf("wrote %d bytes to %s", len(content), path)
	op.CompletedAt = time.Now()

	fom.mu.Lock()
	fom.operations[op.ID] = op
	fom.mu.Unlock()

	return op, nil
}

// EditFile uses AI to intelligently edit a file
func (fom *FileOperationManager) EditFile(req FileEditRequest) (*FileOperation, error) {
	op := &FileOperation{
		ID:   fmt.Sprintf("op-%d", time.Now().UnixNano()),
		Type: "edit",
		Path: req.Path,
		Context: map[string]interface{}{
			"instruction": req.Instruction,
		},
		ManualACK: req.ManualACK,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	// Require manual ACK for edit operations
	if !req.ManualACK {
		op.Status = "failed"
		op.Error = "manual ACK required for edit operations"
		op.CompletedAt = time.Now()
		return op, errors.New(op.Error)
	}

	// Check destructive limit
	if err := fom.CheckDestructiveLimit(); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Validate path
	if err := fom.ValidatePath(req.Path); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Read current file content
	currentContent, err := os.ReadFile(req.Path)
	if err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Check file size
	if int64(len(currentContent)) > fom.maxFileSize {
		op.Status = "failed"
		op.Error = fmt.Sprintf("file too large for AI editing: %d bytes (max %d)", len(currentContent), fom.maxFileSize)
		op.CompletedAt = time.Now()
		return op, errors.New(op.Error)
	}

	// Use LLM to generate the edited content
	prompt := fmt.Sprintf(`You are an expert code editor. Edit the following file based on the instruction.

File Path: %s

Current Content:
%s

Instruction: %s

Provide ONLY the complete edited file content. Do not include explanations or markdown code blocks. Output the raw file content only.`,
		req.Path, string(currentContent), req.Instruction)

	resp, err := fom.llm.Query(prompt)
	if err != nil {
		op.Status = "failed"
		op.Error = fmt.Sprintf("LLM query failed: %v", err)
		op.CompletedAt = time.Now()
		return op, err
	}

	// Clean up the response (remove markdown code blocks if present)
	editedContent := resp.Response
	editedContent = strings.TrimPrefix(editedContent, "```")
	editedContent = strings.TrimSuffix(editedContent, "```")
	editedContent = strings.TrimSpace(editedContent)

	// Write the edited content
	if err := os.WriteFile(req.Path, []byte(editedContent), 0644); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	fom.IncrementDestructiveCount()

	op.Status = "completed"
	op.Result = fmt.Sprintf("edited file %s using AI", req.Path)
	op.Context["original_size"] = len(currentContent)
	op.Context["edited_size"] = len(editedContent)
	op.CompletedAt = time.Now()

	fom.mu.Lock()
	fom.operations[op.ID] = op
	fom.mu.Unlock()

	return op, nil
}

// CreateFile creates a new file
func (fom *FileOperationManager) CreateFile(path, content string, manualACK bool) (*FileOperation, error) {
	op := &FileOperation{
		ID:        fmt.Sprintf("op-%d", time.Now().UnixNano()),
		Type:      "create",
		Path:      path,
		Content:   content,
		ManualACK: manualACK,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	// Require manual ACK for create operations
	if !manualACK {
		op.Status = "failed"
		op.Error = "manual ACK required for create operations"
		op.CompletedAt = time.Now()
		return op, errors.New(op.Error)
	}

	// Check destructive limit
	if err := fom.CheckDestructiveLimit(); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Validate path
	if err := fom.ValidatePath(path); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		op.Status = "failed"
		op.Error = "file already exists"
		op.CompletedAt = time.Now()
		return op, errors.New(op.Error)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Create file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	fom.IncrementDestructiveCount()

	op.Status = "completed"
	op.Result = fmt.Sprintf("created file %s", path)
	op.CompletedAt = time.Now()

	fom.mu.Lock()
	fom.operations[op.ID] = op
	fom.mu.Unlock()

	return op, nil
}

// DeleteFile deletes a file
func (fom *FileOperationManager) DeleteFile(path string, manualACK bool) (*FileOperation, error) {
	op := &FileOperation{
		ID:        fmt.Sprintf("op-%d", time.Now().UnixNano()),
		Type:      "delete",
		Path:      path,
		ManualACK: manualACK,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	// Require manual ACK for delete operations
	if !manualACK {
		op.Status = "failed"
		op.Error = "manual ACK required for delete operations"
		op.CompletedAt = time.Now()
		return op, errors.New(op.Error)
	}

	// Check destructive limit
	if err := fom.CheckDestructiveLimit(); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Validate path
	if err := fom.ValidatePath(path); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Delete file
	if err := os.Remove(path); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	fom.IncrementDestructiveCount()

	op.Status = "completed"
	op.Result = fmt.Sprintf("deleted file %s", path)
	op.CompletedAt = time.Now()

	fom.mu.Lock()
	fom.operations[op.ID] = op
	fom.mu.Unlock()

	return op, nil
}

// ListDirectory lists files in a directory
func (fom *FileOperationManager) ListDirectory(path string) (*FileOperation, error) {
	op := &FileOperation{
		ID:        fmt.Sprintf("op-%d", time.Now().UnixNano()),
		Type:      "list",
		Path:      path,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	// Validate path
	if err := fom.ValidatePath(path); err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Read directory
	entries, err := os.ReadDir(path)
	if err != nil {
		op.Status = "failed"
		op.Error = err.Error()
		op.CompletedAt = time.Now()
		return op, err
	}

	// Build file info list
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		files = append(files, FileInfo{
			Name:         entry.Name(),
			Path:         fullPath,
			Size:         info.Size(),
			Mode:         info.Mode(),
			ModTime:      info.ModTime(),
			IsDir:        entry.IsDir(),
			IsExecutable: info.Mode().Perm()&0111 != 0,
		})
	}

	// Convert to JSON for result
	resultJSON, _ := json.Marshal(files)
	op.Status = "completed"
	op.Result = string(resultJSON)
	op.CompletedAt = time.Now()

	fom.mu.Lock()
	fom.operations[op.ID] = op
	fom.mu.Unlock()

	return op, nil
}

// GetOperation retrieves an operation by ID
func (fom *FileOperationManager) GetOperation(id string) (*FileOperation, error) {
	fom.mu.RLock()
	defer fom.mu.RUnlock()

	op, exists := fom.operations[id]
	if !exists {
		return nil, fmt.Errorf("operation not found: %s", id)
	}

	return op, nil
}

// GetOperations retrieves all operations
func (fom *FileOperationManager) GetOperations() []*FileOperation {
	fom.mu.RLock()
	defer fom.mu.RUnlock()

	ops := make([]*FileOperation, 0, len(fom.operations))
	for _, op := range fom.operations {
		ops = append(ops, op)
	}

	return ops
}

// GetDestructiveCount returns the current destructive operation count
func (fom *FileOperationManager) GetDestructiveCount() int {
	fom.mu.RLock()
	defer fom.mu.RUnlock()

	return fom.destructiveCount
}

// GetDestructiveLimit returns the destructive operation limit
func (fom *FileOperationManager) GetDestructiveLimit() int {
	return fom.destructiveLimit
}

// SetBaseDirectory sets the base directory for safe operations
func (fom *FileOperationManager) SetBaseDirectory(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	fom.mu.Lock()
	defer fom.mu.Unlock()

	fom.baseDir = absDir
	return nil
}
