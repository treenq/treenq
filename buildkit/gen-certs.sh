#!/bin/bash
set -e

# Generate CA certificate
openssl req -new -x509 -days 365 -keyout ca.key -out ca.crt -nodes -subj "/CN=Treenq CA"

# Generate registry certificate with SAN for multiple hosts
cat > registry.conf << 'CONF'
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = registry

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = registry
DNS.2 = localhost
IP.1 = 127.0.0.1
CONF

openssl req -new -keyout registry.key -out registry.csr -nodes -config registry.conf
openssl x509 -req -in registry.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out registry.crt -days 365 -extensions v3_req -extfile registry.conf

# Generate buildkit server certificate with SAN for multiple hosts
cat > buildkit.conf << 'CONF'
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = buildkit

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = buildkit
DNS.2 = localhost
IP.1 = 127.0.0.1
CONF

openssl req -new -keyout server.key -out server.csr -nodes -config buildkit.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -extensions v3_req -extfile buildkit.conf

# Move certificates to appropriate directories
mv registry.crt registry.key ../registry/certs/
mv server.crt server.key certs/
cp ca.crt ../registry/certs/
cp ca.crt certs/

# Clean up
rm ca.key ca.crt ca.srl registry.conf registry.csr buildkit.conf server.csr

echo "Certificates generated successfully!"