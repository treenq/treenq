⏺ BuildKit + Docker Registry TLS Configuration Cheat Sheet

Problem

BuildKit fails to push to HTTPS registry with TLS certificate validation errors.

Key Insights

1. Registry Protocol Detection

- BuildKit defaults to HTTP for custom ports
- Registry hostname format doesn't specify protocol
- Need explicit configuration to force HTTPS

2. Certificate Validation Requirements

- Certificate must match the connecting hostname exactly
- Self-signed certificates need CA configuration
- Certificate valid for registry:5000 ≠ valid for localhost:5005

3. BuildKit Configuration Location

- Config file: /etc/buildkit/buildkitd.toml (inside container)
- Must mount config file in docker-compose
- Syntax: address = [ "tcp://0.0.0.0:1234" ] (array, not string)

Solution Pattern

1. BuildKit Configuration (buildkitd.toml)

debug = true

[grpc]
address = [ "tcp://0.0.0.0:1234" ]

# Secure config for proper hostname

[registry."registry:5000"]
http = false
insecure = false
ca = ["/certs/registry.crt"]

# Relaxed config for localhost access

[registry."localhost:5005"]
http = false
insecure = true
ca = ["/certs/registry.crt"]

2. Docker Compose Mounts

buildkit:
volumes: - ./buildkitd.toml:/etc/buildkit/buildkitd.toml - ./certs:/certs

3. Go Code TLS Configuration

tlsConfig := map[string]\*authprovider.AuthTLSConfig{
a.registry: {
Insecure: !a.registryTLSVerify,
},
}
if a.registryTLSVerify {
tlsConfig[a.registry].RootCAs = []string{a.registryCert}
}

Environment Variables

DOCKER_REGISTRY=localhost:5005 # or registry:5000
REGISTRY_TLS_VERIFY=true
REGISTRY_CERT=./certs/registry.crt
REGISTRY_AUTH_TYPE=basic
REGISTRY_AUTH_USERNAME=testuser
REGISTRY_AUTH_PASSWORD=testpassword

Common Errors & Fixes

| Error                                   | Cause               | Fix                                    |
| --------------------------------------- | ------------------- | -------------------------------------- |
| HTTP request to HTTPS server            | BuildKit using HTTP | Add registry config with http = false  |
| certificate signed by unknown authority | Missing CA          | Set ca = ["/path/to/cert"]             |
| certificate valid for X, not Y          | Hostname mismatch   | Use insecure = true or fix certificate |
| Can't convert string to []string        | Wrong TOML syntax   | Use address = [ "..." ]                |

Best Practices

- Use proper hostname (registry:5000) for production
- Use insecure = true only for localhost in development
- Mount both config and certificates to BuildKit container
- Restart BuildKit container after config changes
