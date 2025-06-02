package domain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/treenq/treenq/pkg/vel"
)

// --- Mock Implementations (assuming similar to removeSecret_test.go) ---
// For brevity, not repeating MockKube, MockStore, TestHandler struct definitions here.
// Assume they are available or will be defined identically as in removeSecret_test.go.
// If these tests are in the same package, these might not need to be redefined.
// However, `go test` compiles package by package, test files in the same package can see each other.
// So the mocks from removeSecret_test.go should be available.

// --- Tests for RevealSecrets ---

func TestRevealSecrets_Success_SingleKey(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore) // Uses TestHandler from removeSecret_test.go
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockStore.RepositorySecretKeyExistsFunc = func(ctx context.Context, repoID, key, userDisplayName string) (bool, error) {
		if repoID != "repo1" || key != "key1" || userDisplayName != "test-user" {
			return false, fmt.Errorf("db.RepositorySecretKeyExists called with unexpected args")
		}
		return true, nil
	}
	mockKube.GetSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key string) (string, error) {
		if repoID != "repo1" || key != "key1" || space != "test-user" {
			return "", fmt.Errorf("kube.GetSecret called with unexpected args")
		}
		return "value1", nil
	}

	req := RevealSecretsRequest{RepoID: "repo1", Keys: []string{"key1"}}
	resp, err := h.RevealSecrets(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expectedValues := map[string]string{"key1": "value1"}
	if !reflect.DeepEqual(resp.Values, expectedValues) {
		t.Errorf("Expected values %v, got %v", expectedValues, resp.Values)
	}
}

func TestRevealSecrets_Success_MultipleKeys(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	secrets := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	mockStore.RepositorySecretKeyExistsFunc = func(ctx context.Context, repoID, key, userDisplayName string) (bool, error) {
		if _, exists := secrets[key]; !exists {
			return false, fmt.Errorf("db.RepositorySecretKeyExists called for unexpected key: %s", key)
		}
		return true, nil
	}
	mockKube.GetSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key string) (string, error) {
		val, exists := secrets[key]
		if !exists {
			return "", fmt.Errorf("kube.GetSecret called for unexpected key: %s", key)
		}
		return val, nil
	}

	req := RevealSecretsRequest{RepoID: "repo1", Keys: []string{"key1", "key2"}}
	resp, err := h.RevealSecrets(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !reflect.DeepEqual(resp.Values, secrets) {
		t.Errorf("Expected values %v, got %v", secrets, resp.Values)
	}
}

func TestRevealSecrets_KeyNotFound(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockStore.RepositorySecretKeyExistsFunc = func(ctx context.Context, repoID, key, userDisplayName string) (bool, error) {
		return false, nil // Key does not exist
	}

	req := RevealSecretsRequest{RepoID: "repo1", Keys: []string{"key1"}}
	_, velErr := h.RevealSecrets(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr.Code != "SECRET_DOESNT_EXIST" {
		t.Errorf("Expected error code 'SECRET_DOESNT_EXIST', got '%s'", velErr.Code)
	}
	if velErr.Message != "secret key does not exist: key1" {
		t.Errorf("Expected message 'secret key does not exist: key1', got '%s'", velErr.Message)
	}
}

func TestRevealSecrets_DBExistsCheckFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockStore.RepositorySecretKeyExistsFunc = func(ctx context.Context, repoID, key, userDisplayName string) (bool, error) {
		return false, errors.New("db exists check failed")
	}

	req := RevealSecretsRequest{RepoID: "repo1", Keys: []string{"key1"}}
	_, velErr := h.RevealSecrets(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr.Message != "failed to lookup a secret key: key1" {
		t.Errorf("Expected error message 'failed to lookup a secret key: key1', got '%s'", velErr.Message)
	}
	if velErr.Err.Error() != "db exists check failed" {
		t.Errorf("Expected underlying error 'db exists check failed', got '%v'", velErr.Err)
	}
}

func TestRevealSecrets_KubeGetSecretFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockStore.RepositorySecretKeyExistsFunc = func(ctx context.Context, repoID, key, userDisplayName string) (bool, error) {
		return true, nil // Key exists
	}
	mockKube.GetSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key string) (string, error) {
		return "", errors.New("kube get failed")
	}

	req := RevealSecretsRequest{RepoID: "repo1", Keys: []string{"key1"}}
	_, velErr := h.RevealSecrets(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr.Message != "failed to reveal secret: key1" {
		t.Errorf("Expected error message 'failed to reveal secret: key1', got '%s'", velErr.Message)
	}
	if velErr.Err.Error() != "kube get failed" {
		t.Errorf("Expected underlying error 'kube get failed', got '%v'", velErr.Err)
	}
}

func TestRevealSecrets_GetProfileFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)

	expectedErr := &vel.Error{Message: "get profile failed"}
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{}, expectedErr
	}

	req := RevealSecretsRequest{RepoID: "repo1", Keys: []string{"key1"}}
	_, velErr := h.RevealSecrets(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, velErr)
	}
}
