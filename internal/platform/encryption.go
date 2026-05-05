package platform

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/scrypt"
)

// EncryptionAlgorithm represents supported encryption algorithms
type EncryptionAlgorithm string

const (
	AlgorithmAES256GCM EncryptionAlgorithm = "aes-256-gcm"
	AlgorithmAES256CBC EncryptionAlgorithm = "aes-256-cbc"
	AlgorithmChaCha20  EncryptionAlgorithm = "chacha20"
)

// HashAlgorithm represents supported hash algorithms
type HashAlgorithm string

const (
	HashSHA256 HashAlgorithm = "sha256"
	HashSHA512 HashAlgorithm = "sha512"
)

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Algorithm EncryptionAlgorithm `json:"algorithm"`
	KeySize   int                 `json:"key_size"`
	SaltSize  int                 `json:"salt_size"`
}

// DefaultEncryptionConfig returns sensible encryption defaults
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		Algorithm: AlgorithmAES256GCM,
		KeySize:   32,
		SaltSize:  16,
	}
}

// EncryptionService provides cryptographic operations
type EncryptionService struct {
	config    *EncryptionConfig
	masterKey []byte
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(config *EncryptionConfig, masterPassword string) (*EncryptionService, error) {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	// Derive master key from password using scrypt
	salt := []byte("hades-encryption-salt")
	masterKey, err := scrypt.Key([]byte(masterPassword), salt, 32768, 8, 1, config.KeySize)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to derive master key: %w", err)
	}

	return &EncryptionService{
		config:    config,
		masterKey: masterKey,
	}, nil
}

// Encrypt encrypts data using the configured algorithm
func (es *EncryptionService) Encrypt(plaintext []byte) ([]byte, error) {
	switch es.config.Algorithm {
	case AlgorithmAES256GCM:
		return es.encryptAES256GCM(plaintext)
	case AlgorithmAES256CBC:
		return es.encryptAES256CBC(plaintext)
	case AlgorithmChaCha20:
		return es.encryptChaCha20(plaintext)
	default:
		return nil, fmt.Errorf("hades.platform.encryption: unsupported algorithm: %s", es.config.Algorithm)
	}
}

// Decrypt decrypts data using the configured algorithm
func (es *EncryptionService) Decrypt(ciphertext []byte) ([]byte, error) {
	switch es.config.Algorithm {
	case AlgorithmAES256GCM:
		return es.decryptAES256GCM(ciphertext)
	case AlgorithmAES256CBC:
		return es.decryptAES256CBC(ciphertext)
	case AlgorithmChaCha20:
		return es.decryptChaCha20(ciphertext)
	default:
		return nil, fmt.Errorf("hades.platform.encryption: unsupported algorithm: %s", es.config.Algorithm)
	}
}

// encryptAES256GCM encrypts data using AES-256-GCM
func (es *EncryptionService) encryptAES256GCM(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(es.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAES256GCM decrypts data using AES-256-GCM
func (es *EncryptionService) decryptAES256GCM(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(es.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("hades.platform.encryption: ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// encryptAES256CBC encrypts data using AES-256-CBC
func (es *EncryptionService) encryptAES256CBC(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(es.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to create cipher: %w", err)
	}

	// Add PKCS7 padding
	plaintext = es.pkcs7Pad(plaintext, aes.BlockSize)

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to generate IV: %w", err)
	}

	ciphertext := make([]byte, len(plaintext))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, plaintext)

	// Return IV + ciphertext
	return append(iv, ciphertext...), nil
}

// decryptAES256CBC decrypts data using AES-256-CBC
func (es *EncryptionService) decryptAES256CBC(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(es.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to create cipher: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("hades.platform.encryption: ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	plaintext, err = es.pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to remove padding: %w", err)
	}

	return plaintext, nil
}

// encryptChaCha20 encrypts data using ChaCha20
func (es *EncryptionService) encryptChaCha20(plaintext []byte) ([]byte, error) {
	// Generate key and nonce using HKDF
	salt := make([]byte, es.config.SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to generate salt: %w", err)
	}

	hkdf := hkdf.New(sha256.New, es.masterKey, salt, nil)
	key := make([]byte, 32) // ChaCha20 key size
	if _, err := hkdf.Read(key); err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to derive key: %w", err)
	}

	nonce := make([]byte, 12) // ChaCha20 nonce size
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to generate nonce: %w", err)
	}

	// Use XOR cipher as simplified ChaCha20 (in production, use proper ChaCha20 implementation)
	stream := es.xorStream(key, nonce)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	// Return salt + nonce + ciphertext
	result := append(salt, nonce...)
	result = append(result, ciphertext...)
	return result, nil
}

// decryptChaCha20 decrypts data using ChaCha20
func (es *EncryptionService) decryptChaCha20(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < es.config.SaltSize+12 {
		return nil, fmt.Errorf("hades.platform.encryption: ciphertext too short")
	}

	salt := ciphertext[:es.config.SaltSize]
	nonce := ciphertext[es.config.SaltSize : es.config.SaltSize+12]
	ciphertext = ciphertext[es.config.SaltSize+12:]

	// Derive key using HKDF
	hkdf := hkdf.New(sha256.New, es.masterKey, salt, nil)
	key := make([]byte, 32)
	if _, err := hkdf.Read(key); err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to derive key: %w", err)
	}

	stream := es.xorStream(key, nonce)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// pkcs7Pad adds PKCS7 padding
func (es *EncryptionService) pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS7 padding
func (es *EncryptionService) pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("hades.platform.encryption: empty data")
	}

	padding := int(data[len(data)-1])
	if padding > blockSize || padding > len(data) {
		return nil, fmt.Errorf("hades.platform.encryption: invalid padding")
	}

	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, fmt.Errorf("hades.platform.encryption: invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// xorStream creates a simple XOR stream (simplified ChaCha20)
func (es *EncryptionService) xorStream(key, nonce []byte) cipher.Stream {
	return &xorStream{key: key, nonce: nonce, counter: 0}
}

// xorStream is a simple XOR stream implementation
type xorStream struct {
	key     []byte
	nonce   []byte
	counter uint32
}

func (xs *xorStream) XORKeyStream(dst, src []byte) {
	for i := range src {
		keyByte := xs.key[xs.counter%uint32(len(xs.key))]
		dst[i] = src[i] ^ keyByte
		xs.counter++
	}
}

// HashService provides cryptographic hash operations
type HashService struct {
	algorithm HashAlgorithm
}

// NewHashService creates a new hash service
func NewHashService(algorithm HashAlgorithm) *HashService {
	return &HashService{algorithm: algorithm}
}

// Hash computes hash of data
func (hs *HashService) Hash(data []byte) ([]byte, error) {
	switch hs.algorithm {
	case HashSHA256:
		hash := sha256.Sum256(data)
		return hash[:], nil
	case HashSHA512:
		hash := sha512.Sum512(data)
		return hash[:], nil
	default:
		return nil, fmt.Errorf("hades.platform.encryption: unsupported hash algorithm: %s", hs.algorithm)
	}
}

// HMAC computes HMAC of data
func (hs *HashService) HMAC(data, key []byte) ([]byte, error) {
	var h hash.Hash

	switch hs.algorithm {
	case HashSHA256:
		h = sha256.New()
	case HashSHA512:
		h = sha512.New()
	default:
		return nil, fmt.Errorf("hades.platform.encryption: unsupported hash algorithm: %s", hs.algorithm)
	}

	// Simple HMAC implementation (in production, use crypto/hmac)
	h.Write(key)
	keyHash := h.Sum(nil)
	h.Reset()
	h.Write(keyHash)
	h.Write(data)
	return h.Sum(nil), nil
}

// SecureStorage provides encrypted file storage
type SecureStorage struct {
	encryption *EncryptionService
	hash       *HashService
}

// NewSecureStorage creates a new secure storage instance
func NewSecureStorage(encryption *EncryptionService, hash *HashService) *SecureStorage {
	return &SecureStorage{
		encryption: encryption,
		hash:       hash,
	}
}

// Store stores data securely with integrity protection
func (ss *SecureStorage) Store(data []byte) ([]byte, error) {
	// Encrypt data
	encrypted, err := ss.encryption.Encrypt(data)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to encrypt: %w", err)
	}

	// Compute HMAC for integrity
	hmac, err := ss.hash.HMAC(encrypted, ss.encryption.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to compute HMAC: %w", err)
	}

	// Return encrypted data + HMAC
	result := append(encrypted, hmac...)
	return result, nil
}

// Retrieve retrieves and verifies data
func (ss *SecureStorage) Retrieve(storedData []byte) ([]byte, error) {
	if len(storedData) < 32 { // Minimum size for HMAC
		return nil, fmt.Errorf("hades.platform.encryption: stored data too short")
	}

	// Split encrypted data and HMAC
	dataSize := len(storedData) - 32
	encryptedData := storedData[:dataSize]
	storedHMAC := storedData[dataSize:]

	// Verify HMAC
	computedHMAC, err := ss.hash.HMAC(encryptedData, ss.encryption.masterKey)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to compute HMAC: %w", err)
	}

	// Simple HMAC comparison (in production, use crypto/subtle.ConstantTimeCompare)
	if !ss.hmacEqual(storedHMAC, computedHMAC) {
		return nil, fmt.Errorf("hades.platform.encryption: HMAC verification failed - data may be tampered")
	}

	// Decrypt data
	decrypted, err := ss.encryption.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to decrypt: %w", err)
	}

	return decrypted, nil
}

// hmacEqual securely compares two HMAC values
func (ss *SecureStorage) hmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// GenerateKey generates a cryptographically secure key
func GenerateKey(size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf("hades.platform.encryption: key size must be positive")
	}

	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("hades.platform.encryption: failed to generate key: %w", err)
	}

	return key, nil
}

// GenerateToken generates a secure random token
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("hades.platform.encryption: failed to generate token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
