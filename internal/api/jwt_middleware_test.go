package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func resetAuthForTest() {
	authConfigOnce = sync.Once{}
	authConfigErr = nil
	jwtSecret = nil
	devPasswords = nil
}

func signedToken(t *testing.T, secret string, exp time.Time) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  1,
		"username": "admin",
		"role":     "Administrator",
		"exp":      exp.Unix(),
	})
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func TestGetClaimsFromRequestValid(t *testing.T) {
	resetAuthForTest()
	t.Setenv("HADES_JWT_SECRET", "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, "test-secret", time.Now().Add(time.Hour)))

	claims, ok := getClaimsFromRequest(req)
	if !ok {
		t.Fatalf("expected claims to be extracted")
	}
	if claims.Username != "admin" {
		t.Fatalf("expected admin claims, got %s", claims.Username)
	}
}

func TestGetClaimsFromRequestExpired(t *testing.T) {
	resetAuthForTest()
	t.Setenv("HADES_JWT_SECRET", "test-secret")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, "test-secret", time.Now().Add(-time.Minute)))

	_, ok := getClaimsFromRequest(req)
	if ok {
		t.Fatalf("expected expired token to be rejected")
	}
}

func TestJWTMiddlewareRejectsMalformedHeader(t *testing.T) {
	resetAuthForTest()
	t.Setenv("HADES_JWT_SECRET", "test-secret")
	srv := &Server{}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/protected", nil)
	req.Header.Set("Authorization", "invalid-header")

	srv.JWTMiddleware(next).ServeHTTP(rr, req)
	if called {
		t.Fatalf("expected next handler not to be called")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", rr.Code)
	}
}

func TestRequireAdminRejectsNonAdmin(t *testing.T) {
	srv := &Server{}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", "Security Analyst"))

	srv.RequireAdmin(next).ServeHTTP(rr, req)
	if called {
		t.Fatalf("expected next handler not to be called for non-admin")
	}
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", rr.Code)
	}
}
