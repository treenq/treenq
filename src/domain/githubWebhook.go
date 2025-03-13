package domain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/pkg/vel"
)

type GithubWebhookRequest struct {
	// After holds a latest commit SHA
	After        string       `json:"after"`
	Installation Installation `json:"installation"`
	Sender       Sender       `json:"sender"`

	// installation only fields
	Action              string                `json:"action"`
	Repositories        []InstalledRepository `json:"repositories"`
	RepositoriesAdded   []InstalledRepository `json:"repositories_added"`
	RepositoriesRemoved []InstalledRepository `json:"repositories_removed"`

	// commits only
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
}

type Sender struct {
	Login string `json:"login"`
}

type Installation struct {
	ID      int                 `json:"id"`
	Account InstallationAccount `json:"account"`
}

type InstallationAccount struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Login string `json:"login"`
}

type Repository struct {
	// Fields come from github api

	ID            int    `json:"id"`
	CloneUrl      string `json:"clone_url"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`

	// fields managed by treenq

	// InstallationID defiens a github app Installation id
	InstallationID int `json:"installationID"`
	// TreenqID is an internal identifier
	TreenqID string `json:"treenqID"`
	// Status describes whether a repo is actively deployed or suspended
	Status string `json:"status"`
	// Connected describes whether the repo is used as an app
	Connected bool `json:"connected"`
}

const (
	StatusRepoActive    = "active"
	StatusRepoSuspended = "suspended"
)

type InstalledRepository struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

type BuildArtifactRequest struct {
	Name       string
	Path       string
	Dockerfile string
	Tag        string
}

type Image struct {
	// Registry is a registry name in the cloud provider
	Registry string
	// Repository is a facto name of the image
	Repository string
	// Tag is a version of the image
	Tag string
}

func (i Image) Image() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
}

func (i Image) FullPath() string {
	return fmt.Sprintf("%s/%s:%s", i.Registry, i.Repository, i.Tag)
}

type GithubWebhookResponse struct{}

type GitRepo struct {
	Dir string
	Sha string
}

type AppDeployment struct {
	ID string
	// RepoID is a reference to a repository id
	RepoID string
	// Space is a treenq space definition
	Space tqsdk.Space
	// Sha is a commit sha a user requested to deploy or given from a github webhook
	Sha string
	// BuildTag is a docker build image or an image created using buildpacks
	BuildTag string
	// UserDisplayName is a user loging, comes from a user token or github hook Sender
	UserDisplayName string
	// CreatedAt marks the start of the deployment (might not fit the exact start of the execution)
	CreatedAt time.Time
}

func (h *Handler) GithubWebhook(ctx context.Context, req GithubWebhookRequest) (GithubWebhookResponse, *vel.Error) {
	// Save installation id link to a profile
	if req.Action == "created" {
		err := h.db.LinkGithub(ctx, req.Installation.ID, req.Sender.Login, req.Repositories)
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Message: "failed to link github",
				Err:     err,
			}
		}
		return GithubWebhookResponse{}, nil
	}
	if req.Action == "added" {
		err := h.db.SaveGithubRepos(ctx, req.Installation.ID, req.Sender.Login, req.RepositoriesAdded)
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Message: "failed to save github repos",
				Err:     err,
			}
		}
		return GithubWebhookResponse{}, nil

	}
	if req.Action == "removed" {
		err := h.db.RemoveGithubRepos(ctx, req.Installation.ID, req.RepositoriesRemoved)
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Message: "failed to remove github repos",
				Err:     err,
			}
		}
		return GithubWebhookResponse{}, nil

	}

	// new commit to the default branch
	if req.Action == "" && req.Ref == "refs/heads/"+req.Repository.DefaultBranch {
		repo, err := h.db.GetRepoByGithub(ctx, req.Repository.ID)
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Message: "failed to get treenq repo by github",
				Err:     err,
			}
		}
		req.Repository.InstallationID = req.Installation.ID
		req.Repository.TreenqID = repo.TreenqID
		req.Repository.Status = repo.Status
		req.Repository.Connected = repo.Connected

		// TODO: update repo it changes
		// if repo.HasDiff(req.Repository) {
		// 	h.db.UpdateRepository(ctx, req.Repository)
		// }

		return GithubWebhookResponse{}, h.deployRepo(
			ctx,
			req.Sender.Login,
			req.Repository,
		)
	}

	return GithubWebhookResponse{}, nil
}

func (h *Handler) deployRepo(ctx context.Context, userDisplayName string, repo Repository) *vel.Error {
	if !repo.Connected {
		return nil
	}
	if repo.Status != StatusRepoActive {
		// not expected case, suspended repos won't send any events,
		// but in case we introduce a new status it must handle it
		return nil
	}

	token := ""
	if repo.Private {
		var err error
		token, err = h.githubClient.IssueAccessToken(repo.InstallationID)
		if err != nil {
			return &vel.Error{
				Message: "failed to issue github access token",
				Err:     err,
			}
		}
	}

	gitRepo, err := h.git.Clone(repo.CloneUrl, repo.InstallationID, repo.TreenqID, token)
	if err != nil {
		return &vel.Error{
			Message: "failed to close git repo",
			Err:     err,
		}
	}
	defer os.RemoveAll(gitRepo.Dir)

	appSpace, err := h.extractor.ExtractConfig(gitRepo.Dir)
	if err != nil {
		return &vel.Error{
			Message: "failed to extract config",
			Err:     err,
		}
	}

	dockerFilePath := filepath.Join(gitRepo.Dir, appSpace.Service.DockerfilePath)
	buildRequest := BuildArtifactRequest{
		Name:       appSpace.Service.Name,
		Path:       gitRepo.Dir,
		Dockerfile: dockerFilePath,
		Tag:        "latest",
	}
	image, err := h.docker.Build(ctx, buildRequest)
	if err != nil {
		return &vel.Error{
			Message: "failed to build an image",
			Err:     err,
		}
	}

	_, err = h.db.SaveDeployment(ctx, AppDeployment{
		RepoID:          repo.TreenqID,
		Space:           appSpace,
		Sha:             gitRepo.Sha,
		BuildTag:        buildRequest.Tag,
		UserDisplayName: userDisplayName,
	})
	if err != nil {
		return &vel.Error{
			Message: "failed to save deployment",
			Err:     err,
		}
	}

	appKubeDef := h.kube.DefineApp(ctx, repo.TreenqID, appSpace, image)
	if err := h.kube.Apply(ctx, h.kubeConfig, appKubeDef); err != nil {
		return &vel.Error{
			Message: "failed to apply kube definition",
			Err:     err,
		}
	}
	return nil
}
