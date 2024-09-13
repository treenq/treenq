package repo

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-git/go-git/v5"
)

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{dir: dir}
}

func (g *Git) Clone(urlStr string, installationID, repoID int, accessToken string) (string, error) {
	dir := filepath.Join(g.dir, strconv.Itoa(installationID), strconv.Itoa(repoID))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create clone directory: %s", err)
		}
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	if accessToken != "" {
		u.User = url.UserPassword("x-access-token", accessToken)
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL:      u.String(),
		Progress: os.Stdout,
	})
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return "", fmt.Errorf("error while cloning the repo: %s", err)
		}

		r, err := git.PlainOpen(dir)
		if err != nil {
			return "", fmt.Errorf("error while opening the repo: %s", err)
		}
		w, err := r.Worktree()
		if err != nil {
			return "", fmt.Errorf("error while getting worktree: %s", err)
		}
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return "", fmt.Errorf("error while pulling latest: %s", err)
		}
	}
	return dir, nil
}
