package auxiliary

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"hades-v2/pkg/sdk"
)

// APIServerFixed represents a working API server
type APIServerFixed struct {
	*sdk.BaseModule
	port      int
	authToken string
	server    *http.Server
}

// NewAPIServerFixed creates a new API server instance
func NewAPIServerFixed() *APIServerFixed {
	return &APIServerFixed{
		BaseModule: sdk.NewBaseModule(
			"api_server_fixed",
			"REST API server for external integrations (FIXED)",
			sdk.CategoryReporting,
		),
	}
}

// Execute starts the API server
func (as *APIServerFixed) Execute(ctx context.Context) error {
	as.SetStatus(sdk.StatusRunning)
	defer as.SetStatus(sdk.StatusIdle)

	if err := as.validateConfig(); err != nil {
		return fmt.Errorf("hades.auxiliary.api_server_fixed: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", as.healthHandler)
	mux.HandleFunc("/api/modules", as.modulesHandler)

	// Create middleware chain
	handler := as.authMiddleware(mux)

	as.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", as.port),
		Handler: handler,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("API server listening on port %d\n", as.port)
		if err := as.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- fmt.Errorf("hades.auxiliary.api_server_fixed: %w", err)
		}
		close(serverErr)
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		fmt.Printf("Shutting down API server\n")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := as.server.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Warning: failed to shutdown server: %v\n", err)
		}
		return ctx.Err()
	case err := <-serverErr:
		return err
	}
}

// SetPort configures the server port
func (as *APIServerFixed) SetPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("hades.auxiliary.api_server_fixed: port must be between 1 and 65535")
	}
	as.port = port
	return nil
}

// SetAuthToken configures the authentication token
func (as *APIServerFixed) SetAuthToken(token string) error {
	if token == "" {
		return fmt.Errorf("hades.auxiliary.api_server_fixed: auth token cannot be empty")
	}
	as.authToken = token
	return nil
}

// GetResult returns server status
func (as *APIServerFixed) GetResult() string {
	return fmt.Sprintf("API server running on port %d", as.port)
}

// validateConfig ensures server configuration is valid
func (as *APIServerFixed) validateConfig() error {
	if as.port == 0 {
		return fmt.Errorf("hades.auxiliary.api_server_fixed: port not configured")
	}
	if as.authToken == "" {
		return fmt.Errorf("hades.auxiliary.api_server_fixed: auth token not configured")
	}
	return nil
}

// authMiddleware provides token-based authentication
func (as *APIServerFixed) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Debug logging
		if _, err := fmt.Fprintf(os.Stderr, "DEBUG: Request received for %s\n", r.URL.Path); err != nil {
			fmt.Printf("Warning: failed to write debug log: %v\n", err)
		}
		if _, err := fmt.Fprintf(os.Stderr, "DEBUG: Method: %s\n", r.Method); err != nil {
			fmt.Printf("Warning: failed to write debug log: %v\n", err)
		}

		// Extract token from multiple possible headers
		var token string
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				token = authHeader
			}
		} else if token = r.Header.Get("X-API-Token"); token == "" {
			token = r.Header.Get("Token")
		}

		if _, err := fmt.Fprintf(os.Stderr, "DEBUG: Auth token received: '%s', expected: '%s'\n", token, as.authToken); err != nil {
			fmt.Printf("Warning: failed to write debug log: %v\n", err)
		}

		if token != as.authToken {
			if _, err := fmt.Fprintf(os.Stderr, "DEBUG: Authentication failed!\n"); err != nil {
				fmt.Printf("Warning: failed to write debug log: %v\n", err)
			}
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if _, err := fmt.Fprintf(os.Stderr, "DEBUG: Authentication successful!\n"); err != nil {
			fmt.Printf("Warning: failed to write debug log: %v\n", err)
		}
		next.ServeHTTP(w, r)
	})
}

// healthHandler returns health status
func (as *APIServerFixed) healthHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(os.Stderr, "DEBUG: Health handler called\n"); err != nil {
		fmt.Printf("Warning: failed to write debug log: %v\n", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","server":"api_server_fixed"}`, time.Now().UTC().Format(time.RFC3339)); err != nil {
		fmt.Printf("Warning: failed to write health response: %v\n", err)
	}
}

// modulesHandler returns module information
func (as *APIServerFixed) modulesHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(os.Stderr, "DEBUG: Modules handler called\n"); err != nil {
		fmt.Printf("Warning: failed to write debug log: %v\n", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, `{"modules":["tcp_scanner","cloud_scanner","reverse_shell","api_server_fixed"],"count":4}`); err != nil {
		fmt.Printf("Warning: failed to write modules response: %v\n", err)
	}
}
