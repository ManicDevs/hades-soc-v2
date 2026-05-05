package sdk

import (
	"context"
	"fmt"
)

// ModuleStatus represents the operational state of a module
type ModuleStatus string

const (
	StatusIdle      ModuleStatus = "idle"
	StatusRunning   ModuleStatus = "running"
	StatusCompleted ModuleStatus = "completed"
	StatusFailed    ModuleStatus = "failed"
	StatusCancelled ModuleStatus = "cancelled"
)

// ModuleCategory represents the functional classification of a module
type ModuleCategory string

const (
	CategoryReconnaissance   ModuleCategory = "reconnaissance"
	CategoryScanning         ModuleCategory = "scanning"
	CategoryExploitation     ModuleCategory = "exploitation"
	CategoryPostExploitation ModuleCategory = "post_exploitation"
	CategoryPersistence      ModuleCategory = "persistence"
	CategoryEvasion          ModuleCategory = "evasion"
	CategoryReporting        ModuleCategory = "reporting"
)

// Module defines the contract for all Hades modules
type Module interface {
	// Execute runs the module with the provided context
	Execute(ctx context.Context) error

	// Name returns the unique identifier for this module
	Name() string

	// Category returns the functional classification
	Category() ModuleCategory

	// Description provides a brief summary of module purpose
	Description() string

	// Status returns the current operational state
	Status() ModuleStatus

	// SetStatus updates the module's operational state
	SetStatus(status ModuleStatus)
}

// BaseModule provides common functionality for all modules
type BaseModule struct {
	name        string
	category    ModuleCategory
	description string
	status      ModuleStatus
}

// NewBaseModule creates a new base module instance
func NewBaseModule(name, description string, category ModuleCategory) *BaseModule {
	return &BaseModule{
		name:        name,
		category:    category,
		description: description,
		status:      StatusIdle,
	}
}

// Name returns the module's unique identifier
func (bm *BaseModule) Name() string {
	return bm.name
}

// Category returns the module's functional classification
func (bm *BaseModule) Category() ModuleCategory {
	return bm.category
}

// Description returns the module's purpose summary
func (bm *BaseModule) Description() string {
	return bm.description
}

// Status returns the current operational state
func (bm *BaseModule) Status() ModuleStatus {
	return bm.status
}

// SetStatus updates the module's operational state
func (bm *BaseModule) SetStatus(status ModuleStatus) {
	bm.status = status
}

// ModuleRegistry manages the collection of available modules
type ModuleRegistry struct {
	modules map[string]Module
}

// NewModuleRegistry creates an empty module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[string]Module),
	}
}

// Register adds a module to the registry
func (mr *ModuleRegistry) Register(module Module) error {
	if module == nil {
		return fmt.Errorf("hades.sdk: cannot register nil module")
	}

	name := module.Name()
	if name == "" {
		return fmt.Errorf("hades.sdk: module name cannot be empty")
	}

	if _, exists := mr.modules[name]; exists {
		return fmt.Errorf("hades.sdk: module '%s' already registered", name)
	}

	mr.modules[name] = module
	return nil
}

// Get retrieves a module by name
func (mr *ModuleRegistry) Get(name string) (Module, error) {
	module, exists := mr.modules[name]
	if !exists {
		return nil, fmt.Errorf("hades.sdk: module '%s' not found", name)
	}
	return module, nil
}

// List returns all registered module names
func (mr *ModuleRegistry) List() []string {
	names := make([]string, 0, len(mr.modules))
	for name := range mr.modules {
		names = append(names, name)
	}
	return names
}

// ListByCategory returns modules filtered by category
func (mr *ModuleRegistry) ListByCategory(category ModuleCategory) []Module {
	modules := make([]Module, 0)
	for _, module := range mr.modules {
		if module.Category() == category {
			modules = append(modules, module)
		}
	}
	return modules
}
