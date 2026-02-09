# Systemd Deployment Guide

This guide covers deploying secrets-sync as a systemd service on Linux systems.

## Overview

Running secrets-sync as a systemd service provides:
- Automatic startup on boot
- Automatic restart on failure
- Centralized logging via journald
- Resource limits and security hardening
- Integration with system management tools

## Prerequisites

- Linux system with systemd (most modern distributions)
- Root access for installation
- Go 1.25.6+ (for building from source)

## Quick Installation

### Automated Installation

```bash
# Build and install in one command
sudo make install-systemd
```

This will:
1. Build the binary
2. Create `secrets-sync` system user and group
3. Install to `/usr/local/bin/secrets-sync`
4. Create config directory `/etc/secrets-sync`
5. Generate sample config
6. Install systemd unit file
7. Enable the service

**Important**: You must manually create output directories before starting the service.

### Manual Installation

See [examples/systemd/README.md](../examples/systemd/README.md) for step-by-step manual installation.

## Configuration

### 1. Edit Configuration File

```bash
sudo nano /etc/secrets-sync/config.yaml
```

Example configuration:

```yaml
secretStore:
  address: "https://vault.example.com:8200"
  authMethod: "token"
  token: "${VAULT_TOKEN}"

secrets:
  - name: "app-secrets"
    key: "app/prod/secrets"
    mountPath: "secret"
    kvVersion: "v2"
    refreshInterval: "5m"
    template:
      data:
        db_password: '{{ .db_password }}'
        api_key: '{{ .api_key }}'
    files:
      - path: "/var/secrets/db_password"
        mode: "0600"
      - path: "/var/secrets/api_key"
        mode: "0600"
```

### 2. Create Output Directories

**Important**: Create directories before starting the service:

```bash
# Create directory for secrets
sudo mkdir -p /var/secrets

# Set ownership to secrets-sync user
sudo chown secrets-sync:secrets-sync /var/secrets

# Set permissions (750 = owner+group can access)
sudo chmod 750 /var/secrets
```

### 3. Update Systemd Unit File

Edit `/etc/systemd/system/secrets-sync.service` and add your output paths:

```ini
# Filesystem protection
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/secrets
```

Reload systemd after editing:
```bash
sudo systemctl daemon-reload
```

### 4. Set Environment Variables

```bash
sudo nano /etc/default/secrets-sync
```

For token authentication:
```bash
VAULT_ADDR=https://vault.example.com:8200
VAULT_TOKEN=s.xxxxxxxxxxxxx
LOG_LEVEL=info
```

For AppRole authentication:
```bash
VAULT_ADDR=https://vault.example.com:8200
VAULT_ROLE_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
VAULT_SECRET_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
LOG_LEVEL=info
```

### 3. Customize Unit File (Optional)

Edit `/etc/systemd/system/secrets-sync.service` to adjust:

**Network Access**:
```ini
# Allow access to your Vault server
IPAddressAllow=10.0.0.0/8
IPAddressAllow=192.168.1.100
```

**File Paths**:
```ini
# Add paths where secrets will be written
# These directories must exist and be owned by secrets-sync user
ReadWritePaths=/var/secrets
ReadWritePaths=/app/secrets
ReadWritePaths=/etc/nginx/ssl
```

Create directories first:
```bash
sudo mkdir -p /var/secrets /app/secrets /etc/nginx/ssl
sudo chown secrets-sync:secrets-sync /var/secrets /app/secrets
sudo chown secrets-sync:nginx /etc/nginx/ssl
sudo chmod 750 /var/secrets /app/secrets /etc/nginx/ssl
```

**Sharing Secrets with Other Services**:
```bash
# Add other service users to secrets-sync group
sudo usermod -a -G secrets-sync nginx
sudo usermod -a -G secrets-sync postgres

# Set appropriate file permissions in config
# mode: "0640" allows group read access
```

## Service Management

### Start the Service

```bash
sudo systemctl start secrets-sync
```

### Enable Auto-Start on Boot

```bash
sudo systemctl enable secrets-sync
```

### Check Status

```bash
sudo systemctl status secrets-sync
```

### View Logs

```bash
# Follow logs in real-time
sudo journalctl -u secrets-sync -f

# View last 100 lines
sudo journalctl -u secrets-sync -n 100

# View logs since boot
sudo journalctl -u secrets-sync -b

# View logs for specific time range
sudo journalctl -u secrets-sync --since "1 hour ago"
```

### Restart Service

```bash
sudo systemctl restart secrets-sync
```

### Stop Service

```bash
sudo systemctl stop secrets-sync
```

### Reload Configuration

If `WATCH_CONFIG=true` is set, the service will automatically reload when the config file changes. Otherwise, restart the service:

```bash
sudo systemctl restart secrets-sync
```

## Health Monitoring

### Check Service Health

```bash
# Using the isready command
secrets-sync isready

# Using HTTP endpoint
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

### View Metrics

```bash
curl http://localhost:8080/metrics
```

### Integration with Monitoring Systems

The service exposes Prometheus metrics on port 8080 (configurable via `METRICS_PORT`):

```bash
# Add to prometheus.yml
scrape_configs:
  - job_name: 'secrets-sync'
    static_configs:
      - targets: ['localhost:8080']
```

## Security Hardening

The provided unit file includes extensive security hardening:

### Filesystem Protection
- `ProtectSystem=strict` - Read-only root filesystem
- `ProtectHome=yes` - No access to home directories
- `PrivateTmp=yes` - Private /tmp directory
- `ReadWritePaths` - Explicit write permissions

### Network Restrictions
- `RestrictAddressFamilies` - Limited to IPv4/IPv6/Unix sockets
- `IPAddressDeny=any` - Deny all by default
- `IPAddressAllow` - Explicit allow list

### Capabilities
- `CapabilityBoundingSet=` - No special capabilities
- `NoNewPrivileges=yes` - Cannot gain privileges

### System Calls
- `SystemCallFilter=@system-service` - Safe system calls only
- `SystemCallFilter=~@privileged` - Block privileged calls

### User Isolation
- `User=secrets-sync` - Dedicated system user
- `Group=secrets-sync` - Dedicated system group
- `ConfigurationDirectory=secrets-sync` - Automatic /etc/secrets-sync creation
- `RuntimeDirectory=secrets-sync` - Automatic /run/secrets-sync creation

**Why Static User Instead of DynamicUser?**

The service uses a static user (`secrets-sync`) rather than `DynamicUser=yes` because:
1. **Persistent UID** - Files maintain correct ownership across reboots/reinstalls
2. **Flexible Paths** - Can write to arbitrary paths (not limited to StateDirectory)
3. **Group Sharing** - Other services can access secrets via group membership
4. **External Access** - Services like nginx/postgres can read the secret files

With `DynamicUser=yes`, the UID changes on each restart, breaking file ownership.

## Troubleshooting

### Service Fails to Start

**Check logs**:
```bash
sudo journalctl -u secrets-sync -n 50 --no-pager
```

**Common issues**:
1. Config file not found or invalid
2. Vault connection issues
3. Permission denied on file paths
4. Network restrictions blocking Vault access

**Validate configuration**:
```bash
secrets-sync --config /etc/secrets-sync/config.yaml validate
```

### Permission Denied Errors

**Symptom**: Service fails with "permission denied" errors

**Solution**: Add required paths to `ReadWritePaths`:
```ini
ReadWritePaths=/var/lib/secrets-sync
ReadWritePaths=/path/to/your/secrets
```

Then reload:
```bash
sudo systemctl daemon-reload
sudo systemctl restart secrets-sync
```

### Network Connection Failures

**Symptom**: Cannot connect to Vault server

**Solution**: Add Vault server IP to `IPAddressAllow`:
```ini
IPAddressAllow=localhost
IPAddressAllow=10.0.0.0/8
IPAddressAllow=your.vault.server.ip
```

Then reload:
```bash
sudo systemctl daemon-reload
sudo systemctl restart secrets-sync
```

### Config File Not Found

**Symptom**: "failed to read config file" error

**Solution**: Ensure config exists and is readable:
```bash
sudo ls -la /etc/secrets-sync/config.yaml
sudo chmod 644 /etc/secrets-sync/config.yaml
```

### Secrets Not Being Written

**Check**:
1. File paths in config are correct
2. Paths are added to `ReadWritePaths` in unit file
3. Vault credentials are correct
4. Secret paths exist in Vault

**Debug**:
```bash
# Test manually
sudo -u secrets-sync /usr/local/bin/secrets-sync \
  --config /etc/secrets-sync/config.yaml

# Check Vault connectivity
curl -H "X-Vault-Token: $VAULT_TOKEN" \
  https://vault.example.com:8200/v1/sys/health
```

## Advanced Configuration

### Sharing Secrets with Other Services

To allow other services (nginx, postgres, etc.) to read secrets:

**Option 1: Group Access (Recommended)**
```bash
# Add service users to secrets-sync group
sudo usermod -a -G secrets-sync nginx
sudo usermod -a -G secrets-sync postgres

# Set group-readable permissions in config
files:
  - path: "/secrets/nginx.crt"
    mode: "0640"  # Owner: rw, Group: r, Other: none
```

**Option 2: World-Readable (Less Secure)**
```yaml
files:
  - path: "/secrets/public.crt"
    mode: "0644"  # Anyone can read
```

**Option 3: Specific Directory Permissions**
```bash
# Pre-create directory with specific ownership
sudo mkdir -p /etc/nginx/ssl
sudo chown secrets-sync:nginx /etc/nginx/ssl
sudo chmod 750 /etc/nginx/ssl

# Add to systemd unit file
ReadWritePaths=/etc/nginx/ssl
```

### Custom User and Group

The installation script automatically creates the `secrets-sync` user and group.

If you need to recreate it manually:
```bash
sudo useradd -r -s /bin/false -d /nonexistent -c "Secrets Sync Service" secrets-sync
```

Create output directories (example):
```bash
sudo mkdir -p /var/secrets
sudo chown secrets-sync:secrets-sync /var/secrets
sudo chmod 750 /var/secrets
```

**Important**: The service will fail to start if output directories don't exist or aren't writable by the secrets-sync user. This is intentional - you must explicitly create and configure the directories you want to use.

### Multiple Instances

Run multiple instances with different configs:

```bash
# Copy unit file
sudo cp /etc/systemd/system/secrets-sync.service \
       /etc/systemd/system/secrets-sync@.service

# Edit to use instance name
ExecStart=/usr/local/bin/secrets-sync --config /etc/secrets-sync/%i.yaml

# Start instances
sudo systemctl start secrets-sync@app1
sudo systemctl start secrets-sync@app2
```

### Resource Limits

Add to unit file:
```ini
[Service]
MemoryLimit=256M
CPUQuota=50%
TasksMax=10
```

### Automatic Restart Policy

Customize restart behavior:
```ini
[Service]
Restart=on-failure
RestartSec=5s
StartLimitBurst=5
StartLimitIntervalSec=60s
```

## Uninstallation

### Automated Uninstallation

```bash
sudo make uninstall-systemd
```

This will:
1. Stop the service
2. Disable the service
3. Remove unit file
4. Remove binary
5. Prompt to remove config directory

### Manual Uninstallation

```bash
# Stop and disable
sudo systemctl stop secrets-sync
sudo systemctl disable secrets-sync

# Remove files
sudo rm /etc/systemd/system/secrets-sync.service
sudo rm /etc/default/secrets-sync
sudo rm /usr/local/bin/secrets-sync
sudo rm -rf /etc/secrets-sync

# Reload systemd
sudo systemctl daemon-reload
```

## Best Practices

1. **Use AppRole for Production**: More secure than tokens
2. **Enable Config Watching**: Set `WATCH_CONFIG=true` for hot reload
3. **Monitor Metrics**: Integrate with Prometheus/Grafana
4. **Rotate Credentials**: Regularly rotate Vault tokens/AppRole credentials
5. **Limit Network Access**: Use `IPAddressAllow` to restrict connections
6. **Use Specific Paths**: Add only required paths to `ReadWritePaths`
7. **Review Logs Regularly**: Check for errors and warnings
8. **Test Configuration**: Always validate config before deploying

## See Also

- [Configuration Guide](configuration.md)
- [Environment Variables](environment-variables.md)
- [Systemd Examples](../examples/systemd/)
- [Main README](../README.md)
