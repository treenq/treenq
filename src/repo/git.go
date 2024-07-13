package repo

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/go-git/go-git/v5"
)

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{dir: dir}
}

func (g *Git) Clone(urlStr string, accesstoken string) (string, error) {
	if _, err := os.Stat(g.dir); os.IsNotExist(err) {
		err = os.MkdirAll(g.dir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create clone directory: %s", err)
		}
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword("x-access-token", accesstoken)

	_, err = git.PlainClone(g.dir, false, &git.CloneOptions{
		URL:      u.String(),
		Progress: os.Stdout,
	})
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return "", fmt.Errorf("error while cloning the repo: %s", err)
		}

		r, err := git.PlainOpen(g.dir)
		if err != nil {
			return "", fmt.Errorf("error while opening the repo: %s", err)
		}
		w, err := r.Worktree()
		if err != nil {
			return "", fmt.Errorf("error while getting worktree: %s", err)
		}
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", fmt.Errorf("error while pulling latest: %s", err)
		}
	}
	return g.dir, nil
}
