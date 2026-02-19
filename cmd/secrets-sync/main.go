package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ohauer/docker-secrets/internal/config"
	"github.com/ohauer/docker-secrets/internal/filewriter"
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
	flag.Usage = printHelp // Override default help
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

	// Log working directory for relative path resolution
	workDir, err := os.Getwd()
	if err != nil {
		logger.Warn("failed to get working directory", zap.Error(err))
		workDir = "unknown"
	}

	// Resolve config path to absolute for logging
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		absConfigPath = configPath
	}

	logger.Info("starting secrets-sync",
		zap.String("config_file", absConfigPath),
		zap.String("working_directory", workDir),
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

	// Create client factory for on-demand client creation
	clientFactory := func(creds config.CredentialSet) (*vault.Client, error) {
		client, err := vault.NewClientWithTLS(cfg.SecretStore.Address, tlsConfig)
		if err != nil {
			return nil, err
		}

		// Set up circuit breaker
		client.WithCircuitBreaker(
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

		// Authenticate with provided credentials
		authConfig := vault.AuthConfig{
			Method:   vault.AuthMethod(creds.AuthMethod),
			Token:    creds.Token,
			RoleID:   creds.RoleID,
			SecretID: creds.SecretID,
		}

		if err := client.Authenticate(authConfig); err != nil {
			return nil, err
		}

		return client, nil
	}

	// Create default client to verify connectivity
	defaultCreds := cfg.SecretStore.GetDefaultCredentials()
	_, err = clientFactory(defaultCreds)
	if err != nil {
		return err
	}

	logger.Info("authenticated to vault",
		zap.String("address", cfg.SecretStore.Address),
		zap.String("auth_method", cfg.SecretStore.AuthMethod),
	)

	// Warn if using HTTP (insecure)
	if strings.HasPrefix(cfg.SecretStore.Address, "http://") &&
		!strings.Contains(cfg.SecretStore.Address, "localhost") &&
		!strings.Contains(cfg.SecretStore.Address, "127.0.0.1") {
		logger.Warn("using insecure HTTP connection to Vault - use HTTPS in production",
			zap.String("address", cfg.SecretStore.Address),
		)
	}

	// Create syncer with client factory
	retryConfig := vault.RetryConfig{
		InitialBackoff: envCfg.InitialBackoff,
		MaxBackoff:     envCfg.MaxBackoff,
		Multiplier:     envCfg.BackoffMultiplier,
		MaxRetries:     3,
	}

	secretSyncer := syncer.NewSecretSyncer(clientFactory, retryConfig)
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
				workDir, _ := os.Getwd()
				if workDir == "" {
					workDir = "unknown"
				}
				absConfigPath, err := filepath.Abs(configPath)
				if err != nil {
					absConfigPath = configPath
				}
				logger.Info("configuration reloaded",
					zap.String("config_file", absConfigPath),
					zap.String("working_directory", workDir),
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

	// Cleanup orphaned .tmp files from previous runs
	var allFilePaths []string
	for _, secret := range cfg.Secrets {
		for _, file := range secret.Files {
			allFilePaths = append(allFilePaths, file.Path)
		}
	}
	outputDirs := filewriter.GetOutputDirectories(allFilePaths)
	if err := filewriter.CleanupOrphanedTempFiles(outputDirs, logger.Get()); err != nil {
		logger.Warn("failed to cleanup orphaned temp files", zap.Error(err))
	}

	logger.Info("docker secrets sync running, waiting for shutdown signal")

	// Wait for signals
	for {
		select {
		case <-shutdownHandler.Wait():
			logger.Info("shutdown signal received, stopping gracefully")

			if err := shutdownHandler.Shutdown(); err != nil {
				logger.Error("shutdown error", zap.Error(err))
				return err
			}

			logger.Info("shutdown complete")
			return nil

		case <-shutdownHandler.WaitReload():
			logger.Info("reload signal (SIGHUP) received, reloading configuration")

			// Log working directory for relative path resolution
			workDir, err := os.Getwd()
			if err != nil {
				logger.Warn("failed to get working directory", zap.Error(err))
				workDir = "unknown"
			}

			// Resolve config path to absolute for logging
			absConfigPath, err := filepath.Abs(configPath)
			if err != nil {
				absConfigPath = configPath
			}

			// Reload configuration
			newCfg, err := config.Load(configPath)
			if err != nil {
				logger.Error("failed to reload configuration", zap.Error(err))
				continue
			}

			// Validate new configuration
			if err := config.Validate(newCfg); err != nil {
				logger.Error("invalid configuration, keeping current config", zap.Error(err))
				continue
			}

			// Stop current scheduler
			scheduler.Stop()

			// Update configuration
			cfg = newCfg
			logger.Info("configuration reloaded",
				zap.String("config_file", absConfigPath),
				zap.String("working_directory", workDir),
				zap.Int("secret_count", len(cfg.Secrets)),
			)

			// Restart scheduler with new secrets
			scheduler = syncer.NewScheduler(secretSyncer)
			for _, secret := range cfg.Secrets {
				scheduler.AddSecret(cfg, secret)
				logger.Info("secret sync restarted",
					zap.String("name", secret.Name),
					zap.Duration("refresh_interval", secret.RefreshInterval),
				)
			}

			metrics.SetSecretsConfigured(len(cfg.Secrets))
		}
	}
}

func isReady() int {
	envCfg := config.LoadEnvConfig()

	if err := health.CheckReadiness(envCfg.StatusFile); err != nil {
		return 1
	}

	return 0
}
