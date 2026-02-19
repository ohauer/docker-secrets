package main

import (
	"fmt"
	"os"

	"github.com/ohauer/docker-secrets/internal/config"
)

func validateConfig(configFile string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("✓ Configuration is valid\n")
	fmt.Printf("  Vault address: %s\n", cfg.SecretStore.Address)
	if cfg.SecretStore.Namespace != "" {
		fmt.Printf("  Namespace:     %s\n", cfg.SecretStore.Namespace)
	}
	fmt.Printf("  Auth method:   %s\n", cfg.SecretStore.AuthMethod)
	fmt.Printf("  Secrets:       %d configured\n", len(cfg.Secrets))

	return nil
}

func runValidate() int {
	configPath := getConfigFile()

	if err := validateConfig(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}
