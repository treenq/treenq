package e2e

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treenq/treenq/client"
)

// install repo
// get repos
// add a repo
// get repos to check a new added
// remove repo
// get repos to check a removed
// connect a branch
// get repos to see a branch
// suspend installation
// get repos to check there are none, but installation has a changed status
// delete installation
// get repos to check there are none

// drop a database

// go:embed ../src/domain/testdata/appInstall.json
var appInstallRequestBody []byte

func TestGithubAppInstallation(t *testing.T) {
	clearDatabase()

	// create a user and obtain its token
	user := client.UserInfo{Email: "test@email.com", DisplayName: "testing"}
	userToken, err := createUser(user)
	require.NoError(t, err, "user must be created")

	ctx := context.Background()
	apiClient := client.NewClient("http://localhost:8000", http.DefaultClient, map[string]string{
		"Authorization": "Bearer " + userToken,
	})

	// install app
	var installAppReq client.GithubWebhookRequest
	err = json.Unmarshal(appInstallRequestBody, &installAppReq)
	require.NoError(t, err, "install app request must be unmarshalled")
	err = apiClient.GithubWebhook(ctx, installAppReq)
	require.NoError(t, err, "installation github webhook must proceed")
	// validate the app has been installed and the repos are saved
	reposResponse, err := apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, []client.InstalledRepository{
		{},
	}, reposResponse.Repos, "installed repositories don't match")

	//
}
