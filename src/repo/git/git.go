package git

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/treenq/treenq/src/domain"
)

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{dir: dir}
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

func (g *Git) Clone(repo domain.Repository, accessToken string, branch, sha, tag string, progress io.Writer) (domain.GitRepo, error) {
	if countNotEmpty(branch, sha, tag) == 0 {
		return domain.GitRepo{}, domain.ErrNoGitCheckoutSpecified
	}
	if countNotEmpty(branch, sha, tag) > 1 {
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
		return domain.GitRepo{}, fmt.Errorf("failed to parse clone URL: %w", err)
	}
	if accessToken != "" {
		u.User = url.UserPassword("x-access-token", accessToken)
	}

	cloneOpts := &git.CloneOptions{
		URL:      u.String(),
		Progress: progress,
	}

	// Optimize based on what we're checking out
	if branch != "" {
		// For branch checkout, use shallow clone with single branch
		cloneOpts.Depth = 1
		cloneOpts.SingleBranch = true
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	} else if tag != "" {
		// For tag checkout, use shallow clone with single tag
		cloneOpts.Depth = 1
		cloneOpts.SingleBranch = true
		cloneOpts.ReferenceName = plumbing.NewTagReferenceName(tag)
	} else if sha != "" {
		// For SHA checkout, we need full history but can skip initial checkout
		cloneOpts.NoCheckout = true
	}

	r, err := git.PlainClone(dir, false, cloneOpts)
	var w *git.Worktree

	if err != nil {
		if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return domain.GitRepo{}, fmt.Errorf("error while cloning the repo: %w", err)
		}

		r, err = git.PlainOpen(dir)
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while opening the repo: %w", err)
		}

		w, err = r.Worktree()
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while getting worktree: %w", err)
		}

		pullOpts := &git.PullOptions{
			Progress: progress,
		}
		if branch != "" {
			// For branch checkout, use shallow clone with single branch
			pullOpts.Depth = 1
			pullOpts.SingleBranch = true
			pullOpts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		} else if tag != "" {
			// For tag checkout, use shallow clone with single tag
			pullOpts.Depth = 1
			pullOpts.SingleBranch = true
			pullOpts.ReferenceName = plumbing.NewTagReferenceName(tag)
		}

		if accessToken != "" {
			pullOpts.Auth = &http.BasicAuth{
				Username: "x-access-token",
				Password: accessToken,
			}
		}

		err = w.Pull(pullOpts)
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return domain.GitRepo{}, fmt.Errorf("error while pulling latest: %w", err)
		}
	}

	if w == nil {
		w, err = r.Worktree()
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error while getting worktree: %w", err)
		}
	}

	if sha != "" {
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(sha),
		})
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error checking out SHA %s: %w", sha, err)
		}
	} else if tag != "" {
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewTagReferenceName(tag),
		})
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error checking out tag %s: %w", tag, err)
		}
	} else {
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
		})
		if err != nil {
			return domain.GitRepo{}, fmt.Errorf("error checking out branch %s: %w", branch, err)
		}
	}

	ref, err := r.Head()
	if err != nil {
		return domain.GitRepo{}, fmt.Errorf("error getting HEAD reference: %w", err)
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
