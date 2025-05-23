package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dbSchema = `
CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    repoId TEXT,
    space JSON,
    sha TEXT,
    buildTag TEXT,
    userDisplayName TEXT,
    status TEXT,
    createdAt TIMESTAMP,
    updatedAt TIMESTAMP 
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE,
    displayName TEXT,
    createdAt TIMESTAMP,
    updatedAt TIMESTAMP
);

CREATE TABLE IF NOT EXISTS installations (
    id TEXT PRIMARY KEY,
    userId TEXT,
    githubId INTEGER,
    createdAt TIMESTAMP,
    FOREIGN KEY (userId) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS installedRepos (
    id TEXT PRIMARY KEY,
    githubId INTEGER UNIQUE,
    fullName TEXT,
    private BOOLEAN,
    branch TEXT,
    installationId TEXT, 
    userId TEXT,
    status TEXT,
    createdAt TIMESTAMP,
    updatedAt TIMESTAMP,
    FOREIGN KEY (userId) REFERENCES users(id),
    FOREIGN KEY (installationId) REFERENCES installations(id)
);
`
	testDBPath = "./test_store.db"
)

// Helper function to set up the test database and return a *Store
func setupTestDB(t *testing.T) (*Store, func()) {
	t.Helper()

	// Remove existing test DB if it exists
	_ = os.Remove(testDBPath)

	db, err := sqlx.Connect("sqlite3", testDBPath)
	require.NoError(t, err, "Failed to connect to test SQLite database")

	_, err = db.Exec(dbSchema)
	require.NoError(t, err, "Failed to apply schema to test database")

	store, err := NewStore(db)
	require.NoError(t, err, "Failed to create store")

	teardown := func() {
		err := db.Close()
		assert.NoError(t, err, "Failed to close test database")
		err = os.Remove(testDBPath)
		assert.NoError(t, err, "Failed to remove test database file")
	}

	return store, teardown
}

// Helper function to insert a test deployment
func insertTestDeployment(t *testing.T, db *sqlx.DB, dep domain.AppDeployment) domain.AppDeployment {
	t.Helper()

	if dep.ID == "" {
		dep.ID = uuid.NewString()
	}
	if dep.CreatedAt.IsZero() {
		// Ensure CreatedAt is set for ordering tests
		dep.CreatedAt = time.Now().UTC().Round(time.Millisecond)
	}

	spacePayload, err := json.Marshal(dep.Space)
	require.NoError(t, err, "Failed to marshal space to JSON")

	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Question).
		Insert("deployments").
		Columns("id", "repoId", "space", "sha", "buildTag", "userDisplayName", "status", "createdAt").
		Values(dep.ID, dep.RepoID, string(spacePayload), dep.Sha, dep.BuildTag, dep.UserDisplayName, dep.Status, dep.CreatedAt).
		ToSql()
	require.NoError(t, err, "Failed to build insert deployment query")

	_, err = db.Exec(query, args...)
	require.NoError(t, err, "Failed to insert test deployment")

	return dep
}

func TestStore_GetDeploymentHistory(t *testing.T) {
	store, teardown := setupTestDB(t)
	defer teardown()

	ctx := context.Background()

	// Mock time.Now for consistent CreatedAt values
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	now = func() time.Time {
		return fixedTime
	}
	// Restore original now func after test
	defer func() { now = time.Now }()


	t.Run("NoDeployments", func(t *testing.T) {
		repoID := "repo-with-no-deployments"
		history, err := store.GetDeploymentHistory(ctx, repoID)
		require.NoError(t, err)
		assert.Empty(t, history, "Expected no deployments for an empty repo")
	})

	t.Run("SingleDeployment", func(t *testing.T) {
		repoID := "single-dep-repo"
		expectedSpace := tqsdk.Space{Name: "my-app", Config: map[string]interface{}{"key": "value"}}
		depToInsert := domain.AppDeployment{
			RepoID:          repoID,
			Space:           expectedSpace,
			Sha:             "abcdef123456",
			BuildTag:        "v1.0.0",
			UserDisplayName: "test-user",
			Status:          "SUCCESS",
			CreatedAt:       now().Add(-time.Hour), // ensure it's in the past for ordering tests
		}
		insertedDep := insertTestDeployment(t, store.db, depToInsert)

		history, err := store.GetDeploymentHistory(ctx, repoID)
		require.NoError(t, err)
		require.Len(t, history, 1, "Expected one deployment")

		retrievedDep := history[0]
		assert.Equal(t, insertedDep.ID, retrievedDep.ID)
		assert.Equal(t, insertedDep.RepoID, retrievedDep.RepoID)
		assert.Equal(t, insertedDep.Sha, retrievedDep.Sha)
		assert.Equal(t, insertedDep.BuildTag, retrievedDep.BuildTag)
		assert.Equal(t, insertedDep.UserDisplayName, retrievedDep.UserDisplayName)
		assert.Equal(t, insertedDep.Status, retrievedDep.Status)
		assert.WithinDuration(t, insertedDep.CreatedAt, retrievedDep.CreatedAt, time.Millisecond)
		assert.Equal(t, expectedSpace, retrievedDep.Space, "Unmarshalled space should match")
	})

	t.Run("MultipleDeployments", func(t *testing.T) {
		repoID := "multi-dep-repo"
		dep1 := domain.AppDeployment{
			RepoID:          repoID,
			Space:           tqsdk.Space{Name: "app1"},
			Sha:             "sha1",
			BuildTag:        "tag1",
			UserDisplayName: "user1",
			Status:          "SUCCESS",
			CreatedAt:       now().Add(-2 * time.Hour),
		}
		dep2 := domain.AppDeployment{ // Newer
			RepoID:          repoID,
			Space:           tqsdk.Space{Name: "app2"},
			Sha:             "sha2",
			BuildTag:        "tag2",
			UserDisplayName: "user2",
			Status:          "FAILED",
			CreatedAt:       now().Add(-1 * time.Hour),
		}
		dep3 := domain.AppDeployment{ // Oldest
			RepoID:          repoID,
			Space:           tqsdk.Space{Name: "app3"},
			Sha:             "sha3",
			BuildTag:        "tag3",
			UserDisplayName: "user3",
			Status:          "PENDING",
			CreatedAt:       now().Add(-3 * time.Hour),
		}

		insertTestDeployment(t, store.db, dep1)
		insertTestDeployment(t, store.db, dep2)
		insertTestDeployment(t, store.db, dep3)

		expectedOrder := []domain.AppDeployment{dep2, dep1, dep3} // Newest to oldest

		history, err := store.GetDeploymentHistory(ctx, repoID)
		require.NoError(t, err)
		require.Len(t, history, 3, "Expected three deployments")

		for i, expected := range expectedOrder {
			assert.Equal(t, expected.ID, history[i].ID, "Deployment order or content mismatch at index %d", i)
			assert.Equal(t, expected.Space.Name, history[i].Space.Name, "Space mismatch for deployment %s", expected.ID)
		}
	})

	t.Run("DifferentRepos", func(t *testing.T) {
		repoID1 := "diff-repo-1"
		repoID2 := "diff-repo-2"

		depRepo1 := domain.AppDeployment{
			RepoID: repoID1, Space: tqsdk.Space{Name: "app-repo1"}, Sha: "sha-r1", CreatedAt: now().Add(-time.Minute),
		}
		depRepo2 := domain.AppDeployment{
			RepoID: repoID2, Space: tqsdk.Space{Name: "app-repo2"}, Sha: "sha-r2", CreatedAt: now().Add(-time.Minute),
		}

		insertTestDeployment(t, store.db, depRepo1)
		insertTestDeployment(t, store.db, depRepo2)

		// Check repo1
		history1, err1 := store.GetDeploymentHistory(ctx, repoID1)
		require.NoError(t, err1)
		require.Len(t, history1, 1)
		assert.Equal(t, depRepo1.ID, history1[0].ID)
		assert.Equal(t, repoID1, history1[0].RepoID)

		// Check repo2
		history2, err2 := store.GetDeploymentHistory(ctx, repoID2)
		require.NoError(t, err2)
		require.Len(t, history2, 1)
		assert.Equal(t, depRepo2.ID, history2[0].ID)
		assert.Equal(t, repoID2, history2[0].RepoID)
	})

	t.Run("Limit", func(t *testing.T) {
		repoID := "limit-test-repo"
		var expectedLatestIDs []string

		for i := 0; i < 25; i++ {
			dep := domain.AppDeployment{
				RepoID:    repoID,
				Space:     tqsdk.Space{Name: fmt.Sprintf("app-%d", i)},
				Sha:       fmt.Sprintf("sha-%d", i),
				CreatedAt: now().Add(time.Duration(i) * time.Minute), // Ensure distinct CreatedAt for ordering
			}
			insertedDep := insertTestDeployment(t, store.db, dep)
			// The 20 latest will be i=24 down to i=5
			if i >= 5 {
				expectedLatestIDs = append(expectedLatestIDs, insertedDep.ID)
			}
		}
		// Reverse to get descending order (newest first)
		sort.SliceStable(expectedLatestIDs, func(i, j int) bool {
			return expectedLatestIDs[i] > expectedLatestIDs[j] // This sort is actually not needed as we are checking against CreatedAt
		})


		history, err := store.GetDeploymentHistory(ctx, repoID)
		require.NoError(t, err)
		assert.Len(t, history, 20, "Expected to retrieve only 20 deployments due to limit")

		// Verify that the retrieved deployments are indeed the latest 20
		// Sort history by CreatedAt ascending to match insertion order for easier comparison
		sort.SliceStable(history, func(i, j int) bool {
			return history[i].CreatedAt.Before(history[j].CreatedAt)
		})
        // The history slice now contains the 20 latest deployments, from oldest of the latest to newest of the latest.
        // We inserted 25 deployments, with CreatedAt increasing from i=0 to i=24.
        // So the latest 20 are those with i from 5 to 24.
		for i := 0; i < 20; i++ {
			originalIndex := i + 5 // deployments with index 5 to 24
			assert.Equal(t, fmt.Sprintf("sha-%d", originalIndex), history[i].Sha, "Mismatch in limited results order")
		}
	})

	t.Run("CorrectFilteringByRepoId", func(t *testing.T) {
		// This test ensures that the Where clause uses repoId and not id.
		// We create two deployments:
		// 1. A deployment where its `ID` matches `targetRepoID` but `RepoID` is different.
		// 2. A deployment where its `RepoID` matches `targetRepoID`.
		// We expect only the second deployment to be returned.

		targetRepoID := "target-repo-for-filtering-test"
		otherRepoID := "other-repo-filtering"

		// Deployment 1: ID matches targetRepoID, but RepoID does not. Should NOT be returned.
		depWithMatchingID := domain.AppDeployment{
			ID:              targetRepoID, // This is the key: ID is the same as the repo we query for
			RepoID:          otherRepoID,  // But it belongs to another repo
			Space:           tqsdk.Space{Name: "app-matching-id"},
			Sha:             "sha-matching-id",
			UserDisplayName: "user-filter",
			Status:          "SUCCESS",
			CreatedAt:       now().Add(-10 * time.Minute),
		}
		insertTestDeployment(t, store.db, depWithMatchingID)

		// Deployment 2: RepoID matches targetRepoID. Should BE returned.
		depWithMatchingRepoID := domain.AppDeployment{
			ID:              uuid.NewString(), // Different ID
			RepoID:          targetRepoID,   // Belongs to the target repo
			Space:           tqsdk.Space{Name: "app-matching-repo-id"},
			Sha:             "sha-matching-repo-id",
			UserDisplayName: "user-filter",
			Status:          "SUCCESS",
			CreatedAt:       now().Add(-5 * time.Minute),
		}
		insertTestDeployment(t, store.db, depWithMatchingRepoID)

		history, err := store.GetDeploymentHistory(ctx, targetRepoID)
		require.NoError(t, err)
		require.Len(t, history, 1, "Expected only one deployment for the targetRepoID")
		assert.Equal(t, depWithMatchingRepoID.ID, history[0].ID, "The wrong deployment was returned based on ID instead of RepoID")
		assert.Equal(t, targetRepoID, history[0].RepoID)
	})
}

// Minimal mock for sql.Result if needed for other tests, not strictly for GetDeploymentHistory
type mockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (m *mockResult) LastInsertId() (int64, error) { return m.lastInsertID, nil }
func (m *mockResult) RowsAffected() (int64, error) { return m.rowsAffected, nil }

// Minimal mock for sql.Row if needed
type mockRow struct {
	err error
}

func (m *mockRow) Scan(dest ...interface{}) error {
	if m.err != nil {
		return m.err
	}
	// Potentially populate dest with mock data if needed for specific scan tests
	if len(dest) > 0 {
		if idPtr, ok := dest[0].(*string); ok {
			*idPtr = "mock-id"
		}
	}
	return nil
}
func (m *mockRow) Err() error { return m.err }
type mockQuerier struct {
	ExecContextFunc func(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContextFunc func(ctx context.Context, query string, args ...any) *sql.Row
}
func (m *mockQuerier) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
    if m.ExecContextFunc != nil {
        return m.ExecContextFunc(ctx, query, args...)
    }
    return &mockResult{rowsAffected: 1}, nil
}
func (m *mockQuerier) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
    if m.QueryRowContextFunc != nil {
        return m.QueryRowContextFunc(ctx, query, args...)
    }
    // This is tricky because sql.Row is a concrete type, not an interface.
    // For real testing of functions using QueryRowContext, you'd need a live DB
    // or a more elaborate mocking setup for the DB driver itself.
    // For now, returning a row that will produce an error if scanned.
    // Or, ensure tests using this mockQuerier directly set QueryRowContextFunc.
    // For GetUserIDByDisplayName, it expects to scan a string.
    // Let's simulate sql.ErrNoRows for safety in tests that might use this indirectly.
    // This part is not directly used by GetDeploymentHistory but is part of store.go
    // and good to have basic mocks for.
    return &sql.Row{} // This will likely cause `sql: no rows in result set` if Scan is called
}
