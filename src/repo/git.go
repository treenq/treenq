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
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", fmt.Errorf("failed to create a tmp clone directory %s", err)
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword("x-access-token", accesstoken)

	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      u.String(),
		Progress: os.Stdout,
	})
	if err != nil {
		if err != git.ErrRepositoryAlreadyExists {
			return "", fmt.Errorf("error while cloning the repo %s", err)
		} else {
			w, err := r.Worktree()
			if err != nil {
				return "", fmt.Errorf("error while getting worktree %s", err)
			}
			err = w.Pull(&git.PullOptions{RemoteName: "origin"})
			if err != nil {
				return "", fmt.Errorf("error while pulling latest %s", err)
			}

		}
	}
	return dir, err
}
