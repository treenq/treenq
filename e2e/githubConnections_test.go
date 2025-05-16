package e2e

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"testing"
	"time"

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

	treenqInstallationID := reposResponse.Installation
	require.NotEmpty(t, treenqInstallationID)
	assert.Equal(t, client.GetReposResponse{
		Installation: treenqInstallationID,
		Repos: []client.Repository{
			{
				TreenqID: reposResponse.Repos[0].TreenqID,
				ID:       805585115,
				FullName: "treenq/useless",
				Private:  false,
				Status:   "active",
			},
		},
	}, reposResponse, "installed repositories don't match")

	// add repos
	var addRepoReq client.GithubWebhookRequest
	err = json.Unmarshal(repoAddedRequestBody, &addRepoReq)
	require.NoError(t, err, "add repo request must be unmarshalled")
	err = githubHookClient.GithubWebhook(ctx, addRepoReq)
	require.NoError(t, err, "add repo webhook must be processed")

	// validate the repo has been added
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repos must be available after added repo")

	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.Repository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/useless",
			Private:  false,
			Status:   "active",
			Branch:   "",
		},
		{
			TreenqID: reposResponse.Repos[1].TreenqID,
			ID:       805584540,
			FullName: "treenq/useless-cli",
			Private:  false,
			Status:   "active",
			Branch:   "",
		},
	}}, reposResponse, "installed repositories don't match")

	branchName := "test-branch"
	connectRepoRes, err := apiClient.ConnectRepoBranch(ctx, client.ConnectBranchRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
		Branch: branchName,
	})
	require.NoError(t, err, "connect branch must succeed")
	require.Equal(t, client.ConnectBranchResponse{
		Repo: client.Repository{
			TreenqID:       reposResponse.Repos[0].TreenqID,
			InstallationID: reposResponse.Repos[0].InstallationID,
			ID:             805585115,
			Branch:         branchName,
			FullName:       reposResponse.Repos[0].FullName,
			Private:        reposResponse.Repos[0].Private,
			Status:         reposResponse.Repos[0].Status,
		},
	}, connectRepoRes, "connect repo branch response doesn't match")

	var mergeMainReq client.GithubWebhookRequest
	err = json.Unmarshal(repoBranchMergeMainRequestBody, &mergeMainReq)
	require.NoError(t, err, "merge main request must be unmarshalled")
	err = githubHookClient.GithubWebhook(ctx, mergeMainReq)
	require.NoError(t, err, "merge main webhook must be processed")
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repos must be available after merge main")

	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.Repository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/useless",
			Private:  false,
			Status:   "active",
			Branch:   branchName,
		},
		{
			TreenqID: reposResponse.Repos[1].TreenqID,
			ID:       805584540,
			FullName: "treenq/useless-cli",
			Private:  false,
			Status:   "active",
			Branch:   "",
		},
	}}, reposResponse, "installed repositories don't match")

	// remove a repo
	var removeRepoReq client.GithubWebhookRequest
	err = json.Unmarshal(repoRemoveRequestBody, &removeRepoReq)
	require.NoError(t, err, "remove app request must be unmarshalled")
	err = githubHookClient.GithubWebhook(ctx, removeRepoReq)
	require.NoError(t, err, "remove repo github webhook must processed")
	// validate the repo has been removed
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.Repository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/useless",
			Private:  false,
			Status:   "active",
			Branch:   branchName,
		},
	}}, reposResponse, "installed repositories don't match")

	// test another user can't connect a branch
	connectRepoResponse, err := anotherApiClient.ConnectRepoBranch(ctx, client.ConnectBranchRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
		Branch: reposResponse.Repos[0].Branch,
	})
	assert.Equal(t, err, &client.Error{
		Code: "REPO_NOT_FOUND",
	})
	assert.Equal(t, connectRepoResponse, client.ConnectBranchResponse{})
	// connect a Branch
	connectRepoResponse, err = apiClient.ConnectRepoBranch(ctx, client.ConnectBranchRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
		Branch: reposResponse.Repos[0].Branch,
	})
	assert.NoError(t, err)
	assert.Equal(t, connectRepoResponse, client.ConnectBranchResponse{
		Repo: client.Repository{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/useless",
			Private:  false,
			Status:   "active",
			Branch:   branchName,
		},
	})
	// get repos and make sure there is a connected one
	reposResponse, err = apiClient.GetRepos(ctx)
	require.NoError(t, err, "repositores must be available after app installation")
	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.Repository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/useless",
			Private:  false,
			Status:   "active",
			Branch:   branchName,
		},
	}}, reposResponse, "installed repositories don't match")

	createdDeployment, err := apiClient.Deploy(ctx, client.DeployRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	require.NotEmpty(t, createdDeployment.DeploymentID)
	require.NoError(t, err, "failed to deploys app")

	// wait for deployment done
	for range 20 {
		time.Sleep(time.Second * 2)
		deployment, err := apiClient.GetDeployment(ctx, client.GetDeploymentRequest{
			DeploymentID: createdDeployment.DeploymentID,
		})
		require.NoError(t, err)
		require.NotEqual(t, "failed", deployment.Deployment.Status)
		if deployment.Deployment.Status != "done" {
			continue
		}

		req, err := http.NewRequest("GET", "http://localhost:8000/getBuildProgress?deploymentID="+createdDeployment.DeploymentID, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+userToken)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			var progressMessage client.GetBuildProgressResponse
			err = json.Unmarshal(line, &progressMessage)
			require.NoError(t, err)

			t.Logf("%+v", progressMessage)
			// t.Log(string(line))
		}

		return

	}

	t.Log("status must be done to this moment")
	t.FailNow()
}
