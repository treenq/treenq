package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/treenq/treenq/pkg/vel"
)

// --- Mock Implementations (assuming similar to removeSecret_test.go) ---
// Re-using MockKube, MockStore, TestHandler from other _test.go files in the same package.

// --- Tests for SetSecrets ---

func TestSetSecrets_Success_SingleSecret(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	secretToSet := SecretKVPair{Key: "key1", Value: "value1"}

	mockKube.StoreSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key, value string) error {
		if repoID != "repo1" || key != secretToSet.Key || value != secretToSet.Value || space != "test-user" {
			return fmt.Errorf("kube.StoreSecret called with unexpected args: space=%s, repoID=%s, key=%s, value=%s", space, repoID, key, value)
		}
		return nil
	}
	mockStore.SaveSecretFunc = func(ctx context.Context, repoID, key, userDisplayName string) error {
		if repoID != "repo1" || key != secretToSet.Key || userDisplayName != "test-user" {
			return fmt.Errorf("db.SaveSecret called with unexpected args: repoID=%s, key=%s, user=%s", repoID, key, userDisplayName)
		}
		return nil
	}

	req := SetSecretsRequest{RepoID: "repo1", Secrets: []SecretKVPair{secretToSet}}
	_, err := h.SetSecrets(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSetSecrets_Success_MultipleSecrets(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	secretsToSet := []SecretKVPair{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}
	
	callCountKube := 0
	mockKube.StoreSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key, value string) error {
		found := false
		for _, s := range secretsToSet {
			if s.Key == key && s.Value == value {
				found = true
				break
			}
		}
		if !found || repoID != "repo1" || space != "test-user" {
			return fmt.Errorf("kube.StoreSecret received unexpected call for key %s", key)
		}
		callCountKube++
		return nil
	}

	callCountDB := 0
	mockStore.SaveSecretFunc = func(ctx context.Context, repoID, key, userDisplayName string) error {
		found := false
		for _, s := range secretsToSet {
			if s.Key == key {
				found = true
				break
			}
		}
		if !found || repoID != "repo1" || userDisplayName != "test-user" {
				return fmt.Errorf("db.SaveSecret received unexpected call for key %s", key)
		}
		callCountDB++
		return nil
	}

	req := SetSecretsRequest{RepoID: "repo1", Secrets: secretsToSet}
	_, err := h.SetSecrets(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if callCountKube != len(secretsToSet) {
		t.Errorf("Expected kube.StoreSecret to be called %d times, got %d", len(secretsToSet), callCountKube)
	}
	if callCountDB != len(secretsToSet) {
		t.Errorf("Expected db.SaveSecret to be called %d times, got %d", len(secretsToSet), callCountDB)
	}
}

func TestSetSecrets_KubeStoreFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockKube.StoreSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key, value string) error {
		return errors.New("kube store failed")
	}

	req := SetSecretsRequest{RepoID: "repo1", Secrets: []SecretKVPair{{Key: "key1", Value: "value1"}}}
	_, velErr := h.SetSecrets(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr.Message != "failed to store secret: key1" {
		t.Errorf("Expected error message 'failed to store secret: key1', got '%s'", velErr.Message)
	}
	if velErr.Err.Error() != "kube store failed" {
		t.Errorf("Expected underlying error 'kube store failed', got '%v'", velErr.Err)
	}
}

func TestSetSecrets_DBSaveFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockKube.StoreSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key, value string) error {
		return nil // Kube succeeds
	}
	mockStore.SaveSecretFunc = func(ctx context.Context, repoID, key, userDisplayName string) error {
		return errors.New("db save failed")
	}

	req := SetSecretsRequest{RepoID: "repo1", Secrets: []SecretKVPair{{Key: "key1", Value: "value1"}}}
	_, velErr := h.SetSecrets(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr.Message != "failed to save secret: key1" {
		t.Errorf("Expected error message 'failed to save secret: key1', got '%s'", velErr.Message)
	}
	if velErr.Err.Error() != "db save failed" {
		t.Errorf("Expected underlying error 'db save failed', got '%v'", velErr.Err)
	}
}

func TestSetSecrets_GetProfileFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)

	expectedErr := &vel.Error{Message: "get profile failed"}
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{}, expectedErr
	}

	req := SetSecretsRequest{RepoID: "repo1", Secrets: []SecretKVPair{{Key: "key1", Value: "value1"}}}
	_, velErr := h.SetSecrets(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, velErr)
	}
}
