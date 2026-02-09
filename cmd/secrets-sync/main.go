package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ohauer/docker-secrets/internal/config"
	"github.com/ohauer/docker-secrets/internal/health"
	"github.com/ohauer/docker-secrets/internal/logger"
	"github.com/ohauer/docker-secrets/internal/metrics"
	"github.com/ohauer/docker-secrets/internal/shutdown"
	"github.com/ohauer/docker-secrets/internal/syncer"
	"github.com/ohauer/docker-secrets/internal/tracing"
	"github.com/ohauer/docker-secrets/internal/vault"
	"go.uber.org/zap"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "", "path to config file")
	flag.StringVar(&configFile, "c", "", "path to config file (shorthand)")
}

func main() {
	// Parse flags first
	flag.Parse()

	// Get remaining args after flags
	args := flag.Args()

	// Check for subcommands
	if len(args) > 0 {
		cmd := args[0]
		switch cmd {
		case "help", "-h", "--help":
			printHelp()
			os.Exit(0)
		case "version", "-v", "--version":
			printVersion()
			os.Exit(0)
		case "init":
			printInitConfig()
			os.Exit(0)
		case "validate":
			os.Exit(runValidate())
		case "convert":
			os.Exit(runConvert(args[1:]))
		case "isready":
			os.Exit(isReady())
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
			printUsage()
			os.Exit(1)
		}
	}

	// No command specified, run main service
	if err := run(); err != nil {
		if logger.Get() != nil {
			logger.Error("fatal error", zap.Error(err))
			logger.Sync()
		}
		os.Exit(1)
	}
}

func getConfigFile() string {
	// Precedence: flag > env var > defaults
	if configFile != "" {
		return configFile
	}

	if envFile := os.Getenv("CONFIG_FILE"); envFile != "" {
		return envFile
	}

	// Try default locations
	defaults := []string{"./config.yaml", "/etc/secrets-sync/config.yaml"}
	for _, path := range defaults {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "./config.yaml"
}

func run() error {
	envCfg := config.LoadEnvConfig()
	configPath := getConfigFile()

	if err := logger.Init(envCfg.LogLevel); err != nil {
		return err
	}
	defer logger.Sync()

	logger.Info("starting docker secrets sync",
		zap.String("config_file", configPath),
		zap.Bool("watch_config", envCfg.WatchConfig),
	)

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	logger.Info("configuration loaded",
		zap.Int("secret_count", len(cfg.Secrets)),
	)

	// Initialize tracing if enabled
	var tracingShutdown func()
	if envCfg.EnableTracing {
		shutdown, err := tracing.Init("docker-secrets-sync", envCfg.OTELExporterEndpoint)
		if err != nil {
			logger.Warn("failed to initialize tracing", zap.Error(err))
		} else {
			tracingShutdown = shutdown
			logger.Info("tracing enabled", zap.String("endpoint", envCfg.OTELExporterEndpoint))
		}
	}

	// Create Vault client with TLS configuration
	tlsConfig := &vault.TLSConfig{
		CACert:     cfg.SecretStore.TLSCACert,
		CAPath:     cfg.SecretStore.TLSCAPath,
		ClientCert: cfg.SecretStore.TLSClientCert,
		ClientKey:  cfg.SecretStore.TLSClientKey,
		SkipVerify: cfg.SecretStore.TLSSkipVerify,
	}

	// Override with environment variables if set
	if envCfg.VaultCACert != "" {
		tlsConfig.CACert = envCfg.VaultCACert
	}
	if envCfg.VaultCAPath != "" {
		tlsConfig.CAPath = envCfg.VaultCAPath
	}
	if envCfg.VaultClientCert != "" {
		tlsConfig.ClientCert = envCfg.VaultClientCert
	}
	if envCfg.VaultClientKey != "" {
		tlsConfig.ClientKey = envCfg.VaultClientKey
	}
	if envCfg.VaultSkipVerify {
		tlsConfig.SkipVerify = true
	}

	vaultClient, err := vault.NewClientWithTLS(cfg.SecretStore.Address, tlsConfig)
	if err != nil {
		return err
	}

	// Set up circuit breaker
	vaultClient.WithCircuitBreaker(
		vault.BreakerConfig{
			MaxRequests: uint32(envCfg.CircuitBreakerMaxReqs),
			Interval:    envCfg.CircuitBreakerInterval,
			Timeout:     envCfg.CircuitBreakerTimeout,
		},
		func(from, to string) {
			logger.Info("circuit breaker state changed",
				zap.String("from", from),
				zap.String("to", to),
			)
			metrics.SetCircuitBreakerState("vault-client", to)
		},
	)

	// Authenticate
	authConfig := vault.AuthConfig{
		Method: vault.AuthMethod(cfg.SecretStore.AuthMethod),
	}

	if cfg.SecretStore.Token != "" {
		authConfig.Token = cfg.SecretStore.Token
	}
	if cfg.SecretStore.RoleID != "" {
		authConfig.RoleID = cfg.SecretStore.RoleID
	}
	if cfg.SecretStore.SecretID != "" {
		authConfig.SecretID = cfg.SecretStore.SecretID
	}

	if err := vaultClient.Authenticate(authConfig); err != nil {
		return err
	}

	logger.Info("authenticated to vault",
		zap.String("address", cfg.SecretStore.Address),
		zap.String("auth_method", cfg.SecretStore.AuthMethod),
	)

	// Create syncer
	retryConfig := vault.RetryConfig{
		InitialBackoff: envCfg.InitialBackoff,
		MaxBackoff:     envCfg.MaxBackoff,
		Multiplier:     envCfg.BackoffMultiplier,
		MaxRetries:     3,
	}

	secretSyncer := syncer.NewSecretSyncer(vaultClient, retryConfig)
	scheduler := syncer.NewScheduler(secretSyncer)

	// Set up health status
	status := health.NewStatus(envCfg.StatusFile)

	// Validate metrics port
	if envCfg.MetricsPort < 1025 || envCfg.MetricsPort > 65535 {
		logger.Error("invalid METRICS_PORT: must be between 1025 and 65535, disabling metrics",
			zap.Int("port", envCfg.MetricsPort),
		)
		envCfg.EnableMetrics = false
	}

	// Start metrics server if enabled
	var healthServer *health.Server
	if envCfg.EnableMetrics {
		healthServer = health.NewServer(status, envCfg.MetricsAddr, envCfg.MetricsPort)
		if err := healthServer.Start(); err != nil {
			return err
		}
		logger.Info("metrics server started",
			zap.String("addr", envCfg.MetricsAddr),
			zap.Int("port", envCfg.MetricsPort),
		)
	} else {
		logger.Info("metrics server disabled")
	}

	// Set metrics
	metrics.SetSecretsConfigured(len(cfg.Secrets))

	// Start syncing secrets
	for _, secret := range cfg.Secrets {
		scheduler.AddSecret(cfg, secret)
		logger.Info("secret sync started",
			zap.String("name", secret.Name),
			zap.Duration("refresh_interval", secret.RefreshInterval),
		)
	}

	// Monitor sync results
	go func() {
		syncedCount := 0
		for result := range scheduler.Results() {
			if result.Success {
				syncedCount++
				logger.Info("secret synced successfully",
					zap.String("name", result.SecretName),
					zap.Time("timestamp", result.Timestamp),
				)
				metrics.RecordFetchSuccess(result.SecretName, "")
				metrics.SetSecretsSynced(syncedCount)
			} else {
				logger.Error("secret sync failed",
					zap.String("name", result.SecretName),
					zap.Error(result.Error),
					zap.Time("timestamp", result.Timestamp),
				)
				metrics.RecordFetchError(result.SecretName, "", "sync_error")
			}

			// Update readiness status
			_ = status.SetReady(len(cfg.Secrets), syncedCount)
		}
	}()

	// Set up config watcher if enabled
	if envCfg.WatchConfig {
		watcher, err := config.NewWatcher(
			envCfg.ConfigFile,
			func(newCfg *config.Config) error {
				logger.Info("configuration reloaded",
					zap.Int("secret_count", len(newCfg.Secrets)),
				)
				// Update secrets
				for _, secret := range newCfg.Secrets {
					scheduler.AddSecret(newCfg, secret)
				}
				metrics.SetSecretsConfigured(len(newCfg.Secrets))
				return nil
			},
			func(err error) {
				logger.Error("config watcher error", zap.Error(err))
			},
		)
		if err != nil {
			logger.Warn("failed to create config watcher", zap.Error(err))
		} else {
			watcher.Start()
			defer watcher.Stop()
			logger.Info("config watcher started")
		}
	}

	// Set up graceful shutdown
	shutdownHandler := shutdown.NewHandler(30 * time.Second)
	shutdownHandler.Register(func() error {
		logger.Info("shutting down scheduler")
		scheduler.Stop()
		return nil
	})
	if healthServer != nil {
		shutdownHandler.Register(func() error {
			logger.Info("shutting down metrics server")
			return healthServer.Stop()
		})
	}
	if tracingShutdown != nil {
		shutdownHandler.Register(func() error {
			logger.Info("shutting down tracing")
			tracingShutdown()
			return nil
		})
	}

	logger.Info("docker secrets sync running, waiting for shutdown signal")

	// Wait for shutdown signal
	<-shutdownHandler.Wait()

	logger.Info("shutdown signal received, stopping gracefully")

	if err := shutdownHandler.Shutdown(); err != nil {
		logger.Error("shutdown error", zap.Error(err))
		return err
	}

	logger.Info("shutdown complete")
	return nil
}

func isReady() int {
	envCfg := config.LoadEnvConfig()

	if err := health.CheckReadiness(envCfg.StatusFile); err != nil {
		return 1
	}

	return 0
}
