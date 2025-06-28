package repo

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/treenq/treenq/src/domain"
)

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{dir: dir}
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

	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      u.String(),
		Progress: os.Stdout,
	})
	var w *git.Worktree
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return domain.GitRepo{}, fmt.Errorf("error while cloning the repo: %s", err)
		}

		r, err = git.PlainOpen(dir)
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while opening the repo: %s", err)
		}
		w, err = r.Worktree()
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while getting worktree: %s", err)
		}
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return domain.GitRepo{}, fmt.Errorf("error while pulling latest: %s", err)
		}
	}
	if w == nil {
		w, err = r.Worktree()
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while getting worktree: %s", err)
		}
	}
	if sha != "" {
		w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(sha),
		})
	} else {
		w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
		})
	}
	ref, err := r.Head()
	if err != nil {
		return domain.GitRepo{}, nil
	}
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return domain.GitRepo{}, fmt.Errorf("failed to get a repo commit: %w", err)
	}

	return domain.GitRepo{
		Dir:     dir,
		Sha:     ref.Hash().String(),
		Message: commit.Message,
	}, nil
}
