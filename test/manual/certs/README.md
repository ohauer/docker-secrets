# Test Certificates

This directory contains test certificates for TLS testing.

## Generate Certificates

Run the script to create test certificates:

```bash
./generate-certs.sh
```

This creates:
- `ca-cert.pem` / `ca-key.pem` - Certificate Authority
- `ca-bundle.pem` - CA bundle for trust
- `vault-cert.pem` / `vault-key.pem` - Vault server certificate
- `openbao-cert.pem` / `openbao-key.pem` - OpenBao server certificate
- `client-cert.pem` / `client-key.pem` - Client certificate for mTLS

## Important

⚠️ **These are test certificates only!**
- Generated with weak parameters
- Self-signed
- Never use in production
- Automatically ignored by git

The certificates are regenerated each time you run the script.
