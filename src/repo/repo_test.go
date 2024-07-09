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
	// Create a temporary directory for the mock repository
	tempDir, err := os.MkdirTemp("", "test-repo-clone")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Initialize a new git repository in the mock repo path
	mockRepoPath := filepath.Join(tempDir, "mock-repo")
	repo, err := git.PlainInit(mockRepoPath, false)
	if err != nil {
		t.Fatalf("Failed to initialize mock git repo: %v", err)
	}

	// Create a file in the repository to simulate actual repo content
	readmePath := filepath.Join(mockRepoPath, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repository"), 0644); err != nil {
		t.Fatalf("Failed to write file in mock repo: %v", err)
	}

	// Add the file to the repository and commit
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

	// Create an instance of your Git struct
	gitUtil := NewGit()

	// Use the file URL for local repository
	repoURL := "file://" + mockRepoPath

	// Clone the repository for the first time
	// cloneDir := filepath.Join(tempDir, "clone-destination")
	cloneDir, err := gitUtil.Clone(repoURL, "dummy-access-token")
	if err != nil {
		t.Fatalf("First clone failed: %v", err)
	}
	defer os.RemoveAll(cloneDir) // Clean up

	// Verify that the clone was successful by checking for the README.md file
	clonedReadmePath := filepath.Join(cloneDir, "README.md")
	if _, err := os.Stat(clonedReadmePath); os.IsNotExist(err) {
		t.Fatalf("README.md file not found in clone destination after first clone")
	}

	// Create a new commit in the mock repository
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

	// Clone the repository again to trigger the pull scenario
	// secondCloneDir := filepath.Join(tempDir, "second-clone-destination")
	secondCloneDir, err := gitUtil.Clone(repoURL, "dummy-access-token")
	if err != nil {
		t.Fatalf("Second clone (pull) failed: %v", err)
	}
	defer os.RemoveAll(secondCloneDir) // Clean up

	// Verify that the pull was successful by checking for the NEW_FILE.md file
	clonedNewFilePath := filepath.Join(secondCloneDir, "NEW_FILE.md")
	if _, err := os.Stat(clonedNewFilePath); os.IsNotExist(err) {
		t.Fatalf("NEW_FILE.md file not found in clone destination after pull")
	}
}
