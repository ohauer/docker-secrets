# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of Docker Secrets Sidecar seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Please Do Not

- **Do not** open a public GitHub issue for security vulnerabilities
- **Do not** disclose the vulnerability publicly until it has been addressed
- **Do not** exploit the vulnerability beyond what is necessary to demonstrate it

### How to Report

**Preferred Method: GitHub Security Advisories**

1. Go to the [Security tab](https://github.com/ohauer/docker-secrets/security/advisories)
2. Click "Report a vulnerability"
3. Fill out the form with details about the vulnerability

**Alternative Method: Email**

If you prefer not to use GitHub Security Advisories, you can email security reports to:
- **Email**: [Create a private security advisory on GitHub instead]

### What to Include

Please include the following information in your report:

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** of the vulnerability
- **Affected versions** (if known)
- **Suggested fix** (if you have one)
- **Your contact information** for follow-up questions

### Example Report

```
Subject: [SECURITY] Potential secret exposure in log output

Description:
Under certain error conditions, secret values may be logged to stdout,
potentially exposing sensitive information.

Steps to Reproduce:
1. Configure a secret with invalid template syntax
2. Start the application with LOG_LEVEL=debug
3. Observe error logs containing secret values

Impact:
- Secrets may be exposed in log aggregation systems
- Affects all versions prior to 0.1.0
- Severity: High

Affected Versions:
- All versions < 0.1.0

Suggested Fix:
Sanitize error messages to remove secret values before logging.

Contact:
- GitHub: @username
- Email: security@example.com
```

## Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity
  - Critical: 7-14 days
  - High: 14-30 days
  - Medium: 30-60 days
  - Low: 60-90 days

## Security Update Process

1. **Acknowledgment**: We will acknowledge receipt of your report
2. **Investigation**: We will investigate and validate the vulnerability
3. **Fix Development**: We will develop and test a fix
4. **Disclosure**: We will coordinate disclosure with you
5. **Release**: We will release a security update
6. **Announcement**: We will publish a security advisory

## Security Best Practices

When using Docker Secrets Sidecar, follow these security best practices:

### Configuration

- **Never commit secrets** to version control
- **Use environment variables** for sensitive configuration
- **Restrict file permissions** (use mode 0600 for secret files)
- **Run as non-root** user in containers
- **Use TLS** for Vault connections in production

### Deployment

- **Enable TLS verification** (never use `tlsSkipVerify: true` in production)
- **Use custom CA certificates** for self-signed certificates
- **Implement network policies** to restrict access
- **Monitor logs** for suspicious activity
- **Rotate secrets regularly**

### Container Security

- **Use specific image tags** (not `latest`)
- **Scan images** for vulnerabilities
- **Run with read-only filesystem** where possible
- **Limit container capabilities**
- **Use security contexts** in Kubernetes

### Vault/OpenBao Security

- **Use AppRole** instead of tokens when possible
- **Implement least privilege** access policies
- **Enable audit logging** in Vault
- **Rotate AppRole credentials** regularly
- **Use namespaces** for multi-tenancy

## Known Security Considerations

### Secrets in Memory

Secrets are held in memory during processing. While we take precautions:
- Secrets are not logged
- Memory is not swapped to disk (when using `mlock`)
- Secrets are cleared after use where possible

### File System Security

Secret files are written to the filesystem:
- Use restrictive permissions (0600 by default)
- Mount secret volumes with `noexec,nosuid` options
- Use encrypted filesystems for additional protection
- Clean up secrets on container termination

### Network Security

Communication with Vault/OpenBao:
- Always use TLS in production
- Verify server certificates
- Use mTLS for additional security
- Implement network segmentation

## Security Audit

This project has undergone a security audit. See [SECURITY_AUDIT.md](SECURITY_AUDIT.md) for details.

Key findings:
- No critical vulnerabilities identified
- All recommendations addressed
- OWASP Top 10 compliance verified

## Vulnerability Disclosure Policy

We follow responsible disclosure practices:

1. **Private Disclosure**: Report vulnerabilities privately first
2. **Coordinated Disclosure**: We will work with you on disclosure timing
3. **Public Disclosure**: After a fix is released, we will:
   - Publish a security advisory
   - Credit the reporter (if desired)
   - Update the CHANGELOG
   - Notify users via GitHub releases

## Security Hall of Fame

We recognize security researchers who help improve our security:

<!-- Security researchers will be listed here -->

*No vulnerabilities reported yet.*

## Contact

For security-related questions that are not vulnerabilities:
- Open a [Discussion](https://github.com/ohauer/docker-secrets/discussions)
- Check our [documentation](docs/)

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Vault Security Model](https://www.vaultproject.io/docs/internals/security)
- [Container Security Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)

## Updates to This Policy

This security policy may be updated from time to time. Please check back regularly for updates.

**Last Updated**: 2026-02-02
