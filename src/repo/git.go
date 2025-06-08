package repo

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/treenq/treenq/src/domain"
)

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{dir: dir}
}

func (g *Git) Clone(urlStr string, installationID int, repoID string, accessToken string, progress *domain.ProgressBuf) (domain.GitRepo, error) {
	dir := filepath.Join(g.dir, strconv.Itoa(installationID), repoID)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("failed to create clone directory: %s", err)
		}
	}

	u, err := url.Parse(urlStr)
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
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return domain.GitRepo{}, fmt.Errorf("error while cloning the repo: %s", err)
		}

		r, err = git.PlainOpen(dir)
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while opening the repo: %s", err)
		}
		w, err := r.Worktree()
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while getting worktree: %s", err)
		}
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return domain.GitRepo{}, fmt.Errorf("error while pulling latest: %s", err)
		}
	}
	ref, err := r.Head()
	if err != nil {
		return domain.GitRepo{}, nil
	}

	return domain.GitRepo{
		Dir: dir,
		Sha: ref.Hash().String(),
	}, nil
}
