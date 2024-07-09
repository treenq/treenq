package repo

import (
	"fmt"
	"net/url"
	"os"

	"github.com/go-git/go-git/v5"
)

type Git struct {
}

func NewGit() *Git {
	return &Git{}
}

func (g *Git) Clone(urlStr string, accesstoken string) (string, error) {
	// dir, err := os.MkdirTemp("repos/", "")
	// dir := os.Mkdir("/", fs.ModePerm)
	cloneDir := "allrepos/"
	if _, err := os.Stat(cloneDir); os.IsNotExist(err) {
		err = os.MkdirAll(cloneDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create clone directory: %s", err)
		}
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword("x-access-token", accesstoken)

	_, err = git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:      u.String(),
		Progress: os.Stdout,
	})
	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			return "", fmt.Errorf("error while cloning the repo: %s", err)
		} else {
			r, err := git.PlainOpen(cloneDir)
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
	}
	return cloneDir, nil
}
