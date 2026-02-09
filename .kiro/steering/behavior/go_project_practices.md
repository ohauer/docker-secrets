# Go Project Best Practices

## Version Management

### Version Variables
Always define version variables that can be injected at build time:

```go
var (
    Version   = "0.1.0"
    GitCommit = "dev"
    BuildDate = "unknown"
)
```

### Makefile Version Injection
Use ldflags to inject version info:

```makefile
VERSION=0.1.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)"

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/app ./cmd/app
```

### Version Display Format
- Release builds: `0.1.0`
- Development builds: `dev-<git-hash>`
- Show build date, Go version, OS/Arch

## CLI Command Structure

### Subcommand Pattern
Always implement these standard subcommands:

1. `help` / `-h` / `--help` - Usage information
2. `version` / `-v` / `--version` - Version info
3. `init` - Generate sample configuration
4. `validate` - Validate configuration without running

### Main Function Pattern
```go
func main() {
    if len(os.Args) > 1 {
        switch os.Args[1] {
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
        default:
            printUsage()
            os.Exit(1)
        }
    }
    
    if err := run(); err != nil {
        // Handle error
        os.Exit(1)
    }
}
```

## Configuration Management

### Environment Variable Precedence
Environment variables should always override config file values:

```go
// Load from config file
tlsConfig := &TLSConfig{
    CACert: cfg.SecretStore.TLSCACert,
}

// Override with environment variables
if envCfg.VaultCACert != "" {
    tlsConfig.CACert = envCfg.VaultCACert
}
```

### Configuration Validation
Always validate configuration before use:
- Check required fields
- Validate file paths exist
- Validate paired values (e.g., cert + key)
- Provide clear error messages

### Sample Configuration Generation
The `init` command should generate a fully commented example configuration with:
- All available options
- Explanations for each field
- Common use cases
- Security best practices

## TLS/Certificate Handling

### Certificate Generation Script
Provide a script to generate test certificates:
- CA certificate and key
- Server certificates with SANs
- Client certificates for mTLS
- Proper file permissions (644 for certs, 600 for keys)

### TLS Configuration Fields
Standard TLS configuration should include:
- `tlsCACert` - Custom CA certificate file
- `tlsCAPath` - CA certificate directory
- `tlsSkipVerify` - Skip verification (dev only, warn user)
- `tlsClientCert` - Client certificate (mTLS)
- `tlsClientKey` - Client key (mTLS)

### TLS Validation
- Verify CA cert/path files exist
- Ensure client cert and key are both provided or both empty
- Check file permissions
- Validate certificate files are readable

## Testing

### Test Organization
- Unit tests: `*_test.go` in same package
- Integration tests: Separate test files with build tags
- Use `t.TempDir()` for temporary files
- Clean up resources in defer statements

### Test Certificates
For TLS tests, generate minimal valid certificates:
```go
tmpDir := t.TempDir()
caCert := filepath.Join(tmpDir, "ca.pem")
os.WriteFile(caCert, []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"), 0644)
```

### Race Detection
Always run tests with race detector:
```makefile
test:
	go test -v -race -coverprofile=coverage.out ./...
```

Use `atomic` package for shared counters in tests:
```go
var count int32
atomic.AddInt32(&count, 1)
```

## Docker Compose for Development

### Multiple Service Variants
Provide both HTTP and TLS variants:
```yaml
services:
  vault:        # HTTP on port 8200
  vault-tls:    # HTTPS on port 8201
```

### Health Checks
Use appropriate health check commands:
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-check-certificate", "-q", "-O-", "https://127.0.0.1:8200/v1/sys/health"]
```

### Certificate Mounting
Mount certificates read-only:
```yaml
volumes:
  - ./test-run/certs:/certs:ro
```

## Code Quality

### Trailing Whitespace
Always trim trailing whitespace:
```bash
find . -type f \( -name "*.go" -o -name "*.md" \) -exec sed -i 's/[[:space:]]*$//' {} +
```

### Linting
- Run linter before commits: `make lint`
- Fix all linter issues before merging
- Use `golangci-lint` with standard configuration

### Error Handling
- Always check error returns
- Use `_` to explicitly ignore errors when intentional
- Wrap errors with context: `fmt.Errorf("context: %w", err)`

## Documentation

### README Structure
1. Features (with emojis for visual appeal)
2. Quick Start (minimal steps to get running)
3. Installation
4. Usage (command line examples)
5. Configuration (link to detailed docs)
6. Examples (docker-compose, configs)
7. Troubleshooting (link to guide)

### Separate Documentation Files
- `docs/configuration.md` - Detailed configuration reference
- `docs/environment-variables.md` - All env vars with descriptions
- `docs/troubleshooting.md` - Common issues and solutions

### Configuration Documentation
Always include:
- Required vs optional fields
- Default values
- Examples for common use cases
- Security considerations
- Environment variable equivalents

## Security

### Secrets in Logs
- Never log secret values
- Use structured logging with metadata only
- Log secret names/paths, not contents

### File Permissions
- Default to restrictive permissions (0600)
- Make permissions configurable
- Document security implications

### File Operations Security

#### Atomic Writes with Random Temp Files
Always use atomic writes with unpredictable temp file names:
```go
// GOOD: Random suffix prevents TOCTOU attacks
tmpFile := config.Path + ".tmp." + randomString(8)
os.WriteFile(tmpFile, content, mode)
os.Rename(tmpFile, config.Path)  // Atomic

// BAD: Predictable name allows symlink attacks
tmpFile := config.Path + ".tmp"  // Attacker can pre-create symlink
```

#### Path Validation
Validate all file paths with multiple checks:
```go
func validatePath(path string) error {
    // 1. Check length against OS limits
    if len(path) > MaxPathLen {
        return fmt.Errorf("path too long")
    }
    
    // 2. Reject Windows special paths
    if strings.HasPrefix(path, `\\?\`) {
        return fmt.Errorf("extended paths not allowed")
    }
    if strings.HasPrefix(path, `\\`) {
        return fmt.Errorf("UNC paths not allowed")
    }
    
    // 3. Require absolute paths
    if !filepath.IsAbs(path) {
        return fmt.Errorf("path must be absolute")
    }
    
    // 4. Reject path traversal
    if strings.Contains(path, "..") {
        return fmt.Errorf("path traversal not allowed")
    }
    
    return nil
}
```

#### File Type Validation
Reject symlinks and special files:
```go
func validateFileType(path string) error {
    info, err := os.Lstat(path)  // Don't follow symlinks
    if err != nil {
        if os.IsNotExist(err) {
            return nil  // File doesn't exist yet, OK
        }
        return err
    }
    
    // Reject symlinks
    if info.Mode()&os.ModeSymlink != 0 {
        return fmt.Errorf("symlinks not allowed")
    }
    
    // Reject special files (devices, pipes, sockets)
    if !info.Mode().IsRegular() && !info.Mode().IsDir() {
        return fmt.Errorf("only regular files allowed")
    }
    
    return nil
}
```

#### OS-Specific Path Limits
Use build tags for platform-specific constants:
```go
// limits_unix.go
// +build linux darwin freebsd openbsd netbsd dragonfly solaris aix
const MaxPathLen = 4096  // PATH_MAX

// limits_windows.go
// +build windows
const MaxPathLen = 260   // MAX_PATH (conservative)
```

### Resource Limits
Always enforce limits to prevent DoS:
```go
const (
    MaxSecretSize    = 1 * 1024 * 1024   // 1MB per secret
    MaxResponseSize  = 10 * 1024 * 1024  // 10MB from API
    MaxSecretCount   = 100                // Total secrets
    MinRefreshInterval = 30 * time.Second // Prevent hammering
)
```

### Cleanup Orphaned Files
Clean up temporary files on startup:
```go
func CleanupOrphanedTempFiles(dirs []string) error {
    for _, dir := range dirs {
        pattern := filepath.Join(dir, "*.tmp.*")
        matches, _ := filepath.Glob(pattern)
        for _, file := range matches {
            os.Remove(file)
        }
    }
    return nil
}
```

### Signal Handling
Support graceful shutdown and reload:
```go
// Shutdown signals
signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

// Reload signal (SIGHUP)
signal.Notify(reloadCh, syscall.SIGHUP)

// Systemd integration
ExecReload=/bin/kill -HUP $MAINPID
```

### Fuzzing
Add fuzz tests for all input validation:
```go
func FuzzValidatePath(f *testing.F) {
    f.Add("/tmp/test")
    f.Add("../../../etc/passwd")
    f.Add("/dev/null")
    
    f.Fuzz(func(t *testing.T, path string) {
        _ = validatePath(path)  // Should not panic
    })
}
```

Run fuzzing in CI:
```makefile
fuzz:
	go test -fuzz=. -fuzztime=10s ./...
```

### Container Security
- Use `FROM scratch` for minimal attack surface
- Run as non-root user (UID 65534)
- No shell or package manager in final image
- Static binary compilation

## Project Organization

### Test Environment
Create a `test-run/` directory for testing:
- Add to `.gitignore`
- Include setup scripts
- Provide sample configurations
- Document usage in `test-run/README.md`

### Scripts
Provide helper scripts:
- `setup-*.sh` - Setup development environment
- `run-*.sh` - Run the tool with test config
- `generate-*.sh` - Generate test data/certificates
- `verify-*.sh` - Verify results

All scripts should:
- Be POSIX-compliant
- Include logging with timestamps
- Have clear success/failure messages
- Be executable (`chmod +x`)
