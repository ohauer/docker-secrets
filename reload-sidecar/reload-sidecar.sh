#!/bin/bash
# reload-sidecar.sh - Watch files and reload applications
set -euo pipefail

# Configuration via environment variables
WATCH_PATH="${WATCH_PATH:-/secrets}"
SIGNAL="${SIGNAL:-HUP}"
PROCESS_NAME="${PROCESS_NAME:-}"
PROCESS_PATTERN="${PROCESS_PATTERN:-}"
COMMAND="${COMMAND:-}"
WEBHOOK_URL="${WEBHOOK_URL:-}"
WEBHOOK_METHOD="${WEBHOOK_METHOD:-POST}"
DEBOUNCE_SECONDS="${DEBOUNCE_SECONDS:-2}"
LOG_TIMESTAMP="${LOG_TIMESTAMP:-true}"

# Logging function
log() {
    if [ "$LOG_TIMESTAMP" = "true" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
    else
        echo "$*"
    fi
}

# Validate configuration
if [ -z "$PROCESS_NAME" ] && [ -z "$PROCESS_PATTERN" ] && [ -z "$COMMAND" ] && [ -z "$WEBHOOK_URL" ]; then
    log "ERROR: Must set one of: PROCESS_NAME, PROCESS_PATTERN, COMMAND, or WEBHOOK_URL"
    exit 1
fi

if [ ! -d "$WATCH_PATH" ]; then
    log "ERROR: Watch path does not exist: $WATCH_PATH"
    exit 1
fi

log "Reload Sidecar starting..."
log "Watch path: $WATCH_PATH"
log "Debounce: ${DEBOUNCE_SECONDS}s"

if [ -n "$PROCESS_NAME" ]; then
    log "Mode: Signal to process name '$PROCESS_NAME' (SIG$SIGNAL)"
elif [ -n "$PROCESS_PATTERN" ]; then
    log "Mode: Signal to process pattern '$PROCESS_PATTERN' (SIG$SIGNAL)"
elif [ -n "$COMMAND" ]; then
    log "Mode: Execute command '$COMMAND'"
elif [ -n "$WEBHOOK_URL" ]; then
    log "Mode: HTTP webhook $WEBHOOK_METHOD $WEBHOOK_URL"
fi

# Debounce mechanism
last_reload=0

should_reload() {
    current_time=$(date +%s)
    time_diff=$((current_time - last_reload))

    if [ $time_diff -ge $DEBOUNCE_SECONDS ]; then
        last_reload=$current_time
        return 0
    else
        log "Debouncing... (${time_diff}s since last reload)"
        return 1
    fi
}

# Reload function
reload_application() {
    if ! should_reload; then
        return
    fi

    log "Triggering reload..."

    if [ -n "$WEBHOOK_URL" ]; then
        # HTTP webhook
        if curl -f -X "$WEBHOOK_METHOD" "$WEBHOOK_URL" -m 5 >/dev/null 2>&1; then
            log "✓ Webhook successful"
        else
            log "✗ Webhook failed"
        fi

    elif [ -n "$COMMAND" ]; then
        # Execute command
        if eval "$COMMAND"; then
            log "✓ Command executed successfully"
        else
            log "✗ Command failed"
        fi

    else
        # Send signal to process
        local pattern="${PROCESS_PATTERN:-$PROCESS_NAME}"
        local pids

        if [ -n "$PROCESS_NAME" ]; then
            pids=$(pgrep -x "$PROCESS_NAME" 2>/dev/null || true)
        else
            pids=$(pgrep -f "$PROCESS_PATTERN" 2>/dev/null || true)
        fi

        if [ -z "$pids" ]; then
            log "✗ Process not found: $pattern"
            return
        fi

        for pid in $pids; do
            if kill -$SIGNAL "$pid" 2>/dev/null; then
                log "✓ Sent SIG$SIGNAL to PID $pid"
            else
                log "✗ Failed to send signal to PID $pid (permission denied?)"
            fi
        done
    fi
}

# Initial check - wait for process to start
if [ -n "$PROCESS_NAME" ] || [ -n "$PROCESS_PATTERN" ]; then
    log "Waiting for process to start..."
    for i in {1..30}; do
        pattern="${PROCESS_PATTERN:-$PROCESS_NAME}"
        if [ -n "$PROCESS_NAME" ]; then
            pids=$(pgrep -x "$PROCESS_NAME" 2>/dev/null || true)
        else
            pids=$(pgrep -f "$PROCESS_PATTERN" 2>/dev/null || true)
        fi

        if [ -n "$pids" ]; then
            log "✓ Process found (PID: $pids)"
            break
        fi

        if [ $i -eq 30 ]; then
            log "WARNING: Process not found after 30s, continuing anyway..."
        fi
        sleep 1
    done
fi

log "Watching for changes..."

# Watch for file changes
inotifywait -m -e modify,create,delete,move,attrib "$WATCH_PATH" 2>/dev/null |
while read -r directory events filename; do
    log "Change detected: $filename ($events)"
    reload_application
done
