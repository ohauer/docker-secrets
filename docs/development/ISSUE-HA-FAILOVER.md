---
title: "HA Vault support: multiple addresses with automatic failover"
labels: enhancement, priority-high
---

## Summary

secrets-sync currently connects to a single Vault address. In HA deployments (Raft cluster with 3+ nodes), if that node goes down, syncing fails until the original node recovers — even though other cluster members are healthy and serving the same data.

## Problem

Tested with a 3-node Vault HA cluster (Raft storage, v1.21.4):

1. secrets-sync configured with `address: https://vault-pu-01:8200`
2. vault-pu-01 stopped → vault-pu-03 elected new leader
3. secrets-sync circuit breaker opens, syncs fail for ~90s
4. vault-pu-01 restarted → syncing resumes

Files on disk are preserved during outage (good), but there is no reason for the outage — two healthy nodes were available the entire time.

Full test report: `HA-FAILOVER-TEST.md`

## Proposed Solution

### 1. Multiple addresses in config

```yaml
secretStore:
  # Single address (backward compatible)
  address: "https://vault-01:8200"

  # OR: multiple addresses for HA
  address:
    - "https://vault-01:8200"
    - "https://vault-02:8200"
    - "https://vault-03:8200"
```

Environment variable: `VAULT_ADDR=https://vault-01:8200,https://vault-02:8200`

### 2. Sticky-with-failover strategy

- Stay connected to the current working node
- On failure (connection refused, timeout, circuit breaker open): try next address
- On success: stick to the new node
- Log address switches

### 3. Cluster identity logging

Query `/v1/sys/health` on connect to get `cluster_name`. Log it:

```json
{"msg":"connected to vault","address":"vault-01:8200","cluster_name":"vault-pu"}
{"msg":"vault failover","from":"vault-01:8200","to":"vault-03:8200","cluster_name":"vault-pu"}
```

### 4. Circuit breaker integration

When the circuit breaker transitions to half-open, the probe should try all configured addresses (not just the failed one).

## Files to change

| File | Change |
|------|--------|
| `internal/config/types.go` | `Address` field: support string or []string |
| `internal/config/validator.go` | Validate address list |
| `internal/config/env.go` | `VAULT_ADDR` comma-separated support |
| `internal/vault/client.go` | Address rotation, `ClusterName()` method |
| `cmd/secrets-sync/main.go` | `clientFactory` with multi-address, log cluster_name |

## Design notes

- Vault standbys forward reads to the active node by default, so connecting to any healthy node works
- After address switch, existing token should still be valid (Raft replicates token store) — only re-auth if token is rejected
- `/v1/sys/health` and `/v1/sys/leader` are unauthenticated, safe to call on any node

## Acceptance criteria

- [ ] Single address config still works (backward compatible)
- [ ] Multiple addresses: failover to next on connection failure
- [ ] Log cluster_name on connect
- [ ] Log address switch on failover
- [ ] Circuit breaker half-open probe tries all addresses
- [ ] `VAULT_ADDR` env var supports comma-separated list
- [ ] Unit tests for address rotation logic
- [ ] Integration test with multi-node cluster

## Estimated effort

6-8 hours
