# Future Enhancement: Separate Credentials File

## Problem
When storing configuration in git, credentials should not be committed. While environment variables work, a separate credentials file provides better organization for multiple credential sets.

## Current State
- Credentials are defined inline in `config.yaml`
- Environment variable expansion is supported: `${TEAM_A_TOKEN}`
- Works well but requires many environment variables for multiple teams

## Proposed Solution: External Credentials File

### Configuration Structure

**`config.yaml`** (safe to commit to git):
```yaml
secretStore:
  address: "https://vault.example.com"
  authMethod: "token"
  token: "${VAULT_TOKEN}"

  # Reference to external credentials file
  credentialsFile: "/etc/secrets-sync/credentials.yaml"
  # OR use environment variable
  credentialsFile: "${CREDENTIALS_FILE}"

secrets:
  - name: "team-a-secret"
    credentials: "team-a"  # References credential set from external file
```

**`credentials.yaml`** (in .gitignore, not committed):
```yaml
# Credentials file - DO NOT COMMIT TO GIT
credentials:
  team-a:
    authMethod: "token"
    token: "actual-token-value"

  team-b:
    authMethod: "approle"
    roleId: "actual-role-id"
    secretId: "actual-secret-id"
```

### Implementation Tasks

1. **Add `credentialsFile` field to `SecretStore`**
   - Optional field: `credentialsFile string`
   - Can be absolute or relative path
   - Supports environment variable expansion

2. **Create credentials file loader**
   - New file: `internal/config/credentials_loader.go`
   - Function: `LoadCredentials(path string) (map[string]CredentialSet, error)`
   - Validates credential sets after loading

3. **Merge credentials in config loader**
   - Load main config
   - If `credentialsFile` is set, load external credentials
   - Merge external credentials with inline credentials
   - External file takes precedence over inline

4. **Update validation**
   - Validate `credentialsFile` path exists if specified
   - Validate merged credentials

5. **Add to .gitignore**
   ```
   credentials.yaml
   *-credentials.yaml
   .env
   ```

6. **Create example files**
   - `examples/credentials.yaml.example`
   - Document in README and configuration docs

7. **Update init command**
   - Add commented `credentialsFile` option
   - Generate `credentials.yaml.example`

### Benefits
- Clean separation of config (git) and secrets (not in git)
- Single file for all credentials (easier to manage than many env vars)
- Still supports environment variables for maximum flexibility
- Backward compatible (credentialsFile is optional)

### Estimated Effort
- Implementation: 1-2 hours
- Testing: 30 minutes
- Documentation: 30 minutes
- Total: ~2-3 hours

### Priority
Medium - Nice to have for better credential management in multi-team environments

### Related
- Could be combined with SOPS/git-crypt for encrypted credentials in git
- Could support multiple credential files (e.g., per environment)
