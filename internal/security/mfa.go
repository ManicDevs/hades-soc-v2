package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// MFAService provides Multi-Factor Authentication functionality
type MFAService struct {
	issuer string
}

// MFAConfig holds MFA configuration
type MFAConfig struct {
	Issuer        string        `json:"issuer"`
	SecretLength  int           `json:"secret_length"`
	CodeLength    int           `json:"code_length"`
	WindowSize    int           `json:"window_size"`
	BackupCodes   int           `json:"backup_codes"`
	SessionExpiry time.Duration `json:"session_expiry"`
}

// MFAUser holds user MFA data
type MFAUser struct {
	UserID         string    `json:"user_id"`
	Secret         string    `json:"secret"`
	Enabled        bool      `json:"enabled"`
	BackupCodes    []string  `json:"backup_codes"`
	LastUsed       time.Time `json:"last_used"`
	FailedAttempts int       `json:"failed_attempts"`
	LockedUntil    time.Time `json:"locked_until"`
}

// MFASetupResponse contains setup response data
type MFASetupResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
	ManualKey   string   `json:"manual_key"`
}

// NewMFAService creates a new MFA service
func NewMFAService(issuer string) *MFAService {
	return &MFAService{
		issuer: issuer,
	}
}

// GenerateSecret generates a new TOTP secret
func (m *MFAService) GenerateSecret() (string, error) {
	secret := make([]byte, 20) // 160 bits for Base32 encoding
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// GenerateBackupCodes generates backup codes for MFA
func (m *MFAService) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code := make([]byte, 4) // 8 digits
		if _, err := rand.Read(code); err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		// Convert to 8-digit number
		num := int(code[0])<<24 | int(code[1])<<16 | int(code[2])<<8 | int(code[3])
		codes[i] = fmt.Sprintf("%08d", num%100000000)
	}
	return codes, nil
}

// SetupMFA sets up MFA for a user
func (m *MFAService) SetupMFA(userID string) (*MFASetupResponse, error) {
	secret, err := m.GenerateSecret()
	if err != nil {
		return nil, err
	}

	backupCodes, err := m.GenerateBackupCodes(10)
	if err != nil {
		return nil, err
	}

	// Generate QR code URL
	qrURL := m.buildTOTPURL(secret, userID)

	return &MFASetupResponse{
		Secret:      secret,
		QRCode:      qrURL,
		BackupCodes: backupCodes,
		ManualKey:   qrURL,
	}, nil
}

// buildTOTPURL builds the TOTP URL for QR code generation
func (m *MFAService) buildTOTPURL(secret, userID string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=6&algorithm=SHA1",
		url.QueryEscape(m.issuer),
		url.QueryEscape(userID),
		secret,
		url.QueryEscape(m.issuer))
}

// VerifyCode verifies a TOTP code
func (m *MFAService) VerifyCode(secret, code string) bool {
	// Implement TOTP verification manually
	codeInt, err := strconv.Atoi(code)
	if err != nil || len(code) != 6 {
		return false
	}

	// Get current time counter (30-second intervals)
	counter := uint64(time.Now().Unix() / 30)

	// Check current counter and surrounding counters for time drift
	for i := -1; i <= 1; i++ {
		if m.verifyCodeAtCounter(secret, counter+uint64(i), uint64(codeInt)) {
			return true
		}
	}

	return false
}

// verifyCodeAtCounter verifies code at specific counter
func (m *MFAService) verifyCodeAtCounter(secret string, counter, code uint64) bool {
	// Decode secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return false
	}

	// Generate HMAC
	hash := hmac.New(sha1.New, key)
	if err := binary.Write(hash, binary.BigEndian, counter); err != nil {
		return false
	}
	sum := hash.Sum(nil)

	// Dynamic truncation
	offset := int(sum[len(sum)-1] & 0x0f)
	codeVal := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1]&0xff) << 16) |
		(int(sum[offset+2]&0xff) << 8) |
		int(sum[offset+3]&0xff)

	// Get last 6 digits
	return codeVal%1000000 == int(code)
}

// VerifyBackupCode verifies a backup code
func (m *MFAService) VerifyBackupCode(user *MFAUser, code string) bool {
	for i, backupCode := range user.BackupCodes {
		if backupCode == code {
			// Remove used backup code
			user.BackupCodes = append(user.BackupCodes[:i], user.BackupCodes[i+1:]...)
			user.LastUsed = time.Now()
			return true
		}
	}
	return false
}

// IsUserLocked checks if user is locked out
func (m *MFAService) IsUserLocked(user *MFAUser) bool {
	return user.LockedUntil.After(time.Now())
}

// LockUser locks a user for MFA attempts
func (m *MFAService) LockUser(user *MFAUser, duration time.Duration) {
	user.LockedUntil = time.Now().Add(duration)
	user.FailedAttempts = 0
}

// IncrementFailedAttempts increments failed attempt counter
func (m *MFAService) IncrementFailedAttempts(user *MFAUser) {
	user.FailedAttempts++

	// Lock after 5 failed attempts for 15 minutes
	if user.FailedAttempts >= 5 {
		m.LockUser(user, 15*time.Minute)
	}
}

// ResetFailedAttempts resets failed attempt counter
func (m *MFAService) ResetFailedAttempts(user *MFAUser) {
	user.FailedAttempts = 0
	user.LockedUntil = time.Time{}
}

// ValidateMFA validates MFA for a user
func (m *MFAService) ValidateMFA(user *MFAUser, code string) (bool, string) {
	if !user.Enabled {
		return false, "MFA not enabled for user"
	}

	if m.IsUserLocked(user) {
		return false, "User is locked out due to too many failed attempts"
	}

	// Check if it's a backup code first
	if len(code) == 8 && strings.HasPrefix(code, "0") {
		if m.VerifyBackupCode(user, code) {
			m.ResetFailedAttempts(user)
			return true, "Backup code verified"
		}
		m.IncrementFailedAttempts(user)
		return false, "Invalid backup code"
	}

	// Check TOTP code
	if m.VerifyCode(user.Secret, code) {
		m.ResetFailedAttempts(user)
		user.LastUsed = time.Now()
		return true, "TOTP code verified"
	}

	m.IncrementFailedAttempts(user)
	return false, "Invalid TOTP code"
}

// HashSecret hashes the MFA secret for storage
func (m *MFAService) HashSecret(secret string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}
	return string(hash), nil
}

// VerifyHashedSecret verifies a hashed secret
func (m *MFAService) VerifyHashedSecret(hashedSecret, secret string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secret))
	return err == nil
}

// GetMFASummary returns MFA summary for a user
func (m *MFAService) GetMFASummary(user *MFAUser) map[string]interface{} {
	return map[string]interface{}{
		"enabled":           user.Enabled,
		"last_used":         user.LastUsed,
		"failed_attempts":   user.FailedAttempts,
		"locked_until":      user.LockedUntil,
		"backup_codes_left": len(user.BackupCodes),
		"is_locked":         m.IsUserLocked(user),
	}
}

// GenerateRecoveryCodes generates new recovery codes
func (m *MFAService) GenerateRecoveryCodes(count int) ([]string, error) {
	return m.GenerateBackupCodes(count)
}

// EnableMFA enables MFA for a user
func (m *MFAService) EnableMFA(user *MFAUser, secret string, backupCodes []string) error {
	hashedSecret, err := m.HashSecret(secret)
	if err != nil {
		return err
	}

	user.Secret = hashedSecret
	user.BackupCodes = backupCodes
	user.Enabled = true
	user.FailedAttempts = 0
	user.LockedUntil = time.Time{}
	user.LastUsed = time.Time{}

	return nil
}

// DisableMFA disables MFA for a user
func (m *MFAService) DisableMFA(user *MFAUser) {
	user.Enabled = false
	user.Secret = ""
	user.BackupCodes = nil
	user.FailedAttempts = 0
	user.LockedUntil = time.Time{}
	user.LastUsed = time.Time{}
}

// GenerateQRCodeImage generates QR code image data
func (m *MFAService) GenerateQRCodeImage(secret, userID string) ([]byte, error) {
	// For now return the URL as text - in production you'd use a QR code library
	qrURL := m.buildTOTPURL(secret, userID)
	return []byte(qrURL), nil
}

// ValidateTOTPSetup validates TOTP setup during enrollment
func (m *MFAService) ValidateTOTPSetup(secret, code string) bool {
	return m.VerifyCode(secret, code)
}

// GetRemainingBackupCodes returns remaining backup codes count
func (m *MFAService) GetRemainingBackupCodes(user *MFAUser) int {
	return len(user.BackupCodes)
}

// RegenerateBackupCodes regenerates backup codes for a user
func (m *MFAService) RegenerateBackupCodes(user *MFAUser) ([]string, error) {
	codes, err := m.GenerateBackupCodes(10)
	if err != nil {
		return nil, err
	}
	user.BackupCodes = codes
	return codes, nil
}

// ExportMFAData exports user MFA data (for backup/migration)
func (m *MFAService) ExportMFAData(user *MFAUser) map[string]interface{} {
	return map[string]interface{}{
		"user_id":           user.UserID,
		"enabled":           user.Enabled,
		"last_used":         user.LastUsed,
		"failed_attempts":   user.FailedAttempts,
		"locked_until":      user.LockedUntil,
		"backup_codes":      user.BackupCodes,
		"backup_codes_left": len(user.BackupCodes),
		"mfa_summary":       m.GetMFASummary(user),
	}
}

// ImportMFAData imports user MFA data (for backup/migration)
func (m *MFAService) ImportMFAData(data map[string]interface{}) (*MFAUser, error) {
	user := &MFAUser{
		UserID:  data["user_id"].(string),
		Enabled: data["enabled"].(bool),
	}

	if lastUsed, ok := data["last_used"].(string); ok && lastUsed != "" {
		if t, err := time.Parse(time.RFC3339, lastUsed); err == nil {
			user.LastUsed = t
		}
	}

	if lockedUntil, ok := data["locked_until"].(string); ok && lockedUntil != "" {
		if t, err := time.Parse(time.RFC3339, lockedUntil); err == nil {
			user.LockedUntil = t
		}
	}

	if failedAttempts, ok := data["failed_attempts"].(int); ok {
		user.FailedAttempts = failedAttempts
	}

	if backupCodes, ok := data["backup_codes"].([]string); ok {
		user.BackupCodes = backupCodes
	}

	return user, nil
}
