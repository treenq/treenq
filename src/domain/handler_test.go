package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

// --- MockDB Implementation ---

type MockDB struct {
	// GetUserPATByUserID
	GetUserPATByUserIDFunc             func(ctx context.Context, userID string) (string, error)
	GetUserPATByUserIDCalledWithUserID string

	// CreateUserPAT
	CreateUserPATFunc                   func(ctx context.Context, userID string, encryptedPAT string) error
	CreateUserPATCalledWithUserID       string
	CreateUserPATCalledWithEncryptedPAT string

	// DeleteUserPAT
	DeleteUserPATFunc             func(ctx context.Context, userID string) error
	DeleteUserPATCalledWithUserID string

	// GetAllUserPATs
	GetAllUserPATsFunc func(ctx context.Context) ([]string, error)

	// Other Database methods (not directly used by PAT handlers, but needed for interface completeness)
	GetOrCreateUserFunc     func(ctx context.Context, user UserInfo) (UserInfo, error)
	SaveDeploymentFunc      func(ctx context.Context, def AppDeployment) (AppDeployment, error)
	UpdateDeploymentFunc    func(ctx context.Context, def AppDeployment) error
	GetDeploymentFunc       func(ctx context.Context, deploymentID string) (AppDeployment, error)
	GetDeploymentHistoryFunc func(ctx context.Context, appID string) ([]AppDeployment, error)
	LinkGithubFunc          func(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) (string, error)
	SaveGithubReposFunc     func(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	RemoveGithubReposFunc   func(ctx context.Context, installationID int, repos []InstalledRepository) error
	GetGithubReposFunc      func(ctx context.Context, email string) ([]Repository, error)
	GetInstallationIDFunc   func(ctx context.Context, userID string) (string, int, error)
	SaveInstallationFunc    func(ctx context.Context, userID string, githubID int) (string, error)
	ConnectRepoFunc         func(ctx context.Context, userID, repoID, branchName string) (Repository, error)
	GetRepoByGithubFunc     func(ctx context.Context, githubRepoID int) (Repository, error)
	GetRepoByIDFunc         func(ctx context.Context, userID, repoID string) (Repository, error)
	RepoIsConnectedFunc     func(ctx context.Context, repoID string) (bool, error)
}

// PAT-related methods
func (m *MockDB) GetUserPATByUserID(ctx context.Context, userID string) (string, error) {
	m.GetUserPATByUserIDCalledWithUserID = userID
	if m.GetUserPATByUserIDFunc != nil {
		return m.GetUserPATByUserIDFunc(ctx, userID)
	}
	return "", fmt.Errorf("MockDB: GetUserPATByUserIDFunc not set")
}

func (m *MockDB) CreateUserPAT(ctx context.Context, userID string, encryptedPAT string) error {
	m.CreateUserPATCalledWithUserID = userID
	m.CreateUserPATCalledWithEncryptedPAT = encryptedPAT
	if m.CreateUserPATFunc != nil {
		return m.CreateUserPATFunc(ctx, userID, encryptedPAT)
	}
	return fmt.Errorf("MockDB: CreateUserPATFunc not set")
}

func (m *MockDB) DeleteUserPAT(ctx context.Context, userID string) error {
	m.DeleteUserPATCalledWithUserID = userID
	if m.DeleteUserPATFunc != nil {
		return m.DeleteUserPATFunc(ctx, userID)
	}
	return fmt.Errorf("MockDB: DeleteUserPATFunc not set")
}

func (m *MockDB) GetAllUserPATs(ctx context.Context) ([]string, error) {
	if m.GetAllUserPATsFunc != nil {
		return m.GetAllUserPATsFunc(ctx)
	}
	return nil, fmt.Errorf("MockDB: GetAllUserPATsFunc not set")
}

// Other Database interface methods (minimal implementation)
func (m *MockDB) GetOrCreateUser(ctx context.Context, user UserInfo) (UserInfo, error) {
	if m.GetOrCreateUserFunc != nil {
		return m.GetOrCreateUserFunc(ctx, user)
	}
	return UserInfo{}, errors.New("MockDB: GetOrCreateUser not implemented")
}
func (m *MockDB) SaveDeployment(ctx context.Context, def AppDeployment) (AppDeployment, error) {
	if m.SaveDeploymentFunc != nil {
		return m.SaveDeploymentFunc(ctx, def)
	}
	return AppDeployment{}, errors.New("MockDB: SaveDeployment not implemented")
}
func (m *MockDB) UpdateDeployment(ctx context.Context, def AppDeployment) error {
	if m.UpdateDeploymentFunc != nil {
		return m.UpdateDeploymentFunc(ctx, def)
	}
	return errors.New("MockDB: UpdateDeployment not implemented")
}
func (m *MockDB) GetDeployment(ctx context.Context, deploymentID string) (AppDeployment, error) {
	if m.GetDeploymentFunc != nil {
		return m.GetDeploymentFunc(ctx, deploymentID)
	}
	return AppDeployment{}, errors.New("MockDB: GetDeployment not implemented")
}
func (m *MockDB) GetDeploymentHistory(ctx context.Context, appID string) ([]AppDeployment, error) {
	if m.GetDeploymentHistoryFunc != nil {
		return m.GetDeploymentHistoryFunc(ctx, appID)
	}
	return nil, errors.New("MockDB: GetDeploymentHistory not implemented")
}
func (m *MockDB) LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) (string, error) {
	if m.LinkGithubFunc != nil {
		return m.LinkGithubFunc(ctx, installationID, senderLogin, repos)
	}
	return "", errors.New("MockDB: LinkGithub not implemented")
}
func (m *MockDB) SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error {
	if m.SaveGithubReposFunc != nil {
		return m.SaveGithubReposFunc(ctx, installationID, senderLogin, repos)
	}
	return errors.New("MockDB: SaveGithubRepos not implemented")
}
func (m *MockDB) RemoveGithubRepos(ctx context.Context, installationID int, repos []InstalledRepository) error {
	if m.RemoveGithubReposFunc != nil {
		return m.RemoveGithubReposFunc(ctx, installationID, repos)
	}
	return errors.New("MockDB: RemoveGithubRepos not implemented")
}
func (m *MockDB) GetGithubRepos(ctx context.Context, email string) ([]Repository, error) {
	if m.GetGithubReposFunc != nil {
		return m.GetGithubReposFunc(ctx, email)
	}
	return nil, errors.New("MockDB: GetGithubRepos not implemented")
}
func (m *MockDB) GetInstallationID(ctx context.Context, userID string) (string, int, error) {
	if m.GetInstallationIDFunc != nil {
		return m.GetInstallationIDFunc(ctx, userID)
	}
	return "", 0, errors.New("MockDB: GetInstallationID not implemented")
}
func (m *MockDB) SaveInstallation(ctx context.Context, userID string, githubID int) (string, error) {
	if m.SaveInstallationFunc != nil {
		return m.SaveInstallationFunc(ctx, userID, githubID)
	}
	return "", errors.New("MockDB: SaveInstallation not implemented")
}
func (m *MockDB) ConnectRepo(ctx context.Context, userID, repoID, branchName string) (Repository, error) {
	if m.ConnectRepoFunc != nil {
		return m.ConnectRepoFunc(ctx, userID, repoID, branchName)
	}
	return Repository{}, errors.New("MockDB: ConnectRepo not implemented")
}
func (m *MockDB) GetRepoByGithub(ctx context.Context, githubRepoID int) (Repository, error) {
	if m.GetRepoByGithubFunc != nil {
		return m.GetRepoByGithubFunc(ctx, githubRepoID)
	}
	return Repository{}, errors.New("MockDB: GetRepoByGithub not implemented")
}
func (m *MockDB) GetRepoByID(ctx context.Context, userID, repoID string) (Repository, error) {
	if m.GetRepoByIDFunc != nil {
		return m.GetRepoByIDFunc(ctx, userID, repoID)
	}
	return Repository{}, errors.New("MockDB: GetRepoByID not implemented")
}
func (m *MockDB) RepoIsConnected(ctx context.Context, repoID string) (bool, error) {
	if m.RepoIsConnectedFunc != nil {
		return m.RepoIsConnectedFunc(ctx, repoID)
	}
	return false, errors.New("MockDB: RepoIsConnected not implemented")
}

// --- Test Setup ---
var testPATEncryptionKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f" // 32-byte hex

func setupTestHandler(mockDB *MockDB) *Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // Discard logs during tests
	return NewHandler(mockDB, nil, nil, nil, nil, nil, "", nil, nil, "", 0, logger)
}

func setTestEncryptionKey(t *testing.T) func() {
	t.Helper()
	originalKey := os.Getenv("PAT_ENCRYPTION_KEY")
	os.Setenv("PAT_ENCRYPTION_KEY", testPATEncryptionKey)
	return func() {
		if originalKey == "" {
			os.Unsetenv("PAT_ENCRYPTION_KEY")
		} else {
			os.Setenv("PAT_ENCRYPTION_KEY", originalKey)
		}
	}
}

// --- IssuePAT Tests ---

func TestIssuePAT_Success(t *testing.T) {
	cleanup := setTestEncryptionKey(t)
	defer cleanup()

	mockDB := &MockDB{}
	mockDB.GetUserPATByUserIDFunc = func(ctx context.Context, userID string) (string, error) {
		return "", sql.ErrNoRows // Simulate PAT not found
	}
	mockDB.CreateUserPATFunc = func(ctx context.Context, userID string, encryptedPAT string) error {
		if userID != "placeholder_user_id" {
			t.Errorf("expected placeholder_user_id, got %s", userID)
		}
		if encryptedPAT == "" {
			t.Error("expected non-empty encrypted PAT")
		}
		return nil // Success
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodPost, "/hook", nil)
	rr := httptest.NewRecorder()
	handler.IssuePAT(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusCreated, rr.Body.String())
	}

	var resp IssuePATResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}
	if len(resp.PAT) != PATLength*2 {
		t.Errorf("Expected PAT of length %d, got %d", PATLength*2, len(resp.PAT))
	}
	if mockDB.CreateUserPATCalledWithUserID != "placeholder_user_id" {
		t.Errorf("CreateUserPAT called with wrong userID: got %s", mockDB.CreateUserPATCalledWithUserID)
	}
}

func TestIssuePAT_Conflict(t *testing.T) {
	cleanup := setTestEncryptionKey(t)
	defer cleanup()

	mockDB := &MockDB{}
	mockDB.GetUserPATByUserIDFunc = func(ctx context.Context, userID string) (string, error) {
		return "existing_encrypted_pat", nil // Simulate PAT already exists
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodPost, "/hook", nil)
	rr := httptest.NewRecorder()
	handler.IssuePAT(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusConflict, rr.Body.String())
	}
	expectedMsg := "User already has an active PAT"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

func TestIssuePAT_EncryptError(t *testing.T) {
	// Unset key for this test
	originalKey := os.Getenv("PAT_ENCRYPTION_KEY")
	os.Unsetenv("PAT_ENCRYPTION_KEY")
	defer os.Setenv("PAT_ENCRYPTION_KEY", originalKey)

	mockDB := &MockDB{}
	mockDB.GetUserPATByUserIDFunc = func(ctx context.Context, userID string) (string, error) {
		return "", sql.ErrNoRows
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodPost, "/hook", nil)
	rr := httptest.NewRecorder()
	handler.IssuePAT(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
	}
	expectedMsg := "Failed to encrypt PAT"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

func TestIssuePAT_StoreError(t *testing.T) {
	cleanup := setTestEncryptionKey(t)
	defer cleanup()

	mockDB := &MockDB{}
	mockDB.GetUserPATByUserIDFunc = func(ctx context.Context, userID string) (string, error) {
		return "", sql.ErrNoRows
	}
	mockDB.CreateUserPATFunc = func(ctx context.Context, userID string, encryptedPAT string) error {
		return errors.New("database store failed")
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodPost, "/hook", nil)
	rr := httptest.NewRecorder()
	handler.IssuePAT(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
	}
	expectedMsg := "Failed to store PAT"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

// --- DeletePAT Tests ---

func TestDeletePAT_Success(t *testing.T) {
	mockDB := &MockDB{}
	mockDB.DeleteUserPATFunc = func(ctx context.Context, userID string) error {
		if userID != "placeholder_user_id" {
			t.Errorf("expected placeholder_user_id, got %s", userID)
		}
		return nil // Success
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodDelete, "/hook", nil)
	rr := httptest.NewRecorder()
	handler.DeletePAT(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusNoContent, rr.Body.String())
	}
	if mockDB.DeleteUserPATCalledWithUserID != "placeholder_user_id" {
		t.Errorf("DeleteUserPAT called with wrong userID: got %s", mockDB.DeleteUserPATCalledWithUserID)
	}
}

func TestDeletePAT_DBError(t *testing.T) {
	mockDB := &MockDB{}
	mockDB.DeleteUserPATFunc = func(ctx context.Context, userID string) error {
		return errors.New("database delete failed")
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodDelete, "/hook", nil)
	rr := httptest.NewRecorder()
	handler.DeletePAT(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
	}
	expectedMsg := "Failed to delete PAT"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

// --- ValidatePAT Tests ---

func TestValidatePAT_Success(t *testing.T) {
	cleanup := setTestEncryptionKey(t)
	defer cleanup()

	plainTextPAT := "this_is_a_test_pat_1234567890abcdef"
	encryptedPAT, err := encryptPAT(plainTextPAT)
	if err != nil {
		t.Fatalf("Failed to encrypt PAT for test: %v", err)
	}

	mockDB := &MockDB{}
	mockDB.GetAllUserPATsFunc = func(ctx context.Context) ([]string, error) {
		return []string{encryptedPAT, "another_encrypted_pat"}, nil
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodGet, "/hook?key="+plainTextPAT, nil)
	rr := httptest.NewRecorder()
	handler.ValidatePAT(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}
}

func TestValidatePAT_Invalid(t *testing.T) {
	cleanup := setTestEncryptionKey(t)
	defer cleanup()

	mockDB := &MockDB{}
	mockDB.GetAllUserPATsFunc = func(ctx context.Context) ([]string, error) {
		// Encrypt a *different* PAT to ensure the one being validated isn't accidentally present
		otherPAT, _ := encryptPAT("a_completely_different_pat_0987654321fedcba")
		return []string{otherPAT, "another_encrypted_pat"}, nil
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodGet, "/hook?key=this_is_an_invalid_pat", nil)
	rr := httptest.NewRecorder()
	handler.ValidatePAT(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusUnauthorized, rr.Body.String())
	}
	expectedMsg := "Invalid or expired PAT"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

func TestValidatePAT_NoKey(t *testing.T) {
	handler := setupTestHandler(&MockDB{}) // DB not really used here
	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	rr := httptest.NewRecorder()
	handler.ValidatePAT(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusBadRequest, rr.Body.String())
	}
	expectedMsg := "Missing 'key' query parameter"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

func TestValidatePAT_DBError(t *testing.T) {
	mockDB := &MockDB{}
	mockDB.GetAllUserPATsFunc = func(ctx context.Context) ([]string, error) {
		return nil, errors.New("database query failed")
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodGet, "/hook?key=somekey", nil)
	rr := httptest.NewRecorder()
	handler.ValidatePAT(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusInternalServerError, rr.Body.String())
	}
	expectedMsg := "Failed to validate PAT"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

func TestValidatePAT_DecryptError(t *testing.T) {
	// No need to set PAT_ENCRYPTION_KEY here, as we want decryption to fail with garbage.
	// However, if encryptPAT is called, it would need it. We provide malformed data.
	cleanup := setTestEncryptionKey(t) // Still good practice for consistency if other parts call encrypt
	defer cleanup()

	mockDB := &MockDB{}
	mockDB.GetAllUserPATsFunc = func(ctx context.Context) ([]string, error) {
		return []string{"this_is_not_a_valid_encrypted_hex_string_and_should_cause_decrypt_error"}, nil
	}

	handler := setupTestHandler(mockDB)
	req := httptest.NewRequest(http.MethodGet, "/hook?key=somekey_to_validate", nil)
	rr := httptest.NewRecorder()
	handler.ValidatePAT(rr, req)

	// Decryption errors for individual PATs are logged but don't stop the loop.
	// If no valid PAT is found after checking all, it results in Unauthorized.
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", status, http.StatusUnauthorized, rr.Body.String())
	}
	expectedMsg := "Invalid or expired PAT"
	if !strings.Contains(rr.Body.String(), expectedMsg) {
		t.Errorf("expected response body to contain %q, got %q", expectedMsg, rr.Body.String())
	}
}

// Placeholder for context with UserID if needed in future
// func contextWithUserID(ctx context.Context, userID string) context.Context {
// 	return context.WithValue(ctx, "userID", userID) // Replace "userID" with actual key used in middleware
// }

// Example of using context with UserID in a test, if handler is adapted:
// func TestIssuePAT_Success_WithUserIDInContext(t *testing.T) {
// 	...
// 	req := httptest.NewRequest(http.MethodPost, "/hook", nil)
//   // Assuming "test_user_123" is the ID we want to simulate
// 	ctx := contextWithUserID(req.Context(), "test_user_123")
// 	req = req.WithContext(ctx)
// 	...
//   // Then in mockDB.CreateUserPATFunc, check for "test_user_123"
// }

// Note: The `Database` interface in `handler.go` has many methods.
// The MockDB above implements them with simple error returns.
// This is to satisfy the interface requirement.
// Only PAT-related methods are given functional mock implementations.
// If other methods were called by PAT handlers, their mocks would need to be fleshed out.

// Adjust PATLength based on the actual definition in pat.go if different
// For now, assuming PATLength = 32 as per pat.go, so hex string is 64
const testHandlerPATLength = 32

func TestIssuePATResponse_PATLength(t *testing.T) {
	// This is more of a check on my understanding of PATLength for the test
	if PATLength != testHandlerPATLength {
		t.Logf("Warning: PATLength in domain (%d) differs from test assumption (%d). TestIssuePAT_Success might use the wrong length for PAT in response.", PATLength, testHandlerPATLength)
	}
}
```
