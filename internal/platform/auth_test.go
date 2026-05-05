package platform

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret-key-for-testing-only"

// Test JWT token generation and validation
func TestJWTTokenGenerationAndValidation(t *testing.T) {
	t.Run("GenerateAndValidateValidToken", func(t *testing.T) {
		// Create test user
		user := &User{
			ID:       "test-user-id",
			Username: "testuser",
			Email:    "test@example.com",
			Role:     RoleAdmin,
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     string(user.Role),
			"exp":      time.Now().Add(time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte(testJWTSecret))
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}
		if tokenString == "" {
			t.Fatal("Token string is empty")
		}

		// Validate token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(testJWTSecret), nil
		})

		if err != nil {
			t.Fatalf("Failed to parse token: %v", err)
		}
		if !parsedToken.Valid {
			t.Fatal("Token is not valid")
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("Claims are not of type MapClaims")
		}

		if user.ID != claims["user_id"] {
			t.Errorf("Expected user ID %s, got %v", user.ID, claims["user_id"])
		}
		if user.Username != claims["username"] {
			t.Errorf("Expected username %s, got %v", user.Username, claims["username"])
		}
		if string(user.Role) != claims["role"] {
			t.Errorf("Expected role %s, got %v", string(user.Role), claims["role"])
		}
	})

	t.Run("RejectExpiredToken", func(t *testing.T) {
		// Create expired token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  "test-user-id",
			"username": "testuser",
			"exp":      time.Now().Add(-time.Hour).Unix(), // Expired
		})

		tokenString, err := token.SignedString([]byte(testJWTSecret))
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}

		// Try to validate expired token
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(testJWTSecret), nil
		})

		if err == nil {
			t.Fatal("Expected error for expired token, got nil")
		}
		if parsedToken.Valid {
			t.Fatal("Expected token to be invalid")
		}
		if !strings.Contains(err.Error(), "token is expired") {
			t.Errorf("Expected 'token is expired' error, got: %v", err.Error())
		}
	})

	t.Run("RejectInvalidSignature", func(t *testing.T) {
		// Create token with wrong secret
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  "test-user-id",
			"username": "testuser",
			"exp":      time.Now().Add(time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte("wrong-secret"))
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}

		// Try to validate with correct secret
		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(testJWTSecret), nil
		})

		if err == nil {
			t.Fatal("Expected error for invalid signature, got nil")
		}
		if parsedToken.Valid {
			t.Fatal("Expected token to be invalid")
		}
		// JWT library may return different error types for signature validation
		if !strings.Contains(err.Error(), "signature") {
			t.Errorf("Expected signature error, got: %v", err)
		}
	})
}

// Test AuthManager functionality
func TestAuthManager(t *testing.T) {
	ctx := context.Background()
	config := &AuthConfig{
		SessionTimeout:    1 * time.Hour,
		MaxFailedAttempts: 3,
		LockoutDuration:   5 * time.Minute,
		RequireMFA:        false,
	}

	t.Run("SuccessfulLogin", func(t *testing.T) {
		am := NewAuthManager(config)

		// Create user
		user, err := am.CreateUser(ctx, "testuser", "test@example.com", "password123", RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if "testuser" != user.Username {
			t.Errorf("Expected username testuser, got %s", user.Username)
		}
		if "test@example.com" != user.Email {
			t.Errorf("Expected email test@example.com, got %s", user.Email)
		}
		if RoleAdmin != user.Role {
			t.Errorf("Expected role admin, got %s", user.Role)
		}
		if !user.IsActive {
			t.Error("Expected user to be active")
		}

		// Authenticate
		session, err := am.Authenticate(ctx, "testuser", "password123", "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to authenticate: %v", err)
		}
		if session.Token == "" {
			t.Error("Session token is empty")
		}
		if user.ID != session.UserID {
			t.Errorf("Expected user ID %s, got %s", user.ID, session.UserID)
		}
		if time.Now().After(session.ExpiresAt) {
			t.Error("Session expiration time is in the past")
		}

		// Validate session
		validatedUser, err := am.ValidateSession(ctx, session.Token)
		if err != nil {
			t.Fatalf("Failed to validate session: %v", err)
		}
		if user.ID != validatedUser.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, validatedUser.ID)
		}
		if user.Username != validatedUser.Username {
			t.Errorf("Expected username %s, got %s", user.Username, validatedUser.Username)
		}
	})

	t.Run("FailedLoginWrongPassword", func(t *testing.T) {
		am := NewAuthManager(config)

		// Create user
		_, err := am.CreateUser(ctx, "testuser", "test@example.com", "password123", RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Try to authenticate with wrong password
		session, err := am.Authenticate(ctx, "testuser", "wrongpassword", "127.0.0.1", "test-agent")
		if err == nil {
			t.Fatal("Expected error for wrong password, got nil")
		}
		if session != nil {
			t.Error("Expected nil session for wrong password")
		}
		if !strings.Contains(err.Error(), "invalid credentials") {
			t.Errorf("Expected 'invalid credentials' error, got: %v", err.Error())
		}
	})

	t.Run("AccountLockoutAfterFailedAttempts", func(t *testing.T) {
		am := NewAuthManager(config)

		// Create user
		_, err := am.CreateUser(ctx, "testuser", "test@example.com", "password123", RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Fail login MaxFailedAttempts times
		for i := 0; i < config.MaxFailedAttempts; i++ {
			session, err := am.Authenticate(ctx, "testuser", "wrongpassword", "127.0.0.1", "test-agent")
			if err == nil {
				t.Fatal("Expected error for wrong password, got nil")
			}
			if session != nil {
				t.Error("Expected nil session for wrong password")
			}
			if i < config.MaxFailedAttempts-1 {
				if !strings.Contains(err.Error(), "invalid credentials") {
					t.Errorf("Expected 'invalid credentials' error, got: %v", err.Error())
				}
			}
		}

		// Account should now be locked
		session, err := am.Authenticate(ctx, "testuser", "password123", "127.0.0.1", "test-agent")
		if err == nil {
			t.Fatal("Expected error for locked account, got nil")
		}
		if session != nil {
			t.Error("Expected nil session for locked account")
		}
		if !strings.Contains(err.Error(), "is locked") {
			t.Errorf("Expected 'is locked' error, got: %v", err.Error())
		}

		// Even correct password should fail
		session, err = am.Authenticate(ctx, "testuser", "password123", "127.0.0.1", "test-agent")
		if err == nil {
			t.Fatal("Expected error for locked account with correct password, got nil")
		}
		if session != nil {
			t.Error("Expected nil session for locked account with correct password")
		}
		if !strings.Contains(err.Error(), "is locked") {
			t.Errorf("Expected 'is locked' error, got: %v", err.Error())
		}
	})

	t.Run("SessionCleanup", func(t *testing.T) {
		// Create auth manager with short session timeout
		shortConfig := &AuthConfig{
			SessionTimeout:    100 * time.Millisecond,
			MaxFailedAttempts: 3,
			LockoutDuration:   5 * time.Minute,
			RequireMFA:        false,
		}
		am := NewAuthManager(shortConfig)

		// Create user and authenticate
		_, err := am.CreateUser(ctx, "testuser", "test@example.com", "password123", RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		session, err := am.Authenticate(ctx, "testuser", "password123", "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to authenticate: %v", err)
		}

		// Session should be valid immediately
		_, err = am.ValidateSession(ctx, session.Token)
		if err != nil {
			t.Fatalf("Failed to validate session immediately: %v", err)
		}

		// Wait for session to expire
		time.Sleep(shortConfig.SessionTimeout + 50*time.Millisecond)

		// Session should now be invalid
		_, err = am.ValidateSession(ctx, session.Token)
		if err == nil {
			t.Fatal("Expected error for expired session, got nil")
		}
		if !strings.Contains(err.Error(), "session expired") {
			t.Errorf("Expected 'session expired' error, got: %v", err.Error())
		}

		// Run cleanup to remove expired sessions
		am.CleanupExpiredSessions(ctx)

		// Session should still be invalid after cleanup
		_, err = am.ValidateSession(ctx, session.Token)
		if err == nil {
			t.Fatal("Expected error for expired session after cleanup, got nil")
		}
		if !strings.Contains(err.Error(), "invalid session token") {
			t.Errorf("Expected 'invalid session token' error, got: %v", err.Error())
		}
	})

	t.Run("InvalidateSession", func(t *testing.T) {
		am := NewAuthManager(config)

		// Create user and authenticate
		_, err := am.CreateUser(ctx, "testuser", "test@example.com", "password123", RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		session, err := am.Authenticate(ctx, "testuser", "password123", "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to authenticate: %v", err)
		}

		// Session should be valid
		_, err = am.ValidateSession(ctx, session.Token)
		if err != nil {
			t.Fatalf("Failed to validate session: %v", err)
		}

		// Invalidate session
		err = am.InvalidateSession(ctx, session.Token)
		if err != nil {
			t.Fatalf("Failed to invalidate session: %v", err)
		}

		// Session should now be invalid
		_, err = am.ValidateSession(ctx, session.Token)
		if err == nil {
			t.Fatal("Expected error for invalidated session, got nil")
		}
		if !strings.Contains(err.Error(), "invalid session token") {
			t.Errorf("Expected 'invalid session token' error, got: %v", err.Error())
		}
	})

	t.Run("UserAuthorization", func(t *testing.T) {
		am := NewAuthManager(config)

		// Create users with different roles
		viewerUser := &User{Role: RoleViewer}
		operatorUser := &User{Role: RoleOperator}
		adminUser := &User{Role: RoleAdmin}
		rootUser := &User{Role: RoleRoot}

		// Test role hierarchy
		if am.Authorize(viewerUser, RoleOperator) {
			t.Error("Viewer should not be authorized for operator role")
		}
		if !am.Authorize(viewerUser, RoleViewer) {
			t.Error("Viewer should be authorized for viewer role")
		}

		if !am.Authorize(operatorUser, RoleViewer) {
			t.Error("Operator should be authorized for viewer role")
		}
		if !am.Authorize(operatorUser, RoleOperator) {
			t.Error("Operator should be authorized for operator role")
		}
		if am.Authorize(operatorUser, RoleAdmin) {
			t.Error("Operator should not be authorized for admin role")
		}

		if !am.Authorize(adminUser, RoleViewer) {
			t.Error("Admin should be authorized for viewer role")
		}
		if !am.Authorize(adminUser, RoleOperator) {
			t.Error("Admin should be authorized for operator role")
		}
		if !am.Authorize(adminUser, RoleAdmin) {
			t.Error("Admin should be authorized for admin role")
		}
		if am.Authorize(adminUser, RoleRoot) {
			t.Error("Admin should not be authorized for root role")
		}

		if !am.Authorize(rootUser, RoleViewer) {
			t.Error("Root should be authorized for viewer role")
		}
		if !am.Authorize(rootUser, RoleOperator) {
			t.Error("Root should be authorized for operator role")
		}
		if !am.Authorize(rootUser, RoleAdmin) {
			t.Error("Root should be authorized for admin role")
		}
		if !am.Authorize(rootUser, RoleRoot) {
			t.Error("Root should be authorized for root role")
		}
	})

	t.Run("InactiveUserCannotLogin", func(t *testing.T) {
		am := NewAuthManager(config)

		// Create user
		user, err := am.CreateUser(ctx, "testuser", "test@example.com", "password123", RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Deactivate user
		user.IsActive = false
		am.users["testuser"] = user

		// Try to authenticate with inactive user
		session, err := am.Authenticate(ctx, "testuser", "password123", "127.0.0.1", "test-agent")
		if err == nil {
			t.Fatal("Expected error for inactive user, got nil")
		}
		if session != nil {
			t.Error("Expected nil session for inactive user")
		}
		if !strings.Contains(err.Error(), "account is inactive") {
			t.Errorf("Expected 'account is inactive' error, got: %v", err.Error())
		}
	})

	t.Run("NonExistentUserCannotLogin", func(t *testing.T) {
		am := NewAuthManager(config)

		// Try to authenticate with non-existent user
		session, err := am.Authenticate(ctx, "nonexistent", "password123", "127.0.0.1", "test-agent")
		if err == nil {
			t.Fatal("Expected error for non-existent user, got nil")
		}
		if session != nil {
			t.Error("Expected nil session for non-existent user")
		}
		if !strings.Contains(err.Error(), "invalid credentials") {
			t.Errorf("Expected 'invalid credentials' error, got: %v", err.Error())
		}
	})
}

// Test password hashing and verification
func TestPasswordHashing(t *testing.T) {
	am := NewAuthManager(nil)

	t.Run("HashAndVerifyPassword", func(t *testing.T) {
		password := "test-password-123"

		// Hash password
		hash, err := am.hashPassword(password)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}
		if hash == "" {
			t.Fatal("Hash is empty")
		}
		if password == hash {
			t.Error("Hash should not equal password")
		}

		// Verify correct password
		isValid := am.verifyPassword(password, hash)
		if !isValid {
			t.Error("Password verification should succeed for correct password")
		}

		// Verify wrong password
		isValid = am.verifyPassword("wrong-password", hash)
		if isValid {
			t.Error("Password verification should fail for wrong password")
		}
	})

	t.Run("DifferentHashesForSamePassword", func(t *testing.T) {
		password := "test-password-123"

		// Hash password twice
		hash1, err1 := am.hashPassword(password)
		hash2, err2 := am.hashPassword(password)

		if err1 != nil {
			t.Fatalf("Failed to hash password first time: %v", err1)
		}
		if err2 != nil {
			t.Fatalf("Failed to hash password second time: %v", err2)
		}
		if hash1 == hash2 {
			t.Error("Hashes should be different due to random salt")
		}

		// But both should verify the same password
		if !am.verifyPassword(password, hash1) {
			t.Error("First hash should verify password")
		}
		if !am.verifyPassword(password, hash2) {
			t.Error("Second hash should verify password")
		}
	})
}

// Test token and ID generation
func TestTokenAndIDGeneration(t *testing.T) {
	am := NewAuthManager(nil)

	t.Run("GenerateUniqueTokens", func(t *testing.T) {
		token1 := am.generateToken()
		token2 := am.generateToken()

		if token1 == "" {
			t.Error("Token1 is empty")
		}
		if token2 == "" {
			t.Error("Token2 is empty")
		}
		if token1 == token2 {
			t.Error("Tokens should be different")
		}
		// Tokens should be reasonable length (base64 encoded 32 bytes)
		if len(token1) < 20 || len(token1) > 50 {
			t.Errorf("Token1 length %d is unreasonable", len(token1))
		}
		if len(token2) < 20 || len(token2) > 50 {
			t.Errorf("Token2 length %d is unreasonable", len(token2))
		}
	})

	t.Run("GenerateUniqueIDs", func(t *testing.T) {
		id1 := am.generateID()
		id2 := am.generateID()

		if id1 == "" {
			t.Error("ID1 is empty")
		}
		if id2 == "" {
			t.Error("ID2 is empty")
		}
		if id1 == id2 {
			t.Error("IDs should be different")
		}
		// IDs should be reasonable length (base64 encoded 16 bytes)
		if len(id1) < 10 || len(id1) > 30 {
			t.Errorf("ID1 length %d is unreasonable", len(id1))
		}
		if len(id2) < 10 || len(id2) > 30 {
			t.Errorf("ID2 length %d is unreasonable", len(id2))
		}
	})
}

// Test edge cases and error conditions
func TestAuthManagerEdgeCases(t *testing.T) {
	ctx := context.Background()
	am := NewAuthManager(nil)

	t.Run("CreateUserWithEmptyFields", func(t *testing.T) {
		// Empty username
		user, err := am.CreateUser(ctx, "", "test@example.com", "password123", RoleAdmin)
		if err == nil {
			t.Fatal("Expected error for empty username, got nil")
		}
		if user != nil {
			t.Error("Expected nil user for empty username")
		}
		if !strings.Contains(err.Error(), "username, email, and password required") {
			t.Errorf("Expected 'username, email, and password required' error, got: %v", err.Error())
		}

		// Empty email
		user, err = am.CreateUser(ctx, "testuser", "", "password123", RoleAdmin)
		if err == nil {
			t.Fatal("Expected error for empty email, got nil")
		}
		if user != nil {
			t.Error("Expected nil user for empty email")
		}
		if !strings.Contains(err.Error(), "username, email, and password required") {
			t.Errorf("Expected 'username, email, and password required' error, got: %v", err.Error())
		}

		// Empty password
		user, err = am.CreateUser(ctx, "testuser", "test@example.com", "", RoleAdmin)
		if err == nil {
			t.Fatal("Expected error for empty password, got nil")
		}
		if user != nil {
			t.Error("Expected nil user for empty password")
		}
		if !strings.Contains(err.Error(), "username, email, and password required") {
			t.Errorf("Expected 'username, email, and password required' error, got: %v", err.Error())
		}
	})

	t.Run("CreateDuplicateUser", func(t *testing.T) {
		// Create first user
		user1, err := am.CreateUser(ctx, "testuser", "test1@example.com", "password123", RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}
		if user1 == nil {
			t.Fatal("First user is nil")
		}

		// Try to create duplicate user
		user2, err := am.CreateUser(ctx, "testuser", "test2@example.com", "password456", RoleOperator)
		if err == nil {
			t.Fatal("Expected error for duplicate user, got nil")
		}
		if user2 != nil {
			t.Error("Expected nil user for duplicate user")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected 'already exists' error, got: %v", err.Error())
		}
	})

	t.Run("ValidateInvalidSession", func(t *testing.T) {
		// Validate non-existent session
		user, err := am.ValidateSession(ctx, "non-existent-token")
		if err == nil {
			t.Fatal("Expected error for non-existent session, got nil")
		}
		if user != nil {
			t.Error("Expected nil user for non-existent session")
		}
		if !strings.Contains(err.Error(), "invalid session token") {
			t.Errorf("Expected 'invalid session token' error, got: %v", err.Error())
		}
	})

	t.Run("InvalidateNonExistentSession", func(t *testing.T) {
		// Try to invalidate non-existent session
		err := am.InvalidateSession(ctx, "non-existent-token")
		if err == nil {
			t.Fatal("Expected error for non-existent session, got nil")
		}
		if !strings.Contains(err.Error(), "session not found") {
			t.Errorf("Expected 'session not found' error, got: %v", err.Error())
		}
	})
}
