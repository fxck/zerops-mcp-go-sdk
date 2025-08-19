package transport

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPServerConfig contains configuration for the HTTP server
type HTTPServerConfig struct {
	Host   string
	Port   string
	Server *mcp.Server
}

// StartHTTPServer starts the HTTP server with SSE support
func StartHTTPServer(ctx context.Context, config HTTPServerConfig) error {
	// Create the new HTTP handler that manages API keys per request
	httpHandler := NewHTTPSharedHandler(config.Server)

	// Create HTTP mux for multiple endpoints
	mux := http.NewServeMux()

	// Main MCP endpoint
	mux.Handle("/mcp", httpHandler)

	// Health check endpoint
	mux.HandleFunc("/health", healthCheckHandler)

	// CORS and security headers middleware
	wrappedMux := securityMiddleware(mux)

	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	fmt.Fprintf(os.Stderr, "Starting HTTP server on %s...\n", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: wrappedMux,
	}

	// Graceful shutdown on context cancellation
	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}

// createDirectHandler creates a direct HTTP handler with authentication
func createDirectHandler(config HTTPServerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request details
		log.Printf("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Validate Origin header for security (DNS rebinding protection)
		origin := r.Header.Get("Origin")
		if origin != "" && !isValidOrigin(origin, config.Host) {
			http.Error(w, "Invalid origin", http.StatusForbidden)
			return
		}

		// Validate authentication using ZEROPS_API_KEY (if configured)
		expectedToken := os.Getenv("ZEROPS_API_KEY")
		if expectedToken != "" {
			authHeader := r.Header.Get("Authorization")
			if !validateBearerToken(authHeader, expectedToken) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Use the direct handler
		DirectHTTPHandler(config.Server)(w, r)
	})
}

// createAuthenticatedHandler creates an SSE handler with authentication (deprecated)
func createAuthenticatedHandler(config HTTPServerConfig) http.Handler {
	sseHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		// Log the request for debugging
		log.Printf("SSE Handler: %s %s", request.Method, request.URL.Path)

		// Authentication is handled via ZEROPS_API_KEY which is already
		// validated when creating the Zerops client in main.go
		// The Bearer token in Authorization header should contain the same ZEROPS_API_KEY
		expectedToken := os.Getenv("ZEROPS_API_KEY")

		// Only validate if token is configured
		if expectedToken != "" {
			authHeader := request.Header.Get("Authorization")
			if !validateBearerToken(authHeader, expectedToken) {
				log.Printf("Authentication failed for request from %s", request.RemoteAddr)
				return nil // Will result in 404
			}
		}

		// Always return the server for the /mcp endpoint
		if request.URL.Path == "/mcp" {
			return config.Server
		}

		return config.Server // Return server for any path
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request details
		log.Printf("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Check method
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate Origin header for security (DNS rebinding protection)
		origin := r.Header.Get("Origin")
		if origin != "" && !isValidOrigin(origin, config.Host) {
			http.Error(w, "Invalid origin", http.StatusForbidden)
			return
		}

		// Validate authentication using ZEROPS_API_KEY (if configured)
		expectedToken := os.Getenv("ZEROPS_API_KEY")
		if expectedToken != "" {
			authHeader := r.Header.Get("Authorization")
			if !validateBearerToken(authHeader, expectedToken) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Delegate to SSE handler
		sseHandler.ServeHTTP(w, r)
	})
}

// validateBearerToken validates the Bearer token
func validateBearerToken(authHeader, expectedToken string) bool {
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	return parts[1] == expectedToken
}

// isValidOrigin validates the Origin header
func isValidOrigin(origin, host string) bool {
	// For localhost, allow same origin
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		return strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")
	}

	// For production, implement stricter validation
	// This is a placeholder - adjust based on your security requirements
	return true
}

// healthCheckHandler handles health check requests
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"zerops-mcp","transport":"http"}`))
}

// securityMiddleware adds security headers
func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers for API access
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, MCP-Protocol-Version, Mcp-Session-Id, Accept")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		next.ServeHTTP(w, r)
	})
}
