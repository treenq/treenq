package e2e

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treenq/treenq/client"
)

// connect a branch
// get repos to see a branch
// suspend installation
// get repos to check there are none, but installation has a changed status
// delete installation
// get repos to check there are none

// drop a database

//go:embed testdata/appInstall.json
var appInstallRequestBody []byte

//go:embed testdata/repoAdded.json
var repoAddedRequestBody []byte

//go:embed testdata/repoRemoved.json
var repoRemoveRequestBody []byte

func TestGithubAppInstallation(t *testing.T) {
	clearDatabase()

	// create a user and obtain its token
	user := client.UserInfo{ID: uuid.NewString(), Email: "test@mail.com", DisplayName: "testing"}
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
	require.NoError(t, err, "installation github webhook must processed")
	// validate the app has been installed and the repos are saved
	reposResponse, err := apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, []client.InstalledRepository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/treenq",
			Private:  false,
			Branch:   "",
		},
	}, reposResponse.Repos, "installed repositories don't match")

	// add repos
	var addRepoReq client.GithubWebhookRequest
	err = json.Unmarshal(repoAddedRequestBody, &addRepoReq)
	require.NoError(t, err, "add repo request must be unmarshalled")
	err = apiClient.GithubWebhook(ctx, addRepoReq)
	require.NoError(t, err, "add repo webhook must be processed")
	// validate the repo has been added
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repos must be available after added repo")
	assert.Equal(t, []client.InstalledRepository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/treenq",
			Private:  false,
			Branch:   "",
		},
		{
			TreenqID: reposResponse.Repos[1].TreenqID,
			ID:       805584540,
			FullName: "treenq/treenq-cli",
			Private:  false,
			Branch:   "",
		},
	}, reposResponse.Repos, "installed repositories don't match")

	// remove a repo
	var removeRepoReq client.GithubWebhookRequest
	err = json.Unmarshal(repoRemoveRequestBody, &removeRepoReq)
	require.NoError(t, err, "remove app request must be unmarshalled")
	err = apiClient.GithubWebhook(ctx, removeRepoReq)
	require.NoError(t, err, "remove repo github webhook must processed")
	// validate the repo has been removed
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, []client.InstalledRepository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/treenq",
			Private:  false,
			Branch:   "",
		},
	}, reposResponse.Repos, "installed repositories don't match")
}
