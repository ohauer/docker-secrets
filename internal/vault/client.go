package vault

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/vault/api"
	"github.com/sony/gobreaker"
)

const (
	// MaxResponseSize is the maximum allowed size for Vault responses (10MB)
	MaxResponseSize = 10 * 1024 * 1024
)

// TLSConfig holds TLS configuration for Vault client
type TLSConfig struct {
	CACert     string
	CAPath     string
	ClientCert string
	ClientKey  string
	SkipVerify bool
}

// Client wraps the Vault API client
type Client struct {
	client  *api.Client
	breaker *gobreaker.CircuitBreaker
}

// NewClient creates a new Vault client
func NewClient(address string) (*Client, error) {
	return NewClientWithTLS(address, nil)
}

// NewClientWithTLS creates a new Vault client with TLS configuration
func NewClientWithTLS(address string, tlsConfig *TLSConfig) (*Client, error) {
	config := api.DefaultConfig()
	config.Address = address

	if tlsConfig != nil {
		if err := configureTLS(config, tlsConfig); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Wrap HTTP client to limit response size
	originalTransport := client.CloneConfig().HttpClient.Transport
	client.CloneConfig().HttpClient.Transport = &limitedTransport{
		base:     originalTransport,
		maxBytes: MaxResponseSize,
	}

	return &Client{client: client}, nil
}

func configureTLS(config *api.Config, tlsConfig *TLSConfig) error {
	tlsClientConfig := &api.TLSConfig{
		Insecure: tlsConfig.SkipVerify,
	}

	if tlsConfig.CACert != "" {
		tlsClientConfig.CACert = tlsConfig.CACert
	}

	if tlsConfig.CAPath != "" {
		tlsClientConfig.CAPath = tlsConfig.CAPath
	}

	if tlsConfig.ClientCert != "" {
		tlsClientConfig.ClientCert = tlsConfig.ClientCert
	}

	if tlsConfig.ClientKey != "" {
		tlsClientConfig.ClientKey = tlsConfig.ClientKey
	}

	return config.ConfigureTLS(tlsClientConfig)
}

// GetAPIClient returns the underlying Vault API client
func (c *Client) GetAPIClient() *api.Client {
	return c.client
}

// Ping checks if the Vault server is reachable
func (c *Client) Ping() error {
	_, err := c.executeWithBreaker(func() (interface{}, error) {
		return c.client.Sys().Health()
	})
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	return nil
}

// limitedTransport wraps http.RoundTripper to limit response body size
type limitedTransport struct {
	base     http.RoundTripper
	maxBytes int64
}

func (t *limitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Wrap response body with size limiter
	resp.Body = &limitedReadCloser{
		reader:   io.LimitReader(resp.Body, t.maxBytes),
		closer:   resp.Body,
		maxBytes: t.maxBytes,
	}

	return resp, nil
}

// limitedReadCloser wraps io.Reader with size limit and preserves Close
type limitedReadCloser struct {
	reader   io.Reader
	closer   io.Closer
	maxBytes int64
}

func (r *limitedReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *limitedReadCloser) Close() error {
	return r.closer.Close()
}
