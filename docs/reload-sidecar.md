# Reload Sidecar

Automatically reload third-party containers when secrets change.

## Overview

The reload-sidecar watches secret files and triggers application reloads without requiring code changes to the application. Perfect for third-party containers like Nginx, Apache, Keycloak, PostgreSQL, etc.

## How It Works

```
secrets-sync → updates files → reload-sidecar → sends signal/webhook → app reloads
```

1. `secrets-sync` updates secret files
2. `reload-sidecar` detects file changes (inotify)
3. Triggers reload via signal, command, or webhook
4. Application picks up new secrets without restart

## Quick Start

### Example: Nginx

```yaml
services:
  nginx:
    image: nginx:alpine
    volumes:
      - secrets:/etc/nginx/conf.d:ro
    pid: "service:reload-sidecar"

  reload-sidecar:
    image: secrets-sync:reload
    volumes:
      - secrets:/secrets:ro
    environment:
      WATCH_PATH: /secrets
      PROCESS_NAME: nginx
      SIGNAL: HUP

  secrets-sync:
    image: secrets-sync:latest
    volumes:
      - secrets:/secrets
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WATCH_PATH` | `/secrets` | Directory to watch for changes |
| `PROCESS_NAME` | - | Exact process name to signal |
| `PROCESS_PATTERN` | - | Process pattern to match (regex) |
| `SIGNAL` | `HUP` | Signal to send (HUP, USR1, USR2, etc.) |
| `COMMAND` | - | Command to execute on change |
| `WEBHOOK_URL` | - | HTTP webhook to call |
| `WEBHOOK_METHOD` | `POST` | HTTP method for webhook |
| `DEBOUNCE_SECONDS` | `2` | Minimum seconds between reloads |
| `LOG_TIMESTAMP` | `true` | Include timestamps in logs |

### Reload Methods

#### 1. Signal to Process (Default)

Send Unix signal to process:

```yaml
environment:
  PROCESS_NAME: nginx
  SIGNAL: HUP
```

**Common signals:**
- `HUP` - Reload configuration (Nginx, Apache, PostgreSQL)
- `USR1` - Reopen log files (Nginx)
- `USR2` - Upgrade binary (Nginx)

#### 2. Execute Command

Run custom command:

```yaml
environment:
  COMMAND: "nginx -s reload"
```

#### 3. HTTP Webhook

Call HTTP endpoint:

```yaml
environment:
  WEBHOOK_URL: "http://app:8080/reload"
  WEBHOOK_METHOD: "POST"
```

## Application Examples

### Nginx

```yaml
reload-sidecar:
  environment:
    PROCESS_NAME: nginx
    SIGNAL: HUP
    WATCH_PATH: /etc/nginx/conf.d
```

### Apache

```yaml
reload-sidecar:
  environment:
    PROCESS_NAME: httpd
    SIGNAL: USR1
    WATCH_PATH: /usr/local/apache2/conf
```

### PostgreSQL

```yaml
reload-sidecar:
  environment:
    PROCESS_NAME: postgres
    SIGNAL: HUP
    WATCH_PATH: /var/lib/postgresql/data
```

### Keycloak

```yaml
reload-sidecar:
  environment:
    PROCESS_PATTERN: "java.*keycloak"
    SIGNAL: HUP
    WATCH_PATH: /opt/keycloak/conf
```

### Vault (TLS cert reload)

```yaml
reload-sidecar:
  environment:
    PROCESS_NAME: vault
    SIGNAL: HUP
    WATCH_PATH: /vault/certs
```

### Custom Application (webhook)

```yaml
reload-sidecar:
  environment:
    WEBHOOK_URL: "http://myapp:8080/api/reload"
    WATCH_PATH: /app/config
```

## Docker Compose Setup

### Shared PID Namespace

Required for sending signals between containers:

```yaml
services:
  app:
    image: nginx:alpine
    pid: "service:reload-sidecar"  # Share PID namespace

  reload-sidecar:
    image: secrets-sync:reload
    # This container owns the PID namespace
```

### Complete Example

See `examples/docker-compose.reload-example.yml` for a full working example with Nginx.

## Kubernetes Setup

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-reload
spec:
  shareProcessNamespace: true  # Enable PID sharing

  containers:
  - name: nginx
    image: nginx:alpine
    volumeMounts:
    - name: secrets
      mountPath: /etc/nginx/conf.d
      readOnly: true

  - name: reload-sidecar
    image: secrets-sync:reload
    env:
    - name: WATCH_PATH
      value: /secrets
    - name: PROCESS_NAME
      value: nginx
    - name: SIGNAL
      value: HUP
    volumeMounts:
    - name: secrets
      mountPath: /secrets
      readOnly: true

  - name: secrets-sync
    image: secrets-sync:latest
    volumeMounts:
    - name: secrets
      mountPath: /secrets

  volumes:
  - name: secrets
    emptyDir: {}
```

## Building

```bash
# Build reload-sidecar image
make docker-build-reload

# Or manually
docker build -t secrets-sync:reload -f reload-sidecar/Dockerfile reload-sidecar/
```

## Troubleshooting

### Process not found

**Problem:** `Process not found: nginx`

**Solutions:**
1. Check process name: `docker exec <container> ps aux`
2. Use `PROCESS_PATTERN` instead: `PROCESS_PATTERN="nginx: master"`
3. Ensure PID namespace is shared

### Permission denied sending signal

**Problem:** `Failed to send signal to PID 123 (permission denied?)`

**Solutions:**
1. Ensure PID namespace is shared: `pid: "service:reload-sidecar"`
2. Run containers as same UID
3. Use `COMMAND` or `WEBHOOK_URL` instead

### Too many reloads

**Problem:** Application reloading constantly

**Solutions:**
1. Increase debounce: `DEBOUNCE_SECONDS: 5`
2. Watch specific files, not entire directory
3. Check for file permission changes triggering events

### Reload not working

**Problem:** Files change but app doesn't reload

**Solutions:**
1. Check logs: `docker logs reload-sidecar`
2. Verify watch path: `WATCH_PATH=/correct/path`
3. Test signal manually: `docker exec <container> kill -HUP <pid>`
4. Ensure app supports the signal

## Security Considerations

### Shared PID Namespace

**Risk:** Containers can see each other's processes

**Mitigation:**
- Only share between trusted sidecars
- Use network policies to isolate
- Consider webhook method instead (no PID sharing)

### Signal Permissions

**Risk:** Sending signals requires privileges

**Mitigation:**
- Run as same UID as target container
- Use least-privilege signals (HUP, USR1)
- Avoid SIGKILL, SIGTERM

### File Watching

**Risk:** Watching sensitive directories

**Mitigation:**
- Mount secrets read-only
- Watch specific paths only
- Use debouncing to prevent DoS

## Limitations

### Java Applications

Many Java applications don't support signal-based reload. Options:

1. **Use webhook method:**
   ```yaml
   WEBHOOK_URL: "http://app:8080/actuator/refresh"
   ```

2. **Use healthcheck restart:**
   - Container exits when secrets change
   - Orchestrator restarts it
   - See `docs/development/THIRD_PARTY_RELOAD.md`

### Windows Containers

Signals don't work on Windows. Use:
- Webhook method
- Command execution
- Healthcheck restart

## Performance

- **CPU:** Minimal (inotify is event-driven)
- **Memory:** ~5MB
- **Latency:** <100ms from file change to reload
- **Debouncing:** Prevents reload storms

## Alternatives

If reload-sidecar doesn't fit your needs:

1. **Application-level reload** - App watches files itself
2. **Healthcheck restart** - Container restarts on change
3. **Webhook from secrets-sync** - Direct webhook support (future feature)

See `docs/development/CONTAINER_RELOAD_OPTIONS.md` for detailed comparison.

## Contributing

Improvements welcome! Common additions:
- Support for more applications
- Additional reload methods
- Better error handling
- Metrics/observability

## License

Same as secrets-sync (MIT)
