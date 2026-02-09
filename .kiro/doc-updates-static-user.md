# Documentation Updates - Static User for Systemd

## Summary

Updated documentation to reflect that the systemd service uses a static `secrets-sync` user instead of `DynamicUser=yes`.

## Changes Made

### examples/systemd/README.md

1. **Added Step 1**: Create system user
   - Explains the need for a static user
   - Shows manual user creation command
   - References automated installer

2. **Updated Step 3**: Create configuration
   - Added suggestion to export VAULT_* environment variables before running `init`
   - This generates a complete config.yaml immediately

3. **Added Step 4**: Create output directories
   - Emphasizes the importance of pre-creating directories
   - Shows ownership and permission setup

4. **Removed DynamicUser references**:
   - Removed "User/Group: To run as specific user instead of DynamicUser" section
   - Replaced with explanation of why static user is used

5. **Updated Security section**:
   - Changed "DynamicUser: Runs as ephemeral user" to "Static User: Runs as dedicated secrets-sync system user"
   - Added detailed explanation of why static user is preferred over DynamicUser

## Rationale for Static User

The service uses a static user (`secrets-sync`) rather than `DynamicUser=yes` because:

1. **Persistent UID** - Files maintain correct ownership across reboots/reinstalls
2. **Flexible Paths** - Can write to arbitrary paths (not limited to StateDirectory)
3. **Group Sharing** - Other services can access secrets via group membership
4. **External Access** - Services like nginx/postgres can read the secret files

With `DynamicUser=yes`, the UID changes on each restart, breaking file ownership.

## Files Already Correct

- `docs/systemd-deployment.md` - Already has the static user explanation
- `examples/systemd/secrets-sync.service` - Already uses `User=secrets-sync`
- `scripts/install-systemd.sh` - Already creates the static user
- `Makefile` - Already calls the install script

## User Experience Improvement

The updated documentation now:
1. Clearly states the user must be created first
2. Suggests setting VAULT_* env vars before `init` for a complete config
3. Emphasizes the need to create output directories before starting
4. Explains the technical rationale for using a static user
