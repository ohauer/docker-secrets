package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStatus_SetReady(t *testing.T) {
	tmpDir := t.TempDir()
	statusFile := filepath.Join(tmpDir, ".ready-state")

	status := NewStatus(statusFile)

	if err := status.SetReady(2, 2); err != nil {
		t.Fatalf("failed to set ready: %v", err)
	}

	if !status.IsReady() {
		t.Error("expected status to be ready")
	}

	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		t.Error("status file was not created")
	}
}

func TestStatus_NotReady(t *testing.T) {
	tmpDir := t.TempDir()
	statusFile := filepath.Join(tmpDir, ".ready-state")

	status := NewStatus(statusFile)

	if err := status.SetReady(2, 0); err != nil {
		t.Fatalf("failed to set status: %v", err)
	}

	if status.IsReady() {
		t.Error("expected status to not be ready")
	}

	if _, err := os.Stat(statusFile); !os.IsNotExist(err) {
		t.Error("status file should not exist when not ready")
	}
}

func TestStatus_GetStatus(t *testing.T) {
	status := NewStatus("")

	_ = status.SetReady(5, 3)

	ready, secretCount, syncedCount := status.GetStatus()

	if !ready {
		t.Error("expected ready to be true")
	}

	if secretCount != 5 {
		t.Errorf("expected secret count 5, got %d", secretCount)
	}

	if syncedCount != 3 {
		t.Errorf("expected synced count 3, got %d", syncedCount)
	}
}

func TestHealthHandler(t *testing.T) {
	status := NewStatus("")
	server := NewServer(status, "127.0.0.1", 8080)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response["status"])
	}
}

func TestReadyHandler_Ready(t *testing.T) {
	status := NewStatus("")
	_ = status.SetReady(2, 2)

	server := NewServer(status, "127.0.0.1", 8080)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	server.readyHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response["ready"].(bool) {
		t.Error("expected ready to be true")
	}
}

func TestReadyHandler_NotReady(t *testing.T) {
	status := NewStatus("")
	_ = status.SetReady(2, 0)

	server := NewServer(status, "127.0.0.1", 8080)

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()

	server.readyHandler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["ready"].(bool) {
		t.Error("expected ready to be false")
	}
}

func TestCheckReadiness_Ready(t *testing.T) {
	tmpDir := t.TempDir()
	statusFile := filepath.Join(tmpDir, ".ready-state")

	_ = os.WriteFile(statusFile, []byte("ready"), 0644)

	if err := CheckReadiness(statusFile); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestCheckReadiness_NotReady(t *testing.T) {
	tmpDir := t.TempDir()
	statusFile := filepath.Join(tmpDir, ".ready-state")

	if err := CheckReadiness(statusFile); err == nil {
		t.Error("expected error for missing status file, got nil")
	}
}
