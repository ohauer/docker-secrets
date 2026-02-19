package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSecret_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/secret/data/test/path" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
                "data": {
                    "data": {
                        "username": "testuser",
                        "password": "testpass"
                    }
                }
            }`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	data, err := client.FetchSecret("secret", "test/path", "v2", "")
	if err != nil {
		t.Fatalf("failed to fetch secret: %v", err)
	}

	if data["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got: %v", data["username"])
	}

	if data["password"] != "testpass" {
		t.Errorf("expected password 'testpass', got: %v", data["password"])
	}
}

func TestFetchSecret_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.FetchSecret("secret", "nonexistent", "v2", "")
	if err == nil {
		t.Error("expected error for nonexistent secret, got nil")
	}
}

func TestFetchSecretWithRetry_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
            "data": {
                "data": {
                    "key": "value"
                }
            }
        }`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := RetryConfig{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     2.0,
		MaxRetries:     3,
	}

	ctx := context.Background()
	data, err := client.FetchSecretWithRetry(ctx, "secret", "test/path", "v2", "", config)
	if err != nil {
		t.Fatalf("failed to fetch secret with retry: %v", err)
	}

	if data["key"] != "value" {
		t.Errorf("expected key 'value', got: %v", data["key"])
	}

	if attempts < 2 {
		t.Errorf("expected at least 2 attempts, got: %d", attempts)
	}
}

func TestFetchSecretWithRetry_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := RetryConfig{
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
		MaxRetries:     2,
	}

	ctx := context.Background()
	_, err = client.FetchSecretWithRetry(ctx, "secret", "test/path", "v2", "", config)
	if err == nil {
		t.Error("expected error after max retries, got nil")
	}
}

func TestFetchSecretWithRetry_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := RetryConfig{
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2.0,
		MaxRetries:     5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.FetchSecretWithRetry(ctx, "secret", "test/path", "v2", "", config)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestFetchSecretWithRetry_ExponentialBackoff(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := RetryConfig{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     50 * time.Millisecond,
		Multiplier:     2.0,
		MaxRetries:     3,
	}

	ctx := context.Background()
	start := time.Now()
	_, _ = client.FetchSecretWithRetry(ctx, "secret", "test/path", "v2", "", config)
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("expected at least 10ms elapsed, got: %v", elapsed)
	}

	if attempts < 4 {
		t.Errorf("expected at least 4 attempts, got: %d", attempts)
	}
}

func TestFetchSecret_V1_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/secret/test/path" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
                "data": {
                    "username": "testuser",
                    "password": "testpass"
                }
            }`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	data, err := client.FetchSecret("secret", "test/path", "v1", "")
	if err != nil {
		t.Fatalf("failed to fetch secret: %v", err)
	}

	if data["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got: %v", data["username"])
	}

	if data["password"] != "testpass" {
		t.Errorf("expected password 'testpass', got: %v", data["password"])
	}
}

func TestFetchSecret_V1_PathConstruction(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {"key": "value"}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, _ = client.FetchSecret("secret", "test/path", "v1", "")

	expectedPath := "/v1/secret/test/path"
	if requestedPath != expectedPath {
		t.Errorf("expected path %s, got: %s", expectedPath, requestedPath)
	}
}

func TestFetchSecret_V2_PathConstruction(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": {"data": {"key": "value"}}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, _ = client.FetchSecret("secret", "test/path", "v2", "")

	expectedPath := "/v1/secret/data/test/path"
	if requestedPath != expectedPath {
		t.Errorf("expected path %s, got: %s", expectedPath, requestedPath)
	}
}
