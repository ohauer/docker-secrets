package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// AuthMethod represents the authentication method
type AuthMethod string

const (
	AuthMethodToken   AuthMethod = "token"
	AuthMethodAppRole AuthMethod = "approle"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Method   AuthMethod
	Token    string
	RoleID   string
	SecretID string
}

// RenewalCallback is called on token renewal events
type RenewalCallback func(event string, ttl int)

// Authenticate authenticates the client with Vault and returns the auth secret for renewal
func (c *Client) Authenticate(config AuthConfig) error {
	c.authConfig = &config

	switch config.Method {
	case AuthMethodToken:
		return c.authenticateToken(config.Token)
	case AuthMethodAppRole:
		return c.authenticateAppRole(config.RoleID, config.SecretID)
	default:
		return fmt.Errorf("unsupported auth method: %s", config.Method)
	}
}

// StartRenewal starts automatic token renewal in the background.
// The callback is called on renewal events. Call the returned stop function to cancel.
func (c *Client) StartRenewal(callback RenewalCallback) (func(), error) {
	if c.authSecret == nil {
		// Static token or no renewable secret - nothing to renew
		return func() {}, nil
	}

	if !c.authSecret.Auth.Renewable {
		return func() {}, nil
	}

	watcher, err := c.client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
		Secret: c.authSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create token watcher: %w", err)
	}

	stopped := make(chan struct{})
	go watcher.Start()

	go func() {
		for {
			select {
			case <-stopped:
				return
			case renewal := <-watcher.RenewCh():
				if renewal != nil && callback != nil {
					callback("renewed", renewal.Secret.Auth.LeaseDuration)
				}
			case <-watcher.DoneCh():
				// Check if we were stopped intentionally
				select {
				case <-stopped:
					return
				default:
				}

				// Token can no longer be renewed, try re-authentication
				if c.authConfig != nil && c.authConfig.Method == AuthMethodAppRole {
					if callback != nil {
						callback("re-authenticating", 0)
					}
					if err := c.authenticateAppRole(c.authConfig.RoleID, c.authConfig.SecretID); err != nil {
						if callback != nil {
							callback("re-auth-failed", 0)
						}
						return
					}
					// Restart renewal with new token
					if c.authSecret != nil && c.authSecret.Auth.Renewable {
						newWatcher, err := c.client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
							Secret: c.authSecret,
						})
						if err != nil {
							if callback != nil {
								callback("re-auth-failed", 0)
							}
							return
						}
						if callback != nil {
							callback("re-authenticated", c.authSecret.Auth.LeaseDuration)
						}
						watcher = newWatcher
						go watcher.Start()
						continue
					}
				} else if callback != nil {
					callback("expired", 0)
				}
				return
			}
		}
	}()

	return func() {
		close(stopped)
		watcher.Stop()
	}, nil
}

func (c *Client) authenticateToken(token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	c.client.SetToken(token)

	_, err := c.executeWithBreaker(func() (interface{}, error) {
		return c.client.Auth().Token().LookupSelf()
	})
	if err != nil {
		return fmt.Errorf("token authentication failed: %w", err)
	}

	// No authSecret for static tokens - renewal not possible
	return nil
}

func (c *Client) authenticateAppRole(roleID, secretID string) error {
	if roleID == "" {
		return fmt.Errorf("roleId is required")
	}
	if secretID == "" {
		return fmt.Errorf("secretId is required")
	}

	data := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	result, err := c.executeWithBreaker(func() (interface{}, error) {
		return c.client.Logical().Write("auth/approle/login", data)
	})
	if err != nil {
		return fmt.Errorf("approle authentication failed: %w", err)
	}

	resp, ok := result.(*api.Secret)
	if !ok || resp == nil || resp.Auth == nil {
		return fmt.Errorf("approle authentication returned no token")
	}

	c.client.SetToken(resp.Auth.ClientToken)
	c.authSecret = resp
	return nil
}
