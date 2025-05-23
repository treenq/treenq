package domain

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/pkg/vel"
)

// MockDatabase is a mock type for the Database interface
type MockDatabase struct {
	GetDeploymentHistoryFunc func(ctx context.Context, repoID string) ([]AppDeployment, error)
	// Dummy implementations for other Database interface methods
	GetOrCreateUserFunc     func(ctx context.Context, user UserInfo) (UserInfo, error)
	SaveDeploymentFunc      func(ctx context.Context, def AppDeployment) (AppDeployment, error)
	UpdateDeploymentFunc    func(ctx context.Context, def AppDeployment) error
	GetDeploymentFunc       func(ctx context.Context, deploymentID string) (AppDeployment, error)
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

// Implement the Database interface for MockDatabase
func (m *MockDatabase) GetDeploymentHistory(ctx context.Context, repoID string) ([]AppDeployment, error) {
	if m.GetDeploymentHistoryFunc != nil {
		return m.GetDeploymentHistoryFunc(ctx, repoID)
	}
	return nil, errors.New("GetDeploymentHistoryFunc not implemented")
}

func (m *MockDatabase) GetOrCreateUser(ctx context.Context, user UserInfo) (UserInfo, error) {
	if m.GetOrCreateUserFunc != nil {
		return m.GetOrCreateUserFunc(ctx, user)
	}
	return UserInfo{}, errors.New("GetOrCreateUserFunc not implemented")
}

func (m *MockDatabase) SaveDeployment(ctx context.Context, def AppDeployment) (AppDeployment, error) {
	if m.SaveDeploymentFunc != nil {
		return m.SaveDeploymentFunc(ctx, def)
	}
	return AppDeployment{}, errors.New("SaveDeploymentFunc not implemented")
}

func (m *MockDatabase) UpdateDeployment(ctx context.Context, def AppDeployment) error {
	if m.UpdateDeploymentFunc != nil {
		return m.UpdateDeploymentFunc(ctx, def)
	}
	return errors.New("UpdateDeploymentFunc not implemented")
}

func (m *MockDatabase) GetDeployment(ctx context.Context, deploymentID string) (AppDeployment, error) {
	if m.GetDeploymentFunc != nil {
		return m.GetDeploymentFunc(ctx, deploymentID)
	}
	return AppDeployment{}, errors.New("GetDeploymentFunc not implemented")
}

func (m *MockDatabase) LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) (string, error) {
	if m.LinkGithubFunc != nil {
		return m.LinkGithubFunc(ctx, installationID, senderLogin, repos)
	}
	return "", errors.New("LinkGithubFunc not implemented")
}

func (m *MockDatabase) SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error {
	if m.SaveGithubReposFunc != nil {
		return m.SaveGithubReposFunc(ctx, installationID, senderLogin, repos)
	}
	return errors.New("SaveGithubReposFunc not implemented")
}

func (m *MockDatabase) RemoveGithubRepos(ctx context.Context, installationID int, repos []InstalledRepository) error {
	if m.RemoveGithubReposFunc != nil {
		return m.RemoveGithubReposFunc(ctx, installationID, repos)
	}
	return errors.New("RemoveGithubReposFunc not implemented")
}

func (m *MockDatabase) GetGithubRepos(ctx context.Context, email string) ([]Repository, error) {
	if m.GetGithubReposFunc != nil {
		return m.GetGithubReposFunc(ctx, email)
	}
	return nil, errors.New("GetGithubReposFunc not implemented")
}

func (m *MockDatabase) GetInstallationID(ctx context.Context, userID string) (string, int, error) {
	if m.GetInstallationIDFunc != nil {
		return m.GetInstallationIDFunc(ctx, userID)
	}
	return "", 0, errors.New("GetInstallationIDFunc not implemented")
}

func (m *MockDatabase) SaveInstallation(ctx context.Context, userID string, githubID int) (string, error) {
	if m.SaveInstallationFunc != nil {
		return m.SaveInstallationFunc(ctx, userID, githubID)
	}
	return "", errors.New("SaveInstallationFunc not implemented")
}

func (m *MockDatabase) ConnectRepo(ctx context.Context, userID, repoID, branchName string) (Repository, error) {
	if m.ConnectRepoFunc != nil {
		return m.ConnectRepoFunc(ctx, userID, repoID, branchName)
	}
	return Repository{}, errors.New("ConnectRepoFunc not implemented")
}

func (m *MockDatabase) GetRepoByGithub(ctx context.Context, githubRepoID int) (Repository, error) {
	if m.GetRepoByGithubFunc != nil {
		return m.GetRepoByGithubFunc(ctx, githubRepoID)
	}
	return Repository{}, errors.New("GetRepoByGithubFunc not implemented")
}

func (m *MockDatabase) GetRepoByID(ctx context.Context, userID, repoID string) (Repository, error) {
	if m.GetRepoByIDFunc != nil {
		return m.GetRepoByIDFunc(ctx, userID, repoID)
	}
	return Repository{}, errors.New("GetRepoByIDFunc not implemented")
}

func (m *MockDatabase) RepoIsConnected(ctx context.Context, repoID string) (bool, error) {
	if m.RepoIsConnectedFunc != nil {
		return m.RepoIsConnectedFunc(ctx, repoID)
	}
	return false, errors.New("RepoIsConnectedFunc not implemented")
}

func TestHandler_GetDeploymentHistory(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default() // Use default logger for tests

	t.Run("SuccessfulPath", func(t *testing.T) {
		mockDB := &MockDatabase{}
		fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		mockDeployments := []AppDeployment{
			{
				ID:              "dep1",
				RepoID:          "repo1",
				Space:           tqsdk.Space{Name: "App1"},
				Sha:             "sha1",
				BuildTag:        "tag1-message",
				UserDisplayName: "userA",
				Status:          "SUCCESS",
				CreatedAt:       fixedTime,
			},
			{
				ID:              "dep2",
				RepoID:          "repo1",
				Space:           tqsdk.Space{Name: "App2"},
				Sha:             "sha2",
				BuildTag:        "tag2-message",
				UserDisplayName: "userB",
				Status:          "FAILED",
				CreatedAt:       fixedTime.Add(-time.Hour),
			},
		}
		mockDB.GetDeploymentHistoryFunc = func(ctx context.Context, repoID string) ([]AppDeployment, error) {
			assert.Equal(t, "repo1", repoID, "RepoID passed to DB method should match request")
			return mockDeployments, nil
		}

		handler := NewHandler(mockDB, nil, nil, nil, nil, nil, "", nil, nil, "", 0, logger)

		req := GetDeploymentHistoryRequest{RepoID: "repo1"}
		resp, err := handler.GetDeploymentHistory(ctx, req)

		require.Nil(t, err, "Expected no vel.Error")
		require.NotNil(t, resp, "Response should not be nil")
		require.Len(t, resp.History, 2, "Expected two items in history")

		// Verify item 1
		assert.Equal(t, "dep1", resp.History[0].ID)
		assert.Equal(t, "repo1", resp.History[0].RepoID)
		assert.Equal(t, "sha1", resp.History[0].CommitHash)
		assert.Equal(t, "tag1-message", resp.History[0].Message)
		assert.Equal(t, "userA", resp.History[0].UserDisplayName)
		assert.Equal(t, "SUCCESS", resp.History[0].Status)
		assert.Equal(t, fixedTime, resp.History[0].Timestamp)

		// Verify item 2
		assert.Equal(t, "dep2", resp.History[1].ID)
		assert.Equal(t, "repo1", resp.History[1].RepoID)
		assert.Equal(t, "sha2", resp.History[1].CommitHash)
		assert.Equal(t, "tag2-message", resp.History[1].Message)
		assert.Equal(t, "userB", resp.History[1].UserDisplayName)
		assert.Equal(t, "FAILED", resp.History[1].Status)
		assert.Equal(t, fixedTime.Add(-time.Hour), resp.History[1].Timestamp)
	})

	t.Run("EmptyHistory", func(t *testing.T) {
		mockDB := &MockDatabase{}
		mockDB.GetDeploymentHistoryFunc = func(ctx context.Context, repoID string) ([]AppDeployment, error) {
			assert.Equal(t, "empty-repo", repoID)
			return []AppDeployment{}, nil // Return empty slice
		}

		handler := NewHandler(mockDB, nil, nil, nil, nil, nil, "", nil, nil, "", 0, logger)

		req := GetDeploymentHistoryRequest{RepoID: "empty-repo"}
		resp, err := handler.GetDeploymentHistory(ctx, req)

		require.Nil(t, err, "Expected no vel.Error for empty history")
		require.NotNil(t, resp, "Response should not be nil")
		assert.Empty(t, resp.History, "Expected history to be an empty slice")
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mockDB := &MockDatabase{}
		expectedDBError := errors.New("database exploded")
		mockDB.GetDeploymentHistoryFunc = func(ctx context.Context, repoID string) ([]AppDeployment, error) {
			assert.Equal(t, "error-repo", repoID)
			return nil, expectedDBError // Return an error
		}

		handler := NewHandler(mockDB, nil, nil, nil, nil, nil, "", nil, nil, "", 0, logger)

		req := GetDeploymentHistoryRequest{RepoID: "error-repo"}
		resp, velErr := handler.GetDeploymentHistory(ctx, req)

		require.NotNil(t, velErr, "Expected a vel.Error")
		assert.Equal(t, "failed get deployment history", velErr.Message)
		assert.ErrorIs(t, velErr.Err, expectedDBError, "Underlying error should be the one from the DB")
		assert.Empty(t, resp.History, "History should be empty on error")
	})
}

[end of src/domain/getDeploymentHistory_test.go]
