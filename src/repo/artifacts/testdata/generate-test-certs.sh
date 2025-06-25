#!/bin/bash
# Generate test certificates for TLS registry testing
# This script should be run manually to generate certificates for TLS tests

set -e

CERTS_DIR="$(dirname "$0")/certs"
mkdir -p "$CERTS_DIR"

# Generate CA key and certificate
openssl genrsa -out "$CERTS_DIR/ca.key" 4096
openssl req -new -x509 -key "$CERTS_DIR/ca.key" -sha256 -subj "/CN=Test CA" -days 365 -out "$CERTS_DIR/ca.crt"

# Generate registry key and certificate
openssl genrsa -out "$CERTS_DIR/registry.key" 4096
openssl req -new -key "$CERTS_DIR/registry.key" -out "$CERTS_DIR/registry.csr" -subj "/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
openssl x509 -req -in "$CERTS_DIR/registry.csr" -CA "$CERTS_DIR/ca.crt" -CAkey "$CERTS_DIR/ca.key" \
    -CAcreateserial -out "$CERTS_DIR/registry.crt" -days 365 -sha256 \
    -extensions SAN -extfile <(echo "[SAN]"; echo "subjectAltName=DNS:localhost,IP:127.0.0.1")

# Clean up CSR files
rm -f "$CERTS_DIR"/*.csr "$CERTS_DIR"/*.srl

echo "Test certificates generated successfully in $CERTS_DIR"
echo "Generated files:"
echo "  - ca.crt (CA certificate)"
echo "  - ca.key (CA private key)"  
echo "  - registry.crt (Registry server certificate)"
echo "  - registry.key (Registry private key)"
echo ""
echo "Note: These are self-signed certificates for testing only!"