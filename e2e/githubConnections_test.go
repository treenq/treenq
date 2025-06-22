package e2e

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treenq/treenq/client"
	"github.com/treenq/treenq/src/domain"
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
	user := client.UserInfo{ID: xid.New().String(), Email: "test@mail.com", DisplayName: "testing"}
	userToken, err := createUser(user)
	require.NoError(t, err, "user must be created")

	anotherUser := client.UserInfo{ID: xid.New().String(), Email: "test2@mail.com", DisplayName: "testing2"}
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
		Repos: []client.GithubRepository{
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

	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.GithubRepository{
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

	// not connected repo can't be deployed
	_, err = apiClient.Deploy(ctx, client.DeployRequest{
		RepoID: reposResponse.Repos[1].TreenqID,
	})
	require.Equal(t, err, &client.Error{
		Code: "REPO_IS_NOT_CONNECTED",
	}, "a not connected repo must not be deployable")

	// connect repo
	branchName := "test-branch"
	connectRepoRes, err := apiClient.ConnectRepoBranch(ctx, client.ConnectBranchRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
		Branch: branchName,
	})
	require.NoError(t, err, "connect branch must succeed")
	require.Equal(t, client.ConnectBranchResponse{
		Repo: client.GithubRepository{
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

	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.GithubRepository{
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
	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.GithubRepository{
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
		Repo: client.GithubRepository{
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
	assert.Equal(t, client.GetReposResponse{Installation: treenqInstallationID, Repos: []client.GithubRepository{
		{
			TreenqID: reposResponse.Repos[0].TreenqID,
			ID:       805585115,
			FullName: "treenq/useless",
			Private:  false,
			Status:   "active",
			Branch:   branchName,
		},
	}}, reposResponse, "installed repositories don't match")

	t.Run("test secrets api", func(t *testing.T) {
		t.Parallel()
		testSecretsApi(t, ctx, apiClient, anotherApiClient, connectRepoRes)
	})

	createdDeployment, err := apiClient.Deploy(ctx, client.DeployRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	require.NotEmpty(t, createdDeployment.Deployment.ID)
	require.Equal(t, createdDeployment.Deployment.Status, "run")
	require.NotEmpty(t, createdDeployment.Deployment.CreatedAt)
	require.NoError(t, err, "failed to deploys app")

	// wait for deployment done
	t.Run("read deployment progress", func(t *testing.T) {
		readProgress(t, ctx, createdDeployment, apiClient, userToken)
	})
	t.Run("test qwer.localhost deployed", func(t *testing.T) {
		t.Parallel()
		var qwerErr error
		for range 20 {
			time.Sleep(time.Second * 2)
			req, err := http.NewRequest("GET", "http://localhost:8080", nil)
			require.NoError(t, err, "request for kube:80 must be created")
			req.Host = "qwer.localhost"
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				qwerErr = err
				continue
			}
			if resp.StatusCode != 200 {
				qwerErr = fmt.Errorf("status code is %d", resp.StatusCode)
				continue
			}
			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "body must be read from qwer.localhost")
			resp.Body.Close()
			if string(b) != "Hello, World!\n" {
				qwerErr = fmt.Errorf("body expected to have Hello, World, given: %s", string(b))
				continue
			}

			break

		}
		assert.NoError(t, qwerErr, "failed to get qwer.localhost")
	})

	deployments, err := apiClient.GetDeployments(ctx, client.GetDeploymentsRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	require.Len(t, deployments.Deployments, 1, "1 item expected in history deployment after first successful")
	require.NoError(t, err, "no error expected on deployment history")
	assert.Equal(t, reposResponse.Repos[0].TreenqID, deployments.Deployments[0].RepoID)
	assert.Equal(t, createdDeployment.Deployment.ID, deployments.Deployments[0].ID)
	assert.NotEmpty(t, deployments.Deployments[0].BuildTag)
	assert.NotEmpty(t, deployments.Deployments[0].Sha)
	assert.Equal(t, branchName, deployments.Deployments[0].Branch)
	assert.NotEmpty(t, deployments.Deployments[0].CommitMessage)
	assert.Equal(t, user.DisplayName, deployments.Deployments[0].UserDisplayName)
	assert.EqualValues(t, domain.DeployStatusDone, deployments.Deployments[0].Status)

	rollbackDeploy, err := apiClient.Deploy(ctx, client.DeployRequest{
		RepoID:           reposResponse.Repos[0].TreenqID,
		FromDeploymentID: deployments.Deployments[0].ID,
	})
	require.NotEqual(t, rollbackDeploy.Deployment.ID, createdDeployment.Deployment.ID, "rollback id must not be the same")
	require.NotEmpty(t, rollbackDeploy.Deployment.ID)
	require.NotEmpty(t, rollbackDeploy.Deployment.CreatedAt)
	require.NoError(t, err, "failed to deploys app")

	deployments, err = apiClient.GetDeployments(ctx, client.GetDeploymentsRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	require.Len(t, deployments.Deployments, 2, "1 item expected in history deployment after first successful")
	require.NoError(t, err, "no error expected on deployment history")
	assert.Equal(t, reposResponse.Repos[0].TreenqID, deployments.Deployments[0].RepoID)
	assert.Equal(t, rollbackDeploy.Deployment.ID, deployments.Deployments[0].ID)
	assert.NotEmpty(t, deployments.Deployments[0].BuildTag)
	assert.NotEmpty(t, deployments.Deployments[0].Sha)
	assert.Equal(t, branchName, deployments.Deployments[0].Branch)
	assert.NotEmpty(t, deployments.Deployments[0].CommitMessage)
	assert.Equal(t, user.DisplayName, deployments.Deployments[0].UserDisplayName)
	t.Run("read deployment progress", func(t *testing.T) {
		t.Parallel()
		readProgress(t, ctx, createdDeployment, apiClient, userToken)
	})
	readProgress(t, ctx, rollbackDeploy, apiClient, userToken)
}

func testSecretsApi(t *testing.T, ctx context.Context, apiClient, anotherApiClient *client.Client, connectRepoRes client.ConnectBranchResponse) {
	secrets, err := apiClient.GetSecrets(ctx, client.GetSecretsRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
	})
	require.NoError(t, err, "no error expected on empty secrets list")
	require.Empty(t, secrets.Keys, "secrets are expected to be empty")

	revealSecretResponse, err := apiClient.RevealSecret(ctx, client.RevealSecretRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
		Key:    "SECRET",
	})
	require.Equal(t, err, &client.Error{Code: "SECRET_DOESNT_EXIST"})
	require.Empty(t, revealSecretResponse.Value, "no revealed secret is expected")

	err = apiClient.SetSecret(ctx, client.SetSecretRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
		Key:    "SECRET",
		Value:  "SUPER",
	})
	require.NoError(t, err, "no error expect on set secret")

	secrets, err = apiClient.GetSecrets(ctx, client.GetSecretsRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
	})
	require.NoError(t, err, "no error expected on empty secrets list")
	require.Equal(t, []string{"SECRET"}, secrets.Keys, "secrets are expected to be empty")

	revealSecretResponse, err = apiClient.RevealSecret(ctx, client.RevealSecretRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
		Key:    "SECRET",
	})
	require.NoError(t, err, "must reveal a secret successfully")
	require.Equal(t, "SUPER", revealSecretResponse.Value, "no revealed secret is expected")

	revealSecretResponse, err = apiClient.RevealSecret(ctx, client.RevealSecretRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
		Key:    "SECRET2",
	})
	require.Equal(t, err, &client.Error{Code: "SECRET_DOESNT_EXIST"})
	require.Empty(t, revealSecretResponse.Value, "no revealed secret is expected")

	secrets, err = anotherApiClient.GetSecrets(ctx, client.GetSecretsRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
	})
	require.NoError(t, err, "no error expected on empty secrets list")
	require.Empty(t, secrets.Keys, "secrets are expected to be empty")

	// secrets must not be available to the other users
	require.NoError(t, err, "secret must be removed successfully")
	secrets, err = anotherApiClient.GetSecrets(ctx, client.GetSecretsRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
	})
	require.NoError(t, err, "no error expected on empty secrets list")
	require.Empty(t, secrets.Keys, "secrets are expected to be empty")

	revealSecretResponse, err = anotherApiClient.RevealSecret(ctx, client.RevealSecretRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
		Key:    "SECRET",
	})
	require.Equal(t, err, &client.Error{Code: "SECRET_DOESNT_EXIST"})
	require.Empty(t, revealSecretResponse.Value, "no revealed secret is expected")

	err = apiClient.RemoveSecret(ctx, client.RemoveSecretRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
		Key:    "SECRET",
	})
	require.NoError(t, err, "secret must be removed successfully")
	secrets, err = apiClient.GetSecrets(ctx, client.GetSecretsRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
	})
	require.NoError(t, err, "no error expected on empty secrets list")
	require.Empty(t, secrets.Keys, "secrets are expected to be empty")

	revealSecretResponse, err = apiClient.RevealSecret(ctx, client.RevealSecretRequest{
		RepoID: connectRepoRes.Repo.TreenqID,
		Key:    "SECRET",
	})
	require.Equal(t, err, &client.Error{Code: "SECRET_DOESNT_EXIST"})
	require.Empty(t, revealSecretResponse.Value, "no revealed secret is expected")
}

func readProgress(t *testing.T, ctx context.Context, createdDeployment client.GetDeploymentResponse, apiClient *client.Client, userToken string) {
	progressRead := false
	for range 20 {
		time.Sleep(time.Second * 2)
		deployment, err := apiClient.GetDeployment(ctx, client.GetDeploymentRequest{
			DeploymentID: createdDeployment.Deployment.ID,
		})
		require.NoError(t, err)
		require.NotEqual(t, "failed", deployment.Deployment.Status)
		if deployment.Deployment.Status != "done" {
			continue
		}

		req, err := http.NewRequest("GET", "http://localhost:8000/getBuildProgress?deploymentID="+createdDeployment.Deployment.ID, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+userToken)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		hasFinalMessage := false
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var progressMessage client.GetBuildProgressResponse
			t.Log(line)
			line = strings.TrimPrefix(line, "data: ")
			err = json.Unmarshal([]byte(line), &progressMessage)
			require.NoError(t, err)

			hasFinalMessage = progressMessage.Message.Final
			if hasFinalMessage {
				break
			}

			require.NotEmpty(t, progressMessage.Message.Timestamp)
			require.NotEmpty(t, progressMessage.Message.Level.String())
			require.NotEmpty(t, progressMessage.Message.Payload)
		}

		assert.True(t, hasFinalMessage, "progress build must have a final message")
		progressRead = true
		break
	}

	require.True(t, progressRead, "progress must be read")
}

func testServiceResponds(t *testing.T) {
}
