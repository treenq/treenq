package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/treenq/treenq/pkg/vel"
)

// --- Mock Implementations ---

type MockKube struct {
	RemoveSecretFunc func(ctx context.Context, rawConfig string, space, repoID, key string) error
	GetSecretFunc    func(ctx context.Context, rawConfig string, space, repoID, key string) (string, error)
	StoreSecretFunc  func(ctx context.Context, rawConfig string, space, repoID, key, value string) error
	// Add other methods if Handler uses them and they need mocking for these tests
}

func (m *MockKube) RemoveSecret(ctx context.Context, rawConfig string, space, repoID, key string) error {
	if m.RemoveSecretFunc != nil {
		return m.RemoveSecretFunc(ctx, rawConfig, space, repoID, key)
	}
	return nil
}

func (m *MockKube) GetSecret(ctx context.Context, rawConfig string, space, repoID, key string) (string, error) {
	if m.GetSecretFunc != nil {
		return m.GetSecretFunc(ctx, rawConfig, space, repoID, key)
	}
	return "", nil
}

func (m *MockKube) StoreSecret(ctx context.Context, rawConfig string, space, repoID, key, value string) error {
	if m.StoreSecretFunc != nil {
		return m.StoreSecretFunc(ctx, rawConfig, space, repoID, key, value)
	}
	return nil
}

type MockStore struct {
	RemoveSecretFunc              func(ctx context.Context, repoID, key, userDisplayName string) error
	RepositorySecretKeyExistsFunc func(ctx context.Context, repoID, key, userDisplayName string) (bool, error)
	SaveSecretFunc                func(ctx context.Context, repoID, key, userDisplayName string) error
	// Add other methods if Handler uses them
}

func (m *MockStore) RemoveSecret(ctx context.Context, repoID, key, userDisplayName string) error {
	if m.RemoveSecretFunc != nil {
		return m.RemoveSecretFunc(ctx, repoID, key, userDisplayName)
	}
	return nil
}

func (m *MockStore) RepositorySecretKeyExists(ctx context.Context, repoID, key, userDisplayName string) (bool, error) {
	if m.RepositorySecretKeyExistsFunc != nil {
		return m.RepositorySecretKeyExistsFunc(ctx, repoID, key, userDisplayName)
	}
	return false, nil
}

func (m *MockStore) SaveSecret(ctx context.Context, repoID, key, userDisplayName string) error {
	if m.SaveSecretFunc != nil {
		return m.SaveSecretFunc(ctx, repoID, key, userDisplayName)
	}
	return nil
}

// --- Test Handler with Mocked GetProfile ---
// We need to control GetProfile behavior for these tests.
// We can make GetProfile a function field on the handler for tests,
// or embed Handler and override GetProfile for testing.
// For simplicity, let's assume GetProfile can be mocked directly or is part of an interface.
// If Handler.GetProfile is a concrete method, this will need adjustment.
// Let's assume a simplified Handler for tests or that GetProfile is mockable.

type TestHandler struct {
	*Handler // Embed original handler if it has other methods not being tested directly
	mockKube   *MockKube
	mockStore  *MockStore
	// Store the original GetProfile if we are temporarily overriding it
	// originalGetProfile func(context.Context, struct{}) (GetProfileResponse, *vel.Error)

	// Mocked GetProfile
	GetProfileFunc func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error)
}

// Override GetProfile for the TestHandler
func (th *TestHandler) GetProfile(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
	if th.GetProfileFunc != nil {
		return th.GetProfileFunc(ctx, req)
	}
	// Default mock profile for tests that don't focus on GetProfile failure
	return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
}

// Actual GetProfile might look like this in the real Handler:
// func (h *Handler) GetProfile(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) { ... }


func NewTestHandler(kube *MockKube, store *MockStore) *TestHandler {
	// Create a base Handler. If NewHandler requires concrete types, this will need to be adjusted.
	// For now, assume we can set kube and db fields directly or NewHandler takes interfaces.
	baseHandler := &Handler{
		kube:       kube, // This assumes Handler has KubeInterface field
		db:         store,  // This assumes Handler has StoreInterface field
		kubeConfig: "test-config",
		// other necessary fields for Handler
	}
	return &TestHandler{
		Handler:   baseHandler,
		mockKube:  kube,
		mockStore: store,
	}
}


// --- Tests for RemoveSecret ---

func TestRemoveSecret_Success(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	// Assuming Handler has fields 'kube' and 'db' that accept these mocks
	// and GetProfile can be mocked or is part of an overridable structure.
	// For simplicity, we'll assume a constructor or direct field setting.
	// Also assuming kubeConfig is some string value.
	
	// h := &Handler{kube: mockKube, db: mockStore, kubeConfig: "dummy-config"}
	// Replace with a test handler that allows mocking GetProfile
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}


	mockKube.RemoveSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key string) error {
		if repoID != "repo1" || key != "key1" || space != "test-user" {
			return fmt.Errorf("kube.RemoveSecret called with unexpected args: space=%s, repoID=%s, key=%s", space, repoID, key)
		}
		return nil
	}
	mockStore.RemoveSecretFunc = func(ctx context.Context, repoID, key, userDisplayName string) error {
		if repoID != "repo1" || key != "key1" || userDisplayName != "test-user" {
			return fmt.Errorf("db.RemoveSecret called with unexpected args: repoID=%s, key=%s, user=%s", repoID, key, userDisplayName)
		}
		return nil
	}

	req := RemoveSecretRequest{RepoID: "repo1", Key: "key1"}
	_, err := h.RemoveSecret(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRemoveSecret_KubeRemoveFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockKube.RemoveSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key string) error {
		return errors.New("kube failed")
	}

	req := RemoveSecretRequest{RepoID: "repo1", Key: "key1"}
	_, velErr := h.RemoveSecret(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr.Message != "failed to remove secret from Kubernetes" {
		t.Errorf("Expected error message 'failed to remove secret from Kubernetes', got '%s'", velErr.Message)
	}
	if velErr.Err.Error() != "kube failed" {
		t.Errorf("Expected underlying error 'kube failed', got '%v'", velErr.Err)
	}
}

func TestRemoveSecret_DBRemoveFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{UserInfo: UserInfo{DisplayName: "test-user"}}, nil
	}

	mockKube.RemoveSecretFunc = func(ctx context.Context, rawConfig string, space, repoID, key string) error {
		return nil // Kube succeeds
	}
	mockStore.RemoveSecretFunc = func(ctx context.Context, repoID, key, userDisplayName string) error {
		return errors.New("db failed")
	}

	req := RemoveSecretRequest{RepoID: "repo1", Key: "key1"}
	_, velErr := h.RemoveSecret(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr.Message != "failed to remove secret from database" {
		t.Errorf("Expected error message 'failed to remove secret from database', got '%s'", velErr.Message)
	}
	if velErr.Err.Error() != "db failed" {
		t.Errorf("Expected underlying error 'db failed', got '%v'", velErr.Err)
	}
}

func TestRemoveSecret_GetProfileFails(t *testing.T) {
	mockKube := &MockKube{}
	mockStore := &MockStore{}
	h := NewTestHandler(mockKube, mockStore)
	
	expectedErr := &vel.Error{Message: "get profile failed"}
	h.GetProfileFunc = func(ctx context.Context, req struct{}) (GetProfileResponse, *vel.Error) {
		return GetProfileResponse{}, expectedErr
	}

	req := RemoveSecretRequest{RepoID: "repo1", Key: "key1"}
	_, velErr := h.RemoveSecret(context.Background(), req)

	if velErr == nil {
		t.Fatal("Expected an error, got nil")
	}
	if velErr != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, velErr)
	}
}
