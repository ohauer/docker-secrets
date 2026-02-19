package vault

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	MaxRetries     int
}

// FetchSecretWithRetry fetches a secret with exponential backoff retry
func (c *Client) FetchSecretWithRetry(ctx context.Context, mountPath, secretPath, kvVersion, namespace string, config RetryConfig) (SecretData, error) {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoff):
			}

			backoff = time.Duration(float64(backoff) * config.Multiplier)
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}

		data, err := c.FetchSecret(mountPath, secretPath, kvVersion, namespace)
		if err == nil {
			return data, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed after %d retries: %w", config.MaxRetries, lastErr)
}
