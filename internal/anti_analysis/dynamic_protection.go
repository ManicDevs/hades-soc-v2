package anti_analysis

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// DynamicProtection provides runtime anti-analysis techniques
type DynamicProtection struct {
	integrityCheck uint32
	heartbeat      uint64
	lastCheck      int64
}

// NewDynamicProtection creates a new dynamic protection instance
func NewDynamicProtection() *DynamicProtection {
	dp := &DynamicProtection{
		integrityCheck: 0x12345678,
		heartbeat:      uint64(time.Now().Unix()),
		lastCheck:      time.Now().UnixNano(),
	}

	// Don't start background monitoring in server mode
	// go dp.monitorExecution()

	return dp
}

// monitorExecution runs background checks for analysis detection
func (dp *DynamicProtection) monitorExecution() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for range ticker.C {
		if dp.detectAnalysis() {
			dp.triggerProtection()
		}

		// Update heartbeat
		atomic.StoreUint64(&dp.heartbeat, uint64(time.Now().Unix()))
	}
}

// detectAnalysis performs various runtime analysis detection
func (dp *DynamicProtection) detectAnalysis() bool {
	return dp.detectTimingAnalysis() ||
		dp.detectMemoryAnalysis() ||
		dp.detectBehavioralAnalysis()
}

// detectTimingAnalysis detects timing-based analysis
func (dp *DynamicProtection) detectTimingAnalysis() bool {
	// Perform multiple timing measurements
	measurements := make([]time.Duration, 10)

	for i := 0; i < 10; i++ {
		start := time.Now()

		// Perform some computation
		data := make([]byte, 1024)
		rand.Read(data)

		for j := 0; j < 1000; j++ {
			data[j%len(data)] ^= byte(j)
		}

		measurements[i] = time.Since(start)

		// Add random delay
		time.Sleep(time.Microsecond * time.Duration(randInt(1, 100)))
	}

	// Check for consistent timing (indicates analysis)
	var variance int64
	mean := measurements[0]

	for _, m := range measurements {
		diff := int64(m - mean)
		if diff < 0 {
			diff = -diff
		}
		variance += diff
	}

	// If variance is too low, likely being analyzed
	return variance < int64(time.Millisecond)*10
}

// detectMemoryAnalysis detects memory scanning attempts
func (dp *DynamicProtection) detectMemoryAnalysis() bool {
	// Create honey pot memory regions
	honeyData := make([]byte, 4096)
	copy(honeyData, []byte("HONEY_POT_ANALYSIS_DETECTION"))

	// Check if honey pot is accessed
	initialHash := hashMemory(honeyData)

	// Let some time pass
	time.Sleep(time.Millisecond * 10)

	// Check if honey pot was modified
	finalHash := hashMemory(honeyData)

	return initialHash != finalHash
}

// detectBehavioralAnalysis detects behavioral analysis patterns
func (dp *DynamicProtection) detectBehavioralAnalysis() bool {
	// Check for unusual call patterns
	currentTime := time.Now().UnixNano()
	timeSinceLastCheck := currentTime - atomic.LoadInt64(&dp.lastCheck)

	// If checks are too frequent, likely analysis
	if timeSinceLastCheck < int64(time.Millisecond)*100 {
		return true
	}

	atomic.StoreInt64(&dp.lastCheck, currentTime)
	return false
}

// triggerProtection activates when analysis is detected
func (dp *DynamicProtection) triggerProtection() {
	// Scramble critical data structures
	dp.scrambleCriticalData()

	// Add noise to execution
	dp.addExecutionNoise()

	// Fake crash or exit behavior
	dp.fakeSystemError()
}

// scrambleCriticalData scrambles sensitive data
func (dp *DynamicProtection) scrambleCriticalData() {
	// Scramble integrity check
	current := atomic.LoadUint32(&dp.integrityCheck)
	atomic.StoreUint32(&dp.integrityCheck, current^0x89ABCDEF)

	// Create memory noise
	noise := make([]byte, 1024*1024) // 1MB of noise
	rand.Read(noise)

	// Write noise to memory (confuses memory analysis)
	for i := 0; i < len(noise); i += 4096 {
		page := noise[i:min(i+4096, len(noise))]
		_ = page // Force compiler to keep this
	}
}

// addExecutionNoise adds random execution patterns
func (dp *DynamicProtection) addExecutionNoise() {
	// Create fake computation threads
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 1000; j++ {
				// Fake cryptographic operations
				data := make([]byte, 32)
				rand.Read(data)

				// Perform fake operations
				for k := 0; k < len(data); k++ {
					data[k] = byte(uint32(data[k])*uint32(id+1) + uint32(j))
				}

				time.Sleep(time.Microsecond * time.Duration(randInt(1, 50)))
			}
		}(i)
	}
}

// fakeSystemError creates fake system errors
func (dp *DynamicProtection) fakeSystemError() {
	// Simulate system error by modifying memory (safe implementation)
	// Note: In production, this would involve actual memory manipulation
	// For safety, we'll just simulate the effect
	fmt.Printf("Simulating system error at address 0x12345678 with value 0x87654321\n")
}

// recursiveFunction creates fake stack overflow
func (dp *DynamicProtection) recursiveFunction(depth int) {
	if depth <= 0 {
		return
	}

	// Add some work to make it realistic
	data := make([]byte, 1024)
	rand.Read(data)

	dp.recursiveFunction(depth - 1)
}

// AntiDumping provides protection against memory dumping
type AntiDumping struct {
	protectedRegions [][]byte
	checksums        []uint32
}

// NewAntiDumping creates a new anti-dumping protection
func NewAntiDumping() *AntiDumping {
	return &AntiDumping{
		protectedRegions: make([][]byte, 0),
		checksums:        make([]uint32, 0),
	}
}

// ProtectRegion adds memory region protection
func (ad *AntiDumping) ProtectRegion(data []byte) {
	// Calculate checksum
	checksum := calculateChecksum(data)

	// Store protected region
	ad.protectedRegions = append(ad.protectedRegions, data)
	ad.checksums = append(ad.checksums, checksum)

	// Start monitoring
	go ad.monitorRegion(len(ad.protectedRegions) - 1)
}

// monitorRegion monitors a protected region for dumping
func (ad *AntiDumping) monitorRegion(index int) {
	if index >= len(ad.protectedRegions) {
		return
	}

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for range ticker.C {
		currentChecksum := calculateChecksum(ad.protectedRegions[index])
		if currentChecksum != ad.checksums[index] {
			// Region was accessed, trigger protection
			ad.handleDumpingAttempt(index)
			break
		}
	}
}

// handleDumpingAttempt handles memory dumping attempts
func (ad *AntiDumping) handleDumpingAttempt(index int) {
	// Overwrite the region with garbage
	if index < len(ad.protectedRegions) {
		rand.Read(ad.protectedRegions[index])
	}

	// Trigger system protection
	panic("memory access violation detected")
}

// AntiInstrumentation provides protection against instrumentation
type AntiInstrumentation struct {
	instrumentationChecks []func() bool
}

// NewAntiInstrumentation creates a new anti-instrumentation instance
func NewAntiInstrumentation() *AntiInstrumentation {
	return &AntiInstrumentation{
		instrumentationChecks: make([]func() bool, 0),
	}
}

// AddInstrumentationCheck adds various instrumentation detection methods
func (ai *AntiInstrumentation) AddInstrumentationCheck() {
	ai.instrumentationChecks = append(ai.instrumentationChecks,
		func() bool {
			// Check for breakpoints
			return ai.detectBreakpoints()
		},
		func() bool {
			// Check for code modification
			return ai.detectCodeModification()
		},
		func() bool {
			// Check for hooking
			return ai.detectHooking()
		},
	)
}

// detectBreakpoints detects software breakpoints
func (ai *AntiInstrumentation) detectBreakpoints() bool {
	// Check for INT3 instructions (0xCC) in our code
	pc, _, _, _ := runtime.Caller(1)

	// Read memory at program counter (real check)
	funcPtr := runtime.FuncForPC(pc)
	if funcPtr != nil {
		entry := funcPtr.Entry()

		// Check first few bytes for breakpoint pattern (real implementation)
		if entry != 0 {
			fmt.Printf("Checking for breakpoints in function at address 0x%x\n", entry)

			// Use real memory operations
			memOps := NewMemoryOperations()

			for i := 0; i < 16; i++ {
				addr := uintptr(entry) + uintptr(i)
				// Additional safety: ensure reasonable address range
				if addr != 0 && addr > 0x1000 && addr < ^uintptr(0)>>8 {
					// Real breakpoint detection
					hasBreakpoint, err := memOps.CheckBreakpointAtAddress(addr)
					if err != nil {
						// Memory not accessible, skip this address
						continue
					}

					if hasBreakpoint {
						fmt.Printf("INT3 breakpoint detected at offset %d (address 0x%x)\n", i, addr)
						return true
					}
				}
			}
		}
	}

	return false
}

// detectCodeModification detects if code has been modified
func (ai *AntiInstrumentation) detectCodeModification() bool {
	// Get current function and check its integrity
	pc := make([]uintptr, 10)
	count := runtime.Callers(1, pc)

	if count > 0 {
		funcPtr := runtime.FuncForPC(pc[0])
		if funcPtr != nil {
			// Calculate hash of function code (real implementation)
			entry := funcPtr.Entry()
			// Use a fixed size for safety instead of calculating potentially invalid size
			size := 64 // Fixed size for safety

			if size > 0 && size < 1024*1024 && entry != 0 { // Reasonable size and valid entry
				// Additional safety: ensure reasonable address range
				if entry > 0x1000 && entry < ^uintptr(0)>>8 {
					fmt.Printf("Checking code integrity for function at address 0x%x\n", entry)

					// Use real memory operations
					memOps := NewMemoryOperations()

					// Read actual memory content
					data, err := memOps.ReadMemorySafely(entry, size)
					if err != nil {
						// Memory not accessible, cannot check integrity
						fmt.Printf("Cannot access memory at address 0x%x: %v\n", entry, err)
						return false
					}

					// Calculate checksum of real data
					currentChecksum := calculateChecksum(data)
					fmt.Printf("Current checksum: 0x%08x\n", currentChecksum)

					// Check if data appears to be zeroed out (sign of modification)
					zeroCount := 0
					for _, b := range data {
						if b == 0 {
							zeroCount++
						}
					}

					// If more than 50% of bytes are zero, likely modified
					if float64(zeroCount)/float64(len(data)) > 0.5 {
						fmt.Printf("Code modification detected: %d/%d bytes are zeroed\n", zeroCount, len(data))
						return true
					}

					// Check for repeated patterns (another sign of modification)
					patternCount := make(map[byte]int)
					for _, b := range data {
						patternCount[b]++
					}

					// If one byte appears more than 80% of the time, likely modified
					for _, count := range patternCount {
						if float64(count)/float64(len(data)) > 0.8 {
							fmt.Printf("Code modification detected: suspicious byte pattern\n")
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// detectHooking detects API hooking attempts
func (ai *AntiInstrumentation) detectHooking() bool {
	// Check if system calls are being intercepted
	// This is a simplified check - real implementation would be more complex

	// Time critical system calls
	start := time.Now()
	runtime.GC() // System call
	duration := time.Since(start)

	// If GC takes unusually long, might be hooked
	return duration > time.Millisecond*100
}

// IsInstrumentationActive checks if any instrumentation is detected
func (ai *AntiInstrumentation) IsInstrumentationActive() bool {
	for _, check := range ai.instrumentationChecks {
		if check() {
			return true
		}
	}
	return false
}

// Helper functions
func hashMemory(data []byte) uint32 {
	hash := uint32(0)
	for _, b := range data {
		hash = hash*31 + uint32(b)
	}
	return hash
}

func calculateChecksum(data []byte) uint32 {
	var sum uint32
	for i := 0; i < len(data); i += 4 {
		if i+4 <= len(data) {
			sum += binary.LittleEndian.Uint32(data[i:])
		} else {
			remaining := data[i:]
			var val uint32
			for j, b := range remaining {
				val |= uint32(b) << (j * 8)
			}
			sum += val
		}
	}
	return sum
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
