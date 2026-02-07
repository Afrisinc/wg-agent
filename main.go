// Package main provides the WireGuard Agent API server
// @title WireGuard Agent API
// @version 1.0.0
// @description Secure HTTP API for managing WireGuard peers
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:9999
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/afrisinc/wg-agent/internal/config"
	"github.com/afrisinc/wg-agent/internal/wireguard"
)

type AddPeerRequest struct {
	PublicKey string `json:"public_key"`
	AllowedIP string `json:"allowed_ip"`
	Endpoint  string `json:"endpoint,omitempty"`
}

type RemovePeerRequest struct {
	PublicKey string `json:"public_key"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Validation regexes
var (
	publicKeyRegex = regexp.MustCompile(`^[A-Za-z0-9+/]{43}=$`)
	ipOctetRegex   = regexp.MustCompile(`^\d{1,3}$`)
)

// Middleware to check API key
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" || apiKey != os.Getenv("API_KEY") {
			log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or missing API key",
			})
			return
		}
		next(w, r)
	}
}

// RequestLogger middleware
func requestLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next(w, r)
	}
}

func validatePublicKey(key string) bool {
	return publicKeyRegex.MatchString(key)
}

func validateIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if !ipOctetRegex.MatchString(part) {
			return false
		}
		// Validate octet is between 0-255
		var octet int
		if _, err := fmt.Sscanf(part, "%d", &octet); err != nil || octet > 255 {
			return false
		}
	}
	return true
}

// addPeerHandler adds a new peer to WireGuard
// @Summary Add a new WireGuard peer
// @Description Add a new peer to the WireGuard interface with specified public key and allowed IP
// @Tags Peers
// @Accept json
// @Produce json
// @Param Authorization header string true "API Key" default(Bearer your-api-key)
// @Param request body AddPeerRequest true "Peer configuration"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Server error"
// @Security ApiKeyAuth
// @Router /add-peer [post]
func addPeerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AddPeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	// Validate input
	if !validatePublicKey(req.PublicKey) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_public_key",
			Message: "Public key format is invalid",
		})
		return
	}

	if !validateIP(req.AllowedIP) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_ip",
			Message: "IP address format is invalid",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := wireguard.AddPeer(ctx, req.PublicKey, req.AllowedIP); err != nil {
		log.Printf("Error adding peer: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "add_peer_failed",
			Message: "Failed to add peer to WireGuard",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SuccessResponse{
		Status:  "success",
		Message: fmt.Sprintf("Peer %s added successfully", req.PublicKey[:8]+"..."),
	})
}

// removePeerHandler removes a peer from WireGuard
// @Summary Remove a WireGuard peer
// @Description Remove a peer from the WireGuard interface by public key
// @Tags Peers
// @Accept json
// @Produce json
// @Param Authorization header string true "API Key" default(Bearer your-api-key)
// @Param request body RemovePeerRequest true "Peer public key"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Server error"
// @Security ApiKeyAuth
// @Router /remove-peer [post]
func removePeerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RemovePeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	if !validatePublicKey(req.PublicKey) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_public_key",
			Message: "Public key format is invalid",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := wireguard.RemovePeer(ctx, req.PublicKey); err != nil {
		log.Printf("Error removing peer: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "remove_peer_failed",
			Message: "Failed to remove peer from WireGuard",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SuccessResponse{
		Status:  "success",
		Message: fmt.Sprintf("Peer %s removed successfully", req.PublicKey[:8]+"..."),
	})
}

// statusHandler returns WireGuard interface status
// @Summary Get WireGuard status
// @Description Get the current status of the WireGuard interface including all connected peers
// @Tags Status
// @Produce plain
// @Param Authorization header string true "API Key" default(Bearer your-api-key)
// @Success 200 {string} string "WireGuard status output"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Server error"
// @Security ApiKeyAuth
// @Router /status [get]
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := wireguard.GetStatus(ctx)
	if err != nil {
		log.Printf("Error getting status: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
		return
	}

	_, _ = w.Write(output)
}

// publicKeyHandler returns the server's WireGuard public key
// @Summary Get server public key
// @Description Retrieve the public key of the WireGuard interface
// @Tags Configuration
// @Produce json
// @Param Authorization header string true "API Key" default(Bearer your-api-key)
// @Success 200 {object} map[string]string "Public key"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Server error"
// @Security ApiKeyAuth
// @Router /public-key [get]
func publicKeyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pubKey, err := wireguard.GetPublicKey(ctx)
	if err != nil {
		log.Printf("Error getting public key: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "get_public_key_failed",
			Message: "Failed to retrieve public key",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"public_key": pubKey,
	})
}

// healthHandler returns the health status of the agent
// @Summary Health check
// @Description Check if the WireGuard Agent is running and healthy
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Healthy status"
// @Router /health [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// readinessHandler returns the readiness status of the agent
// @Summary Readiness check
// @Description Check if the WireGuard Agent is ready to handle requests
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Ready status"
// @Failure 503 {object} map[string]string "Not ready"
// @Router /ready [get]
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := wireguard.CheckInterface(ctx); err != nil {
		log.Printf("Readiness check failed: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not_ready"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

// swaggerHandler serves the Swagger UI
func swaggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/swagger/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
  <head>
    <title>WireGuard Agent API - Swagger UI</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui.min.css">
    <style>
      body { margin: 0; background: #fafafa; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-bundle.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-standalone-preset.min.js"></script>
    <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/swagger.json",
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        layout: "BaseLayout"
      })
      window.ui = ui
    }
    </script>
  </body>
</html>`
	_, _ = w.Write([]byte(html))
}

// swaggerJSONHandler serves the Swagger JSON definition
func swaggerJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	swaggerJSON := `{"swagger":"2.0","info":{"title":"WireGuard Agent API","description":"Secure HTTP API for managing WireGuard peers","version":"1.0.0"},"host":"localhost:9999","basePath":"/","schemes":["http","https"],"consumes":["application/json"],"produces":["application/json"],"securityDefinitions":{"ApiKeyAuth":{"type":"apiKey","in":"header","name":"X-API-Key"}},"paths":{"/add-peer":{"post":{"summary":"Add a new WireGuard peer","tags":["Peers"],"parameters":[{"name":"X-API-Key","in":"header","type":"string","required":true},{"name":"body","in":"body","schema":{"$ref":"#/definitions/AddPeerRequest"}}],"responses":{"200":{"description":"Success","schema":{"$ref":"#/definitions/SuccessResponse"}},"400":{"description":"Bad Request"},"401":{"description":"Unauthorized"},"500":{"description":"Server Error"}},"security":[{"ApiKeyAuth":[]}]}},"/remove-peer":{"post":{"summary":"Remove a WireGuard peer","tags":["Peers"],"parameters":[{"name":"X-API-Key","in":"header","type":"string","required":true},{"name":"body","in":"body","schema":{"$ref":"#/definitions/RemovePeerRequest"}}],"responses":{"200":{"description":"Success"},"400":{"description":"Bad Request"},"401":{"description":"Unauthorized"},"500":{"description":"Server Error"}},"security":[{"ApiKeyAuth":[]}]}},"/status":{"get":{"summary":"Get WireGuard status","tags":["Status"],"parameters":[{"name":"X-API-Key","in":"header","type":"string","required":true}],"responses":{"200":{"description":"Success"},"401":{"description":"Unauthorized"},"500":{"description":"Server Error"}},"security":[{"ApiKeyAuth":[]}]}},"/public-key":{"get":{"summary":"Get server public key","tags":["Configuration"],"parameters":[{"name":"X-API-Key","in":"header","type":"string","required":true}],"responses":{"200":{"description":"Success"},"401":{"description":"Unauthorized"},"500":{"description":"Server Error"}},"security":[{"ApiKeyAuth":[]}]}},"/health":{"get":{"summary":"Health check","tags":["Health"],"responses":{"200":{"description":"OK"}}}},"/ready":{"get":{"summary":"Readiness check","tags":["Health"],"responses":{"200":{"description":"OK"},"503":{"description":"Unavailable"}}}}},"definitions":{"AddPeerRequest":{"type":"object","properties":{"public_key":{"type":"string"},"allowed_ip":{"type":"string"},"endpoint":{"type":"string"}},"required":["public_key","allowed_ip"]},"RemovePeerRequest":{"type":"object","properties":{"public_key":{"type":"string"}},"required":["public_key"]},"SuccessResponse":{"type":"object","properties":{"status":{"type":"string"},"message":{"type":"string"}}}}}`
	_, _ = w.Write([]byte(swaggerJSON))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("WireGuard Agent starting on %s:%d", cfg.Host, cfg.Port)
	log.Printf("📚 Swagger UI available at http://%s:%d/swagger/", cfg.Host, cfg.Port)

	mux := http.NewServeMux()

	// Swagger/API Documentation
	mux.HandleFunc("/swagger/", swaggerHandler)
	mux.HandleFunc("/swagger.json", swaggerJSONHandler)

	// Protected endpoints
	mux.HandleFunc("/add-peer", requestLogger(authMiddleware(addPeerHandler)))
	mux.HandleFunc("/remove-peer", requestLogger(authMiddleware(removePeerHandler)))
	mux.HandleFunc("/status", requestLogger(authMiddleware(statusHandler)))
	mux.HandleFunc("/public-key", requestLogger(authMiddleware(publicKeyHandler)))

	// Health check endpoints
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readinessHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutdown signal received, gracefully stopping...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}
