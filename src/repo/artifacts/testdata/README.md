# Docker Artifact Integration Tests

This directory contains test data and configuration for Docker artifact integration tests.

## Test Scenarios

The integration tests cover four main scenarios:

1. **no_tls_no_auth**: Basic HTTP registry without authentication ✅ **WORKING**
2. **tls_no_auth**: HTTPS registry with TLS but no authentication ✅ **WORKING**
3. **tls_with_auth**: HTTPS registry with TLS and basic authentication ✅ **WORKING**
4. **no_tls_with_auth**: HTTP registry with basic authentication ✅ **WORKING**

## Running Tests

### Basic Tests (HTTP, No Auth)
```bash
go test -v ./src/repo/artifacts/
```

### Full Tests (Including TLS)
1. Generate test certificates:
   ```bash
   cd src/repo/artifacts/testdata
   ./generate-test-certs.sh
   ```

2. Run tests:
   ```bash
   go test -v ./src/repo/artifacts/
   ```

## Test Infrastructure

### Files
- `docker-compose-test.yaml`: Basic test setup (HTTP registry + BuildKit)
- `docker-compose-extended-test.yaml`: Extended setup with TLS and auth support
- `buildkitd-test.toml`: BuildKit configuration for HTTP testing
- `buildkitd-tls-test.toml`: BuildKit configuration for TLS testing
- `generate-test-certs.sh`: Script to generate test certificates
- `auth/htpasswd`: Basic auth credentials (testuser:testpass)
- `certs/`: Directory containing test certificates (generated)

### Test User Credentials
- Username: `testuser`
- Password: `testpass`

### Generated Certificates
The certificate generation script creates:
- `ca.crt` / `ca.key`: Root CA certificate and key
- `registry.crt` / `registry.key`: Registry server certificate and key
- `buildkit-server.crt` / `buildkit-server.key`: BuildKit server certificate and key
- `buildkit-client.crt` / `buildkit-client.key`: BuildKit client certificate and key

All certificates are self-signed and valid for:
- DNS: localhost, registry-tls, registry-auth
- IP: 127.0.0.1

## Test Structure

The tests use a table-driven approach with the `testCase` struct:

```go
type testCase struct {
    name              string
    buildkitTLSCA     string
    registryTLSVerify bool
    registryCert      string
    registryUsername  string
    registryPassword  string
}
```

Each test case:
1. Creates a DockerArtifact with specific configuration
2. Tests Inspect operation (should fail for non-existent image)
3. Tests Build operation (should succeed)
4. Tests Inspect operation again (should now find the built image)

## Current Status

### All Tests Working ✅
All four test scenarios are now fully functional:

- **no_tls_no_auth**: HTTP registry without authentication
- **tls_no_auth**: HTTPS registry with TLS but no authentication  
- **tls_with_auth**: HTTPS registry with TLS and basic authentication
- **no_tls_with_auth**: HTTP registry with basic authentication

### Infrastructure Highlights
- ✅ Multiple registry containers (HTTP, HTTPS, with/without auth) using fixed ports
- ✅ Complete TLS certificate generation (self-signed CA + server certs)
- ✅ BuildKit container with certificate volumes mounted and proper registry configuration
- ✅ Proper health checks for all registry types
- ✅ Table-driven test structure with per-scenario configuration
- ✅ Fixed port mapping (15000-15003) for reliable BuildKit registry configuration
- ✅ Proper authentication handling (empty credentials vs. provided credentials)

### Key Solutions Implemented
1. **Fixed Port Mapping**: Using ports 15000-15003 instead of dynamic ports for reliable BuildKit configuration
2. **Registry-Specific BuildKit Config**: Each registry type has dedicated BuildKit configuration with proper TLS/HTTP settings
3. **Conditional Authentication**: Only set authentication when username/password are provided
4. **Proper Error Handling**: Treat 401 Unauthorized as ErrImageNotFound when image doesn't exist

## Notes

- Tests automatically skip TLS scenarios if certificates are not generated
- Uses testcontainers for Docker Compose integration
- BuildKit uses host networking (localhost:1234)
- Registry ports are randomly assigned by testcontainers
- All test images are cleaned up automatically after test completion
- Test structure demonstrates proper table-driven approach for multiple auth scenarios