# Systemd Deployment

This directory contains files for deploying secrets-sync as a systemd service.

## Files

- `secrets-sync.service` - Systemd unit file with security hardening
- `secrets-sync.env.example` - Environment variables template

## Quick Start

### 1. Create system user

The service requires a static system user (not DynamicUser) to maintain persistent file ownership:

```bash
# Create secrets-sync system user and group
sudo useradd -r -s /bin/false -d /nonexistent -c "Secrets Sync Service" secrets-sync
```

Or use the automated installer which creates the user automatically:

```bash
make install-systemd
```

### 2. Install the binary

```bash
# Build from source
make build
sudo cp bin/secrets-sync /usr/local/bin/
sudo chmod +x /usr/local/bin/secrets-sync
```

### 3. Create configuration

```bash
# Create config directory
sudo mkdir -p /etc/secrets-sync

# Optional: Set environment variables first for a complete config
export VAULT_ADDR=https://vault.example.com:8200
export VAULT_TOKEN=your-token-here
# Or for AppRole:
# export VAULT_ROLE_ID=your-role-id
# export VAULT_SECRET_ID=your-secret-id

# Generate sample config (uses VAULT_* env vars if set)
secrets-sync init | sudo tee /etc/secrets-sync/config.yaml

# Edit the config as needed
sudo nano /etc/secrets-sync/config.yaml
```

### 4. Create output directories

**Important**: Create directories before starting the service:

```bash
# Create directory for secrets (adjust path to match your config)
sudo mkdir -p /var/secrets

# Set ownership to secrets-sync user
sudo chown secrets-sync:secrets-sync /var/secrets

# Set permissions (750 = owner+group can access)
sudo chmod 750 /var/secrets
```

### 5. Set up environment (optional)

```bash
# Copy environment template
sudo cp examples/systemd/secrets-sync.env.example /etc/default/secrets-sync

# Edit environment variables
sudo nano /etc/default/secrets-sync
```

### 6. Install and start service

```bash
# Copy unit file
sudo cp examples/systemd/secrets-sync.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable service (start on boot)
sudo systemctl enable secrets-sync

# Start service
sudo systemctl start secrets-sync

# Check status
sudo systemctl status secrets-sync
```

## Configuration

### Unit File Customization

Edit `/etc/systemd/system/secrets-sync.service` to customize:

**Network Access**: Uncomment and adjust IP ranges for your Vault server:
```ini
IPAddressAllow=10.0.0.0/8
IPAddressAllow=192.168.1.100
```

**File Paths**: If secrets need to be written outside the configured paths:
```ini
ReadWritePaths=/var/secrets
ReadWritePaths=/app/secrets
```

**Note**: The service uses a static `secrets-sync` user (not DynamicUser) to maintain persistent file ownership across restarts. This allows:
- Consistent file ownership
- Flexible output paths
- Group-based access sharing with other services

### Environment Variables

Set Vault credentials in `/etc/default/secrets-sync`:

**Token Authentication**:
```bash
VAULT_ADDR=https://vault.example.com:8200
VAULT_TOKEN=s.xxxxxxxxxxxxx
```

**AppRole Authentication**:
```bash
VAULT_ADDR=https://vault.example.com:8200
VAULT_ROLE_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
VAULT_SECRET_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## Management

### View logs
```bash
# Follow logs
sudo journalctl -u secrets-sync -f

# View recent logs
sudo journalctl -u secrets-sync -n 100

# View logs since boot
sudo journalctl -u secrets-sync -b
```

### Control service
```bash
# Start
sudo systemctl start secrets-sync

# Stop
sudo systemctl stop secrets-sync

# Restart
sudo systemctl restart secrets-sync

# Reload config (sends SIGHUP signal)
sudo systemctl reload secrets-sync
# Or manually: sudo kill -HUP $(pidof secrets-sync)

# Check status
sudo systemctl status secrets-sync

# Check if enabled
sudo systemctl is-enabled secrets-sync
```

### Health check
```bash
# Check if service is ready
secrets-sync isready

# Check metrics
curl http://localhost:8080/metrics

# Check health
curl http://localhost:8080/health
```

## Security

The unit file includes extensive security hardening:

- **Static User**: Runs as dedicated `secrets-sync` system user (not DynamicUser)
- **Filesystem Protection**: Read-only root, private /tmp
- **Network Restrictions**: Limited address families, IP filtering
- **Capabilities**: No special capabilities
- **System Calls**: Filtered to safe subset
- **Namespaces**: Restricted

**Why Static User Instead of DynamicUser?**

The service uses a static user (`secrets-sync`) rather than `DynamicUser=yes` because:
1. **Persistent UID** - Files maintain correct ownership across reboots/reinstalls
2. **Flexible Paths** - Can write to arbitrary paths (not limited to StateDirectory)
3. **Group Sharing** - Other services can access secrets via group membership
4. **External Access** - Services like nginx/postgres can read the secret files

With `DynamicUser=yes`, the UID changes on each restart, breaking file ownership.

### Adjusting Security

If you encounter permission issues:

1. **Check logs**: `sudo journalctl -u secrets-sync -n 50`
2. **Verify paths**: Ensure config file exists and is readable
3. **Network access**: Add your Vault IP to `IPAddressAllow`
4. **File permissions**: Add paths to `ReadWritePaths`

## Troubleshooting

### Service fails to start

```bash
# Check detailed status
sudo systemctl status secrets-sync -l

# Check logs
sudo journalctl -u secrets-sync -n 50 --no-pager

# Validate config
secrets-sync --config /etc/secrets-sync/config.yaml validate

# Test manually
sudo -u secrets-sync /usr/local/bin/secrets-sync --config /etc/secrets-sync/config.yaml
```

### Permission denied errors

Add required paths to `ReadWritePaths` in the unit file:
```ini
ReadWritePaths=/var/lib/secrets-sync
ReadWritePaths=/path/to/your/secrets
```

### Network connection issues

Check and adjust `IPAddressAllow` in the unit file:
```ini
IPAddressAllow=localhost
IPAddressAllow=10.0.0.0/8
IPAddressAllow=your.vault.server.ip
```

### Config not found

Ensure config file is in the correct location:
```bash
sudo ls -la /etc/secrets-sync/config.yaml
```

## Uninstall

```bash
# Stop and disable service
sudo systemctl stop secrets-sync
sudo systemctl disable secrets-sync

# Remove files
sudo rm /etc/systemd/system/secrets-sync.service
sudo rm /etc/default/secrets-sync
sudo rm -rf /etc/secrets-sync
sudo rm /usr/local/bin/secrets-sync

# Reload systemd
sudo systemctl daemon-reload
```

Or use the automated uninstaller:

```bash
make uninstall-systemd
```

## See Also

- [Main Documentation](../../README.md)
- [Configuration Guide](../../docs/configuration.md)
- [Environment Variables](../../docs/environment-variables.md)
- [Systemd Deployment Guide](../../docs/systemd-deployment.md)
