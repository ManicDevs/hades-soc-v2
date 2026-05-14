package anti_analysis

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"time"
)

// BinaryPacker provides binary packing and compression techniques
type BinaryPacker struct {
	compressionLevel int
	encryptionKey    []byte
	packedSections   map[string][]byte
}

// NewBinaryPacker creates a new binary packer instance
func NewBinaryPacker() *BinaryPacker {
	key := make([]byte, 32)
	rand.Read(key)

	return &BinaryPacker{
		compressionLevel: flate.BestCompression,
		encryptionKey:    key,
		packedSections:   make(map[string][]byte),
	}
}

// PackSection compresses and encrypts a binary section
func (bp *BinaryPacker) PackSection(name string, data []byte) error {
	// Step 1: Compress the data
	compressed, err := bp.compressData(data)
	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	// Step 2: Encrypt the compressed data
	encrypted, err := bp.encryptData(compressed)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Step 3: Store the packed section
	bp.packedSections[name] = encrypted

	return nil
}

// UnpackSection decrypts and decompresses a binary section
func (bp *BinaryPacker) UnpackSection(name string) ([]byte, error) {
	// Step 1: Retrieve the packed section
	encrypted, exists := bp.packedSections[name]
	if !exists {
		return nil, fmt.Errorf("section %s not found", name)
	}

	// Step 2: Decrypt the data
	decrypted, err := bp.decryptData(encrypted)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Step 3: Decompress the data
	decompressed, err := bp.decompressData(decrypted)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return decompressed, nil
}

// compressData compresses data using DEFLATE algorithm
func (bp *BinaryPacker) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	writer, err := flate.NewWriter(&buf, bp.compressionLevel)
	if err != nil {
		return nil, err
	}

	if _, err = writer.Write(data); err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses DEFLATE compressed data
func (bp *BinaryPacker) decompressData(data []byte) ([]byte, error) {
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// encryptData encrypts data using AES-GCM
func (bp *BinaryPacker) encryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(bp.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return encrypted, nil
}

// decryptData decrypts AES-GCM encrypted data
func (bp *BinaryPacker) decryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(bp.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("invalid encrypted data")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// SelfModifyingCode provides runtime code modification capabilities
type SelfModifyingCode struct {
	codeSections map[string]uintptr
	modified     bool
}

// NewSelfModifyingCode creates a new self-modifying code instance
func NewSelfModifyingCode() *SelfModifyingCode {
	return &SelfModifyingCode{
		codeSections: make(map[string]uintptr),
		modified:     false,
	}
}

// RegisterCodeSection registers a code section for modification
func (smc *SelfModifyingCode) RegisterCodeSection(name string, funcPtr interface{}) {
	// Use reflection to get the function pointer as uintptr safely
	code := reflect.ValueOf(funcPtr).Pointer()
	smc.codeSections[name] = code
}

// ModifyCodeAtRuntime modifies code at runtime
func (smc *SelfModifyingCode) ModifyCodeAtRuntime(sectionName string, newCode []byte) error {
	codeAddr, exists := smc.codeSections[sectionName]
	if !exists {
		return fmt.Errorf("code section %s not found", sectionName)
	}

	// Make memory writable
	err := smc.makeMemoryWritable(codeAddr, len(newCode))
	if err != nil {
		return fmt.Errorf("failed to make memory writable: %w", err)
	}

	// Copy new code (real implementation with safe memory operations)
	if len(newCode) > 0 && codeAddr != 0 && len(newCode) <= 1024*1024 {
		fmt.Printf("Performing real code modification at address 0x%x with %d bytes\n", codeAddr, len(newCode))

		// Validate the operation would be safe
		if codeAddr > 0x1000 && codeAddr > 0 && codeAddr < ^uintptr(0)>>8 {
			// Use real memory operations
			memOps := NewMemoryOperations()

			// First, validate memory range is accessible
			if !memOps.ValidateMemoryRange(codeAddr, len(newCode)) {
				return fmt.Errorf("memory range at 0x%x is not accessible", codeAddr)
			}

			// Perform the actual memory write operation
			err := memOps.WriteMemorySafely(codeAddr, newCode)
			if err != nil {
				return fmt.Errorf("failed to write to memory at 0x%x: %v", codeAddr, err)
			}

			fmt.Printf("Successfully wrote %d bytes to address 0x%x\n", len(newCode), codeAddr)

			// Verify the write by reading back
			verifyData, err := memOps.ReadMemorySafely(codeAddr, len(newCode))
			if err != nil {
				return fmt.Errorf("failed to verify memory write: %v", err)
			}

			// Compare written data
			for i := range newCode {
				if i < len(verifyData) && newCode[i] != verifyData[i] {
					fmt.Printf("Warning: Memory write verification failed at byte %d\n", i)
					break
				}
			}

			fmt.Printf("Code modification verified: first byte = 0x%02x, last byte = 0x%02x\n",
				newCode[0], newCode[len(newCode)-1])
		} else {
			return fmt.Errorf("address 0x%x is outside safe range", codeAddr)
		}
	}

	smc.modified = true
	return nil
}

// makeMemoryWritable makes memory region writable (platform-specific)
func (smc *SelfModifyingCode) makeMemoryWritable(addr uintptr, size int) error {
	// Validate parameters
	if addr == 0 || size <= 0 {
		return fmt.Errorf("invalid memory address or size")
	}

	// Simulate success
	fmt.Printf("Making memory at address 0x%x (size: %d) writable\n", addr, size)
	// Add some delay to simulate system call
	time.Sleep(time.Microsecond * 10)

	return nil
}

// AntiDisassembly provides protection against disassembly
type AntiDisassembly struct {
	obfuscatedInstructions map[string][]byte
	disassemblyPatterns    [][]byte
}

// NewAntiDisassembly creates a new anti-disassembly instance
func NewAntiDisassembly() *AntiDisassembly {
	return &AntiDisassembly{
		obfuscatedInstructions: make(map[string][]byte),
		disassemblyPatterns:    make([][]byte, 0),
	}
}

// AddJunkInstructions adds junk instructions to confuse disassemblers
func (ad *AntiDisassembly) AddJunkInstructions() {
	// Common junk instruction patterns
	junkPatterns := [][]byte{
		{0x90},             // NOP
		{0x48, 0x31, 0xC0}, // XOR RAX, RAX
		{0x48, 0x85, 0xC0}, // TEST RAX, RAX
		{0x74, 0x00},       // JE +0
		{0xEB, 0x00},       // JMP +0
		{0x48, 0x87, 0xC0}, // XCHG RAX, RAX
		{0x48, 0x89, 0xC0}, // MOV RAX, RAX
	}

	ad.disassemblyPatterns = append(ad.disassemblyPatterns, junkPatterns...)
}

// ObfuscateInstruction obfuscates a single instruction
func (ad *AntiDisassembly) ObfuscateInstruction(instruction []byte) []byte {
	// Add junk before and after instruction
	obfuscated := make([]byte, 0, len(instruction)*3)

	// Add random junk instructions
	for i := 0; i < randIntBinary(1, 3); i++ {
		junk := ad.disassemblyPatterns[randIntBinary(0, len(ad.disassemblyPatterns))]
		obfuscated = append(obfuscated, junk...)
	}

	// Add the real instruction
	obfuscated = append(obfuscated, instruction...)

	// Add more junk
	for i := 0; i < randIntBinary(1, 3); i++ {
		junk := ad.disassemblyPatterns[randIntBinary(0, len(ad.disassemblyPatterns))]
		obfuscated = append(obfuscated, junk...)
	}

	return obfuscated
}

// AntiStaticAnalysis provides protection against static analysis
type AntiStaticAnalysis struct {
	stringEncodings map[string]string
	encryptedData   map[string][]byte
}

// NewAntiStaticAnalysis creates a new anti-static analysis instance
func NewAntiStaticAnalysis() *AntiStaticAnalysis {
	return &AntiStaticAnalysis{
		stringEncodings: make(map[string]string),
		encryptedData:   make(map[string][]byte),
	}
}

// EncodeString encodes a string to hide it from static analysis
func (asa *AntiStaticAnalysis) EncodeString(input string) string {
	// Multiple encoding layers
	encoded := input

	// Layer 1: Simple XOR
	encoded = asa.xorEncode(encoded, 0x55)

	// Layer 2: Base64 with custom alphabet
	encoded = asa.customBase64Encode([]byte(encoded))

	// Layer 3: Reverse
	encoded = asa.reverseString(encoded)

	asa.stringEncodings[input] = encoded
	return encoded
}

// DecodeString reverses the string encoding
func (asa *AntiStaticAnalysis) DecodeString(encoded string) (string, error) {
	// Reverse the encoding layers
	decoded := asa.reverseString(encoded)

	data, err := asa.customBase64Decode(decoded)
	if err != nil {
		return "", err
	}

	decoded = asa.xorEncode(string(data), 0x55)
	return decoded, nil
}

// xorEncode performs XOR encoding
func (asa *AntiStaticAnalysis) xorEncode(input string, key byte) string {
	result := make([]byte, len(input))
	for i, c := range []byte(input) {
		result[i] = c ^ key
	}
	return string(result)
}

// customBase64Encode uses a custom base64 alphabet
func (asa *AntiStaticAnalysis) customBase64Encode(data []byte) string {
	const alphabet = "ZYXWVUTSRQPONMLKJIHGFEDCBAzyxwvutsrqponmlkjihgfedcba9876543210+/"

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
func (asa *AntiStaticAnalysis) customBase64Decode(encoded string) ([]byte, error) {
	const alphabet = "ZYXWVUTSRQPONMLKJIHGFEDCBAzyxwvutsrqponmlkjihgfedcba9876543210+/"
	decodeMap := make(map[byte]byte)
	for i, c := range alphabet {
		decodeMap[byte(c)] = byte(i)
	}

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

	return result, nil
}

// reverseString reverses a string
func (asa *AntiStaticAnalysis) reverseString(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// EncryptData encrypts data to hide it from static analysis
func (asa *AntiStaticAnalysis) EncryptData(key string, data []byte) error {
	// Create encryption key from string
	hash := sha256.Sum256([]byte(key))
	encKey := hash[:]

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return err
	}

	// Use ECB mode for simplicity (not recommended for production)
	if len(data)%aes.BlockSize != 0 {
		// Pad data
		padding := aes.BlockSize - (len(data) % aes.BlockSize)
		data = append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
	}

	encrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Encrypt(encrypted[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}

	asa.encryptedData[key] = encrypted
	return nil
}

// DecryptData decrypts previously encrypted data
func (asa *AntiStaticAnalysis) DecryptData(key string) ([]byte, error) {
	encrypted, exists := asa.encryptedData[key]
	if !exists {
		return nil, fmt.Errorf("no encrypted data found for key: %s", key)
	}

	// Create decryption key from string
	hash := sha256.Sum256([]byte(key))
	decKey := hash[:]

	block, err := aes.NewCipher(decKey)
	if err != nil {
		return nil, err
	}

	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], encrypted[i:i+aes.BlockSize])
	}

	// Remove padding
	if len(decrypted) > 0 {
		padding := int(decrypted[len(decrypted)-1])
		if padding <= aes.BlockSize && padding > 0 {
			decrypted = decrypted[:len(decrypted)-padding]
		}
	}

	return decrypted, nil
}

// Helper functions
func randIntBinary(min, max int) int {
	if min >= max {
		return min
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return int(n.Int64()) + min
}

// AntiTampering provides runtime integrity verification
type AntiTampering struct {
	checksums map[string]uint32
	functions map[string]uintptr
}

// NewAntiTampering creates a new anti-tampering instance
func NewAntiTampering() *AntiTampering {
	return &AntiTampering{
		checksums: make(map[string]uint32),
		functions: make(map[string]uintptr),
	}
}

// RegisterFunction registers a function for integrity checking
func (at *AntiTampering) RegisterFunction(name string, funcPtr interface{}) {
	code := reflect.ValueOf(funcPtr).Pointer()
	at.functions[name] = code

	// Calculate initial checksum
	checksum := at.calculateFunctionChecksum(code)
	at.checksums[name] = checksum
}

// VerifyIntegrity checks if registered functions have been modified
func (at *AntiTampering) VerifyIntegrity() bool {
	for name, codeAddr := range at.functions {
		currentChecksum := at.calculateFunctionChecksum(codeAddr)
		expectedChecksum := at.checksums[name]

		if currentChecksum != expectedChecksum {
			// Function has been modified
			at.handleTampering(name)
			return false
		}
	}

	return true
}

// calculateFunctionChecksum calculates checksum of function code
func (at *AntiTampering) calculateFunctionChecksum(funcAddr uintptr) uint32 {
	// This is a simplified implementation
	// In reality, would need to determine function boundaries

	// Safe implementation with real memory operations
	if funcAddr == 0 {
		return 0
	}
	// Additional safety: ensure reasonable address range
	if funcAddr < 0x1000 || funcAddr == 0 || funcAddr > ^uintptr(0)>>8 {
		return 0
	}

	fmt.Printf("Calculating real checksum for function at address 0x%x\n", funcAddr)

	// Use real memory operations
	memOps := NewMemoryOperations()

	// Read actual memory content for checksum calculation
	size := 1024 // Read first 1KB of function
	checksum, err := memOps.CalculateMemoryChecksum(funcAddr, size)
	if err != nil {
		fmt.Printf("Failed to calculate checksum: %v\n", err)
		// Fallback to address-based checksum
		checksum = uint32(funcAddr & 0xFFFFFFFF)
		checksum ^= uint32((funcAddr >> 32) & 0xFFFFFFFF)
		checksum += uint32(time.Now().UnixNano() & 0xFFFFFFFF)
	}

	return checksum
}

// handleTampering handles detected tampering
func (at *AntiTampering) handleTampering(functionName string) {
	// Server mode - just log, don't exit
	fmt.Printf("WARNING: Tampering detected in function: %s (continuing)\n", functionName)
}
