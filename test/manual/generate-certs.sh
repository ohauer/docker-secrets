#!/bin/sh
# Generate test CA and certificates for Vault/OpenBao TLS testing

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_DIR="${SCRIPT_DIR}/certs"
DAYS=3650

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') generate-certs - $1"
}

mkdir -p "${CERT_DIR}"
cd "${CERT_DIR}"

log_message "Generating CA certificate..."
openssl genrsa -out ca-key.pem 4096
openssl req -new -x509 -days ${DAYS} -key ca-key.pem -out ca-cert.pem \
    -subj "/C=US/ST=Test/L=Test/O=Test CA/CN=Test CA"

log_message "Generating Vault server certificate..."
openssl genrsa -out vault-key.pem 4096
openssl req -new -key vault-key.pem -out vault.csr \
    -subj "/C=US/ST=Test/L=Test/O=Vault/CN=vault-dev"

cat > vault-ext.cnf << EOF
subjectAltName = DNS:vault-dev,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

openssl x509 -req -days ${DAYS} -in vault.csr -CA ca-cert.pem -CAkey ca-key.pem \
    -CAcreateserial -out vault-cert.pem -extfile vault-ext.cnf

log_message "Generating OpenBao server certificate..."
openssl genrsa -out openbao-key.pem 4096
openssl req -new -key openbao-key.pem -out openbao.csr \
    -subj "/C=US/ST=Test/L=Test/O=OpenBao/CN=openbao-dev"

cat > openbao-ext.cnf << EOF
subjectAltName = DNS:openbao-dev,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

openssl x509 -req -days ${DAYS} -in openbao.csr -CA ca-cert.pem -CAkey ca-key.pem \
    -CAcreateserial -out openbao-cert.pem -extfile openbao-ext.cnf

log_message "Generating client certificate..."
openssl genrsa -out client-key.pem 4096
openssl req -new -key client-key.pem -out client.csr \
    -subj "/C=US/ST=Test/L=Test/O=Client/CN=secrets-sync"

cat > client-ext.cnf << EOF
extendedKeyUsage = clientAuth
EOF

openssl x509 -req -days ${DAYS} -in client.csr -CA ca-cert.pem -CAkey ca-key.pem \
    -CAcreateserial -out client-cert.pem -extfile client-ext.cnf

log_message "Creating CA bundle..."
cp ca-cert.pem ca-bundle.pem

log_message "Cleaning up temporary files..."
rm -f *.csr *.cnf *.srl

log_message "Setting permissions..."
chmod 644 *.pem
chmod 600 *-key.pem

log_message "Certificate generation complete!"
log_message ""
log_message "Generated files:"
log_message "  CA:      ca-cert.pem, ca-key.pem, ca-bundle.pem"
log_message "  Vault:   vault-cert.pem, vault-key.pem"
log_message "  OpenBao: openbao-cert.pem, openbao-key.pem"
log_message "  Client:  client-cert.pem, client-key.pem"
