package repo

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-repo-clone")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockRepoPath := filepath.Join(tempDir, "mock-repo")
	worktree := newRepo(t, mockRepoPath)

	wd, err := os.Getwd()
	require.NoError(t, err)
	reposDir := filepath.Join(wd, "repos")
	gitUtil := NewGit(reposDir)

	repoURL := "file://" + mockRepoPath

	cloneDir, err := gitUtil.Clone(repoURL, "dummy-access-token")
	require.NoError(t, err)
	defer os.RemoveAll(cloneDir)

	clonedReadmePath := filepath.Join(cloneDir, "README.md")
	_, err = os.Stat(clonedReadmePath)
	require.NoError(t, err)

	addCommit(t, worktree, mockRepoPath)
	secondCloneDir, err := gitUtil.Clone(repoURL, "dummy-access-token")
	require.NoError(t, err)
	defer os.RemoveAll(secondCloneDir) // Clean up

	clonedNewFilePath := filepath.Join(secondCloneDir, "NEW_FILE.md")
	_, err = os.Stat(clonedNewFilePath)
	assert.NoError(t, err)
}

func newRepo(t *testing.T, path string) *git.Worktree {
	repo, err := git.PlainInit(path, false)
	require.NoError(t, err)

	readmePath := filepath.Join(path, "README.md")
	err = os.WriteFile(readmePath, []byte("# Test Repository"), 0644)
	require.NoError(t, err)

	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("README.md")
	require.NoError(t, err)
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	return worktree
}

func addCommit(t *testing.T, worktree *git.Worktree, path string) {
	newFilePath := filepath.Join(path, "NEW_FILE.md")
	err := os.WriteFile(newFilePath, []byte("# New File"), 0644)
	require.NoError(t, err)
	_, err = worktree.Add("NEW_FILE.md")
	require.NoError(t, err)
	_, err = worktree.Commit("Add new file", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
}
