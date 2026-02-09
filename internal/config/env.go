package config

import (
	"os"
	"strconv"
	"time"
)

// EnvConfig holds environment variable configuration
type EnvConfig struct {
	VaultAddr              string
	VaultToken             string
	VaultRoleID            string
	VaultSecretID          string
	VaultCACert            string
	VaultCAPath            string
	VaultSkipVerify        bool
	VaultClientCert        string
	VaultClientKey         string
	ConfigFile             string
	WatchConfig            bool
	CircuitBreakerMaxReqs  int
	CircuitBreakerInterval time.Duration
	CircuitBreakerTimeout  time.Duration
	LogLevel               string
	MetricsAddr            string
	MetricsPort            int
	EnableMetrics          bool
	StatusFile             string
	EnableTracing          bool
	OTELExporterEndpoint   string
	InitialBackoff         time.Duration
	MaxBackoff             time.Duration
	BackoffMultiplier      float64
}

// LoadEnvConfig loads configuration from environment variables
func LoadEnvConfig() *EnvConfig {
	return &EnvConfig{
		VaultAddr:              getEnv("VAULT_ADDR", ""),
		VaultToken:             getEnv("VAULT_TOKEN", ""),
		VaultRoleID:            getEnv("VAULT_ROLE_ID", ""),
		VaultSecretID:          getEnv("VAULT_SECRET_ID", ""),
		VaultCACert:            getEnv("VAULT_CACERT", ""),
		VaultCAPath:            getEnv("VAULT_CAPATH", ""),
		VaultSkipVerify:        getEnvBool("VAULT_SKIP_VERIFY", false),
		VaultClientCert:        getEnv("VAULT_CLIENT_CERT", ""),
		VaultClientKey:         getEnv("VAULT_CLIENT_KEY", ""),
		ConfigFile:             getEnv("CONFIG_FILE", "/config.yaml"),
		WatchConfig:            getEnvBool("WATCH_CONFIG", false),
		CircuitBreakerMaxReqs:  getEnvInt("CIRCUIT_BREAKER_MAX_REQUESTS", 3),
		CircuitBreakerInterval: getEnvDuration("CIRCUIT_BREAKER_INTERVAL", 60*time.Second),
		CircuitBreakerTimeout:  getEnvDuration("CIRCUIT_BREAKER_TIMEOUT", 30*time.Second),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		MetricsAddr:            getEnv("METRICS_ADDR", "127.0.0.1"),
		MetricsPort:            getEnvIntRange("METRICS_PORT", 8080, 1025, 65535),
		EnableMetrics:          getEnvBool("ENABLE_METRICS", true),
		StatusFile:             getEnv("STATUS_FILE", "/tmp/.ready-state"),
		EnableTracing:          getEnvBool("ENABLE_TRACING", false),
		OTELExporterEndpoint:   getEnv("OTEL_EXPORTER_ENDPOINT", ""),
		InitialBackoff:         getEnvDuration("INITIAL_BACKOFF", 1*time.Second),
		MaxBackoff:             getEnvDuration("MAX_BACKOFF", 5*time.Minute),
		BackoffMultiplier:      getEnvFloat("BACKOFF_MULTIPLIER", 2.0),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvIntRange(key string, defaultValue, minValue, maxValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil && i >= minValue && i <= maxValue {
			return i
		}
		// Return -1 to indicate invalid value
		return -1
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		f, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return f
		}
	}
	return defaultValue
}
