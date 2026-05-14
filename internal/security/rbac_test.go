package security

import (
	"testing"
	"time"
)

func TestRBACManagerCreation(t *testing.T) {
	manager := NewRBACManager()
	if manager == nil {
		t.Fatal("NewRBACManager returned nil")
	}
}

func TestRole(t *testing.T) {
	role := Role{
		ID:          "role-001",
		Name:        "Admin",
		Description: "Administrator role",
		Permissions: []string{"read", "write", "delete"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if role.ID != "role-001" {
		t.Errorf("Expected ID 'role-001', got '%s'", role.ID)
	}
	if role.Name != "Admin" {
		t.Errorf("Expected Name 'Admin', got '%s'", role.Name)
	}
	if len(role.Permissions) != 3 {
		t.Errorf("Expected 3 permissions, got %d", len(role.Permissions))
	}
}

func TestPermission(t *testing.T) {
	perm := Permission{
		ID:          "perm-001",
		Name:        "read_users",
		Description: "Permission to read user data",
		Resource:    "users",
		Action:      "read",
		CreatedAt:   time.Now(),
	}

	if perm.ID != "perm-001" {
		t.Errorf("Expected ID 'perm-001', got '%s'", perm.ID)
	}
	if perm.Resource != "users" {
		t.Errorf("Expected Resource 'users', got '%s'", perm.Resource)
	}
	if perm.Action != "read" {
		t.Errorf("Expected Action 'read', got '%s'", perm.Action)
	}
}
