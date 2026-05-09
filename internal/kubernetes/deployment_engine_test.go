package kubernetes

import (
	"testing"
	"time"
)

func TestDeploymentEngineCreation(t *testing.T) {
	engine, err := NewDeploymentEngine(nil)
	if err != nil {
		t.Fatalf("NewDeploymentEngine returned error: %v", err)
	}
	if engine == nil {
		t.Fatal("NewDeploymentEngine returned nil")
	}
}

func TestCluster(t *testing.T) {
	cluster := Cluster{
		ID:        "cluster-001",
		Name:      "prod-cluster",
		Version:   "1.28",
		Region:    "us-west-2",
		Provider:  "aws",
		Status:    "ready",
		Nodes:     make([]*Node, 0),
		Resources: &ClusterResources{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if cluster.ID != "cluster-001" {
		t.Errorf("Expected ID 'cluster-001', got '%s'", cluster.ID)
	}
	if cluster.Name != "prod-cluster" {
		t.Errorf("Expected Name 'prod-cluster', got '%s'", cluster.Name)
	}
}

func TestNode(t *testing.T) {
	node := Node{
		ID:     "node-001",
		Name:   "worker-1",
		Status: "ready",
	}

	if node.ID != "node-001" {
		t.Errorf("Expected ID 'node-001', got '%s'", node.ID)
	}
	if node.Status != "ready" {
		t.Errorf("Expected Status 'ready', got '%s'", node.Status)
	}
}

func TestDeployment(t *testing.T) {
	deployment := Deployment{
		ID:          "deploy-001",
		Name:        "hades-api",
		Namespace:   "hades",
		Replicas:    3,
		ReadyReplicas: 3,
		Image:       "hades:latest",
		Status:      "running",
	}

	if deployment.ID != "deploy-001" {
		t.Errorf("Expected ID 'deploy-001', got '%s'", deployment.ID)
	}
}

func TestService(t *testing.T) {
	service := Service{
		ID:        "svc-001",
		Name:      "hades-svc",
		Namespace: "hades",
		Type:      "ClusterIP",
	}

	if service.ID != "svc-001" {
		t.Errorf("Expected ID 'svc-001', got '%s'", service.ID)
	}
}