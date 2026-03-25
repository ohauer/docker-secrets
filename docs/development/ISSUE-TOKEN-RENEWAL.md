---
title: "Token renewal and re-authentication for long-running instances"
labels: enhancement, priority-high
status: implemented
---

## Summary

secrets-sync now automatically renews Vault tokens before they expire
and re-authenticates via AppRole when renewal is no longer possible.

## Implementation

### Token Renewal (AppRole)
- Uses Vault SDK `LifetimeWatcher` to renew tokens at ~50% TTL
- Runs as background goroutine
- Stops cleanly on shutdown

### Re-authentication (AppRole)
- When token cannot be renewed (max TTL reached), re-authenticates
- Gets fresh token via AppRole login
- Restarts renewal with new token

### Static Tokens
- No renewal possible (no auth secret)
- Logs warning when token expires

### Log Output
```json
{"msg":"vault token renewed","ttl_seconds":384}
{"msg":"vault token expired, re-authenticating"}
{"msg":"vault re-authenticated successfully","ttl_seconds":768}
```

## Files Changed

| File | Change |
|------|--------|
| `internal/vault/auth.go` | Added `StartRenewal()`, `RenewalCallback`, re-auth logic |
| `internal/vault/client.go` | Added `authSecret`, `authConfig` fields to Client |
| `cmd/secrets-sync/main.go` | Start renewal after auth, register shutdown |
