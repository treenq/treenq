# Docker Artifact Integration Tests

This directory contains test data and configuration for Docker artifact integration tests.

## Test Scenarios

The integration tests cover four main scenarios using table-driven tests:

1. **no_tls_no_auth**: HTTP registry without authentication ✅
2. **tls_no_auth**: HTTPS registry with TLS but no authentication ✅  
3. **tls_with_auth**: HTTPS registry with TLS and basic authentication ✅
4. **no_tls_with_auth**: HTTP registry with basic authentication ✅

## Running Tests

### All Tests
```bash
go test -v ./src/repo/artifacts/
```

### TLS Tests Setup
For TLS test scenarios, generate certificates first:
```bash
cd src/repo/artifacts/testdata
./generate-test-certs.sh
go test -v ./src/repo/artifacts/
```

## Test Infrastructure

### Files Structure
```
testdata/
├── docker-compose.yaml                    # Base BuildKit service
├── docker-compose.no_tls_no_auth.yaml    # HTTP registry, no auth
├── docker-compose.tls_no_auth.yaml       # HTTPS registry, no auth  
├── docker-compose.tls_with_auth.yaml     # HTTPS registry, with auth
├── docker-compose.no_tls_with_auth.yaml  # HTTP registry, with auth
├── generate-test-certs.sh                # Certificate generation script
├── auth/htpasswd                         # Basic auth credentials
└── certs/                                # Generated certificates (optional)
    ├── ca.crt                            # CA certificate  
    ├── ca.key                            # CA private key
    ├── registry.crt                      # Registry server certificate
    └── registry.key                      # Registry private key
```

### Test User Credentials
- Username: `testuser`
- Password: `testpass`

## Test Structure

Table-driven tests using the `testCase` struct:

```go
type testCase struct {
    name              string  // Test scenario name
    registryPort      string  // Fixed port (15000-15003)
    buildkitTLSCA     string  // BuildKit TLS CA (unused in current tests)
    registryTLSVerify bool    // Enable TLS verification for registry
    registryCert      string  // Registry CA certificate path
    registryUsername  string  // Registry auth username
    registryPassword  string  // Registry auth password
}
```

Each test case:
1. Starts dedicated docker-compose stack for the scenario
2. Tests Inspect operation (should return ErrImageNotFound for non-existent image)
3. Tests Build operation (should succeed and push to registry)
4. Tests Inspect operation again (should now find the built image)
5. Cleans up containers automatically

## Architecture

### Port Mapping
- **15000**: HTTP registry, no auth
- **15001**: HTTPS registry, no auth
- **15002**: HTTPS registry, with auth  
- **15003**: HTTP registry, with auth
- **1234**: BuildKit daemon (shared across all tests)

### Certificate Usage
- Only registry certificates are generated and used
- BuildKit runs without TLS (simplified setup)
- Registry containers mount certificates for HTTPS scenarios

## Notes

- Tests use separate docker-compose files per scenario for isolation
- Fixed port mapping eliminates dynamic port issues
- Certificate generation is optional - TLS tests skip if certificates don't exist
- All test containers are automatically cleaned up after each test
- Uses testcontainers for reliable container lifecycle management