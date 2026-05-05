package versioning

import (
	"net/http"
	"time"
)

// Integration with existing server
type ServerIntegration struct {
	manager *VersionManager
	server  interface{} // Will be the actual Server type
}

// NewServerIntegration creates a new server integration
func NewServerIntegration(manager *VersionManager) *ServerIntegration {
	return &ServerIntegration{
		manager: manager,
	}
}

// Middleware returns the versioning middleware for integration
func (si *ServerIntegration) Middleware(next http.Handler) http.Handler {
	return si.manager.Middleware(next)
}

// SetupVersionedRoutes sets up versioned routes on the existing server
func (si *ServerIntegration) SetupVersionedRoutes(router interface{}) {
	// This will be implemented to work with mux.Router
	// For now, we'll add the routes directly to the server
}

// GetVersionManager returns the version manager
func (si *ServerIntegration) GetVersionManager() *VersionManager {
	return si.manager
}

// Default configuration for the versioning system
func DefaultConfig() ManagerConfig {
	return ManagerConfig{
		DefaultVersion:    "v2",
		PreferredVersion:  "v2",
		SupportedVersions: []string{"v1", "v2", "v3"},
		DeprecationPolicy: DeprecationPolicy{
			Name:         "Standard Policy",
			Deprecation:  6 * 30 * 24 * time.Hour,  // 6 months
			Sunset:       12 * 30 * 24 * time.Hour, // 12 months
			Notification: "email, banner, api-header",
		},
		VersionHeaders: map[string]string{
			"documentation": "X-API-Documentation",
			"support":       "X-API-Support",
			"feedback":      "X-API-Feedback",
		},
	}
}
