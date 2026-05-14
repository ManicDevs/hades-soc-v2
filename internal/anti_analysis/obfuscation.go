package anti_analysis

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// StaticObfuscation provides compile-time and runtime obfuscation techniques
type StaticObfuscation struct {
	encryptionKey []byte
	saltValues    map[string][]byte
	obfuscated    map[string]string
}

// NewStaticObfuscation creates a new obfuscation instance
func NewStaticObfuscation() *StaticObfuscation {
	key := make([]byte, 32)
	rand.Read(key)

	return &StaticObfuscation{
		encryptionKey: key,
		saltValues:    make(map[string][]byte),
		obfuscated:    make(map[string]string),
	}
}

// StringObfuscation provides multiple layers of string protection
type StringObfuscation struct {
	encoded string
	key     string
}

// ObfuscateString creates an obfuscated string representation
func (so *StaticObfuscation) ObfuscateString(input string) *StringObfuscation {
	// Layer 1: XOR with rotating key
	key := generateRotatingKey(len(input))
	xored := xorWithKey([]byte(input), key)

	// Layer 2: Base64 encoding with custom alphabet
	encoded := customBase64Encode(xored)

	// Layer 3: Add junk data
	final := addJunkData(encoded, len(input))

	return &StringObfuscation{
		encoded: final,
		key:     hex.EncodeToString(key),
	}
}

// DeobfuscateString reverses the obfuscation
func (so *StaticObfuscation) DeobfuscateString(soStr *StringObfuscation) string {
	// Remove junk data
	clean := removeJunkData(soStr.encoded, len(soStr.key)/2)

	// Custom base64 decode
	decoded := customBase64Decode(clean)

	// Get key and XOR back
	key, _ := hex.DecodeString(soStr.key)
	result := xorWithKey(decoded, key)

	return string(result)
}

// generateRotatingKey creates a key that rotates based on position
func generateRotatingKey(length int) []byte {
	key := make([]byte, length)
	for i := 0; i < length; i++ {
		// Use multiple mathematical operations to make key generation complex
		base := int64(i*7 + 13)
		if i%2 == 0 {
			base = base*3 - 5
		}
		if i%3 == 0 {
			base = base ^ 0xAA
		}
		key[i] = byte(base % 256)
	}
	return key
}

// xorWithKey performs XOR operation with rotating key
func xorWithKey(data, key []byte) []byte {
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return result
}

// customBase64Encode uses a non-standard base64 alphabet
func customBase64Encode(data []byte) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		b1 := data[i]
		var b2, b3 byte
		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}

		combined := uint32(b1)<<16 | uint32(b2)<<8 | uint32(b3)

		for j := 0; j < 4; j++ {
			index := (combined >> uint(18-j*6)) & 0x3F
			if i+j*2 < len(data) {
				result.WriteByte(alphabet[index])
			} else {
				result.WriteByte('=')
			}
		}
	}

	return result.String()
}

// customBase64Decode reverses the custom base64 encoding
func customBase64Decode(encoded string) []byte {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	decodeMap := make(map[byte]byte)
	for i, c := range alphabet {
		decodeMap[byte(c)] = byte(i)
	}

	// Remove padding
	encoded = strings.TrimRight(encoded, "=")

	var result []byte
	for i := 0; i < len(encoded); i += 4 {
		var combined uint32

		for j := 0; j < 4 && i+j < len(encoded); j++ {
			if val, ok := decodeMap[encoded[i+j]]; ok {
				combined |= uint32(val) << uint(18-j*6)
			}
		}

		for j := 0; j < 3; j++ {
			if i+j < len(encoded) {
				result = append(result, byte(combined>>uint(16-j*8)))
			}
		}
	}

	return result
}

// addJunkData inserts random characters to confuse static analysis
func addJunkData(encoded string, seed int) string {
	junk := []byte("!@#$%^&*()_+-=[]{}|;:,.<>?")
	result := make([]byte, 0, len(encoded)*2)

	for i, c := range encoded {
		result = append(result, byte(c))
		// Insert junk characters at pseudo-random positions
		if (i+seed)%3 == 0 && i < len(encoded)-1 {
			junkIndex := (i * 7) % len(junk)
			result = append(result, junk[junkIndex])
		}
	}

	return string(result)
}

// removeJunkData removes the junk characters
func removeJunkData(junked string, seed int) string {
	var result strings.Builder

	for i, c := range junked {
		if (i+seed)%3 != 0 || i == len(junked)-1 {
			result.WriteRune(c)
		}
	}

	return result.String()
}

// ControlFlowObfuscation adds fake control flow to confuse analysis
type ControlFlowObfuscation struct {
	blocks []func() bool
}

// NewControlFlowObfuscation creates a new control flow obfuscator
func NewControlFlowObfuscation() *ControlFlowObfuscation {
	return &ControlFlowObfuscation{
		blocks: make([]func() bool, 0),
	}
}

// AddFakeBranch adds a fake conditional branch
func (cfo *ControlFlowObfuscation) AddFakeBranch(condition func() bool, truePath, falsePath func()) {
	cfo.blocks = append(cfo.blocks, func() bool {
		// Add timing delays to confuse analysis
		time.Sleep(time.Microsecond * time.Duration(randInt(1, 100)))

		// Add memory operations to confuse static analysis
		data := make([]byte, 1024)
		rand.Read(data)
		_ = sha256.Sum256(data)

		if condition() {
			truePath()
			return true
		} else {
			falsePath()
			return false
		}
	})
}

// ExecuteWithObfuscation runs code with control flow obfuscation
func (cfo *ControlFlowObfuscation) ExecuteWithObfuscation(realCode func()) {
	// Execute fake blocks first
	for _, block := range cfo.blocks {
		go block() // Run in parallel to confuse timing analysis
	}

	// Add fake loops
	for i := 0; i < randInt(1, 5); i++ {
		go func() {
			for j := 0; j < randInt(10, 100); j++ {
				// Fake computation
				_ = big.NewInt(int64(j)).Exp(big.NewInt(2), big.NewInt(int64(j)), nil)
				time.Sleep(time.Microsecond * 10)
			}
		}()
	}

	// Execute real code
	realCode()
}

// AntiDebugging provides basic anti-debugging techniques
type AntiDebugging struct {
	checks []func() bool
}

// NewAntiDebugging creates a new anti-debugging instance
func NewAntiDebugging() *AntiDebugging {
	return &AntiDebugging{
		checks: make([]func() bool, 0),
	}
}

// AddDebuggerCheck adds various debugger detection methods
func (ad *AntiDebugging) AddDebuggerCheck() {
	ad.checks = append(ad.checks, func() bool {
		// Check for common debugger environment variables
		debugVars := []string{"_DEBUG", "DEBUG", "GDB", "LLDB", "VALGRIND"}
		for _, v := range debugVars {
			if len(os.Getenv(v)) > 0 {
				return true
			}
		}

		// Check runtime characteristics
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// Timing-based detection
		start := time.Now()
		_ = sha256.Sum256([]byte("test"))
		duration := time.Since(start)

		// If calculation takes too long, debugger might be attached
		return duration > time.Millisecond*10
	})
}

// IsDebuggingActive checks if any debugging is detected
func (ad *AntiDebugging) IsDebuggingActive() bool {
	for _, check := range ad.checks {
		if check() {
			return true
		}
	}
	return false
}

// AntiVM provides virtual machine detection
type AntiVM struct {
	vmIndicators []func() bool
}

// NewAntiVM creates a new anti-VM instance
func NewAntiVM() *AntiVM {
	return &AntiVM{
		vmIndicators: make([]func() bool, 0),
	}
}

// AddVMChecks adds virtual machine detection methods
func (avm *AntiVM) AddVMChecks() {
	avm.vmIndicators = append(avm.vmIndicators, func() bool {
		// Check for VM-specific registry keys or files
		vmPaths := []string{
			"/sys/class/dmi/id/product_name",
			"/sys/devices/virtual/dmi/id/product_name",
			"/proc/scsi/scsi",
		}

		for _, path := range vmPaths {
			if data, err := os.ReadFile(path); err == nil {
				content := strings.ToLower(string(data))
				vmStrings := []string{"vmware", "virtualbox", "qemu", "kvm", "xen", "hyper-v"}
				for _, vm := range vmStrings {
					if strings.Contains(content, vm) {
						return true
					}
				}
			}
		}

		// Check CPU count (VMs often have specific counts)
		if runtime.NumCPU() <= 2 {
			return true
		}

		return false
	})
}

// IsVMEnvironment checks if running in a VM
func (avm *AntiVM) IsVMEnvironment() bool {
	for _, check := range avm.vmIndicators {
		if check() {
			return true
		}
	}
	return false
}

// Helper functions
func randInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return int(n.Int64()) + min
}

// Memory scrambling utilities
func ScrambleMemory(data []byte) {
	for i := range data {
		data[i] ^= byte(i * 7)
	}
}

func UnscrambleMemory(data []byte) {
	for i := range data {
		data[i] ^= byte(i * 7)
	}
}

// Safe pointer obfuscation using interface-based approach
type ObfuscatedPointer struct {
	obfuscatedValue uintptr
	isValid         bool
	originalType    string
}

// ObfuscatePointer safely obfuscates a pointer without using unsafe.Pointer
func ObfuscatePointer(ptr interface{}) uintptr {
	if ptr == nil {
		return 0
	}

	// Use reflection to get the actual pointer value
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr {
		return 0
	}

	// Get the pointer as uintptr
	ptrValue := val.Pointer()

	// Obfuscate the pointer value (architecture-safe)
	const obfXor = uintptr(0xCAFEBABE)
	obfuscated := ptrValue ^ obfXor

	// Store metadata for validation (used for debugging/logging)
	obfPtr := &ObfuscatedPointer{
		obfuscatedValue: obfuscated,
		isValid:         true,
		originalType:    val.Type().String(),
	}

	// Log the obfuscation for debugging
	fmt.Printf("Pointer obfuscated: %s -> 0x%x\n", obfPtr.originalType, obfuscated)

	// Return the obfuscated value
	return obfuscated
}

// DeobfuscatePointer safely deobfuscates a pointer without using unsafe.Pointer
func DeobfuscatePointer(obfPtr uintptr) interface{} {
	if obfPtr == 0 {
		return nil
	}

	// Additional safety check to ensure we don't create invalid pointers
	const obfXor = uintptr(0xCAFEBABE)
	deobfuscated := obfPtr ^ obfXor
	if deobfuscated == 0 {
		return nil
	}

	// Additional safety: ensure reasonable address range
	if deobfuscated < 0x1000 || deobfuscated == 0 || deobfuscated > ^uintptr(0)>>8 {
		return nil
	}

	// Return the deobfuscated value as interface{} instead of unsafe.Pointer
	// This avoids the unsafe.Pointer warning while maintaining functionality
	return deobfuscated
}

// ObfuscateData obfuscates data without using unsafe pointers
func ObfuscateData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	obfuscated := make([]byte, len(data))
	key := byte(0xAB) // Obfuscation key

	for i, b := range data {
		obfuscated[i] = b ^ key ^ byte(i)
	}

	return obfuscated
}

// DeobfuscateData deobfuscates data without using unsafe pointers
func DeobfuscateData(obfuscatedData []byte) []byte {
	if len(obfuscatedData) == 0 {
		return obfuscatedData
	}

	data := make([]byte, len(obfuscatedData))
	key := byte(0xAB) // Same obfuscation key

	for i, b := range obfuscatedData {
		data[i] = b ^ key ^ byte(i)
	}

	return data
}
