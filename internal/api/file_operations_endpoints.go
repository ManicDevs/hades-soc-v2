package api

import (
	"encoding/json"
	"log"
	"net/http"

	"hades-v2/internal/ai"
)

// FileOperationsEndpoints provides file operation API endpoints
type FileOperationsEndpoints struct {
	fileManager *ai.FileOperationManager
	router      *http.ServeMux
}

// NewFileOperationsEndpoints creates new file operations endpoints
func NewFileOperationsEndpoints(baseDir string) (*FileOperationsEndpoints, error) {
	fileManager := ai.NewFileOperationManager(baseDir)

	endpoints := &FileOperationsEndpoints{
		fileManager: fileManager,
		router:      http.NewServeMux(),
	}

	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers file operation API routes
func (foe *FileOperationsEndpoints) registerRoutes() {
	foe.router.HandleFunc("/read", foe.handleReadFile)
	foe.router.HandleFunc("/write", foe.handleWriteFile)
	foe.router.HandleFunc("/edit", foe.handleEditFile)
	foe.router.HandleFunc("/create", foe.handleCreateFile)
	foe.router.HandleFunc("/delete", foe.handleDeleteFile)
	foe.router.HandleFunc("/list", foe.handleListDirectory)
	foe.router.HandleFunc("/operations", foe.handleGetOperations)
	foe.router.HandleFunc("/operations/", foe.handleGetOperation)
	foe.router.HandleFunc("/status", foe.handleStatus)
}

// handleReadFile handles file read requests
func (foe *FileOperationsEndpoints) handleReadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	op, err := foe.fileManager.ReadFile(request.Path)
	if err != nil {
		log.Printf("File read failed: %v", err)
		WriteJSONResponse(w, op)
		return
	}

	WriteJSONResponse(w, op)
}

// handleWriteFile handles file write requests
func (foe *FileOperationsEndpoints) handleWriteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Path      string `json:"path"`
		Content   string `json:"content"`
		ManualACK bool   `json:"manual_ack"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	op, err := foe.fileManager.WriteFile(request.Path, request.Content, request.ManualACK)
	if err != nil {
		log.Printf("File write failed: %v", err)
		WriteJSONResponse(w, op)
		return
	}

	WriteJSONResponse(w, op)
}

// handleEditFile handles AI-assisted file edit requests
func (foe *FileOperationsEndpoints) handleEditFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request ai.FileEditRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	if request.Instruction == "" {
		http.Error(w, "Instruction is required", http.StatusBadRequest)
		return
	}

	op, err := foe.fileManager.EditFile(request)
	if err != nil {
		log.Printf("File edit failed: %v", err)
		WriteJSONResponse(w, op)
		return
	}

	WriteJSONResponse(w, op)
}

// handleCreateFile handles file creation requests
func (foe *FileOperationsEndpoints) handleCreateFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Path      string `json:"path"`
		Content   string `json:"content"`
		ManualACK bool   `json:"manual_ack"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	op, err := foe.fileManager.CreateFile(request.Path, request.Content, request.ManualACK)
	if err != nil {
		log.Printf("File creation failed: %v", err)
		WriteJSONResponse(w, op)
		return
	}

	WriteJSONResponse(w, op)
}

// handleDeleteFile handles file deletion requests
func (foe *FileOperationsEndpoints) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Path      string `json:"path"`
		ManualACK bool   `json:"manual_ack"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	op, err := foe.fileManager.DeleteFile(request.Path, request.ManualACK)
	if err != nil {
		log.Printf("File deletion failed: %v", err)
		WriteJSONResponse(w, op)
		return
	}

	WriteJSONResponse(w, op)
}

// handleListDirectory handles directory listing requests
func (foe *FileOperationsEndpoints) handleListDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	op, err := foe.fileManager.ListDirectory(request.Path)
	if err != nil {
		log.Printf("Directory listing failed: %v", err)
		WriteJSONResponse(w, op)
		return
	}

	WriteJSONResponse(w, op)
}

// handleGetOperations handles retrieving all operations
func (foe *FileOperationsEndpoints) handleGetOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ops := foe.fileManager.GetOperations()

	response := map[string]interface{}{
		"operations": ops,
		"count":      len(ops),
	}

	WriteJSONResponse(w, response)
}

// handleGetOperation handles retrieving a specific operation
func (foe *FileOperationsEndpoints) handleGetOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract operation ID from URL path
	// URL format: /operations/{id}
	id := r.URL.Path[len("/operations/"):]
	if id == "" {
		http.Error(w, "Operation ID is required", http.StatusBadRequest)
		return
	}

	op, err := foe.fileManager.GetOperation(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	WriteJSONResponse(w, op)
}

// handleStatus handles status requests
func (foe *FileOperationsEndpoints) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"destructive_count": foe.fileManager.GetDestructiveCount(),
		"destructive_limit": foe.fileManager.GetDestructiveLimit(),
		"status":            "active",
	}

	WriteJSONResponse(w, response)
}

// GetRouter returns the file operations router
func (foe *FileOperationsEndpoints) GetRouter() *http.ServeMux {
	return foe.router
}

// GetFileManager returns the file operation manager
func (foe *FileOperationsEndpoints) GetFileManager() *ai.FileOperationManager {
	return foe.fileManager
}
