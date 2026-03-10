# Container Reload for Third-Party Applications

## Problem
Third-party containers (Keycloak, Vault, Nginx, etc.) can't be modified to watch files or expose reload endpoints.

## Solution: Reload Sidecar Pattern

### Architecture
```
┌─────────────────────────────────────────────────┐
│ Pod / Compose Service                           │
│                                                 │
│  ┌──────────────┐  ┌──────────────┐           │
│  │ secrets-sync │  │ reload-proxy │           │
│  │              │  │              │           │
│  │ Syncs secrets│─▶│ Watches files│           │
│  │ from Vault   │  │ Sends signals│           │
│  └──────────────┘  └──────┬───────┘           │
│                            │ SIGHUP             │
│                            ▼                    │
│                    ┌──────────────┐            │
│                    │ Application  │            │
│                    │ (Nginx, etc) │            │
│                    └──────────────┘            │
│                                                 │
│  Shared: /secrets (volume)                     │
│  Shared: PID namespace (for signals)           │
└─────────────────────────────────────────────────┘
```

## Option 1: Reload Sidecar (RECOMMENDED)

**How:** Small sidecar watches secret files and sends reload signal to main container

**Pros:**
- Works with ANY third-party container
- No app modifications needed
- Minimal privileges (shared PID namespace only)
- Reusable across all apps

**Implementation:**

### reload-sidecar Container
```dockerfile
FROM alpine:latest

RUN apk add --no-cache inotify-tools procps

COPY reload-sidecar.sh /reload-sidecar.sh
RUN chmod +x /reload-sidecar.sh

ENTRYPOINT ["/reload-sidecar.sh"]
```

### reload-sidecar.sh
```bash
#!/bin/sh
set -e

# Configuration via environment variables
WATCH_PATH="${WATCH_PATH:-/secrets}"
SIGNAL="${SIGNAL:-HUP}"
PROCESS_NAME="${PROCESS_NAME:-nginx}"
COMMAND="${COMMAND:-}"

echo "Watching $WATCH_PATH for changes..."
echo "Will send SIG$SIGNAL to process: $PROCESS_NAME"

inotifywait -m -e modify,create,delete,move "$WATCH_PATH" |
while read -r directory events filename; do
    echo "Detected change: $directory$filename ($events)"

    if [ -n "$COMMAND" ]; then
        # Execute custom command
        echo "Executing: $COMMAND"
        eval "$COMMAND"
    else
        # Send signal to process
        PID=$(pgrep -f "$PROCESS_NAME" | head -1)
        if [ -n "$PID" ]; then
            echo "Sending SIG$SIGNAL to PID $PID ($PROCESS_NAME)"
            kill -$SIGNAL "$PID"
        else
            echo "Warning: Process $PROCESS_NAME not found"
        fi
    fi
done
```

### docker-compose.yml
```yaml
services:
  nginx:
    image: nginx:alpine
    volumes:
      - secrets:/etc/nginx/conf.d:ro
    # Share PID namespace with reload-sidecar
    pid: "service:reload-sidecar"

  reload-sidecar:
    image: reload-sidecar:latest
    volumes:
      - secrets:/secrets:ro
    environment:
      WATCH_PATH: /secrets
      PROCESS_NAME: nginx
      SIGNAL: HUP
    # This container owns the PID namespace
    pid: host  # Or use a shared namespace

  secrets-sync:
    image: secrets-sync:latest
    volumes:
      - secrets:/secrets
    environment:
      VAULT_ADDR: http://vault:8200
      VAULT_TOKEN: ${VAULT_TOKEN}

volumes:
  secrets:
```

## Option 2: Exec Wrapper (Simpler, Less Secure)

**How:** Wrapper script in main container that watches and reloads

**Pros:**
- No separate sidecar
- No PID namespace sharing

**Cons:**
- Requires custom entrypoint
- Must rebuild container image

### Wrapper Script
```bash
#!/bin/sh
# entrypoint-wrapper.sh

# Start the main application in background
nginx -g 'daemon off;' &
APP_PID=$!

# Watch for secret changes
inotifywait -m -e modify /etc/nginx/conf.d |
while read -r directory events filename; do
    echo "Config changed, reloading nginx..."
    nginx -s reload
done

# Wait for main process
wait $APP_PID
```

### Dockerfile
```dockerfile
FROM nginx:alpine

RUN apk add --no-cache inotify-tools

COPY entrypoint-wrapper.sh /entrypoint-wrapper.sh
RUN chmod +x /entrypoint-wrapper.sh

ENTRYPOINT ["/entrypoint-wrapper.sh"]
```

## Option 3: Docker Compose Healthcheck Restart

**How:** Container checks secret mtime, exits if changed. Compose restarts it.

**Pros:**
- No code changes
- No PID sharing
- Works with unmodified images

**Cons:**
- Full restart (not reload)
- Brief downtime

### docker-compose.yml
```yaml
services:
  nginx:
    image: nginx:alpine
    volumes:
      - secrets:/etc/nginx/conf.d:ro
      - ./check-secrets.sh:/check-secrets.sh:ro
    healthcheck:
      test: ["/check-secrets.sh"]
      interval: 10s
      timeout: 5s
      retries: 1
    restart: unless-stopped

  secrets-sync:
    image: secrets-sync:latest
    volumes:
      - secrets:/secrets

volumes:
  secrets:
```

### check-secrets.sh
```bash
#!/bin/sh
MTIME_FILE="/tmp/secrets-mtime"
WATCH_FILE="/etc/nginx/conf.d/default.conf"

LAST_MTIME=$(cat "$MTIME_FILE" 2>/dev/null || echo 0)
CURR_MTIME=$(stat -c %Y "$WATCH_FILE" 2>/dev/null || echo 0)

if [ "$CURR_MTIME" != "$LAST_MTIME" ] && [ "$LAST_MTIME" != "0" ]; then
    echo "Secrets changed, triggering restart..."
    exit 1  # Fail healthcheck to trigger restart
fi

echo "$CURR_MTIME" > "$MTIME_FILE"
exit 0
```

## Application-Specific Examples

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

### Keycloak
```yaml
reload-sidecar:
  environment:
    PROCESS_NAME: java
    SIGNAL: HUP
    WATCH_PATH: /opt/keycloak/conf
```

### PostgreSQL
```yaml
reload-sidecar:
  environment:
    PROCESS_NAME: postgres
    SIGNAL: HUP
    WATCH_PATH: /var/lib/postgresql/data
```

### Vault (reload TLS certs)
```yaml
reload-sidecar:
  environment:
    PROCESS_NAME: vault
    SIGNAL: HUP
    WATCH_PATH: /vault/certs
```

### Custom Command (for apps without signal support)
```yaml
reload-sidecar:
  environment:
    WATCH_PATH: /secrets
    COMMAND: "curl -X POST http://app:8080/reload"
```

## Kubernetes Example

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-with-secrets
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
    image: reload-sidecar:latest
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
    env:
    - name: VAULT_ADDR
      value: http://vault:8200
    volumeMounts:
    - name: secrets
      mountPath: /secrets

  volumes:
  - name: secrets
    emptyDir: {}
```

## Security Considerations

### Shared PID Namespace
- **Risk:** Containers can see each other's processes
- **Mitigation:** Only share between trusted sidecars
- **Alternative:** Use healthcheck restart (no PID sharing)

### Signal Permissions
- **Risk:** Sending signals requires same user or capabilities
- **Mitigation:** Run sidecars as same UID as main container
- **Alternative:** Use exec wrapper or healthcheck restart

## Recommendation

**For Production: Use Reload Sidecar (#1)**
- Works with any third-party container
- Minimal privileges (PID namespace only)
- Reusable across all applications
- Clean separation of concerns

**For Development: Use Healthcheck Restart (#3)**
- Simplest to set up
- No PID sharing needed
- Brief downtime acceptable in dev

## Implementation in secrets-sync

Add optional sidecar image to project:

```
docker-secrets/
├── Dockerfile              # Main secrets-sync
├── Dockerfile.reload       # Reload sidecar
└── scripts/
    └── reload-sidecar.sh
```

Users can deploy both:
```yaml
services:
  app:
    image: nginx:alpine
    pid: "service:reload-sidecar"

  reload-sidecar:
    image: secrets-sync:reload
    environment:
      PROCESS_NAME: nginx
      SIGNAL: HUP

  secrets-sync:
    image: secrets-sync:latest
```

This provides a complete solution for third-party containers without modification!
