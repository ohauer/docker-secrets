package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Status represents the health status
type Status struct {
	Ready       bool   `json:"ready"`
	SecretCount int    `json:"secret_count"`
	SyncedCount int    `json:"synced_count"`
	StatusFile  string `json:"-"`
	mu          sync.RWMutex
}

// NewStatus creates a new status tracker
func NewStatus(statusFile string) *Status {
	return &Status{
		Ready:      false,
		StatusFile: statusFile,
	}
}

// SetReady marks the service as ready
func (s *Status) SetReady(secretCount, syncedCount int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Ready = syncedCount > 0
	s.SecretCount = secretCount
	s.SyncedCount = syncedCount

	if s.StatusFile != "" {
		if s.Ready {
			if err := os.WriteFile(s.StatusFile, []byte("ready"), 0644); err != nil {
				return fmt.Errorf("failed to write status file: %w", err)
			}
		} else {
			_ = os.Remove(s.StatusFile)
		}
	}

	return nil
}

// IsReady returns whether the service is ready
func (s *Status) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Ready
}

// GetStatus returns the current status
func (s *Status) GetStatus() (bool, int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Ready, s.SecretCount, s.SyncedCount
}

// Server provides HTTP health endpoints
type Server struct {
	status *Status
	addr   string
	port   int
	server *http.Server
}

// NewServer creates a new health server
func NewServer(status *Status, addr string, port int) *Server {
	return &Server{
		status: status,
		addr:   addr,
		port:   port,
	}
}

// Start starts the health server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/ready", s.readyHandler)
	mux.Handle("/metrics", promhttp.Handler())

	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.addr, s.port),
		Handler: mux,
	}

	go func() {
		_ = s.server.ListenAndServe()
	}()

	return nil
}

// Stop stops the health server
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	ready, secretCount, syncedCount := s.status.GetStatus()

	w.Header().Set("Content-Type", "application/json")

	if ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ready":        ready,
		"secret_count": secretCount,
		"synced_count": syncedCount,
	})
}
