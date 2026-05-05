package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hades-v2/internal/database"
	"hades-v2/internal/kubernetes"
)

// KubernetesEndpoints provides Kubernetes API endpoints
type KubernetesEndpoints struct {
	deploymentEngine *kubernetes.DeploymentEngine
	router           *http.ServeMux
}

// NewKubernetesEndpoints creates new Kubernetes endpoints
func NewKubernetesEndpoints(db interface{}) (*KubernetesEndpoints, error) {
	// Create deployment engine
	deploymentEngine, err := kubernetes.NewDeploymentEngine(db.(database.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment engine: %w", err)
	}

	endpoints := &KubernetesEndpoints{
		deploymentEngine: deploymentEngine,
		router:           http.NewServeMux(),
	}

	// Register Kubernetes routes
	endpoints.registerRoutes()

	return endpoints, nil
}

// registerRoutes registers Kubernetes API routes
func (ke *KubernetesEndpoints) registerRoutes() {
	ke.router.HandleFunc("/api/v2/kubernetes/clusters", ke.handleGetClusters)
	ke.router.HandleFunc("/api/v2/kubernetes/deployments", ke.handleDeployments)
	ke.router.HandleFunc("/api/v2/kubernetes/scale", ke.handleScaleDeployment)
	ke.router.HandleFunc("/api/v2/kubernetes/services", ke.handleGetServices)
	ke.router.HandleFunc("/api/v2/kubernetes/autoscalers", ke.handleGetAutoscalers)
	ke.router.HandleFunc("/api/v2/kubernetes/status", ke.handleGetStatus)
}

// handleGetClusters handles getting clusters
func (ke *KubernetesEndpoints) handleGetClusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clusters := ke.deploymentEngine.GetClusters()

	response := map[string]interface{}{
		"clusters":  clusters,
		"count":     len(clusters),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleDeployments handles deployment requests (GET and POST)
func (ke *KubernetesEndpoints) handleDeployments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ke.handleGetDeployments(w, r)
	case http.MethodPost:
		ke.handleDeployApplication(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetDeployments handles getting deployments
func (ke *KubernetesEndpoints) handleGetDeployments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	deployments := ke.deploymentEngine.GetDeployments()

	response := map[string]interface{}{
		"deployments": deployments,
		"count":       len(deployments),
		"timestamp":   time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleDeployApplication handles deploying applications
func (ke *KubernetesEndpoints) handleDeployApplication(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var deployment kubernetes.Deployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if deployment.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if deployment.Image == "" {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}

	// Deploy application
	err := ke.deploymentEngine.DeployApplication(r.Context(), &deployment)
	if err != nil {
		http.Error(w, "Failed to deploy application", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"deployment_id": deployment.ID,
		"timestamp":     time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleScaleDeployment handles scaling deployments
func (ke *KubernetesEndpoints) handleScaleDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		DeploymentID string `json:"deployment_id"`
		Replicas     int32  `json:"replicas"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if request.DeploymentID == "" {
		http.Error(w, "Deployment ID is required", http.StatusBadRequest)
		return
	}
	if request.Replicas < 0 {
		http.Error(w, "Replicas must be non-negative", http.StatusBadRequest)
		return
	}

	// Scale deployment
	err := ke.deploymentEngine.ScaleDeployment(r.Context(), request.DeploymentID, request.Replicas)
	if err != nil {
		http.Error(w, "Failed to scale deployment", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"deployment_id": request.DeploymentID,
		"replicas":      request.Replicas,
		"timestamp":     time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetServices handles getting services
func (ke *KubernetesEndpoints) handleGetServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services := ke.deploymentEngine.GetServices()

	response := map[string]interface{}{
		"services":  services,
		"count":     len(services),
		"timestamp": time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetAutoscalers handles getting autoscalers
func (ke *KubernetesEndpoints) handleGetAutoscalers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	autoscalers := ke.deploymentEngine.GetAutoscalers()

	response := map[string]interface{}{
		"autoscalers": autoscalers,
		"count":       len(autoscalers),
		"timestamp":   time.Now(),
	}

	WriteJSONResponse(w, response)
}

// handleGetStatus handles getting Kubernetes status
func (ke *KubernetesEndpoints) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := ke.deploymentEngine.GetEngineStatus()

	WriteJSONResponse(w, status)
}

// GetRouter returns the Kubernetes endpoints router
func (ke *KubernetesEndpoints) GetRouter() *http.ServeMux {
	return ke.router
}

// GetDeploymentEngine returns the deployment engine
func (ke *KubernetesEndpoints) GetDeploymentEngine() *kubernetes.DeploymentEngine {
	return ke.deploymentEngine
}
