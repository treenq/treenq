⏺ BuildKit + Docker Registry TLS Configuration Cheat Sheet

Problem

BuildKit fails to push to HTTPS registry with TLS certificate validation errors, and BuildKit client connections need secure TLS configuration.

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

4. BuildKit Client TLS

- BuildKit daemon can run with TLS enabled for client connections
- **Server-only TLS**: Client verifies server certificate (recommended)
- **Self-signed certificates**: Simple approach, client trusts server cert directly
- **Mutual TLS**: Both client and server verify each other (more complex)

Directory Structure

```
registry/
├── auth/
│   └── htpasswd
└── certs/
    ├── registry.crt
    ├── registry.key
    └── registry.cert

buildkit/
├── certs/
│   ├── server.crt      # Self-signed server certificate
│   └── server.key      # Server private key
├── buildkitd.toml      # BuildKit daemon config
└── entrypoint.sh       # Container entrypoint
```

Solution Pattern

1. BuildKit Configuration (buildkit/buildkitd.toml)

```toml
debug = true

[grpc]
address = ["tcp://0.0.0.0:1234"]

# TLS configuration for BuildKit daemon (server-only TLS)
[grpc.tls]
cert = "/buildkit/server.crt"
key = "/buildkit/server.key"
# Note: No 'ca' field = server-only TLS (client verifies server)
# To enable mutual TLS, add: ca = "/buildkit/ca.crt"

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
```

2. Docker Compose Mounts

```yaml
buildkit:
  volumes:
    - ./buildkit/entrypoint.sh:/entrypoint.sh
    - ./buildkit/buildkitd.toml:/etc/buildkit/buildkitd.toml
    - ./registry/certs:/certs
    - ./buildkit/certs:/buildkit
```

3. Go Code TLS Configuration

Registry TLS:
```go
tlsConfig := map[string]*authprovider.AuthTLSConfig{
	a.registry: {
		Insecure: !a.registryTLSVerify,
	},
}
if a.registryTLSVerify {
	tlsConfig[a.registry].RootCAs = []string{a.registryCert}
}
```

BuildKit Client TLS (Server-only):
```go
var clientOpts []client.ClientOpt
if a.buildkitTLSCA != "" {
	// Extract hostname from tcp://localhost:1234 -> localhost
	u, err := url.Parse(a.buildkitHost)
	if err != nil {
		return image, fmt.Errorf("given invalid buildkit host: %w", err)
	}
	host := u.Hostname()
	clientOpts = append(clientOpts, client.WithServerConfig(host, a.buildkitTLSCA))
}

c, err := client.New(ctx, a.buildkitHost, clientOpts...)
```

Environment Variables

```bash
# Registry configuration
DOCKER_REGISTRY=localhost:5005 # or registry:5000
REGISTRY_TLS_VERIFY=true
REGISTRY_CERT=./registry/certs/registry.crt
REGISTRY_AUTH_TYPE=basic
REGISTRY_AUTH_USERNAME=testuser
REGISTRY_AUTH_PASSWORD=testpassword

# BuildKit configuration  
BUILDKIT_HOST=tcp://localhost:1234
BUILDKIT_TLS_CA=./buildkit/certs/server.crt
```

Common Errors & Fixes

| Error                                   | Cause               | Fix                                    |
| --------------------------------------- | ------------------- | -------------------------------------- |
| HTTP request to HTTPS server            | BuildKit using HTTP | Add registry config with http = false  |
| certificate signed by unknown authority | Missing CA          | Set ca = ["/path/to/cert"]             |
| certificate valid for X, not Y          | Hostname mismatch   | Use insecure = true or fix certificate |
| certificate required                    | Mutual TLS enabled  | Remove ca field from [grpc.tls] or provide client cert |
| Can't convert string to []string        | Wrong TOML syntax   | Use address = [ "..." ]                |

Best Practices

- **Use server-only TLS** for most cases (simpler, secure enough)
- **Use mutual TLS** only when you need client authentication
- Use proper hostname (registry:5000) for production
- Use insecure = true only for localhost in development  
- Mount both config and certificates to BuildKit container
- Restart BuildKit container after config changes
- Extract hostname from BuildKit URL for proper server name verification
