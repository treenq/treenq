package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/dennypenta/vel"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"golang.org/x/exp/maps"
)

var (
	ErrDeployStatusMustBeString         = errors.New("deploy status must be string")
	ErrImageNotFound                    = errors.New("image not found")
	ErrNoGitCheckoutSpecified           = errors.New("git branch or sha must be specified")
	ErrGitBranchAndShaMutuallyExclusive = errors.New("git branch and sha and tag are mutually exclusive")
	ErrSecretNotFound                   = errors.New("secret not found")
	ErrSpaceNotFound                    = errors.New("space not found")
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
	Ref        string           `json:"ref"`
	Repository GithubRepository `json:"repository"`
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

type InstalledRepository = GithubRepository

type Repository interface {
	CloneURL() string
	Location(root string) string
}

type GithubRepository struct {
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

// CloneUrl implements gives a provider's clone url
func (r GithubRepository) CloneURL() string {
	return fmt.Sprintf("https://github.com/%s.git", r.FullName)
}

// Location gives a defined location where the repository is placed on a disk
func (r GithubRepository) Location(root string) string {
	return filepath.Join(root, strconv.Itoa(r.InstallationID), r.TreenqID)
}

const (
	StatusRepoActive    = "active"
	StatusRepoSuspended = "suspended"
)

type BuildArtifactRequest struct {
	Name          string
	Path          string
	Dockerfile    string
	DockerContext string
	Tag           string
	DeploymentID  string
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
	// Message defines a commit message
	Message string
}

type AppDeployment struct {
	ID string `json:"id"`
	// FromDeploymentID defines a deployment from which it was inherited,
	// used to specify rollbacks
	FromDeploymentID string `json:"fromDeploymentID"`
	// RepoID is a reference to a repository id
	RepoID string `json:"repoID"`
	// Space is a treenq space definition
	Space tqsdk.Space `json:"space"`
	// Sha is a commit sha a user requested to deploy or given from a github webhook
	Sha string `json:"sha"`
	// Branch is a git branch deployed if sha is not specified directly
	Branch string `json:"branch"`
	// CommitMessage defines a commit message
	CommitMessage string `json:"commitMessage"`
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

func (d AppDeployment) IsZero() bool {
	return d.ID == ""
}

type DeployStatus string

const (
	DeployStatusRunning DeployStatus = "run"
	DeployStatusDone    DeployStatus = "done"
	DeployStatusFailed  DeployStatus = "failed"
)

func (h *Handler) GithubWebhook(ctx context.Context, req GithubWebhookRequest) (GithubWebhookResponse, *vel.Error) {
	// Save installation id link to a profile
	if req.Action == "created" {
		_, err := h.db.LinkGithub(ctx, req.Installation.ID, req.Sender.Login, req.Repositories)
		if err != nil {
			if errors.Is(err, ErrInstallationNotFound) {
				return GithubWebhookResponse{}, &vel.Error{
					Code: "INSTALLATION_NOT_FOUND",
				}
			}
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
	if req.Action == "deleted" {
		err := h.removeInstallation(ctx, req.Installation.ID, req.Sender.Login, req.Repositories)
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Message: "failed to remove installation",
				Err:     err,
			}
		}
		return GithubWebhookResponse{}, nil
	}

	if req.Action == "" {
		repo, err := h.db.GetRepoByGithub(ctx, req.Repository.ID)
		if err != nil {
			if errors.Is(err, ErrRepoNotFound) {
				return GithubWebhookResponse{}, &vel.Error{
					Code: "REPO_NOT_FOUND",
				}
			}
			return GithubWebhookResponse{}, &vel.Error{
				Message: "failed to get github repo by github id",
				Err:     err,
			}
		}

		if req.Ref == "refs/heads/"+repo.Branch {
			space, err := h.githubClient.GetRepoSpace(ctx, req.Installation.ID, repo.FullName, repo.Branch)
			if err != nil {
				if errors.Is(err, ErrNoTqJsonFound) {
					return GithubWebhookResponse{}, &vel.Error{
						Code: "TQ_JSON_NOT_FOUND",
					}
				}
				if errors.Is(err, ErrTqIsNotValidJson) {
					return GithubWebhookResponse{}, &vel.Error{
						Code: "TQ_JSON_INVALID",
					}
				}
				return GithubWebhookResponse{}, &vel.Error{
					Message: "failed to get space from github",
				}
			}

			if err := h.db.SaveSpace(ctx, repo.TreenqID, space); err != nil {
				return GithubWebhookResponse{}, &vel.Error{
					Message: "failed to save space",
				}
			}
		}
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

func countNotEmpty(vals ...string) int {
	notEmpty := 0
	for i := range vals {
		if vals[i] != "" {
			notEmpty += 1
		}
	}

	return notEmpty
}

func (h *Handler) deployRepo(ctx context.Context, userDisplayName string, workspace Workspace, repo GithubRepository, fromDeploymentID, branch, sha, tag string) (AppDeployment, *vel.Error) {
	// validate the repo must run
	if repo.Branch == "" {
		return AppDeployment{}, &vel.Error{
			Code: "REPO_IS_NOT_CONNECTED",
		}
	}
	if repo.Status != StatusRepoActive {
		// not expected case, suspended repos won't send any events,
		// but in case we introduce a new status it must handle it
		return AppDeployment{}, &vel.Error{
			Code: "REPO_IS_NOT_ACTIVE",
		}
	}

	notEmptyDeployMarks := countNotEmpty(branch, sha, tag)
	if notEmptyDeployMarks > 1 {
		return AppDeployment{}, &vel.Error{
			Code: "ONLY_BRANCH_OR_SHA_OR_TAG_ALLOWED",
		}
	}

	if notEmptyDeployMarks == 0 {
		branch = repo.Branch
	}

	// Create initial deployment with "init" status
	deployment := AppDeployment{
		RepoID:           repo.TreenqID,
		UserDisplayName:  userDisplayName,
		Status:           DeployStatusRunning,
		Space:            tqsdk.Space{},
		FromDeploymentID: fromDeploymentID,
		Branch:           branch,
		Sha:              sha,
		BuildTag:         tag,
	}

	if fromDeploymentID != "" {
		fromDeployment, err := h.db.GetDeployment(ctx, workspace.ID, fromDeploymentID)
		if err != nil {
			if errors.Is(err, ErrDeploymentNotFound) {
				return AppDeployment{}, &vel.Error{
					Code: "DEPLOYMENT_NOT_FOUND",
				}
			}
			return AppDeployment{}, &vel.Error{
				Message: "failed to get deployment to rollback to",
				Err:     err,
			}
		}
		deployment.Sha = fromDeployment.Sha
		deployment.Branch = fromDeployment.Branch
		deployment.CommitMessage = fromDeployment.CommitMessage
		deployment.BuildTag = fromDeployment.BuildTag
		deployment.Space = fromDeployment.Space
	}

	deployment, err := h.db.SaveDeployment(ctx, deployment)
	if err != nil {
		return AppDeployment{}, &vel.Error{
			Code: "FAILED_CREATE_DEPLOYMENT",
			Err:  err,
		}
	}

	go func() {
		// prepare local goroutine scope
		deployment := deployment
		workspace := workspace
		ctx := context.WithoutCancel(ctx)
		ctx, cancel := context.WithTimeout(ctx, time.Second*300)
		defer cancel()

		// start the build process
		deployment, err := h.buildApp(ctx, deployment, repo, workspace)
		if err != nil {
			// update deployment status to failed
			deployment.Status = DeployStatusFailed
			if updateErr := h.db.UpdateDeployment(ctx, deployment); updateErr != nil {
				// TODO: log error
				log.Println("[ERROR] failed update deployment", updateErr)
			}
			log.Println("[ERROR] failed to build app", err)
		}

		// update deployment status to done
		deployment.Status = DeployStatusDone
		if err := h.db.UpdateDeployment(ctx, deployment); err != nil {
			// TODO: log error
			log.Println("[ERROR] failed mark deployment as done", err)
		}
	}()

	return deployment, nil
}

type ProgressBuf struct {
	Bufs map[string]buf

	mx sync.RWMutex
}

func NewProgressBuf() *ProgressBuf {
	return &ProgressBuf{
		Bufs: make(map[string]buf),
	}
}

type buf struct {
	WriteAt time.Time
	Content []ProgressMessage
	Subs    []Subscriber
}

type ProgressMessage struct {
	Payload    string        `json:"payload"`
	Level      slog.Level    `json:"level"`
	Final      bool          `json:"final"`
	Timestamp  time.Time     `json:"timestamp"`
	Deployment AppDeployment `json:"deployment,omitzero"`
	ErrorCode  string        `json:"errorCode,omitempty"`
}

type Subscriber struct {
	out    chan ProgressMessage
	done   <-chan struct{}
	closed bool
}

func (b *ProgressBuf) Get(ctx context.Context, deploymentID string) <-chan ProgressMessage {
	b.mx.Lock()

	out := make(chan ProgressMessage)

	go func() {
		defer b.mx.Unlock()

		buf, ok := b.Bufs[deploymentID]
		if !ok {
			out <- ProgressMessage{
				ErrorCode: "NO_LOGS",
			}
			close(out)
			return
		}
		for _, m := range buf.Content {
			select {
			case out <- m:
				if m.Final {
					return
				}
			case <-ctx.Done():
				return
			}
		}
		buf.Subs = append(buf.Subs, Subscriber{
			out:  out,
			done: ctx.Done(),
		})
		b.Bufs[deploymentID] = buf
	}()

	return out
}

func (b *ProgressBuf) Append(deploymentID string, m ProgressMessage) {
	m.Timestamp = time.Now().UTC()
	b.mx.Lock()
	defer b.mx.Unlock()

	buf := b.Bufs[deploymentID]
	buf.WriteAt = time.Now()
	buf.Content = append(buf.Content, m)
	for i, sub := range buf.Subs {
		select {
		case <-sub.done:
			close(sub.out)
			buf.Subs[i].closed = true
			continue
		case <-time.After(time.Second):
			close(sub.out)
			buf.Subs[i].closed = true
			continue
		case sub.out <- m:
		}

		if m.Final {
			close(sub.out)
			buf.Subs[i].closed = true
			continue
		}
	}

	copiedSubs := make([]Subscriber, 0, len(buf.Subs))
	for i := range buf.Subs {
		if buf.Subs[i].closed {
			continue
		}

		copiedSubs = append(copiedSubs, buf.Subs[i])
	}
	buf.Subs = copiedSubs
	b.Bufs[deploymentID] = buf

	if len(b.Bufs) >= 100 {
		b.clean()
	}
}

func (b *ProgressBuf) clean() {
	maps.DeleteFunc(b.Bufs, func(k string, v buf) bool {
		return time.Since(v.WriteAt) > (time.Minute * 5)
	})
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
	w.buf.Append(w.deploymentID, ProgressMessage{
		Payload: string(buf),
		Level:   w.level,
	})
	return len(buf), nil
}

var progress = &ProgressBuf{Bufs: make(map[string]buf)}

func (h *Handler) buildApp(ctx context.Context, deployment AppDeployment, repo GithubRepository, workspace Workspace) (AppDeployment, *vel.Error) {
	if deployment.FromDeploymentID != "" {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "inspecting an image",
			Level:   slog.LevelDebug,
		})
		image, err := h.docker.Inspect(ctx, deployment)
		if err != nil {
			if errors.Is(err, ErrImageNotFound) {
				progress.Append(deployment.ID, ProgressMessage{
					Payload: "image not found, build is required",
					Level:   slog.LevelWarn,
				})
				return h.buildFromRepo(ctx, deployment, repo, workspace)
			}
			progress.Append(deployment.ID, ProgressMessage{
				Payload: "failed to inspect an iamge",
				Level:   slog.LevelError,
			})
			return AppDeployment{}, &vel.Error{
				Message: "failed to inspect an image",
				Err:     err,
			}
		}
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "image has been inspected",
			Level:   slog.LevelInfo,
		})

		return h.applyImage(ctx, repo.TreenqID, deployment, image, workspace)
	}

	return h.buildFromRepo(ctx, deployment, repo, workspace)
}

func (h *Handler) buildFromRepo(ctx context.Context, deployment AppDeployment, repo GithubRepository, workspace Workspace) (AppDeployment, *vel.Error) {
	token := ""
	if repo.Private {
		var err error
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "private repository detected, issuing github access token",
			Level:   slog.LevelDebug,
		})
		token, err = h.githubClient.IssueAccessToken(repo.InstallationID)
		if err != nil {
			progress.Append(deployment.ID, ProgressMessage{
				Payload: "failed to issue a github access token: " + err.Error(),
				Level:   slog.LevelError,
			})
			return AppDeployment{}, &vel.Error{
				Message: "failed to issue github access token",
				Err:     err,
			}
		}
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "issued github access token",
			Level:   slog.LevelInfo,
		})
	}

	progress.Append(deployment.ID, ProgressMessage{
		Payload: "cloning github repository",
		Level:   slog.LevelDebug,
	})
	gitRepo, err := h.git.Clone(repo, token, deployment.Branch, deployment.Sha, deployment.BuildTag, progress.AsWriter(deployment.ID, slog.LevelInfo))
	if err != nil {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "failed to clone github repository: " + err.Error(),
			Level:   slog.LevelError,
		})
		return AppDeployment{}, &vel.Error{
			Message: "failed to clone git repo",
			Err:     err,
		}
	}
	deployment.Sha = gitRepo.Sha
	deployment.CommitMessage = gitRepo.Message
	if deployment.BuildTag == "" {
		deployment.BuildTag = gitRepo.Sha
	}
	progress.Append(deployment.ID, ProgressMessage{
		Payload:    "cloned github repository",
		Level:      slog.LevelInfo,
		Deployment: deployment,
	})

	defer os.RemoveAll(gitRepo.Dir)

	progress.Append(deployment.ID, ProgressMessage{
		Payload: "extracting treenq config",
		Level:   slog.LevelDebug,
	})
	var appSpace tqsdk.Space
	if deployment.Space.Service.Name != "" {
		appSpace = deployment.Space
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "reusing tq config from referenced deployment",
			Level:   slog.LevelInfo,
		})
	} else {
		appSpace, err = h.extractor.ExtractConfig(gitRepo.Dir)
		if err != nil {
			progress.Append(deployment.ID, ProgressMessage{
				Payload: "failed to extract treenq config: " + err.Error(),
				Level:   slog.LevelError,
			})
			if errors.Is(err, ErrNoTqJsonFound) {
				return AppDeployment{}, &vel.Error{
					Message: "failed to extract config",
					Code:    "NO_TQ_JSON_FOUND",
				}
			}
			return AppDeployment{}, &vel.Error{
				Message: "failed to extract config",
				Err:     err,
			}
		}
		deployment.Space = appSpace
		progress.Append(deployment.ID, ProgressMessage{
			Payload:    "extracted treenq config",
			Level:      slog.LevelInfo,
			Deployment: deployment,
		})
	}
	marshalledSpaceConfig, err := json.Marshal(deployment.Space)
	if err == nil {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: string(marshalledSpaceConfig),
			Level:   slog.LevelDebug,
		})
	}

	dockerContext := appSpace.Service.DockerContext
	dockerContext = filepath.Join(gitRepo.Dir, dockerContext)
	dockerFilePath := filepath.Join(gitRepo.Dir, appSpace.Service.DockerfilePath)

	buildRequest := BuildArtifactRequest{
		Name:          appSpace.Service.Name,
		DockerContext: dockerContext,
		Path:          gitRepo.Dir,
		Dockerfile:    dockerFilePath,
		Tag:           deployment.BuildTag,
		DeploymentID:  deployment.ID,
	}
	progress.Append(deployment.ID, ProgressMessage{
		Payload: "build image",
		Level:   slog.LevelDebug,
	})
	progress.Append(deployment.ID, ProgressMessage{
		Payload: fmt.Sprintf("%+v", buildRequest),
		Level:   slog.LevelDebug,
	})
	image, err := h.docker.Build(ctx, buildRequest, progress)
	if err != nil {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "failed to build image: " + err.Error(),
			Level:   slog.LevelError,
		})
		return AppDeployment{}, &vel.Error{
			Message: "failed to build an image",
			Err:     err,
		}
	}
	deployment.BuildTag = image.Tag
	progress.Append(deployment.ID, ProgressMessage{
		Payload:    "built image: " + image.FullPath(),
		Level:      slog.LevelInfo,
		Deployment: deployment,
	})

	progress.Append(deployment.ID, ProgressMessage{
		Payload: "updating deployment state",
		Level:   slog.LevelDebug,
	})
	err = h.db.UpdateDeployment(ctx, deployment)
	if err != nil {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "failed to update deployment state" + err.Error(),
			Level:   slog.LevelError,
		})
		return AppDeployment{}, &vel.Error{
			Message: "failed to save deployment",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, ProgressMessage{
		Payload: "updated deployment state",
		Level:   slog.LevelInfo,
	})

	return h.applyImage(ctx, repo.TreenqID, deployment, image, workspace)
}

func (h *Handler) applyImage(ctx context.Context, repoID string, deployment AppDeployment, image Image, workspace Workspace) (AppDeployment, *vel.Error) {
	progress.Append(deployment.ID, ProgressMessage{
		Payload: "get avilable secret keys",
		Level:   slog.LevelDebug,
	})
	secretKeys, err := h.db.GetRepositorySecretKeys(ctx, repoID, workspace.ID)
	if err != nil {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "failed to get repo secret keys" + err.Error(),
			Level:   slog.LevelError,
		})
		return AppDeployment{}, &vel.Error{
			Message: "failed to get repo secret keys",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, ProgressMessage{
		Payload: "retrieved available secret keys",
		Level:   slog.LevelInfo,
		Final:   true,
	})

	progress.Append(deployment.ID, ProgressMessage{
		Payload: fmt.Sprintf("apply new image: %+v", image),
		Level:   slog.LevelDebug,
	})
	appKubeDef, err := h.kube.DefineApp(ctx, repoID, workspace.Name, deployment.Space, image, secretKeys)
	if err != nil {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "failed to define app" + err.Error(),
			Level:   slog.LevelError,
		})
		return AppDeployment{}, &vel.Error{
			Message: "failed to define app",
			Err:     err,
		}
	}
	if err := h.kube.Apply(ctx, h.kubeConfig, appKubeDef); err != nil {
		progress.Append(deployment.ID, ProgressMessage{
			Payload: "failed to apply new image" + err.Error(),
			Level:   slog.LevelError,
		})
		return AppDeployment{}, &vel.Error{
			Message: "failed to apply app definition",
			Err:     err,
		}
	}
	progress.Append(deployment.ID, ProgressMessage{
		Payload: "applied new image",
		Level:   slog.LevelInfo,
		Final:   true,
	})
	return deployment, nil
}

func (h *Handler) removeInstallation(ctx context.Context, installationID int, userDisplayName string, repos []InstalledRepository) *vel.Error {
	// Remove Kubernetes namespaces first
	for _, repo := range repos {
		treenqRepo, err := h.db.GetRepoByGithub(ctx, repo.ID)
		if err != nil {
			if errors.Is(err, ErrRepoNotFound) {
				continue
			}
			return &vel.Error{
				Message: fmt.Sprintf("failed to get repo %d", repo.ID),
				Err:     err,
			}
		}

		workspace, err := h.db.GetWorkspaceByUserDisplayName(ctx, userDisplayName)
		if err != nil {
			if errors.Is(err, ErrWorkspaceNotFound) {
				return &vel.Error{
					Code: "WORKSPACE_NOT_FOUND",
				}
			}
			return &vel.Error{
				Message: "failed to get workspace info",
				Err:     err,
			}
		}

		if err := h.kube.RemoveNamespace(ctx, h.kubeConfig, treenqRepo.TreenqID, workspace.Name); err != nil {
			return &vel.Error{
				Message: "failed to remove namespace",
				Err:     err,
			}
		}
	}

	// Remove all database data (this handles repos, deployments, secrets, spaces, and installation)
	err := h.db.RemoveInstallation(ctx, installationID)
	if err != nil {
		return &vel.Error{
			Message: "failed to remove installation from database",
			Err:     err,
		}
	}

	return nil
}
