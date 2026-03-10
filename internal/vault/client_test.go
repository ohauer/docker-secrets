package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.GetAPIClient() == nil {
		t.Fatal("expected API client, got nil")
	}
}

func TestClient_Ping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"initialized":true,"sealed":false}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.Ping(); err != nil {
		t.Errorf("ping failed: %v", err)
	}
}

func TestClient_AuthenticateToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/auth/token/lookup-self" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"id":"test-token"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := AuthConfig{
		Method: AuthMethodToken,
		Token:  "test-token",
	}

	if err := client.Authenticate(config); err != nil {
		t.Errorf("token authentication failed: %v", err)
	}
}

func TestClient_AuthenticateToken_Empty(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := AuthConfig{
		Method: AuthMethodToken,
		Token:  "",
	}

	if err := client.Authenticate(config); err == nil {
		t.Error("expected error for empty token, got nil")
	}
}

func TestClient_AuthenticateAppRole_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/auth/approle/login" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"auth":{"client_token":"test-token"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := AuthConfig{
		Method:   AuthMethodAppRole,
		RoleID:   "test-role-id",
		SecretID: "test-secret-id",
	}

	if err := client.Authenticate(config); err != nil {
		t.Errorf("approle authentication failed: %v", err)
	}

	if client.GetAPIClient().Token() != "test-token" {
		t.Errorf("expected token 'test-token', got: %s", client.GetAPIClient().Token())
	}
}

func TestClient_AuthenticateAppRole_MissingCredentials(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	tests := []struct {
		name     string
		roleID   string
		secretID string
	}{
		{"missing roleID", "", "secret"},
		{"missing secretID", "role", ""},
		{"missing both", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AuthConfig{
				Method:   AuthMethodAppRole,
				RoleID:   tt.roleID,
				SecretID: tt.secretID,
			}

			if err := client.Authenticate(config); err == nil {
				t.Error("expected error for missing credentials, got nil")
			}
		})
	}
}

func TestClient_AuthenticateUnsupportedMethod(t *testing.T) {
	client, err := NewClient(TestVaultAddr)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	config := AuthConfig{
		Method: "unsupported",
	}

	if err := client.Authenticate(config); err == nil {
		t.Error("expected error for unsupported auth method, got nil")
	}
}
