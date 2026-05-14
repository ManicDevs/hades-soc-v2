package anti_analysis

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// AntiAnalysisManager coordinates all anti-analysis techniques
type AntiAnalysisManager struct {
	staticObfuscation   *StaticObfuscation
	dynamicProtection   *DynamicProtection
	binaryPacker        *BinaryPacker
	antiDisassembly     *AntiDisassembly
	antiStaticAnalysis  *AntiStaticAnalysis
	antiTampering       *AntiTampering
	antiDebugging       *AntiDebugging
	antiVM              *AntiVM
	antiInstrumentation *AntiInstrumentation

	enabled     bool
	protection  sync.RWMutex
	initialized bool
}

// NewAntiAnalysisManager creates a comprehensive anti-analysis manager
func NewAntiAnalysisManager() *AntiAnalysisManager {
	manager := &AntiAnalysisManager{
		staticObfuscation:   NewStaticObfuscation(),
		dynamicProtection:   NewDynamicProtection(),
		binaryPacker:        NewBinaryPacker(),
		antiDisassembly:     NewAntiDisassembly(),
		antiStaticAnalysis:  NewAntiStaticAnalysis(),
		antiTampering:       NewAntiTampering(),
		antiDebugging:       NewAntiDebugging(),
		antiVM:              NewAntiVM(),
		antiInstrumentation: NewAntiInstrumentation(),
		enabled:             false, // Disabled by default for server mode
	}

	// Don't auto-initialize protections in server mode
	// manager.initializeProtections()

	return manager
}

// initializeProtections sets up all anti-analysis mechanisms
func (aam *AntiAnalysisManager) initializeProtections() {
	aam.protection.Lock()
	defer aam.protection.Unlock()

	if aam.initialized {
		return
	}

	// Initialize static protections
	aam.antiDisassembly.AddJunkInstructions()

	// Initialize dynamic protections
	aam.antiDebugging.AddDebuggerCheck()
	aam.antiVM.AddVMChecks()
	aam.antiInstrumentation.AddInstrumentationCheck()

	// Register critical functions for integrity checking
	aam.registerCriticalFunctions()

	// Start background monitoring
	go aam.backgroundMonitoring()

	aam.initialized = true
}

// registerCriticalFunctions registers important functions for tampering detection
func (aam *AntiAnalysisManager) registerCriticalFunctions() {
	// Register key functions for integrity checking
	aam.antiTampering.RegisterFunction("backgroundMonitoring",
		aam.backgroundMonitoring)
	aam.antiTampering.RegisterFunction("detectAnalysis",
		aam.detectAnalysis)
	aam.antiTampering.RegisterFunction("triggerProtection",
		aam.triggerProtection)
}

// backgroundMonitoring runs continuous background checks
func (aam *AntiAnalysisManager) backgroundMonitoring() {
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for range ticker.C {
		if !aam.enabled {
			continue
		}

		// Perform comprehensive analysis detection
		if aam.detectAnalysis() {
			aam.triggerProtection()
		}

		// Verify integrity
		if !aam.antiTampering.VerifyIntegrity() {
			aam.handleTampering()
		}

		// Add execution noise
		aam.addExecutionNoise()
	}
}

// detectAnalysis performs comprehensive analysis detection
func (aam *AntiAnalysisManager) detectAnalysis() bool {
	return aam.antiDebugging.IsDebuggingActive() ||
		aam.antiVM.IsVMEnvironment() ||
		aam.antiInstrumentation.IsInstrumentationActive()
}

// triggerProtection activates when analysis is detected
func (aam *AntiAnalysisManager) triggerProtection() {
	// Server mode - never crash
	fmt.Printf("WARNING: Analysis detected but continuing in server mode\n")
}

// fakeSystemCrash creates a fake system crash
func (aam *AntiAnalysisManager) fakeSystemCrash() {
	// Don't actually crash in server mode
	fmt.Printf("WARNING: Fake crash triggered (would panic in non-server mode)\n")
}

// corruptMemory corrupts memory regions
func (aam *AntiAnalysisManager) corruptMemory() {
	// Create memory noise
	junk := make([]byte, 1024*1024) // 1MB
	for i := range junk {
		junk[i] = byte(i * 7)
	}

	// Force garbage collection to confuse memory analysis
	runtime.GC()
}

// enterInfiniteLoop creates an infinite loop
func (aam *AntiAnalysisManager) enterInfiniteLoop() {
	for {
		time.Sleep(time.Second)
		runtime.Gosched() // Be nice to the scheduler
	}
}

// triggerGarbageCollection triggers aggressive garbage collection
func (aam *AntiAnalysisManager) triggerGarbageCollection() {
	for i := 0; i < 100; i++ {
		runtime.GC()
		time.Sleep(time.Millisecond * 10)
	}
}

// scrambleExecution scrambles the execution flow
func (aam *AntiAnalysisManager) scrambleExecution() {
	// Create fake computation threads
	for i := 0; i < 20; i++ {
		go func(id int) {
			for j := 0; j < 1000; j++ {
				// Fake cryptographic operations
				data := make([]byte, 64)
				for k := range data {
					data[k] = byte(uint32(k) * uint32(id+j))
				}
				time.Sleep(time.Microsecond * time.Duration(randIntBinary(1, 100)))
			}
		}(i)
	}
}

// handleTampering handles detected tampering attempts
func (aam *AntiAnalysisManager) handleTampering() {
	// Just log the tampering, don't crash the server
	fmt.Printf("WARNING: Tampering detected but continuing (server mode)\n")
}

// addExecutionNoise adds random execution patterns
func (aam *AntiAnalysisManager) addExecutionNoise() {
	// Add timing noise
	time.Sleep(time.Microsecond * time.Duration(randIntBinary(1, 1000)))

	// Add memory operations
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i * randIntBinary(1, 255))
	}
	_ = data // Use the data to avoid unused variable error

	// Add CPU operations
	for i := 0; i < randIntBinary(10, 100); i++ {
		_ = i * i * i // Fake computation
	}
}

// ProtectString obfuscates a string using multiple techniques
func (aam *AntiAnalysisManager) ProtectString(input string) string {
	aam.protection.RLock()
	defer aam.protection.RUnlock()

	if !aam.enabled {
		return input
	}

	// Use static obfuscation
	obfuscated := aam.staticObfuscation.ObfuscateString(input)

	// Also encode with anti-static analysis
	encoded := aam.antiStaticAnalysis.EncodeString(input)

	// Return the more complex obfuscation
	if len(encoded) > len(obfuscated.encoded) {
		return encoded
	}

	return obfuscated.encoded
}

// UnprotectString reverses string protection
func (aam *AntiAnalysisManager) UnprotectString(protected string) (string, error) {
	aam.protection.RLock()
	defer aam.protection.RUnlock()

	if !aam.enabled {
		return protected, nil
	}

	// Try anti-static analysis first
	if decoded, err := aam.antiStaticAnalysis.DecodeString(protected); err == nil {
		return decoded, nil
	}

	// Try static obfuscation
	obfuscated := &StringObfuscation{encoded: protected}
	return aam.staticObfuscation.DeobfuscateString(obfuscated), nil
}

// ProtectData encrypts and packs sensitive data
func (aam *AntiAnalysisManager) ProtectData(key string, data []byte) error {
	aam.protection.Lock()
	defer aam.protection.Unlock()

	if !aam.enabled {
		return nil
	}

	// Encrypt with anti-static analysis
	if err := aam.antiStaticAnalysis.EncryptData(key, data); err != nil {
		return err
	}

	// Pack with binary packer
	return aam.binaryPacker.PackSection(key, data)
}

// UnprotectData decrypts and unpacks protected data
func (aam *AntiAnalysisManager) UnprotectData(key string) ([]byte, error) {
	aam.protection.RLock()
	defer aam.protection.RUnlock()

	if !aam.enabled {
		return nil, fmt.Errorf("protection not enabled")
	}

	// Unpack with binary packer
	data, err := aam.binaryPacker.UnpackSection(key)
	if err != nil {
		return nil, err
	}

	// Decrypt with anti-static analysis using the unpacked data
	decrypted, err := aam.antiStaticAnalysis.DecryptData(key)
	if err != nil {
		return nil, err
	}

	// Combine data from both sources for added complexity
	combined := make([]byte, len(data)+len(decrypted))
	copy(combined, data)
	copy(combined[len(data):], decrypted)

	return combined, nil
}

// ObfuscateCode obfuscates code instructions
func (aam *AntiAnalysisManager) ObfuscateCode(code []byte) []byte {
	aam.protection.RLock()
	defer aam.protection.RUnlock()

	if !aam.enabled {
		return code
	}

	return aam.antiDisassembly.ObfuscateInstruction(code)
}

// Enable enables anti-analysis protections
func (aam *AntiAnalysisManager) Enable() {
	aam.protection.Lock()
	defer aam.protection.Unlock()

	aam.enabled = true
}

// Disable disables anti-analysis protections
func (aam *AntiAnalysisManager) Disable() {
	aam.protection.Lock()
	defer aam.protection.Unlock()

	aam.enabled = false
}

// IsEnabled checks if protections are enabled
func (aam *AntiAnalysisManager) IsEnabled() bool {
	aam.protection.RLock()
	defer aam.protection.RUnlock()

	return aam.enabled
}

// GetStatus returns the current status of all protections
func (aam *AntiAnalysisManager) GetStatus() map[string]interface{} {
	aam.protection.RLock()
	defer aam.protection.RUnlock()

	return map[string]interface{}{
		"enabled":                  aam.enabled,
		"initialized":              aam.initialized,
		"debugging_detected":       aam.antiDebugging.IsDebuggingActive(),
		"vm_detected":              aam.antiVM.IsVMEnvironment(),
		"instrumentation_detected": aam.antiInstrumentation.IsInstrumentationActive(),
		"integrity_valid":          aam.antiTampering.VerifyIntegrity(),
		"packed_sections":          len(aam.binaryPacker.packedSections),
		"obfuscated_strings":       len(aam.antiStaticAnalysis.stringEncodings),
		"encrypted_data":           len(aam.antiStaticAnalysis.encryptedData),
	}
}

// Global anti-analysis manager instance
var globalAntiAnalysisManager *AntiAnalysisManager
var once sync.Once

// GetGlobalAntiAnalysisManager returns the singleton instance
func GetGlobalAntiAnalysisManager() *AntiAnalysisManager {
	once.Do(func() {
		globalAntiAnalysisManager = NewAntiAnalysisManager()
	})
	return globalAntiAnalysisManager
}

// Convenience functions for global access

// ProtectStringGlobal protects a string using the global manager
func ProtectStringGlobal(input string) string {
	return GetGlobalAntiAnalysisManager().ProtectString(input)
}

// UnprotectStringGlobal unprotects a string using the global manager
func UnprotectStringGlobal(protected string) (string, error) {
	return GetGlobalAntiAnalysisManager().UnprotectString(protected)
}

// ProtectDataGlobal protects data using the global manager
func ProtectDataGlobal(key string, data []byte) error {
	return GetGlobalAntiAnalysisManager().ProtectData(key, data)
}

// UnprotectDataGlobal unprotects data using the global manager
func UnprotectDataGlobal(key string) ([]byte, error) {
	return GetGlobalAntiAnalysisManager().UnprotectData(key)
}

// ObfuscateCodeGlobal obfuscates code using the global manager
func ObfuscateCodeGlobal(code []byte) []byte {
	return GetGlobalAntiAnalysisManager().ObfuscateCode(code)
}

// EnableGlobal enables global anti-analysis protections
func EnableGlobal() {
	GetGlobalAntiAnalysisManager().Enable()
}

// DisableGlobal disables global anti-analysis protections
func DisableGlobal() {
	GetGlobalAntiAnalysisManager().Disable()
}

// GetStatusGlobal returns global protection status
func GetStatusGlobal() map[string]interface{} {
	return GetGlobalAntiAnalysisManager().GetStatus()
}
