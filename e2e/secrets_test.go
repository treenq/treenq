package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baseURL should be configured for your test environment
var baseURL = "http://localhost:8080" // Replace with actual or env var

// testRepoID is a placeholder for a repository ID used in tests.
const testRepoID = "e2e-test-repo-secrets"

// authToken is a placeholder for authentication. In a real scenario,
// this would be dynamically obtained or managed by a test helper.
var authToken = "test-auth-token" // Replace with actual token or retrieval logic

// Client for making HTTP requests
var httpClient = &http.Client{Timeout: 10 * time.Second}

// SetSecretPayload matches the request structure for setting a secret.
type SetSecretPayload struct {
	SecretKey   string `json:"secret_key"`
	SecretValue string `json:"secret_value"`
}

// ViewSecretResponse matches the response structure for viewing a secret.
type ViewSecretResponse struct {
	SecretValue string `json:"secret_value"`
}

// Helper function to create and execute HTTP requests with authentication.
func makeRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonBody)
	}

	fullURL := baseURL + path
	req, err := http.NewRequest(method, fullURL, reqBody)
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := httpClient.Do(req)
	require.NoError(t, err, "Failed to execute request")

	return resp
}

// TestSecretsAPI_Lifecycle covers the full lifecycle of secrets management.
func TestSecretsAPI_Lifecycle(t *testing.T) {
	// Setup: Ensure baseURL is set, potentially from an environment variable
	envBaseURL := os.Getenv("E2E_BASE_URL")
	if envBaseURL != "" {
		baseURL = envBaseURL
	}
	// Similarly for authToken, if needed
	envAuthToken := os.Getenv("E2E_AUTH_TOKEN")
	if envAuthToken != "" {
		authToken = envAuthToken
	}

	// --- 1. Initial Get Secrets (should be empty) ---
	t.Run("InitialGetSecrets", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		resp := makeRequest(t, http.MethodGet, path, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected OK status for initial get secrets")

		var secretKeys []string
		err := json.NewDecoder(resp.Body).Decode(&secretKeys)
		require.NoError(t, err, "Failed to decode response body for initial get secrets")
		assert.Empty(t, secretKeys, "Expected no secret keys initially")
	})

	secretKey1 := "testKey" + fmt.Sprintf("%d", time.Now().UnixNano()) // Unique key
	secretValue1 := "testValue123"
	updatedSecretValue1 := "updatedTestValue456"

	// --- 2. Set Secret ---
	t.Run("SetSecret1", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		payload := SetSecretPayload{SecretKey: secretKey1, SecretValue: secretValue1}
		resp := makeRequest(t, http.MethodPost, path, payload)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected Created status for setting secret")
		// Optionally, check response body if any (e.g., {"status": "secret set successfully"})
	})

	// --- 3. Get Secrets After Set ---
	t.Run("GetSecretsAfterSet1", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		resp := makeRequest(t, http.MethodGet, path, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected OK status for get secrets after set")

		var secretKeys []string
		err := json.NewDecoder(resp.Body).Decode(&secretKeys)
		require.NoError(t, err, "Failed to decode response body for get secrets after set")
		assert.Contains(t, secretKeys, secretKey1, "Expected to find the newly set secret key")
	})

	// --- 4. View Secret ---
	t.Run("ViewSecret1", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets/%s", testRepoID, secretKey1)
		resp := makeRequest(t, http.MethodGet, path, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected OK status for viewing secret")

		var viewResp ViewSecretResponse
		err := json.NewDecoder(resp.Body).Decode(&viewResp)
		require.NoError(t, err, "Failed to decode response body for viewing secret")
		assert.Equal(t, secretValue1, viewResp.SecretValue, "Secret value does not match")
	})

	// --- 5. Set (Update) Existing Secret ---
	t.Run("UpdateSecret1", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		payload := SetSecretPayload{SecretKey: secretKey1, SecretValue: updatedSecretValue1}
		resp := makeRequest(t, http.MethodPost, path, payload)
		defer resp.Body.Close()

		// The API returns StatusCreated for both create and update.
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected Created status for updating secret")
	})

	// --- 6. View Updated Secret ---
	t.Run("ViewUpdatedSecret1", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets/%s", testRepoID, secretKey1)
		resp := makeRequest(t, http.MethodGet, path, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected OK status for viewing updated secret")

		var viewResp ViewSecretResponse
		err := json.NewDecoder(resp.Body).Decode(&viewResp)
		require.NoError(t, err, "Failed to decode response body for viewing updated secret")
		assert.Equal(t, updatedSecretValue1, viewResp.SecretValue, "Updated secret value does not match")
	})

	// --- 7. View Non-Existent Secret ---
	t.Run("ViewNonExistentSecret", func(t *testing.T) {
		nonExistentKey := "nonExistentKey" + fmt.Sprintf("%d", time.Now().UnixNano())
		path := fmt.Sprintf("/repositories/%s/secrets/%s", testRepoID, nonExistentKey)
		resp := makeRequest(t, http.MethodGet, path, nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Expected Not Found status for non-existent secret")
	})

	// --- 8. Set another secret to test listing multiple ---
	secretKey2 := "testKeySecond" + fmt.Sprintf("%d", time.Now().UnixNano())
	secretValue2 := "testValueSecond789"
	t.Run("SetSecret2", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		payload := SetSecretPayload{SecretKey: secretKey2, SecretValue: secretValue2}
		resp := makeRequest(t, http.MethodPost, path, payload)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// --- 9. Get Secrets After Setting Multiple ---
	t.Run("GetSecretsAfterSetMultiple", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		resp := makeRequest(t, http.MethodGet, path, nil)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var secretKeys []string
		err := json.NewDecoder(resp.Body).Decode(&secretKeys)
		require.NoError(t, err)
		assert.Contains(t, secretKeys, secretKey1, "Expected to find secretKey1")
		assert.Contains(t, secretKeys, secretKey2, "Expected to find secretKey2")
	})

	// --- 10. Set secret with empty key ---
	t.Run("SetSecretEmptyKey", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		payload := SetSecretPayload{SecretKey: "", SecretValue: "someValue"}
		resp := makeRequest(t, http.MethodPost, path, payload)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected Bad Request for empty secret key")
	})

	// --- 11. Set secret with empty value ---
	t.Run("SetSecretEmptyValue", func(t *testing.T) {
		path := fmt.Sprintf("/repositories/%s/secrets", testRepoID)
		payload := SetSecretPayload{SecretKey: "someKey", SecretValue: ""}
		resp := makeRequest(t, http.MethodPost, path, payload)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected Bad Request for empty secret value")
	})

}

// TODO: Add more specific test cases if needed, for example:
// - Test with invalid repoID format (if there are specific validation rules)
// - Test concurrent access (if relevant and feasible within E2E scope)
// - Test behavior with very long keys or values (boundary conditions)
// - Test authentication errors (e.g., missing token, invalid token) - this might require a helper
//   that allows sending requests without auth or with bad auth.

// Helper to simulate getting an auth token (replace with actual logic if available)
// func getAuthTokenForTestUser() string {
//    // In a real test suite, this would call an auth endpoint or use a pre-provisioned token.
//    return "dummy-e2e-auth-token"
// }

// Example of how to run this test:
// Ensure the server is running, then:
// E2E_BASE_URL=http://your-running-app-host E2E_AUTH_TOKEN=your-valid-token go test -v ./e2e -run TestSecretsAPI_Lifecycle
// If K8s is involved, ensure the test environment has credentials configured for the K8s Go client.
// The `testRepoID` namespace in K8s might need to be created or cleaned up as part of test setup/teardown.
// For the purpose of this task, such cleanup is outside the scope of the Go code itself.
// The `makeRequest` helper assumes Bearer token authentication.
// If using `os.Getenv` for baseURL and authToken, these need to be set when running the test.
// For example: `E2E_BASE_URL=http://localhost:8080 E2E_AUTH_TOKEN=your_token go test ./e2e -run TestSecretsAPI_Lifecycle`
// The unique key generation `fmt.Sprintf("%d", time.Now().UnixNano())` is to avoid collisions if tests are run multiple times
// without cleaning up the backend state.
