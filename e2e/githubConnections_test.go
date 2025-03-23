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

//go:embed testdata/appInstall.json
var appInstallRequestBody []byte

//go:embed testdata/repoAdded.json
var repoAddedRequestBody []byte

//go:embed testdata/repoRemoved.json
var repoRemoveRequestBody []byte

//go:embed testdata/branchMergeMain.json
var repoBranchMergeMainRequestBody []byte

func TestGithubAppInstallation(t *testing.T) {
	clearDatabase()

	// create a user and obtain its token
	user := client.UserInfo{ID: uuid.NewString(), Email: "test@mail.com", DisplayName: "testing"}
	userToken, err := createUser(user)
	require.NoError(t, err, "user must be created")

	anotherUser := client.UserInfo{ID: uuid.NewString(), Email: "test2@mail.com", DisplayName: "testing2"}
	anotherToken, err := createUser(anotherUser)
	require.NoError(t, err, "user must be created")

	ctx := context.Background()
	githubHookClient := client.NewClient("http://localhost:8000", http.DefaultClient, map[string]string{})
	apiClient := client.NewClient("http://localhost:8000", http.DefaultClient, map[string]string{
		"Authorization": "Bearer " + userToken,
	})
	anotherApiClient := client.NewClient("http://localhost:8000", http.DefaultClient, map[string]string{
		"Authorization": "Bearer " + anotherToken,
	})

	// install app
	var installAppReq client.GithubWebhookRequest
	err = json.Unmarshal(appInstallRequestBody, &installAppReq)
	require.NoError(t, err, "install app request must be unmarshalled")
	err = githubHookClient.GithubWebhook(ctx, installAppReq)
	require.NoError(t, err, "installation github webhook must processed")
	// validate the app has been installed and the repos are saved
	reposResponse, err := apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, []client.Repository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/useless",
			Private:  false,
			Status:   "active",
		},
	}, reposResponse.Repos, "installed repositories don't match")

	// add repos
	var addRepoReq client.GithubWebhookRequest
	err = json.Unmarshal(repoAddedRequestBody, &addRepoReq)
	require.NoError(t, err, "add repo request must be unmarshalled")
	err = githubHookClient.GithubWebhook(ctx, addRepoReq)
	require.NoError(t, err, "add repo webhook must be processed")
	// validate the repo has been added
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repos must be available after added repo")
	assert.Equal(t, []client.Repository{
		{
			TreenqID:      reposResponse.Repos[0].TreenqID,
			ID:            805585115,
			FullName:      "treenq/useless",
			Private:       false,
			Status:        "active",
			DefaultBranch: "",
		},
		{
			TreenqID:      reposResponse.Repos[1].TreenqID,
			ID:            805584540,
			FullName:      "treenq/useless-cli",
			Private:       false,
			Status:        "active",
			DefaultBranch: "",
		},
	}, reposResponse.Repos, "installed repositories don't match")

	var mergeMainReq client.GithubWebhookRequest
	err = json.Unmarshal(repoBranchMergeMainRequestBody, &mergeMainReq)
	require.NoError(t, err, "merge main request must be unmarshalled")
	err = githubHookClient.GithubWebhook(ctx, mergeMainReq)
	require.NoError(t, err, "merge main webhook must be processed")
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repos must be available after merge main")

	assert.Equal(t, []client.Repository{
		{
			TreenqID:      reposResponse.Repos[0].TreenqID,
			ID:            805585115,
			FullName:      "treenq/useless",
			Private:       false,
			Status:        "active",
			DefaultBranch: "main",
		},
		{
			TreenqID:      reposResponse.Repos[1].TreenqID,
			ID:            805584540,
			FullName:      "treenq/useless-cli",
			Private:       false,
			Status:        "active",
			DefaultBranch: "",
		},
	}, reposResponse.Repos, "installed repositories don't match")

	// remove a repo
	var removeRepoReq client.GithubWebhookRequest
	err = json.Unmarshal(repoRemoveRequestBody, &removeRepoReq)
	require.NoError(t, err, "remove app request must be unmarshalled")
	err = githubHookClient.GithubWebhook(ctx, removeRepoReq)
	require.NoError(t, err, "remove repo github webhook must processed")
	// validate the repo has been removed
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, []client.Repository{
		{
			TreenqID:      reposResponse.Repos[0].TreenqID,
			ID:            805585115,
			FullName:      "treenq/useless",
			Private:       false,
			Status:        "active",
			DefaultBranch: "main",
		},
	}, reposResponse.Repos, "installed repositories don't match")

	// test another user can't connect a branch
	connectRepoResponse, err := anotherApiClient.ConnectRepoBranch(ctx, client.ConnectBranchRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	assert.Equal(t, err, &client.Error{
		Code: "REPO_NOT_FOUND",
	})
	assert.Equal(t, connectRepoResponse, client.ConnectBranchResponse{})
	// connect a Branch
	connectRepoResponse, err = apiClient.ConnectRepoBranch(ctx, client.ConnectBranchRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	assert.NoError(t, err)
	assert.Equal(t, connectRepoResponse, client.ConnectBranchResponse{
		Repo: client.Repository{
			TreenqID:      reposResponse.Repos[0].TreenqID,
			ID:            805585115,
			FullName:      "treenq/useless",
			Private:       false,
			Status:        "active",
			Connected:     true,
			DefaultBranch: "main",
		},
	})
	// get repos and make sure there is a connected one
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, []client.Repository{
		{
			TreenqID:      reposResponse.Repos[0].TreenqID,
			ID:            805585115,
			FullName:      "treenq/useless",
			Private:       false,
			Status:        "active",
			Connected:     true,
			DefaultBranch: "main",
		},
	}, reposResponse.Repos, "installed repositories don't match")

	err = apiClient.Deploy(ctx, client.DeployRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	require.NoError(t, err, "failed to deploys app")
}
