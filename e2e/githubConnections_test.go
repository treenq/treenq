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

	treenqInstallationExists := reposResponse.Installation
	require.True(t, treenqInstallationExists)
	assert.Equal(t, client.GetReposResponse{
		Installation: treenqInstallationExists,
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

	assert.Equal(t, client.GetReposResponse{Installation: true, Repos: []client.GithubRepository{
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
	branchName := "main"
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

	assert.Equal(t, client.GetReposResponse{Installation: true, Repos: []client.GithubRepository{
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
	assert.Equal(t, client.GetReposResponse{Installation: true, Repos: []client.GithubRepository{
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
	assert.Equal(t, client.GetReposResponse{Installation: true, Repos: []client.GithubRepository{
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

	createdDeployment := testDeploymentValidation(t, apiClient, userToken, serviceValidateRequest{
		req: client.DeployRequest{
			RepoID: reposResponse.Repos[0].TreenqID,
		},
		expectedBody: "Hello, main\n",
	})
	assert.Equal(t, createdDeployment.Deployment.RepoID, reposResponse.Repos[0].TreenqID)
	assert.EqualValues(t, createdDeployment.Deployment.Space, client.Space{
		Service: client.Service{
			Name:     "treenq-e2e-sample",
			HttpPort: 8000,
			ReleaseOn: client.ReleaseOn{
				Branch: "main",
			},
			DockerfilePath: "Dockerfile",
			DockerContext:  ".",
			Replicas:       1,
			ComputationResource: client.ComputationResource{
				CpuUnits:   1000,
				MemoryMibs: 2048,
				DiskGibs:   20,
			},
		},
	})
	assert.Len(t, createdDeployment.Deployment.Sha, 40)
	assert.Equal(t, createdDeployment.Deployment.Branch, "main")
	assert.NotEmpty(t, createdDeployment.Deployment.CommitMessage)
	assert.Equal(t, createdDeployment.Deployment.BuildTag, createdDeployment.Deployment.Sha)
	assert.Equal(t, createdDeployment.Deployment.UserDisplayName, "testing")

	branchDeployment := testDeploymentValidation(t, apiClient, userToken, serviceValidateRequest{
		req: client.DeployRequest{
			RepoID: reposResponse.Repos[0].TreenqID,
			Branch: "main2",
		},
		expectedBody: "Hello, main2\n",
	})
	assert.Equal(t, branchDeployment.Deployment.RepoID, reposResponse.Repos[0].TreenqID)
	assert.Len(t, branchDeployment.Deployment.Sha, 40)
	assert.Equal(t, branchDeployment.Deployment.Branch, "main2")
	assert.NotEmpty(t, branchDeployment.Deployment.CommitMessage)
	assert.Equal(t, branchDeployment.Deployment.BuildTag, branchDeployment.Deployment.Sha)
	assert.Equal(t, branchDeployment.Deployment.UserDisplayName, "testing")

	tagDeployment := testDeploymentValidation(t, apiClient, userToken, serviceValidateRequest{
		req: client.DeployRequest{
			RepoID: reposResponse.Repos[0].TreenqID,
			Tag:    "v1.0.0",
		},
		expectedBody: "Hello, tag 1.0.0!\n",
	})
	assert.Equal(t, tagDeployment.Deployment.RepoID, reposResponse.Repos[0].TreenqID)
	assert.Len(t, tagDeployment.Deployment.Sha, 40)
	assert.Equal(t, tagDeployment.Deployment.Branch, "")
	assert.NotEmpty(t, tagDeployment.Deployment.CommitMessage)
	assert.Equal(t, tagDeployment.Deployment.BuildTag, "v1.0.0")
	assert.Equal(t, tagDeployment.Deployment.UserDisplayName, "testing")

	shaDeployment := testDeploymentValidation(t, apiClient, userToken, serviceValidateRequest{
		req: client.DeployRequest{
			RepoID: reposResponse.Repos[0].TreenqID,
			Sha:    "32bddc1b2e31fa3b92963c21e9110f321927a0d3",
		},
		expectedBody: "Hello, 754ca2d77143d30c5c70743b0472dd223ce8212b!\n",
	})
	assert.Equal(t, shaDeployment.Deployment.RepoID, reposResponse.Repos[0].TreenqID)
	assert.Len(t, shaDeployment.Deployment.Sha, 40)
	assert.Equal(t, shaDeployment.Deployment.Branch, "")
	assert.NotEmpty(t, shaDeployment.Deployment.CommitMessage)
	assert.Equal(t, shaDeployment.Deployment.BuildTag, shaDeployment.Deployment.Sha)
	assert.Equal(t, shaDeployment.Deployment.UserDisplayName, "testing")

	rollbackDeploy := testDeploymentValidation(t, apiClient, userToken, serviceValidateRequest{
		req: client.DeployRequest{
			RepoID:           reposResponse.Repos[0].TreenqID,
			FromDeploymentID: createdDeployment.Deployment.ID,
		},
		expectedBody: "Hello, main\n",
	})
	assert.Equal(t, rollbackDeploy.Deployment.RepoID, reposResponse.Repos[0].TreenqID)
	assert.Len(t, rollbackDeploy.Deployment.Sha, 40)
	assert.Equal(t, rollbackDeploy.Deployment.Branch, "main")
	assert.NotEmpty(t, rollbackDeploy.Deployment.CommitMessage)
	assert.Equal(t, rollbackDeploy.Deployment.BuildTag, rollbackDeploy.Deployment.Sha)
	assert.Equal(t, rollbackDeploy.Deployment.UserDisplayName, "testing")

	deployments, err := apiClient.GetDeployments(ctx, client.GetDeploymentsRequest{
		RepoID: reposResponse.Repos[0].TreenqID,
	})
	require.NoError(t, err, "deployments list must be given")

	for _, deployment := range deployments.Deployments {
		assert.EqualValues(t, deployment.RepoID, reposResponse.Repos[0].TreenqID)
		assert.EqualValues(t, deployment.Status, "done")
		assert.NotEmpty(t, deployment.CreatedAt, "createdAt of a deployment must not be empty")
		assert.NotEmpty(t, deployment.UpdatedAt, "updatedAt of a deployment must not be empty")
	}
	assert.Equal(t, deployments.Deployments, []client.AppDeployment{
		rollbackDeploy.Deployment,
		shaDeployment.Deployment,
		tagDeployment.Deployment,
		branchDeployment.Deployment,
		createdDeployment.Deployment,
	})
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

type serviceValidateRequest struct {
	req          client.DeployRequest
	expectedBody string
}

func testDeploymentValidation(
	t *testing.T,
	apiClient *client.Client,
	userToken string,
	testCase serviceValidateRequest,
) client.GetDeploymentResponse {
	ctx := context.Background()

	deployment, err := apiClient.Deploy(ctx, testCase.req)
	require.NoError(t, err, "deployment must succeed")
	require.NotEmpty(t, deployment.Deployment.ID, "deployment ID must not be empty")
	require.Equal(t, deployment.Deployment.Status, "run")
	require.NotEmpty(t, deployment.Deployment.CreatedAt)
	require.NoError(t, err, "failed to deploys app")

	readProgress(t, ctx, deployment, apiClient, userToken)
	validateDeployedServiceResponse(t, testCase.req.RepoID+".localhost", testCase.expectedBody, 200)
	doneDeployment, err := apiClient.GetDeployment(ctx, client.GetDeploymentRequest{
		DeploymentID: deployment.Deployment.ID,
	})
	require.NoError(t, err, "deployment must be found")
	return doneDeployment
}

func validateDeployedServiceResponse(t *testing.T, expectedHost, expectedBody string, expectedStatus int) {
	var lastErr error

	for range 20 {
		time.Sleep(time.Second * 2)

		req, err := http.NewRequest("GET", "http://localhost:8080", nil)
		require.NoError(t, err, "request creation must succeed")
		req.Host = expectedHost

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		defer resp.Body.Close()
		if resp.StatusCode != expectedStatus {
			lastErr = fmt.Errorf("expected status %d, got %d", expectedStatus, resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "body read must succeed")

		if string(body) != expectedBody {
			lastErr = fmt.Errorf("expected body %q, got %q", expectedBody, string(body))
			continue
		}

		return
	}

	require.NoError(t, lastErr, "service validation failed")
}
