# Contributing to Docker Secrets Sidecar

Thank you for your interest in contributing to Docker Secrets Sidecar! We welcome contributions from the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/docker-secrets.git`
3. Add upstream remote: `git remote add upstream https://github.com/ohauer/docker-secrets.git`
4. Create a feature branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.25.6 or higher
- Docker and Docker Compose
- Make
- golangci-lint (for linting)

### Build and Test

```bash
# Install dependencies
go mod download

# Build the binary
make build

# Run tests
make test

# Run linter
make lint

# Run all checks
make all
```

### Development Environment

```bash
# Start Vault for testing
docker compose -f docker-compose.vault.yml up -d

# Generate test certificates
cd test-run && ./generate-certs.sh

# Run the tool
CONFIG_FILE=test-run/config.yaml ./bin/secrets-sync
```

## How to Contribute

### Types of Contributions

We welcome:

- **Bug fixes** - Fix issues reported in GitHub Issues
- **Features** - Add new functionality (discuss in an issue first)
- **Documentation** - Improve or add documentation
- **Tests** - Add or improve test coverage
- **Examples** - Add usage examples or tutorials

### Before You Start

1. Check existing [issues](https://github.com/ohauer/docker-secrets/issues) and [pull requests](https://github.com/ohauer/docker-secrets/pulls)
2. For major changes, open an issue first to discuss your proposal
3. Ensure your changes align with the project's goals and architecture

## Coding Standards

We follow the guidelines in [`.kiro/steering/behavior/go_project_practices.md`](.kiro/steering/behavior/go_project_practices.md).

### Key Standards

- **Go Style**: Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- **Formatting**: Run `make fmt` before committing
- **Linting**: All code must pass `make lint` with zero issues
- **Testing**: Maintain or improve test coverage
- **Documentation**: Update docs for any user-facing changes
- **Commit Messages**: Use clear, descriptive commit messages

### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat: add support for Kubernetes auth method

- Implement Kubernetes authentication
- Add configuration options
- Update documentation

Closes #123
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/vault -v

# Run with race detection
go test -race ./...

# Generate coverage report
make coverage
```

### Writing Tests

- Write tests for all new code
- Maintain or improve coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Use `t.TempDir()` for temporary files

Example:
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "result", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFeature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Pull Request Process

1. **Update your branch** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Ensure all checks pass**:
   ```bash
   make all
   ```

3. **Update documentation** if needed:
   - README.md
   - docs/ directory
   - Code comments
   - CHANGELOG.md (if exists)

4. **Create a pull request**:
   - Use a clear, descriptive title
   - Fill out the PR template completely
   - Reference related issues
   - Add screenshots for UI changes
   - Request review from maintainers

5. **Address review feedback**:
   - Make requested changes
   - Push updates to your branch
   - Respond to comments

6. **Merge requirements**:
   - All CI checks must pass
   - At least one maintainer approval
   - No unresolved conversations
   - Up to date with main branch

## Reporting Bugs

Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md) when reporting bugs.

### Before Reporting

1. Check if the bug has already been reported
2. Verify you're using the latest version
3. Test with a minimal reproduction case

### Include in Your Report

- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)
- Relevant logs or error messages
- Configuration files (sanitized)

## Suggesting Enhancements

Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md) for enhancement suggestions.

### Before Suggesting

1. Check if the feature has been requested
2. Ensure it aligns with project goals
3. Consider if it could be a plugin/extension

### Include in Your Suggestion

- Clear description of the feature
- Use cases and benefits
- Possible implementation approach
- Examples from similar projects

## Development Tips

### Useful Commands

```bash
# Format code
make fmt

# Run linter
make lint

# Build Docker image
make docker-build

# Clean build artifacts
make clean

# View all make targets
make help
```

### Debugging

```bash
# Run with debug logging
LOG_LEVEL=debug ./bin/secrets-sync

# Run with race detector
go run -race ./cmd/secrets-sync

# Profile the application
go test -cpuprofile=cpu.prof -memprofile=mem.prof
```

### Project Structure

```
.
├── cmd/secrets-sync/     # Main application
├── internal/             # Internal packages
│   ├── config/          # Configuration
│   ├── vault/           # Vault client
│   ├── syncer/          # Secret synchronization
│   └── ...
├── docs/                # Documentation
├── examples/            # Usage examples
├── test-run/            # Test environment
└── .kiro/steering/      # Project guidelines
```

## Questions?

- Open a [Discussion](https://github.com/ohauer/docker-secrets/discussions)
- Join our community chat (if available)
- Check existing documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- GitHub contributors page
- Release notes
- Project documentation

Thank you for contributing! 🎉
