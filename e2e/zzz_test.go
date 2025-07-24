package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/treenq/treenq/client"
	"github.com/treenq/treenq/pkg/auth"
	"github.com/treenq/treenq/src/resources"
)

var (
	db         *sqlx.DB
	tableNames = []string{
		"deployments",
		"secrets",
		"spaces",
		"installedRepos",
		"installations",
		"workspaceUsers",
		"users",
		"workspaces",
	}
	privateKey = `-----BEGIN PRIVATE KEY-----
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
9Dq6z1VElaxFL5Hrz8GLh7537KKOZg==
-----END PRIVATE KEY-----
`
	publicKey = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAvwbQLKdgUAQgqlY63UzB
E5uilzN5wvJQ3yqfjuQrKLbim6f5Zd9EbuObPY/f36SYKF6oD5fhWNLpJakrueOX
A6GHk/3F2jjsNUe0ySWEXZcRJOhMtHf1eALbwqSrg4LUoekvnw86vokBFrYKpp//
7D6gKGyVuROaWrhWKEj8iAhyefi6DwbH092ZNm5WL9l/fa1wfbb3jo1MqZj6YmmL
RTEtUX4q/C9Vu+BVbvwQ8CouEXFljW/Y16VM65mJyirI5LcGu55Sm/JgGpsafvGY
iMgZhigQtyaTqAhwBE7GaPq77EPY3o1HQ5ua/PPkRucIOErklacXfb0odf7xKeRb
o/rHiHoJXreBSmuMEAmRHchdOJEJyfuTig/pCvpubE1/AooRomJUiJH4DlGWsOMJ
p8u+FhCgaH4ujOGA0jQfmNV733e+gkBbskY+PnKyStLU9i333nDsVmqkHRE9huic
5nVhZavMOcF0xUr7WZ//ZLz/PL7xqC7/0aLynapKu6XN/ioR7UkZLWGi8Sa4MY/0
E1xMb1fflBCwOSJ8TJb4LgLiCyr5gJBy17O2AqDpQJfSdzafCjb8wMQWagVcK6+p
QcUxsLirFRIbYzTUu3Czn6vFCJhnHQC364//Y5mCcon3dbcs5ooB32v+Mlt2guBa
0p1+Hpzq+cBcwHuA0xiXTPUCAwEAAQ==
-----END PUBLIC KEY-----
`
)

func openDB() {
	var err error
	db, err = resources.OpenDB("postgres://postgres@localhost:5432/tq?sslmode=disable", "../migrations")
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	openDB()

	m.Run()
}

// clearDatabase truncates all tables in the database to prepare it for the next test suite
func clearDatabase() {
	// If we have specific table names, truncate them one by one
	if len(tableNames) > 0 {
		for _, tableName := range tableNames {
			query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName)
			if _, err := db.Exec(query); err != nil {
				err = fmt.Errorf("failed to truncate table %s: %w", tableName, err)
				panic(err)
			}
		}
	}
}

func createUser(userInfo client.UserInfo) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert user
	_, err = tx.Exec(`
		INSERT INTO users (id, email, displayName)
		VALUES ($1, $2, $3)
	`, userInfo.ID, userInfo.Email, userInfo.DisplayName)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	// Create a default workspace for the user
	workspaceID := userInfo.ID
	_, err = tx.Exec(`
		INSERT INTO workspaces (id, name, githubOrgName)
		VALUES ($1, $2, $3)
	`, workspaceID, "default", "")
	if err != nil {
		return "", fmt.Errorf("failed to create workspace: %w", err)
	}

	// Link user to workspace
	_, err = tx.Exec(`
		INSERT INTO workspaceUsers (workspaceId, userId, role)
		VALUES ($1, $2, 'owner')
	`, workspaceID, userInfo.ID)
	if err != nil {
		return "", fmt.Errorf("failed to link user to workspace: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	issuer := auth.NewJwtIssuer("treenq-api", []byte(privateKey), []byte(publicKey), 24*time.Hour)
	token, err := issuer.GenerateJwtToken(map[string]any{
		"id":          userInfo.ID,
		"email":       userInfo.Email,
		"displayName": userInfo.DisplayName,
		"workspaces":  []string{workspaceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
