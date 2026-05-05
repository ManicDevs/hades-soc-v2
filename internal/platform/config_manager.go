package platform

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigManager handles hot-swappable configurations
type ConfigManager struct {
	configPath    string
	currentConfig *DistributedDatabaseConfig
	mu            sync.RWMutex
	watchers      []chan *DistributedDatabaseConfig
	lastModified  time.Time
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) (*ConfigManager, error) {
	cm := &ConfigManager{
		configPath: configPath,
		watchers:   make([]chan *DistributedDatabaseConfig, 0),
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load initial configuration
	if err := cm.loadConfig(); err != nil {
		log.Printf("Warning: Failed to load initial config, using default: %v", err)
		cm.currentConfig = DefaultDistributedDatabaseConfig()
	}

	// Start configuration watcher
	go cm.watchConfig()

	return cm, nil
}

// loadConfig loads configuration from file
func (cm *ConfigManager) loadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if it doesn't exist
			defaultConfig := DefaultDistributedDatabaseConfig()
			if err := cm.saveConfig(defaultConfig); err != nil {
				return fmt.Errorf("failed to save default config: %w", err)
			}
			cm.currentConfig = defaultConfig
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config DistributedDatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cm.currentConfig = &config

	// Get file modification time
	if info, err := os.Stat(cm.configPath); err == nil {
		cm.lastModified = info.ModTime()
	}

	return nil
}

// saveConfig saves configuration to file
func (cm *ConfigManager) saveConfig(config *DistributedDatabaseConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// watchConfig watches for configuration changes
func (cm *ConfigManager) watchConfig() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cm.checkConfigChanges()
	}
}

// checkConfigChanges checks if configuration file has changed
func (cm *ConfigManager) checkConfigChanges() {
	info, err := os.Stat(cm.configPath)
	if err != nil {
		return
	}

	if info.ModTime().After(cm.lastModified) {
		if err := cm.loadConfig(); err != nil {
			log.Printf("Error reloading config: %v", err)
			return
		}

		// Notify all watchers
		for _, watcher := range cm.watchers {
			select {
			case watcher <- cm.currentConfig:
			default:
				// Non-blocking send, skip if channel is full
			}
		}

		log.Printf("Configuration reloaded from %s", cm.configPath)
	}
}

// GetCurrentConfig returns the current configuration
func (cm *ConfigManager) GetCurrentConfig() *DistributedDatabaseConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy to avoid race conditions
	configCopy := *cm.currentConfig
	return &configCopy
}

// UpdateConfig updates the current configuration
func (cm *ConfigManager) UpdateConfig(config *DistributedDatabaseConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := cm.saveConfig(config); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}

	cm.currentConfig = config

	// Get file modification time
	if info, err := os.Stat(cm.configPath); err == nil {
		cm.lastModified = info.ModTime()
	}

	// Notify all watchers
	for _, watcher := range cm.watchers {
		select {
		case watcher <- cm.currentConfig:
		default:
		}
	}

	log.Printf("Configuration updated and saved to %s", cm.configPath)
	return nil
}

// AddWatcher adds a configuration change watcher
func (cm *ConfigManager) AddWatcher() chan *DistributedDatabaseConfig {
	watcher := make(chan *DistributedDatabaseConfig, 1)
	cm.mu.Lock()
	cm.watchers = append(cm.watchers, watcher)
	cm.mu.Unlock()
	return watcher
}

// RemoveWatcher removes a configuration change watcher
func (cm *ConfigManager) RemoveWatcher(watcher chan *DistributedDatabaseConfig) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, w := range cm.watchers {
		if w == watcher {
			cm.watchers = append(cm.watchers[:i], cm.watchers[i+1:]...)
			close(watcher)
			break
		}
	}
}

// LoadEnvironmentConfig loads configuration based on environment
func (cm *ConfigManager) LoadEnvironmentConfig(env string) error {
	configPath := filepath.Join(filepath.Dir(cm.configPath), fmt.Sprintf("%s.json", env))

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("environment config %s does not exist", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read environment config: %w", err)
	}

	var config DistributedDatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal environment config: %w", err)
	}

	return cm.UpdateConfig(&config)
}

// GetEnvironmentConfigPath returns the path for an environment-specific config
func (cm *ConfigManager) GetEnvironmentConfigPath(env string) string {
	return filepath.Join(filepath.Dir(cm.configPath), fmt.Sprintf("%s.json", env))
}

// ValidateConfig validates a configuration
func (cm *ConfigManager) ValidateConfig(config *DistributedDatabaseConfig) error {
	if config.Type == "" {
		return fmt.Errorf("database type cannot be empty")
	}

	if config.Type == "postgresql" || config.Type == "mysql" {
		if config.Primary.Host == "" {
			return fmt.Errorf("primary host cannot be empty for %s", config.Type)
		}
		if config.Primary.Port == 0 {
			return fmt.Errorf("primary port cannot be empty for %s", config.Type)
		}
		if config.Primary.Database == "" {
			return fmt.Errorf("primary database cannot be empty for %s", config.Type)
		}
		if config.Primary.Username == "" {
			return fmt.Errorf("primary username cannot be empty for %s", config.Type)
		}
	}

	if config.HealthCheckInterval <= 0 {
		return fmt.Errorf("health check interval must be positive")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if config.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection timeout must be positive")
	}

	return nil
}

// Global config manager instance
var globalConfigManager *ConfigManager
var configManagerOnce sync.Once

// GetConfigManager returns the global configuration manager
func GetConfigManager() *ConfigManager {
	configManagerOnce.Do(func() {
		configPath := filepath.Join(os.Getenv("HOME"), ".hades", "database", "config.json")
		var err error
		globalConfigManager, err = NewConfigManager(configPath)
		if err != nil {
			log.Printf("Failed to create config manager: %v", err)
			// Create a fallback config manager
			fallbackManager, err := NewConfigManager("/tmp/hades_config.json")
			if err != nil {
				log.Printf("Failed to create fallback config manager: %v", err)
			} else {
				globalConfigManager = fallbackManager
			}
		}
	})
	return globalConfigManager
}
