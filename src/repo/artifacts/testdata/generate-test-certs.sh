#!/bin/bash
# Generate test certificates for TLS testing
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
    -addext "subjectAltName=DNS:localhost,DNS:registry-tls,DNS:registry-auth,IP:127.0.0.1"
openssl x509 -req -in "$CERTS_DIR/registry.csr" -CA "$CERTS_DIR/ca.crt" -CAkey "$CERTS_DIR/ca.key" \
    -CAcreateserial -out "$CERTS_DIR/registry.crt" -days 365 -sha256 \
    -extensions SAN -extfile <(echo "[SAN]"; echo "subjectAltName=DNS:localhost,DNS:registry-tls,DNS:registry-auth,IP:127.0.0.1")

# Generate BuildKit server key and certificate  
openssl genrsa -out "$CERTS_DIR/buildkit-server.key" 4096
openssl req -new -key "$CERTS_DIR/buildkit-server.key" -out "$CERTS_DIR/buildkit-server.csr" -subj "/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
openssl x509 -req -in "$CERTS_DIR/buildkit-server.csr" -CA "$CERTS_DIR/ca.crt" -CAkey "$CERTS_DIR/ca.key" \
    -CAcreateserial -out "$CERTS_DIR/buildkit-server.crt" -days 365 -sha256 \
    -extensions SAN -extfile <(echo "[SAN]"; echo "subjectAltName=DNS:localhost,IP:127.0.0.1")

# Generate BuildKit client key and certificate (for mutual TLS if needed)
openssl genrsa -out "$CERTS_DIR/buildkit-client.key" 4096
openssl req -new -key "$CERTS_DIR/buildkit-client.key" -out "$CERTS_DIR/buildkit-client.csr" -subj "/CN=buildkit-client"
openssl x509 -req -in "$CERTS_DIR/buildkit-client.csr" -CA "$CERTS_DIR/ca.crt" -CAkey "$CERTS_DIR/ca.key" \
    -CAcreateserial -out "$CERTS_DIR/buildkit-client.crt" -days 365 -sha256

# Copy CA cert with different names for different use cases
cp "$CERTS_DIR/ca.crt" "$CERTS_DIR/registry-ca.crt"
cp "$CERTS_DIR/ca.crt" "$CERTS_DIR/buildkit-ca.crt"

# Clean up CSR files
rm -f "$CERTS_DIR"/*.csr "$CERTS_DIR"/*.srl

echo "Test certificates generated successfully in $CERTS_DIR"
echo "Note: These are self-signed certificates for testing only!"