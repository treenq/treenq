package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/pkg/vel"
	"golang.org/x/exp/maps"
)

var ErrNoConfigFileFound = errors.New("no config file found")

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

type InstalledRepository = Repository

type Repository struct {
	// Fields come from github api

	ID int `json:"id"`
	// CloneUrl field exists, but it's missing in InstalledRepository,
	// as a result we can't save it in a database on installing a github app,
	// so currently for github repos we rely on building url from FullName
	// CloneUrl      string `json:"clone_url"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	// DefaultBranch string `json:"default_branch"`

	// fields managed by treenq

	Branch string `json:"branch"`
	// InstallationID defines a github app Installation id
	InstallationID int `json:"installationID"`
	// TreenqID is an internal identifier
	TreenqID string `json:"treenqID"`
	// Status describes whether a repo is actively deployed or suspended
	Status string `json:"status"`
}

func (r Repository) CloneUrl() string {
	return fmt.Sprintf("https://github.com/%s.git", r.FullName)
}

const (
	StatusRepoActive    = "active"
	StatusRepoSuspended = "suspended"
)

type BuildArtifactRequest struct {
	Name         string
	Path         string
	Dockerfile   string
	Tag          string
	DeploymentID string
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
	ID string `json:"id"`
	// RepoID is a reference to a repository id
	RepoID string `json:"repoID"`
	// Space is a treenq space definition
	Space tqsdk.Space `json:"space"`
	// Sha is a commit sha a user requested to deploy or given from a github webhook
	Sha string `json:"sha"`
	// BuildTag is a docker build image or an image created using buildpacks
	BuildTag string `json:"buildTag"`
	// UserDisplayName is a user loging, comes from a user token or github hook Sender
	UserDisplayName string `json:"userDisplayName"`
	// CreatedAt marks the start of the deployment (might not fit the exact start of the execution)
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// Status describes the status of the deployment
	Status DeployStatus `json:"status"`
}

type DeployStatus string

const (
	DeployStatusInit   DeployStatus = "init"
	DeployStatusDone   DeployStatus = "done"
	DeployStatusFailed DeployStatus = "failed"
)

func (h *Handler) GithubWebhook(ctx context.Context, req GithubWebhookRequest) (GithubWebhookResponse, *vel.Error) {
	// Save installation id link to a profile
	if req.Action == "created" {
		_, err := h.db.LinkGithub(ctx, req.Installation.ID, req.Sender.Login, req.Repositories)
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
	// if req.Action == "" && req.Ref == "refs/heads/"+req.Repository.Branch {
	// 	repo, err := h.db.GetRepoByGithub(ctx, req.Repository.ID)
	// 	if err != nil {
	// 		return GithubWebhookResponse{}, &vel.Error{
	// 			Message: "failed to get treenq repo by github",
	// 			Err:     err,
	// 		}
	// 	}
	//
	// 	req.Repository.InstallationID = req.Installation.ID
	// 	req.Repository.TreenqID = repo.TreenqID
	// 	req.Repository.Status = repo.Status
	//
	// 	return GithubWebhookResponse{}, h.deployRepo(
	// 		ctx,
	// 		req.Sender.Login,
	// 		req.Repository,
	// 	)
	// }

	return GithubWebhookResponse{}, nil
}

func (h *Handler) deployRepo(ctx context.Context, userDisplayName string, repo Repository) (string, *vel.Error) {
	// validate the repo must run
	if repo.Branch == "" {
		return "", nil
	}
	if repo.Status != StatusRepoActive {
		// not expected case, suspended repos won't send any events,
		// but in case we introduce a new status it must handle it
		return "", nil
	}

	// Create initial deployment with "init" status
	deployment := AppDeployment{
		RepoID:          repo.TreenqID,
		UserDisplayName: userDisplayName,
		Status:          DeployStatusInit,
		Space:           tqsdk.Space{},
	}

	deployment, err := h.db.SaveDeployment(ctx, deployment)
	if err != nil {
		return "", &vel.Error{
			Code: "FAILED_CREATE_DEPLOYMENT",
			Err:  err,
		}
	}

	go func() {
		ctx := context.WithoutCancel(ctx)
		ctx, cancel := context.WithTimeout(ctx, time.Second*300)
		defer cancel()
		// Start the build process
		if err := h.buildApp(ctx, deployment, repo); err != nil {
			// Update deployment status to failed
			deployment.Status = DeployStatusFailed
			if updateErr := h.db.UpdateDeployment(ctx, deployment); updateErr != nil {
				// TODO: log error
				log.Println("[ERROR] failed update deployment", updateErr)
			}
			log.Println("[ERROR] failed to build app", err)
		}

		// Update deployment status to done
		deployment.Status = DeployStatusDone
		if err := h.db.UpdateDeployment(ctx, deployment); err != nil {
			// TODO: log error
			log.Println("[ERROR] failed mark deployment as done", err)
		}
	}()

	return deployment.ID, nil
}

type ProgressBuf struct {
	Bufs map[string]buf

	mx sync.RWMutex
}

type buf struct {
	WriteAt time.Time
	Content []message
}

type message struct {
	Payload []byte
	Level   slog.Level
}

func (b *ProgressBuf) Append(deploymentID string, content []byte, level slog.Level) {
	b.mx.Lock()
	defer b.mx.Unlock()

	buf := b.Bufs[deploymentID]
	buf.WriteAt = time.Now()
	buf.Content = append(buf.Content, message{
		Payload: content,
		Level:   level,
	})
	b.Bufs[deploymentID] = buf

	if len(b.Bufs) >= 100 {
		b.clean()
	}
}

func (b *ProgressBuf) AsWriter(deploymentID string, level slog.Level) io.Writer {
	return &progressWriter{
		deploymentID: deploymentID,
		level:        level,
		buf:          b,
	}
}

type progressWriter struct {
	deploymentID string
	level        slog.Level
	buf          *ProgressBuf
}

func (w *progressWriter) Write(buf []byte) (int, error) {
	w.buf.Append(w.deploymentID, buf, w.level)
	return len(buf), nil
}

func (b *ProgressBuf) clean() {
	maps.DeleteFunc(b.Bufs, func(k string, v buf) bool {
		return time.Since(v.WriteAt) > (time.Minute * 20)
	})
}

var progress = &ProgressBuf{Bufs: make(map[string]buf)}

func (h *Handler) buildApp(ctx context.Context, deployment AppDeployment, repo Repository) *vel.Error {
	token := ""
	if repo.Private {
		var err error
		progress.Append(deployment.ID, []byte("private repository detected, issuing github access token"), slog.LevelDebug)
		token, err = h.githubClient.IssueAccessToken(repo.InstallationID)
		if err != nil {
			progress.Append(deployment.ID, []byte("failed to issue a github access token: "+err.Error()), slog.LevelError)
			return &vel.Error{
				Message: "failed to issue github access token",
				Err:     err,
			}
		}
		progress.Append(deployment.ID, []byte("issued github access token"), slog.LevelInfo)
	}

	progress.Append(deployment.ID, []byte("cloning github repository"), slog.LevelDebug)
	gitRepo, err := h.git.Clone(repo.CloneUrl(), repo.InstallationID, repo.TreenqID, token)
	if err != nil {
		progress.Append(deployment.ID, []byte("failed to clone github repository: "+err.Error()), slog.LevelError)
		return &vel.Error{
			Message: "failed to clone git repo",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, []byte("cloned github repository"), slog.LevelInfo)

	defer os.RemoveAll(gitRepo.Dir)

	progress.Append(deployment.ID, []byte("extracting treenq config"), slog.LevelDebug)
	appSpace, err := h.extractor.ExtractConfig(gitRepo.Dir)
	if err != nil {
		progress.Append(deployment.ID, []byte("failed to extract treenq config: "+err.Error()), slog.LevelError)
		if errors.Is(err, ErrNoConfigFileFound) {
			return &vel.Error{
				Message: "failed to extract config",
				Code:    "NO_TQ_CONFIG_FOUND",
			}
		}
		return &vel.Error{
			Message: "failed to extract config",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, []byte("extracted treenq config"), slog.LevelInfo)

	dockerFilePath := filepath.Join(gitRepo.Dir, appSpace.Service.DockerfilePath)
	buildRequest := BuildArtifactRequest{
		Name:         appSpace.Service.Name,
		Path:         gitRepo.Dir,
		Dockerfile:   dockerFilePath,
		Tag:          gitRepo.Sha,
		DeploymentID: deployment.ID,
	}
	progress.Append(deployment.ID, []byte("build image"), slog.LevelDebug)
	progress.Append(deployment.ID, []byte(fmt.Appendf(nil, "%+v", buildRequest)), slog.LevelDebug)
	image, err := h.docker.Build(ctx, buildRequest, progress)
	if err != nil {
		progress.Append(deployment.ID, []byte("failed to build image: "+err.Error()), slog.LevelError)
		return &vel.Error{
			Message: "failed to build an image",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, []byte("built image"), slog.LevelInfo)

	deployment.Space = appSpace
	deployment.BuildTag = image.Tag
	deployment.Sha = gitRepo.Sha
	progress.Append(deployment.ID, []byte("updating deployment state"), slog.LevelDebug)
	progress.Append(deployment.ID, []byte(fmt.Appendf(nil, "%+v", deployment)), slog.LevelDebug)
	err = h.db.UpdateDeployment(ctx, deployment)
	if err != nil {
		progress.Append(deployment.ID, []byte("failed to update deployment state"+err.Error()), slog.LevelError)
		return &vel.Error{
			Message: "failed to save deployment",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, []byte("updated deployment state"), slog.LevelInfo)

	progress.Append(deployment.ID, []byte("apply new image"), slog.LevelDebug)
	progress.Append(deployment.ID, []byte(fmt.Appendf(nil, "%+v", image)), slog.LevelDebug)
	appKubeDef := h.kube.DefineApp(ctx, repo.TreenqID, appSpace, image)
	if err := h.kube.Apply(ctx, h.kubeConfig, appKubeDef); err != nil {
		progress.Append(deployment.ID, []byte("failed to apply new image"+err.Error()), slog.LevelError)
		return &vel.Error{
			Message: "failed to apply kube definition",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, []byte("applied new image"), slog.LevelInfo)

	return nil
}
