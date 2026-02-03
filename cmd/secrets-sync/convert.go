package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ExternalSecret represents the external-secrets.io format
type ExternalSecret struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		RefreshInterval string `yaml:"refreshInterval"`
		SecretStoreRef  struct {
			Kind string `yaml:"kind"`
			Name string `yaml:"name"`
		} `yaml:"secretStoreRef"`
		Target struct {
			Name     string `yaml:"name"`
			Template struct {
				Data map[string]string `yaml:"data"`
			} `yaml:"template"`
		} `yaml:"target"`
		Data []struct {
			SecretKey string `yaml:"secretKey"`
			RemoteRef struct {
				Key      string `yaml:"key"`
				Property string `yaml:"property"`
			} `yaml:"remoteRef"`
		} `yaml:"data"`
		DataFrom []struct {
			Extract struct {
				Key string `yaml:"key"`
			} `yaml:"extract"`
		} `yaml:"dataFrom"`
	} `yaml:"spec"`
}

// ConvertConfig holds conversion parameters
type ConvertConfig struct {
	MountPath      string
	KVVersion      string
	OutputDir      string
	AutoDetectMount bool
}

// detectMountPath tries to infer mount path from secret key
func detectMountPath(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) == 0 {
		return "secret"
	}

	// Common patterns
	firstSegment := parts[0]
	switch firstSegment {
	case "common":
		return "devops"
	case "artifactory":
		return "artifactory"
	case "infra":
		return "infra"
	default:
		return "secret"
	}
}

func convertExternalSecret(inputFile string, cfg ConvertConfig) error {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var es ExternalSecret
	if err := yaml.Unmarshal(data, &es); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if es.Kind != "ExternalSecret" {
		return fmt.Errorf("not an ExternalSecret (kind: %s)", es.Kind)
	}

	// Build secret configuration
	secretName := es.Spec.Target.Name
	if secretName == "" {
		secretName = es.Metadata.Name
	}

	refreshInterval := es.Spec.RefreshInterval
	if refreshInterval == "" {
		refreshInterval = "30m"
	}

	// Determine mount path
	mountPath := cfg.MountPath
	var key string

	// Handle dataFrom.extract (pulls all fields)
	if len(es.Spec.DataFrom) > 0 {
		key = es.Spec.DataFrom[0].Extract.Key
	} else if len(es.Spec.Data) > 0 {
		key = es.Spec.Data[0].RemoteRef.Key
	}

	// Auto-detect mount path if enabled
	if cfg.AutoDetectMount && key != "" {
		mountPath = detectMountPath(key)
	}

	fmt.Printf("\n# Converted from: %s\n", filepath.Base(inputFile))
	fmt.Printf("  - name: %q\n", secretName)

	// Handle dataFrom.extract (pulls all fields)
	if len(es.Spec.DataFrom) > 0 {
		fmt.Printf("    key: %q\n", key)
		fmt.Printf("    mountPath: %q\n", mountPath)
		fmt.Printf("    kvVersion: %q\n", cfg.KVVersion)
		fmt.Printf("    refreshInterval: %q\n", refreshInterval)
		fmt.Printf("    # Note: Uses dataFrom.extract - pulls ALL fields from secret\n")
		fmt.Printf("    # Check actual fields with: vault kv get %s/%s\n", cfg.MountPath, key)
		fmt.Printf("    template:\n")
		fmt.Printf("      data:\n")

		// Use template if provided
		if len(es.Spec.Target.Template.Data) > 0 {
			for k, v := range es.Spec.Target.Template.Data {
				fmt.Printf("        %s: %q\n", k, v)
			}
		} else {
			fmt.Printf("        # TODO: Add template mappings based on actual secret fields\n")
			fmt.Printf("        example-field: '{{ .fieldName }}'\n")
		}

		fmt.Printf("    files:\n")
		if len(es.Spec.Target.Template.Data) > 0 {
			for k := range es.Spec.Target.Template.Data {
				fmt.Printf("      - path: %q\n", filepath.Join(cfg.OutputDir, secretName, k))
				fmt.Printf("        mode: \"0600\"\n")
			}
		} else {
			fmt.Printf("      - path: %q\n", filepath.Join(cfg.OutputDir, secretName, "data"))
			fmt.Printf("        mode: \"0600\"\n")
		}
		return nil
	}

	// Handle data[] (specific fields)
	if len(es.Spec.Data) > 0 {
		fmt.Printf("    key: %q\n", key)
		fmt.Printf("    mountPath: %q\n", mountPath)
		fmt.Printf("    kvVersion: %q\n", cfg.KVVersion)
		fmt.Printf("    refreshInterval: %q\n", refreshInterval)
		fmt.Printf("    template:\n")
		fmt.Printf("      data:\n")

		for _, d := range es.Spec.Data {
			if d.RemoteRef.Property != "" {
				fmt.Printf("        %s: '{{ .%s }}'\n", d.SecretKey, d.RemoteRef.Property)
			} else {
				fmt.Printf("        %s: '{{ . }}'\n", d.SecretKey)
			}
		}

		fmt.Printf("    files:\n")
		for _, d := range es.Spec.Data {
			fmt.Printf("      - path: %q\n", filepath.Join(cfg.OutputDir, secretName, d.SecretKey))
			fmt.Printf("        mode: \"0600\"\n")
		}
		return nil
	}

	return fmt.Errorf("no data or dataFrom found in ExternalSecret")
}

func runConvert(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: secrets-sync convert <external-secret-files...> [options]\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --mount-path <path>   KV mount path (default: secret)\n")
		fmt.Fprintf(os.Stderr, "  --kv-version <v1|v2>  KV version (default: v2)\n")
		fmt.Fprintf(os.Stderr, "  --output-dir <dir>    Output directory for secrets (default: ./secrets)\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  secrets-sync convert external-secret.yaml --mount-path devops\n")
		return 1
	}

	cfg := ConvertConfig{
		MountPath:      "secret",
		KVVersion:      "v2",
		OutputDir:      "./secrets",
		AutoDetectMount: true, // Enable by default
	}

	var files []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--mount-path":
			if i+1 < len(args) {
				cfg.MountPath = args[i+1]
				cfg.AutoDetectMount = false // Disable auto-detect if explicitly set
				i++
			}
		case "--kv-version":
			if i+1 < len(args) {
				cfg.KVVersion = args[i+1]
				i++
			}
		case "--output-dir":
			if i+1 < len(args) {
				cfg.OutputDir = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(arg, "--") {
				files = append(files, arg)
			}
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no input files specified\n")
		return 1
	}

	fmt.Println("# Generated configuration from external-secrets")
	fmt.Println("# Review and adjust template fields as needed")
	fmt.Println()
	fmt.Println("secrets:")

	for _, file := range files {
		if err := convertExternalSecret(file, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error converting %s: %v\n", file, err)
			continue
		}
	}

	return 0
}
