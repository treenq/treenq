package repo

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/treenq/treenq/src/domain"
)

type Git struct {
	dir         string
	progressBuf *domain.ProgressBuf
}

func NewGit(dir string, progressBuf *domain.ProgressBuf) *Git {
	return &Git{
		dir:         dir,
		progressBuf: progressBuf,
	}
}

func (g *Git) Clone(repo domain.Repository, accessToken string, branch, sha string) (domain.GitRepo, error) {
	if branch == "" && sha == "" {
		return domain.GitRepo{}, domain.ErrNoGitCheckoutSpecified
	}
	if branch != "" && sha != "" {
		return domain.GitRepo{}, domain.ErrGitBranchAndShaMutuallyExclusive
	}

	dir := repo.Location(g.dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("failed to create clone directory: %s", err)
		}
	}

	u, err := url.Parse(repo.CloneURL())
	if err != nil {
		return domain.GitRepo{}, err
	}
	if accessToken != "" {
		u.User = url.UserPassword("x-access-token", accessToken)
	}

	// Create a progress writer using the deployment ID from the repository
	// Using the repository TreenqID as the deployment ID
	progressWriter := g.progressBuf.AsWriter(fmt.Sprintf("repo-%s", dir), slog.LevelInfo)
	g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
		Payload: fmt.Sprintf("Cloning repository from %s", repo.CloneURL()),
		Level:   slog.LevelInfo,
	})

	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      u.String(),
		Progress: progressWriter,
	})
	var w *git.Worktree
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			errMsg := fmt.Sprintf("error while cloning the repo: %s", err)
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: errMsg,
				Level:   slog.LevelError,
			})
			return domain.GitRepo{}, fmt.Errorf(errMsg)
		}

		g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
			Payload: "Repository already exists, opening and pulling latest changes",
			Level:   slog.LevelInfo,
		})

		r, err = git.PlainOpen(dir)
		if err != nil {
			errMsg := fmt.Sprintf("error while opening the repo: %s", err)
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: errMsg,
				Level:   slog.LevelError,
			})
			return domain.GitRepo{}, fmt.Errorf(errMsg)
		}
		w, err = r.Worktree()
		if err != nil {
			errMsg := fmt.Sprintf("error while getting worktree: %s", err)
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: errMsg,
				Level:   slog.LevelError,
			})
			return domain.GitRepo{}, fmt.Errorf(errMsg)
		}
		err = w.Pull(&git.PullOptions{
			RemoteName: "origin",
			Progress:   g.progressBuf.AsWriter(fmt.Sprintf("repo-%s", dir), slog.LevelInfo),
		})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			errMsg := fmt.Sprintf("error while pulling latest: %s", err)
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: errMsg,
				Level:   slog.LevelError,
			})
			return domain.GitRepo{}, fmt.Errorf(errMsg)
		} else if errors.Is(err, git.NoErrAlreadyUpToDate) {
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: "Repository already up-to-date",
				Level:   slog.LevelInfo,
			})
		}
	}
	if w == nil {
		w, err = r.Worktree()
		if err != nil {
			errMsg := fmt.Sprintf("error while getting worktree: %s", err)
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: errMsg,
				Level:   slog.LevelError,
			})
			return domain.GitRepo{}, fmt.Errorf(errMsg)
		}
	}
	if sha != "" {
		g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
			Payload: fmt.Sprintf("Checking out commit: %s", sha),
			Level:   slog.LevelInfo,
		})
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(sha),
		})
		if err != nil {
			errMsg := fmt.Sprintf("error while checking out commit %s: %s", sha, err)
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: errMsg,
				Level:   slog.LevelError,
			})
			return domain.GitRepo{}, fmt.Errorf(errMsg)
		}
	} else {
		g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
			Payload: fmt.Sprintf("Checking out branch: %s", branch),
			Level:   slog.LevelInfo,
		})
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
		})
		if err != nil {
			errMsg := fmt.Sprintf("error while checking out branch %s: %s", branch, err)
			g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
				Payload: errMsg,
				Level:   slog.LevelError,
			})
			return domain.GitRepo{}, fmt.Errorf(errMsg)
		}
	}
	ref, err := r.Head()
	if err != nil {
		errMsg := fmt.Sprintf("error while getting HEAD reference: %s", err)
		g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
			Payload: errMsg,
			Level:   slog.LevelError,
		})
		return domain.GitRepo{}, fmt.Errorf(errMsg)
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		errMsg := fmt.Sprintf("failed to get a repo commit: %s", err)
		g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
			Payload: errMsg,
			Level:   slog.LevelError,
		})
		return domain.GitRepo{}, fmt.Errorf(errMsg)
	}

	result := domain.GitRepo{
		Dir:     dir,
		Sha:     ref.Hash().String(),
		Message: commit.Message,
	}
	
	g.progressBuf.Append(fmt.Sprintf("repo-%s", dir), domain.ProgressMessage{
		Payload: fmt.Sprintf("Successfully cloned repository at commit %s: %s", result.Sha, result.Message),
		Level:   slog.LevelInfo,
		Final:   true,
	})
	
	return result, nil
}
