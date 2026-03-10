# Container Restart/Reload Options for Secrets-Sync

## Problem
When secrets-sync updates secrets, application containers need to reload/restart to pick up changes.

## Constraint
**Must work as unprivileged user** (no root, no docker socket access)

## Options (Ranked by Security & Simplicity)

### 1. ✅ Application-Level Reload (RECOMMENDED)

**How:** Application watches secret files and reloads on change

**Pros:**
- No privileges needed
- No external dependencies
- Fast, efficient
- Works everywhere (Docker, Podman, Kubernetes)

**Cons:**
- Requires app support (inotify/fsnotify)

**Example (Go app):**
```go
watcher, _ := fsnotify.NewWatcher()
watcher.Add("/secrets")
for event := range watcher.Events {
    if event.Op&fsnotify.Write == fsnotify.Write {
        reloadConfig()
    }
}
```

**Example (Nginx):**
```bash
# In container entrypoint
inotifywait -m /secrets -e modify |
while read path action file; do
    nginx -s reload
done
```

---

### 2. ✅ HTTP Webhook (RECOMMENDED)

**How:** secrets-sync calls webhook after updating secrets

**Pros:**
- No privileges needed
- Language-agnostic
- Clean separation of concerns
- Works with any orchestrator

**Cons:**
- Requires app to expose endpoint
- Network call overhead

**Implementation in secrets-sync:**
```yaml
secrets:
  - name: "app-config"
    key: "app/config"
    files:
      - path: "/secrets/config.json"
    onUpdate:
      webhook:
        url: "http://app:8080/reload"
        method: "POST"
        timeout: "5s"
```

**Example (App side):**
```go
http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
    reloadConfig()
    w.WriteHeader(200)
})
```

---

### 3. ✅ Container Healthcheck + Exit

**How:** Container checks secret file mtime, exits if changed. Orchestrator restarts it.

**Pros:**
- No privileges needed
- Works with any orchestrator
- Simple to implement

**Cons:**
- Full restart (not reload)
- Brief downtime
- Restart loop if secrets keep changing

**Example (docker-compose):**
```yaml
services:
  app:
    image: myapp
    volumes:
      - secrets:/secrets:ro
    healthcheck:
      test: ["CMD", "/check-secrets.sh"]
      interval: 10s
    restart: unless-stopped

  secrets-sidecar:
    image: secrets-sync
    volumes:
      - secrets:/secrets
```

**check-secrets.sh:**
```bash
#!/bin/sh
LAST_MTIME=$(cat /tmp/secrets-mtime 2>/dev/null || echo 0)
CURR_MTIME=$(stat -c %Y /secrets/config.json)
if [ "$CURR_MTIME" != "$LAST_MTIME" ]; then
    echo $CURR_MTIME > /tmp/secrets-mtime
    exit 1  # Trigger restart
fi
```

---

### 4. ⚠️ Docker Exec (Requires Privileges)

**How:** secrets-sync runs command in target container via docker exec

**Pros:**
- Flexible
- Works with any app

**Cons:**
- **Requires docker socket access** (security risk)
- Needs container name/ID
- Platform-specific (Docker/Podman)

**NOT RECOMMENDED** for unprivileged deployment

---

### 5. ⚠️ Signal via Shared PID Namespace (Requires Privileges)

**How:** Send SIGHUP to process in another container

**Pros:**
- Standard Unix signal
- Fast

**Cons:**
- **Requires shared PID namespace** (security risk)
- Needs to know target PID
- App must handle signal

**NOT RECOMMENDED** for unprivileged deployment

---

### 6. ⚠️ Docker Socket Access (Requires Privileges)

**How:** Mount docker.sock and restart container via API

**Pros:**
- Full control
- Works with any app

**Cons:**
- **MAJOR SECURITY RISK** (root-equivalent access)
- Violates unprivileged constraint
- Platform-specific

**NEVER RECOMMENDED** for production

---

## Recommended Approach

### For New Applications
**Use Application-Level Reload (#1)**
- Most secure
- Most efficient
- Best user experience

### For Existing Applications
**Use HTTP Webhook (#2)**
- No app code changes (just add endpoint)
- Clean, secure
- Works everywhere

### For Legacy Applications
**Use Healthcheck + Exit (#3)**
- No app changes needed
- Brief downtime acceptable
- Simple to implement

---

## Implementation Plan for secrets-sync

### Phase 1: Webhook Support
Add webhook configuration to secret definition:

```yaml
secrets:
  - name: "app-config"
    key: "app/config"
    files:
      - path: "/secrets/config.json"
    onUpdate:
      webhook:
        url: "http://app:8080/reload"
        method: "POST"
        headers:
          X-Secret-Name: "app-config"
        timeout: "5s"
        retries: 3
```

### Phase 2: Command Execution (Optional)
For containers that share filesystem:

```yaml
secrets:
  - name: "nginx-config"
    key: "nginx/config"
    files:
      - path: "/etc/nginx/conf.d/default.conf"
    onUpdate:
      exec:
        command: ["nginx", "-s", "reload"]
        timeout: "10s"
```

This runs the command in the **secrets-sync container**, not the target container (no privileges needed).

### Phase 3: File Marker (Optional)
Write a marker file that apps can watch:

```yaml
secrets:
  - name: "app-config"
    key: "app/config"
    files:
      - path: "/secrets/config.json"
    onUpdate:
      marker:
        path: "/secrets/.updated"
        content: "{{ .timestamp }}"
```

App watches `/secrets/.updated` for changes.

---

## Examples

### Example 1: Nginx with Webhook
```yaml
# docker-compose.yml
services:
  nginx:
    image: nginx:alpine
    volumes:
      - secrets:/etc/nginx/conf.d:ro
      - ./reload.sh:/reload.sh
    command: sh -c "nginx -g 'daemon off;' & /reload.sh"

  secrets-sidecar:
    image: secrets-sync
    volumes:
      - secrets:/secrets
    environment:
      CONFIG_FILE: /config.yaml
```

```bash
# reload.sh
#!/bin/sh
while true; do
    nc -l -p 8080 -e sh -c 'nginx -s reload; echo "HTTP/1.1 200 OK\n\nReloaded"'
done
```

### Example 2: Go App with fsnotify
```go
// main.go
func watchSecrets() {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add("/secrets")

    for {
        select {
        case event := <-watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                log.Println("Secrets updated, reloading...")
                reloadConfig()
            }
        }
    }
}
```

### Example 3: Python App with watchdog
```python
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

class SecretHandler(FileSystemEventHandler):
    def on_modified(self, event):
        if event.src_path.endswith('.json'):
            reload_config()

observer = Observer()
observer.schedule(SecretHandler(), '/secrets', recursive=False)
observer.start()
```

---

## Security Considerations

### ✅ Safe (Unprivileged)
- Application-level reload
- HTTP webhooks
- Healthcheck + exit
- File watching (inotify)

### ⚠️ Requires Privileges
- Docker socket access
- Shared PID namespace
- Container exec

### 🔒 Best Practices
1. Never mount docker.sock in production
2. Use read-only volume mounts for secrets
3. Run containers as non-root user
4. Use network policies to restrict webhook access
5. Validate webhook responses
6. Set reasonable timeouts

---

## Recommendation

**Implement Webhook Support First**
- Most flexible
- Works with any language
- No privileges needed
- Clean architecture

**Document Application-Level Reload**
- Best practice for new apps
- Include examples for common frameworks
- Provide helper libraries if needed

**Avoid Docker Socket Access**
- Security risk
- Violates unprivileged constraint
- Platform-specific
