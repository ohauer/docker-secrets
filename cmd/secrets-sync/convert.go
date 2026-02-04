package main

import (
	"fmt"
	"os"
	"os/exec"
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
	MountPath       string
	KVVersion       string
	OutputDir       string
	AutoDetectMount bool
	QueryVault      bool
	VaultAddr       string
	VaultToken      string
	VaultRoleID     string
	VaultSecretID   string
}

// getVaultToken obtains a token from AppRole if roleId and secretId are provided
func getVaultToken(vaultAddr, roleId, secretId string) (string, error) {
	if vaultAddr == "" || roleId == "" || secretId == "" {
		return "", fmt.Errorf("vault address, role_id, and secret_id required")
	}

	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("vault write -format=json auth/approle/login role_id=%s secret_id=%s 2>/dev/null | jq -r '.auth.client_token' 2>/dev/null",
			roleId, secretId))
	cmd.Env = append(os.Environ(), "VAULT_ADDR="+vaultAddr)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to authenticate with approle: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" || token == "null" {
		return "", fmt.Errorf("authentication failed: no token returned")
	}

	return token, nil
}

// queryVaultFields queries Vault to get actual field names for a secret
func queryVaultFields(mountPath, key, vaultAddr, vaultToken string) ([]string, error) {
	if vaultAddr == "" || vaultToken == "" {
		return nil, fmt.Errorf("vault address and token required")
	}

	// Use vault CLI to query field names
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("vault kv get -format=json %s/%s 2>/dev/null | jq -r '.data.data | keys[]' 2>/dev/null",
			mountPath, key))
	cmd.Env = append(os.Environ(),
		"VAULT_ADDR="+vaultAddr,
		"VAULT_TOKEN="+vaultToken,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query vault: %w", err)
	}

	fields := strings.Split(strings.TrimSpace(string(output)), "\n")
	var validFields []string
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			validFields = append(validFields, f)
		}
	}

	if len(validFields) == 0 {
		return nil, fmt.Errorf("no fields found")
	}

	return validFields, nil
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

	// Try to parse as a List first
	var list struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Items      []yaml.Node `yaml:"items"`
	}
	
	if err := yaml.Unmarshal(data, &list); err == nil && list.Kind == "List" {
		// Handle Kubernetes List with multiple ExternalSecrets
		for i, item := range list.Items {
			var es ExternalSecret
			if err := item.Decode(&es); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to parse item %d in %s: %v\n", i, inputFile, err)
				continue
			}
			if es.Kind == "ExternalSecret" {
				if err := convertSingleSecret(es, filepath.Base(inputFile), cfg); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to convert item %d in %s: %v\n", i, inputFile, err)
				}
			}
		}
		return nil
	}

	// Try to parse as multi-document YAML
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	count := 0
	for {
		var es ExternalSecret
		if err := decoder.Decode(&es); err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Skip non-ExternalSecret documents
			continue
		}

		if es.Kind == "ExternalSecret" {
			if err := convertSingleSecret(es, filepath.Base(inputFile), cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to convert document %d in %s: %v\n", count, inputFile, err)
			}
			count++
		}
	}

	if count > 0 {
		return nil
	}

	return fmt.Errorf("no ExternalSecret documents found in file")
}

func convertSingleSecret(es ExternalSecret, sourceFile string, cfg ConvertConfig) error {

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

	fmt.Printf("\n# Converted from: %s (secret: %s)\n", sourceFile, secretName)
	
	// Handle dataFrom.extract (pulls all fields)
	if len(es.Spec.DataFrom) > 0 {
		// Try to query vault for actual field names
		var fields []string
		queryFailed := false
		if cfg.QueryVault {
			queriedFields, err := queryVaultFields(mountPath, key, cfg.VaultAddr, cfg.VaultToken)
			if err == nil && len(queriedFields) > 0 {
				fields = queriedFields
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Failed to query %s/%s: %v\n", mountPath, key, err)
				queryFailed = true
			}
		}

		// Comment out the entire secret if query failed and no template provided
		commentPrefix := ""
		if queryFailed && len(es.Spec.Target.Template.Data) == 0 {
			commentPrefix = "# "
			fmt.Printf("# WARNING: Vault query failed - secret commented out, needs manual field mapping\n")
		}

		fmt.Printf("%s  - name: %q\n", commentPrefix, secretName)

		fmt.Printf("%s    key: %q\n", commentPrefix, key)
		fmt.Printf("%s    mountPath: %q\n", commentPrefix, mountPath)
		fmt.Printf("%s    kvVersion: %q\n", commentPrefix, cfg.KVVersion)
		fmt.Printf("%s    refreshInterval: %q\n", commentPrefix, refreshInterval)

		if len(fields) > 0 {
			fmt.Printf("    # Fields queried from Vault\n")
		}

		fmt.Printf("%s    template:\n", commentPrefix)
		fmt.Printf("%s      data:\n", commentPrefix)

		// Use template if provided in external-secret
		if len(es.Spec.Target.Template.Data) > 0 {
			for k, v := range es.Spec.Target.Template.Data {
				fmt.Printf("%s        %s: %q\n", commentPrefix, k, v)
			}
		} else if len(fields) > 0 {
			// Use queried fields from Vault
			for _, field := range fields {
				fmt.Printf("%s        %s: '{{ .%s }}'\n", commentPrefix, field, field)
			}
		} else {
			// Fallback: commented out placeholder
			fmt.Printf("%s        # TODO: Add field mappings, e.g.: username: '{{ .username }}'\n", commentPrefix)
		}

		fmt.Printf("%s    files:\n", commentPrefix)
		if len(es.Spec.Target.Template.Data) > 0 {
			for k := range es.Spec.Target.Template.Data {
				fmt.Printf("%s      - path: %q\n", commentPrefix, filepath.Join(cfg.OutputDir, secretName, k))
				fmt.Printf("%s        mode: \"0600\"\n", commentPrefix)
			}
		} else if len(fields) > 0 {
			// Create one file per field
			for _, field := range fields {
				fmt.Printf("%s      - path: %q\n", commentPrefix, filepath.Join(cfg.OutputDir, secretName, field))
				fmt.Printf("%s        mode: \"0600\"\n", commentPrefix)
			}
		} else {
			// Fallback: commented out placeholder
			fmt.Printf("%s      - path: %q\n", commentPrefix, filepath.Join(cfg.OutputDir, secretName, "field1"))
			fmt.Printf("%s        mode: \"0600\"\n", commentPrefix)
		}
		return nil
	}

	// Handle data[] (specific fields)
	if len(es.Spec.Data) > 0 {
		fmt.Printf("  - name: %q\n", secretName)
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
		fmt.Fprintf(os.Stderr, "  --mount-path <path>      KV mount path (default: auto-detect)\n")
		fmt.Fprintf(os.Stderr, "  --kv-version <v1|v2>     KV version (default: v2)\n")
		fmt.Fprintf(os.Stderr, "  --output-dir <dir>       Output directory for secrets (default: ./secrets)\n")
		fmt.Fprintf(os.Stderr, "  --query-vault            Query Vault for actual field names (requires vault CLI)\n")
		fmt.Fprintf(os.Stderr, "  --vault-addr <url>       Vault address (default: $VAULT_ADDR)\n")
		fmt.Fprintf(os.Stderr, "  --vault-token <token>    Vault token (default: $VAULT_TOKEN)\n")
		fmt.Fprintf(os.Stderr, "  --vault-role-id <id>     Vault AppRole role_id (default: $VAULT_ROLE_ID)\n")
		fmt.Fprintf(os.Stderr, "  --vault-secret-id <id>   Vault AppRole secret_id (default: $VAULT_SECRET_ID)\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  secrets-sync convert external-secret.yaml --query-vault\n")
		fmt.Fprintf(os.Stderr, "  secrets-sync convert external-secret.yaml --query-vault --vault-role-id <id> --vault-secret-id <id>\n")
		return 1
	}

	cfg := ConvertConfig{
		MountPath:       "secret",
		KVVersion:       "v2",
		OutputDir:       "./secrets",
		AutoDetectMount: true,
		QueryVault:      false,
		VaultAddr:       os.Getenv("VAULT_ADDR"),
		VaultToken:      os.Getenv("VAULT_TOKEN"),
		VaultRoleID:     os.Getenv("VAULT_ROLE_ID"),
		VaultSecretID:   os.Getenv("VAULT_SECRET_ID"),
	}

	var files []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--mount-path":
			if i+1 < len(args) {
				cfg.MountPath = args[i+1]
				cfg.AutoDetectMount = false
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
		case "--query-vault":
			cfg.QueryVault = true
		case "--vault-addr":
			if i+1 < len(args) {
				cfg.VaultAddr = args[i+1]
				i++
			}
		case "--vault-token":
			if i+1 < len(args) {
				cfg.VaultToken = args[i+1]
				i++
			}
		case "--vault-role-id":
			if i+1 < len(args) {
				cfg.VaultRoleID = args[i+1]
				i++
			}
		case "--vault-secret-id":
			if i+1 < len(args) {
				cfg.VaultSecretID = args[i+1]
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

	// If query-vault is enabled, ensure we have credentials
	if cfg.QueryVault {
		// Try to get token from AppRole if provided
		if cfg.VaultToken == "" && cfg.VaultRoleID != "" && cfg.VaultSecretID != "" {
			token, err := getVaultToken(cfg.VaultAddr, cfg.VaultRoleID, cfg.VaultSecretID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to authenticate with AppRole: %v\n", err)
				return 1
			}
			cfg.VaultToken = token
		}

		// Verify we have required credentials
		if cfg.VaultAddr == "" || cfg.VaultToken == "" {
			fmt.Fprintf(os.Stderr, "Error: --query-vault requires vault credentials\n")
			fmt.Fprintf(os.Stderr, "Provide either:\n")
			fmt.Fprintf(os.Stderr, "  - VAULT_ADDR and VAULT_TOKEN environment variables\n")
			fmt.Fprintf(os.Stderr, "  - --vault-addr and --vault-token flags\n")
			fmt.Fprintf(os.Stderr, "  - --vault-addr, --vault-role-id, and --vault-secret-id flags\n")
			return 1
		}
	}

	fmt.Println("# Generated configuration from external-secrets")
	fmt.Println("# Review and adjust template fields as needed")
	fmt.Println()

	// Generate secretStore section if vault credentials are available
	if cfg.QueryVault && cfg.VaultAddr != "" {
		fmt.Println("secretStore:")
		fmt.Printf("  address: %q\n", cfg.VaultAddr)
		
		// Use AppRole if role_id/secret_id were provided, otherwise token
		if cfg.VaultRoleID != "" && cfg.VaultSecretID != "" {
			fmt.Println("  authMethod: \"approle\"")
			fmt.Println("  roleId: \"${VAULT_ROLE_ID}\"")
			fmt.Println("  secretId: \"${VAULT_SECRET_ID}\"")
		} else {
			fmt.Println("  authMethod: \"token\"")
			fmt.Println("  token: \"${VAULT_TOKEN}\"")
		}
		fmt.Println()
	}

	fmt.Println("secrets:")

	for _, file := range files {
		if err := convertExternalSecret(file, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error converting %s: %v\n", file, err)
			continue
		}
	}

	return 0
}
