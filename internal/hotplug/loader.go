package hotplug

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"plugin"
	"sync"
	"time"

	"hades-v2/internal/anti_analysis"
)

type DynamicModule interface {
	Name() string
	Version() string
	Initialize() error
	Execute(args []string) error
	Unload() error
}

type LoadedModule struct {
	Name     string
	Version  string
	Path     string
	Checksum string
	LoadedAt time.Time
	Module   DynamicModule
	plugin   *plugin.Plugin
}

type HotplugLoader struct {
	modules    map[string]*LoadedModule
	mu         sync.RWMutex
	watchPaths []string
	enabled    bool

	antiAnalysis *anti_analysis.AntiAnalysisManager
}

var (
	globalLoader *HotplugLoader
	loaderOnce   sync.Once
)

func NewHotplugLoader() *HotplugLoader {
	return &HotplugLoader{
		modules:      make(map[string]*LoadedModule),
		watchPaths:   make([]string, 0),
		enabled:      true,
		antiAnalysis: anti_analysis.GetGlobalAntiAnalysisManager(),
	}
}

func GetGlobalLoader() *HotplugLoader {
	loaderOnce.Do(func() {
		globalLoader = NewHotplugLoader()
	})
	return globalLoader
}

func (hl *HotplugLoader) LoadModule(path string) (*LoadedModule, error) {
	hl.mu.Lock()
	defer hl.mu.Unlock()

	if !hl.enabled {
		return nil, fmt.Errorf("hotplug loading is disabled")
	}

	if _, exists := hl.modules[path]; exists {
		return nil, fmt.Errorf("module already loaded: %s", path)
	}

	plug, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	sym, err := plug.Lookup("Module")
	if err != nil {
		return nil, fmt.Errorf("failed to find Module symbol: %w", err)
	}

	dynamicModule, ok := sym.(DynamicModule)
	if !ok {
		return nil, fmt.Errorf("plugin does not implement DynamicModule interface")
	}

	checksum := hl.computeChecksum(path)

	if hl.antiAnalysis != nil && hl.antiAnalysis.IsEnabled() {
		protectedPath := hl.antiAnalysis.ProtectString(path)
		if err := hl.antiAnalysis.ProtectData(protectedPath, []byte(checksum)); err != nil {
			return nil, fmt.Errorf("failed to protect module: %w", err)
		}
	}

	if err := dynamicModule.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize module: %w", err)
	}

	loaded := &LoadedModule{
		Name:     dynamicModule.Name(),
		Version:  dynamicModule.Version(),
		Path:     path,
		Checksum: checksum,
		LoadedAt: time.Now(),
		Module:   dynamicModule,
		plugin:   plug,
	}

	hl.modules[path] = loaded

	return loaded, nil
}

func (hl *HotplugLoader) UnloadModule(path string) error {
	hl.mu.Lock()
	defer hl.mu.Unlock()

	loaded, exists := hl.modules[path]
	if !exists {
		return fmt.Errorf("module not loaded: %s", path)
	}

	if err := loaded.Module.Unload(); err != nil {
		return fmt.Errorf("failed to unload module: %w", err)
	}

	delete(hl.modules, path)

	return nil
}

func (hl *HotplugLoader) ReloadModule(path string) (*LoadedModule, error) {
	if err := hl.UnloadModule(path); err != nil {
		return nil, fmt.Errorf("failed to unload existing module: %w", err)
	}

	return hl.LoadModule(path)
}

func (hl *HotplugLoader) GetModule(name string) *LoadedModule {
	hl.mu.RLock()
	defer hl.mu.RUnlock()

	for _, mod := range hl.modules {
		if mod.Name == name {
			return mod
		}
	}

	return nil
}

func (hl *HotplugLoader) ListModules() []*LoadedModule {
	hl.mu.RLock()
	defer hl.mu.RUnlock()

	modules := make([]*LoadedModule, 0, len(hl.modules))
	for _, mod := range hl.modules {
		modules = append(modules, mod)
	}

	return modules
}

func (hl *HotplugLoader) computeChecksum(path string) string {
	data := []byte(path)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (hl *HotplugLoader) Enable() {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	hl.enabled = true
}

func (hl *HotplugLoader) Disable() {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	hl.enabled = false
}

func (hl *HotplugLoader) IsEnabled() bool {
	hl.mu.RLock()
	defer hl.mu.RUnlock()
	return hl.enabled
}

func (hl *HotplugLoader) GetStatus() map[string]interface{} {
	hl.mu.RLock()
	defer hl.mu.RUnlock()

	modules := make([]map[string]interface{}, 0, len(hl.modules))
	for _, mod := range hl.modules {
		modules = append(modules, map[string]interface{}{
			"name":      mod.Name,
			"version":   mod.Version,
			"path":      mod.Path,
			"checksum":  mod.Checksum,
			"loaded_at": mod.LoadedAt,
		})
	}

	return map[string]interface{}{
		"enabled": hl.enabled,
		"count":   len(hl.modules),
		"modules": modules,
	}
}

func (hl *HotplugLoader) WatchPath(path string) {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	hl.watchPaths = append(hl.watchPaths, path)
}
