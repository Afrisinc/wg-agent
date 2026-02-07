package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidatePublicKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid public key",
			key:     "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc=",
			wantErr: false,
		},
		{
			name:    "invalid key too short",
			key:     "short",
			wantErr: true,
		},
		{
			name:    "invalid key empty",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validatePublicKey(tt.key)
			if got == tt.wantErr {
				t.Errorf("validatePublicKey() = %v, want %v", got, !tt.wantErr)
			}
		})
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "valid IP",
			ip:      "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "invalid IP too few octets",
			ip:      "192.168.1",
			wantErr: true,
		},
		{
			name:    "invalid IP with letters",
			ip:      "192.168.1.a",
			wantErr: true,
		},
		{
			name:    "invalid IP empty",
			ip:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateIP(tt.ip)
			if got == tt.wantErr {
				t.Errorf("validateIP() = %v, want %v", got, !tt.wantErr)
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	os.Setenv("API_KEY", "test-key-123")

	t.Run("missing API key", func(t *testing.T) {
		handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("invalid API key", func(t *testing.T) {
		handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("valid API key", func(t *testing.T) {
		handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", "test-key-123")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestAddPeerHandler(t *testing.T) {
	os.Setenv("API_KEY", "test-key")

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/add-peer", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()

		addPeerHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid public key", func(t *testing.T) {
		reqBody := AddPeerRequest{
			PublicKey: "invalid",
			AllowedIP: "192.168.1.1",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/add-peer", bytes.NewReader(body))
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()

		addPeerHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid IP", func(t *testing.T) {
		reqBody := AddPeerRequest{
			PublicKey: "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc=",
			AllowedIP: "invalid-ip",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/add-peer", bytes.NewReader(body))
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()

		addPeerHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Errorf("Expected healthy status, got %s", resp["status"])
	}
}
