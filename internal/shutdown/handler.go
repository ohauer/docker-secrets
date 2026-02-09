package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Handler manages graceful shutdown
type Handler struct {
	timeout  time.Duration
	handlers []func() error
	mu       sync.Mutex
	sigCh    chan os.Signal
	reloadCh chan os.Signal
}

// NewHandler creates a new shutdown handler
func NewHandler(timeout time.Duration) *Handler {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	reloadCh := make(chan os.Signal, 1)
	signal.Notify(reloadCh, syscall.SIGHUP)

	return &Handler{
		timeout:  timeout,
		handlers: make([]func() error, 0),
		sigCh:    sigCh,
		reloadCh: reloadCh,
	}
}

// Register registers a shutdown handler
func (h *Handler) Register(handler func() error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, handler)
}

// Wait waits for shutdown signal
func (h *Handler) Wait() <-chan os.Signal {
	return h.sigCh
}

// WaitReload waits for reload signal (SIGHUP)
func (h *Handler) WaitReload() <-chan os.Signal {
	return h.reloadCh
}

// Shutdown executes all registered handlers
func (h *Handler) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		var errs []error
		for _, handler := range h.handlers {
			if err := handler(); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			done <- fmt.Errorf("shutdown errors: %v", errs)
		} else {
			done <- nil
		}
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// Stop stops listening for signals
func (h *Handler) Stop() {
	signal.Stop(h.sigCh)
	signal.Stop(h.reloadCh)
	close(h.sigCh)
	close(h.reloadCh)
}
