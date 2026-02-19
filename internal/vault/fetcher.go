package vault

import (
	"fmt"
	"path"

	"github.com/hashicorp/vault/api"
)

// SecretData represents the data retrieved from Vault
type SecretData map[string]interface{}

// FetchSecret fetches a secret from Vault KV v1 or v2
func (c *Client) FetchSecret(mountPath, secretPath, kvVersion, namespace string) (SecretData, error) {
	var fullPath string
	if kvVersion == "v2" {
		fullPath = path.Join(mountPath, "data", secretPath)
	} else {
		fullPath = path.Join(mountPath, secretPath)
	}

	result, err := c.executeWithBreaker(func() (interface{}, error) {
		// Set namespace if provided
		if namespace != "" {
			c.client.SetNamespace(namespace)
		}
		return c.client.Logical().Read(fullPath)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("secret not found at path: %s", secretPath)
	}

	secret, ok := result.(*api.Secret)
	if !ok || secret == nil {
		return nil, fmt.Errorf("invalid secret response")
	}

	if secret.Data == nil {
		return nil, fmt.Errorf("secret has no data")
	}

	if kvVersion == "v2" {
		data, ok := secret.Data["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid secret data format for KV v2")
		}
		return SecretData(data), nil
	}

	return SecretData(secret.Data), nil
}
