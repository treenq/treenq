package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver for sqlx
)

var testStore *Store

// setupTestDB initializes a connection to the test database.
// It expects the TEST_DB_DSN environment variable to be set.
// Example DSN: "postgres://user:password@localhost:5432/test_db_name?sslmode=disable"
func setupTestDB(t *testing.T) *sqlx.DB {
	dbDSN := os.Getenv("TEST_DB_DSN")
	if dbDSN == "" {
		t.Skip("TEST_DB_DSN not set, skipping repo integration tests")
	}

	db, err := sqlx.Connect("postgres", dbDSN)
	if err != nil {
		t.Fatalf("Failed to connect to test DB using DSN '%s': %v", dbDSN, err)
	}

	// Optional: Ping to ensure connection is live
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test DB: %v", err)
	}

	return db
}

// createTestUser inserts a dummy user into the 'users' table and returns the user's ID.
// This is necessary to satisfy foreign key constraints for 'user_pats.user_id'.
func createTestUser(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	var userID string
	// Using unique email and displayName for each test user to avoid conflicts
	// if tests run in parallel or if table is not cleared perfectly.
	email := fmt.Sprintf("testuser_%s@example.com", uuid.NewString())
	displayName := fmt.Sprintf("Test User %s", uuid.NewString())

	query := "INSERT INTO users (id, email, displayName) VALUES ($1, $2, $3) RETURNING id"
	newUserID := uuid.NewString() // Generate UUID for the user

	err := db.QueryRowxContext(context.Background(), query, newUserID, email, displayName).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	if userID != newUserID {
		t.Fatalf("Returned user ID '%s' does not match generated user ID '%s'", userID, newUserID)
	}
	return userID
}

// clearTables removes all records from user_pats.
// It's called before each test that modifies these tables to ensure a clean state.
// Specific users are deleted by tests that create them.
func clearUserPatsTable(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), "DELETE FROM user_pats")
	if err != nil {
		t.Fatalf("Failed to clear user_pats table: %v", err)
	}
}

func TestMain(m *testing.M) {
	dbDSN := os.Getenv("TEST_DB_DSN")
	if dbDSN == "" {
		fmt.Println("TEST_DB_DSN not set, skipping repo integration tests.")
		os.Exit(0) // Exit successfully, indicating tests were skipped
	}

	db := sqlx.MustConnect("postgres", dbDSN) // Panics on connection failure
	defer db.Close()

	var err error
	testStore, err = NewStore(db) // NewStore uses squirrel which is fine
	if err != nil {
		fmt.Printf("Failed to create store for tests: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// TestStore_CreateAndGetUserPAT tests creating a PAT and then retrieving it.
func TestStore_CreateAndGetUserPAT(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db) // Clear PATs table first
	testUserID := createTestUser(t, db)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", testUserID)

	encryptedPAT := "encrypted_pat_for_test_1"

	err := testStore.CreateUserPAT(ctx, testUserID, encryptedPAT)
	if err != nil {
		t.Fatalf("CreateUserPAT failed: %v", err)
	}

	retrievedPAT, err := testStore.GetUserPATByUserID(ctx, testUserID)
	if err != nil {
		t.Fatalf("GetUserPATByUserID failed: %v", err)
	}
	if retrievedPAT != encryptedPAT {
		t.Errorf("Retrieved PAT mismatch: got '%s', want '%s'", retrievedPAT, encryptedPAT)
	}
}

// TestStore_CreateUserPAT_Duplicate tests the UNIQUE constraint on user_id.
func TestStore_CreateUserPAT_Duplicate(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db)
	testUserID := createTestUser(t, db)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", testUserID)

	encryptedPAT1 := "first_encrypted_pat_for_duplicate_test"
	err := testStore.CreateUserPAT(ctx, testUserID, encryptedPAT1)
	if err != nil {
		t.Fatalf("First CreateUserPAT failed: %v", err)
	}

	encryptedPAT2 := "second_encrypted_pat_for_duplicate_test"
	err = testStore.CreateUserPAT(ctx, testUserID, encryptedPAT2)
	if err == nil {
		t.Fatal("Second CreateUserPAT should have failed due to UNIQUE constraint, but it succeeded.")
	}

	// Check if the error is a unique constraint violation
	// Error messages can be database-specific. For PostgreSQL, it includes "duplicate key value violates unique constraint".
	// Example: "pq: duplicate key value violates unique constraint \"user_pats_user_id_key\""
	if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
		!strings.Contains(err.Error(), "user_pats_user_id_key") { // Check for constraint name too
		t.Errorf("Expected a unique constraint violation error, but got: %v", err)
	}

	// Verify that the first PAT is still the one stored
	retrievedPAT, getErr := testStore.GetUserPATByUserID(ctx, testUserID)
	if getErr != nil {
		t.Fatalf("GetUserPATByUserID after duplicate attempt failed: %v", getErr)
	}
	if retrievedPAT != encryptedPAT1 {
		t.Errorf("PAT was overwritten or changed after duplicate attempt. Got '%s', want '%s'", retrievedPAT, encryptedPAT1)
	}
}

// TestStore_GetUserPAT_NotFound tests retrieving a PAT for a non-existent user ID.
func TestStore_GetUserPAT_NotFound(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db)
	// DO NOT create a user here, or create one and use a different ID for query

	nonExistentUserID := uuid.NewString() // A user ID that certainly does not exist

	_, err := testStore.GetUserPATByUserID(ctx, nonExistentUserID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Expected sql.ErrNoRows for non-existent user ID, but got: %v", err)
	}
}

// TestStore_DeleteUserPAT tests deleting an existing PAT.
func TestStore_DeleteUserPAT(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db)
	testUserID := createTestUser(t, db)
	// No defer for user deletion here, as we want to see PAT deleted for an existing user

	encryptedPAT := "pat_to_be_deleted"
	err := testStore.CreateUserPAT(ctx, testUserID, encryptedPAT)
	if err != nil {
		t.Fatalf("CreateUserPAT for deletion test failed: %v", err)
	}

	err = testStore.DeleteUserPAT(ctx, testUserID)
	if err != nil {
		t.Fatalf("DeleteUserPAT failed: %v", err)
	}

	_, err = testStore.GetUserPATByUserID(ctx, testUserID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Expected sql.ErrNoRows after deleting PAT, but got: %v", err)
	}

	// Clean up the user now
	_, delUserErr := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", testUserID)
	if delUserErr != nil {
		t.Errorf("Failed to delete test user %s after test: %v", testUserID, delUserErr)
	}
}

// TestStore_DeleteUserPAT_NotFound tests deleting a PAT for a non-existent user ID.
func TestStore_DeleteUserPAT_NotFound(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db)

	nonExistentUserID := uuid.NewString()

	err := testStore.DeleteUserPAT(ctx, nonExistentUserID)
	// As per store.go, DeleteUserPAT returns sql.ErrNoRows if RowsAffected is 0
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Expected sql.ErrNoRows when deleting PAT for non-existent user, but got: %v", err)
	}
}

// TestStore_GetAllUserPATs tests retrieving all PATs.
func TestStore_GetAllUserPATs(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db) // Ensure table is empty at the start

	// 1. Test with no PATs
	allPATs, err := testStore.GetAllUserPATs(ctx)
	if err != nil {
		t.Fatalf("GetAllUserPATs failed when table is empty: %v", err)
	}
	if len(allPATs) != 0 {
		t.Errorf("Expected 0 PATs when table is empty, got %d", len(allPATs))
	}

	// 2. Create some users and PATs
	user1ID := createTestUser(t, db)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user1ID)
	pat1 := "all_pats_test_pat_1"
	err = testStore.CreateUserPAT(ctx, user1ID, pat1)
	if err != nil {
		t.Fatalf("CreateUserPAT for user1 failed: %v", err)
	}

	user2ID := createTestUser(t, db)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user2ID)
	pat2 := "all_pats_test_pat_2"
	err = testStore.CreateUserPAT(ctx, user2ID, pat2)
	if err != nil {
		t.Fatalf("CreateUserPAT for user2 failed: %v", err)
	}

	// 3. Retrieve all PATs and verify
	allPATs, err = testStore.GetAllUserPATs(ctx)
	if err != nil {
		t.Fatalf("GetAllUserPATs failed with data: %v", err)
	}
	if len(allPATs) != 2 {
		t.Errorf("Expected 2 PATs, got %d. Retrieved: %v", len(allPATs), allPATs)
	}

	// Sort for consistent comparison as order isn't guaranteed by query
	sort.Strings(allPATs)
	expectedPATs := []string{pat1, pat2}
	sort.Strings(expectedPATs)

	if allPATs[0] != expectedPATs[0] || allPATs[1] != expectedPATs[1] {
		t.Errorf("Retrieved PATs mismatch. Got %v, want %v", allPATs, expectedPATs)
	}

	// 4. Test after deleting one PAT
	err = testStore.DeleteUserPAT(ctx, user1ID)
	if err != nil {
		t.Fatalf("DeleteUserPAT for user1 failed: %v", err)
	}
	allPATs, err = testStore.GetAllUserPATs(ctx)
	if err != nil {
		t.Fatalf("GetAllUserPATs after delete failed: %v", err)
	}
	if len(allPATs) != 1 {
		t.Errorf("Expected 1 PAT after deletion, got %d", len(allPATs))
	}
	if allPATs[0] != pat2 {
		t.Errorf("Expected remaining PAT to be '%s', got '%s'", pat2, allPATs[0])
	}
}

// TestStore_CreateUserPAT_EmptyUserID tests creating a PAT with an empty user ID.
// This should ideally be caught by DB constraints if user_id is NOT NULL,
// or by application logic if appropriate. The current CreateUserPAT in store.go
// doesn't explicitly check for empty userID before querying.
// PostgreSQL will give "null value in column "user_id" violates not-null constraint"
func TestStore_CreateUserPAT_EmptyUserID(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db)
	// We don't create a user, and pass an empty string as userID

	err := testStore.CreateUserPAT(ctx, "", "pat_for_empty_userid")
	if err == nil {
		t.Fatal("CreateUserPAT with empty userID should have failed, but it succeeded.")
	}
	// Check for a not-null violation error (specific message might vary by DB)
	// For PostgreSQL: "null value in column \"user_id\" of relation \"user_pats\" violates not-null constraint"
	// The squirrel builder or DB driver might also return errors for invalid inputs before hitting DB.
	if !strings.Contains(err.Error(), "null value in column") && !strings.Contains(err.Error(), "violates not-null constraint") {
		t.Logf("Warning: CreateUserPAT with empty userID returned an error, but not the expected not-null violation. Error: %v", err)
		// This might still be acceptable if the error is sensible, e.g. from the query builder for an invalid parameter.
	}
}

// TestStore_CreateUserPAT_NonExistentUserID tests creating a PAT with a non-existent user ID.
// This should fail due to foreign key constraint on user_pats.user_id referencing users.id.
func TestStore_CreateUserPAT_NonExistentUserID(t *testing.T) {
	if testStore == nil {
		t.Skip("testStore not initialized")
	}
	ctx := context.Background()
	db := testStore.db
	clearUserPatsTable(t, db)

	nonExistentUserID := uuid.NewString() // This user is not created in the 'users' table

	err := testStore.CreateUserPAT(ctx, nonExistentUserID, "pat_for_non_existent_user")
	if err == nil {
		t.Fatal("CreateUserPAT with non-existent userID should have failed due to FK constraint, but it succeeded.")
	}
	// For PostgreSQL, the error is typically:
	// "pq: insert or update on table "user_pats" violates foreign key constraint "user_pats_user_id_fkey""
	if !strings.Contains(err.Error(), "violates foreign key constraint") &&
		!strings.Contains(err.Error(), "user_pats_user_id_fkey") { // Check for constraint name too
		t.Errorf("Expected a foreign key constraint violation error, but got: %v", err)
	}
}
```
