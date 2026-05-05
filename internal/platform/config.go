package platform

import (
	"os"
	"strconv"
	"strings"
)

// EnvironmentConfig holds environment-specific configuration
type EnvironmentConfig struct {
	Environment         string
	NodeEnv             string
	WebPort             int
	APIPort             int
	AuthPort            int
	DashboardPort       int
	EnableDevAccess     bool
	RequireAuth         bool
	EnableMFA           bool
	SessionTimeout      int
	LogLevel            string
	EnableDebugLogging  bool
	EnableRegistration  bool
	EnablePasswordReset bool
	EnableAuditLogging  bool
	CORSAllowedOrigins  []string
}

// LoadEnvironmentConfig loads configuration from environment variables
func LoadEnvironmentConfig() *EnvironmentConfig {
	config := &EnvironmentConfig{
		Environment:         getEnv("ENVIRONMENT", "development"),
		NodeEnv:             getEnv("NODE_ENV", "development"),
		WebPort:             getEnvInt("WEB_PORT", 3000),
		APIPort:             getEnvInt("API_PORT", 8443),
		AuthPort:            getEnvInt("AUTH_PORT", 8444),
		DashboardPort:       getEnvInt("DASHBOARD_PORT", 8445),
		EnableDevAccess:     getEnvBool("ENABLE_DEV_ACCESS", true),
		RequireAuth:         getEnvBool("REQUIRE_AUTHENTICATION", false),
		EnableMFA:           getEnvBool("ENABLE_MFA", false),
		SessionTimeout:      getEnvInt("SESSION_TIMEOUT", 7200),
		LogLevel:            getEnv("LOG_LEVEL", "debug"),
		EnableDebugLogging:  getEnvBool("ENABLE_DEBUG_LOGGING", true),
		EnableRegistration:  getEnvBool("ENABLE_REGISTRATION", true),
		EnablePasswordReset: getEnvBool("ENABLE_PASSWORD_RESET", true),
		EnableAuditLogging:  getEnvBool("ENABLE_AUDIT_LOGGING", true),
	}

	// Parse CORS origins
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	config.CORSAllowedOrigins = strings.Split(corsOrigins, ",")

	// Environment-specific overrides
	switch strings.ToLower(config.Environment) {
	case "production":
		config.EnableDevAccess = false
		config.RequireAuth = true
		config.EnableMFA = true
		config.SessionTimeout = 3600
		config.LogLevel = "error"
		config.EnableDebugLogging = false
		config.EnableRegistration = false
	case "staging":
		config.RequireAuth = true
		config.EnableMFA = true
		config.SessionTimeout = 1800
		config.LogLevel = "info"
		config.EnableRegistration = false
	case "qa":
		config.RequireAuth = true
		config.EnableMFA = false
		config.SessionTimeout = 1200
		config.LogLevel = "debug"
	case "testing":
		config.RequireAuth = false
		config.EnableMFA = false
		config.SessionTimeout = 999999
		config.LogLevel = "debug"
		config.EnableAuditLogging = false
	}

	return config
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// IsProduction checks if running in production environment
func (c *EnvironmentConfig) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

// IsDevelopment checks if running in development environment
func (c *EnvironmentConfig) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

// IsNonProduction checks if running in non-production environment
func (c *EnvironmentConfig) IsNonProduction() bool {
	return !c.IsProduction()
}
