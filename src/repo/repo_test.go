package repo

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestClone(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-repo-clone")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockRepoPath := filepath.Join(tempDir, "mock-repo")
	repo, err := git.PlainInit(mockRepoPath, false)
	if err != nil {
		t.Fatalf("Failed to initialize mock git repo: %v", err)
	}

	readmePath := filepath.Join(mockRepoPath, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository"), 0644); err != nil {
		t.Fatalf("Failed to write file in mock repo: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}
	_, err = worktree.Add("README.md")
	if err != nil {
		t.Fatalf("Failed to add file to worktree: %v", err)
	}
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Failed to commit file: %v", err)
	}

	gitUtil := NewGit()

	repoURL := "file://" + mockRepoPath

	cloneDir, err := gitUtil.Clone(repoURL, "dummy-access-token")
	if err != nil {
		t.Fatalf("First clone failed: %v", err)
	}
	defer os.RemoveAll(cloneDir)

	clonedReadmePath := filepath.Join(cloneDir, "README.md")
	if _, err := os.Stat(clonedReadmePath); os.IsNotExist(err) {
		t.Fatalf("README.md file not found in clone destination after first clone")
	}

	newFilePath := filepath.Join(mockRepoPath, "NEW_FILE.md")
	if err := os.WriteFile(newFilePath, []byte("# New File"), 0644); err != nil {
		t.Fatalf("Failed to write new file in mock repo: %v", err)
	}
	_, err = worktree.Add("NEW_FILE.md")
	if err != nil {
		t.Fatalf("Failed to add new file to worktree: %v", err)
	}
	_, err = worktree.Commit("Add new file", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Failed to commit new file: %v", err)
	}

	secondCloneDir, err := gitUtil.Clone(repoURL, "dummy-access-token")
	if err != nil {
		t.Fatalf("Second clone (pull) failed: %v", err)
	}
	defer os.RemoveAll(secondCloneDir) // Clean up

	clonedNewFilePath := filepath.Join(secondCloneDir, "NEW_FILE.md")
	if _, err := os.Stat(clonedNewFilePath); os.IsNotExist(err) {
		t.Fatalf("NEW_FILE.md file not found in clone destination after pull")
	}
}
