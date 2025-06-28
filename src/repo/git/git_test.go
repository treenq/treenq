package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treenq/treenq/src/domain"
)

type TestingRepository struct {
	ID   string
	Path string
}

func (r TestingRepository) CloneURL() string {
	return "file://" + r.Path
}

func (r TestingRepository) Location(root string) string {
	return filepath.Join(root, r.ID)
}

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
	gitComponent := NewGit(reposDir)

	repo := TestingRepository{
		ID:   "1",
		Path: mockRepoPath,
	}
	os.RemoveAll(repo.Location(reposDir))

	_, err = gitComponent.Clone(repo, "", "", "", "")
	assert.Equal(t, domain.ErrNoGitCheckoutSpecified, err, "must give an error if no branch or sha passed")

	_, err = gitComponent.Clone(repo, "", "main", "1234", "")
	assert.Equal(t, domain.ErrGitBranchAndShaMutuallyExclusive, err, "must give an error if branch AND sha passed")

	_, err = gitComponent.Clone(repo, "", "main", "", "v1.0.0")
	assert.Equal(t, domain.ErrGitBranchAndShaMutuallyExclusive, err, "must give an error if branch AND tag passed")

	_, err = gitComponent.Clone(repo, "", "", "1234", "v1.0.0")
	assert.Equal(t, domain.ErrGitBranchAndShaMutuallyExclusive, err, "must give an error if sha AND tag passed")

	_, err = gitComponent.Clone(repo, "", "main", "1234", "v1.0.0")
	assert.Equal(t, domain.ErrGitBranchAndShaMutuallyExclusive, err, "must give an error if all three passed")

	firstGitRepo, err := gitComponent.Clone(repo, "dummy-access-token", "master", "", "")
	require.NoError(t, err)
	defer os.RemoveAll(firstGitRepo.Dir)
	assert.Equal(t, len(firstGitRepo.Sha), 40)
	initialSHA := firstGitRepo.Sha

	clonedReadmePath := filepath.Join(firstGitRepo.Dir, "README.md")
	_, err = os.Stat(clonedReadmePath)
	require.NoError(t, err)

	addCommit(t, worktree, mockRepoPath)
	sameGitRepo, err := gitComponent.Clone(repo, "dummy-access-token", "master", "", "")
	require.NoError(t, err)
	defer os.RemoveAll(sameGitRepo.Dir) // Clean up
	latestSHA := sameGitRepo.Sha

	clonedNewFilePath := filepath.Join(sameGitRepo.Dir, "NEW_FILE.md")
	_, err = os.Stat(clonedNewFilePath)
	assert.NoError(t, err)
	assert.Equal(t, len(sameGitRepo.Sha), 40)

	// --- Checkout to the initial commit and verify ---
	checkoutRepo, err := gitComponent.Clone(repo, "dummy-access-token", "", initialSHA, "")
	require.NoError(t, err)
	defer os.RemoveAll(checkoutRepo.Dir)
	assert.Equal(t, initialSHA, checkoutRepo.Sha)

	// README.md should exist (from initial commit)
	readmePath := filepath.Join(checkoutRepo.Dir, "README.md")
	_, err = os.Stat(readmePath)
	assert.NoError(t, err)

	// NEW_FILE.md should NOT exist (added in later commit)
	newFilePath := filepath.Join(checkoutRepo.Dir, "NEW_FILE.md")
	_, err = os.Stat(newFilePath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	// --- Checkout to a master branch
	checkoutRepo, err = gitComponent.Clone(repo, "dummy-access-token", "master", "", "")
	require.NoError(t, err)
	defer os.RemoveAll(checkoutRepo.Dir)
	assert.Equal(t, latestSHA, checkoutRepo.Sha)
	assert.NotEmpty(t, checkoutRepo.Message, "expected not empty commit message")

	// README.md should exist (from initial commit)
	readmePath = filepath.Join(checkoutRepo.Dir, "README.md")
	_, err = os.Stat(readmePath)
	assert.NoError(t, err)

	// NEW_FILE.md should exist (added in later commit)
	newFilePath = filepath.Join(checkoutRepo.Dir, "NEW_FILE.md")
	_, err = os.Stat(newFilePath)
	assert.NoError(t, err)

	// Test tag checkout
	gitRepo, err := git.PlainOpen(mockRepoPath)
	require.NoError(t, err)
	addTag(t, gitRepo, "v1.0.0")

	tagSHA := latestSHA

	// Add another commit after the tag to make sure tag checkout works correctly
	addThirdCommit(t, worktree, mockRepoPath)
	postTagRepo, err := gitComponent.Clone(repo, "dummy-access-token", "master", "", "")
	require.NoError(t, err)
	defer os.RemoveAll(postTagRepo.Dir)
	postTagSHA := postTagRepo.Sha
	assert.NotEqual(t, tagSHA, postTagSHA, "post-tag commit should have different SHA")

	tagCheckoutRepo, err := gitComponent.Clone(TestingRepository{ID: "tag-test", Path: mockRepoPath}, "dummy-access-token", "", "", "v1.0.0")
	require.NoError(t, err)
	defer os.RemoveAll(tagCheckoutRepo.Dir)
	assert.Equal(t, tagSHA, tagCheckoutRepo.Sha)

	// Both files should exist at tag
	readmePath = filepath.Join(tagCheckoutRepo.Dir, "README.md")
	_, err = os.Stat(readmePath)
	assert.NoError(t, err)
	newFilePath = filepath.Join(tagCheckoutRepo.Dir, "NEW_FILE.md")
	_, err = os.Stat(newFilePath)
	assert.NoError(t, err)
	// new commit is not on the tag
	thirdFilePath := filepath.Join(tagCheckoutRepo.Dir, "THIRD_FILE.md")
	_, err = os.Stat(thirdFilePath)
	assert.True(t, os.IsNotExist(err), "file shouldn't exist")
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

func addThirdCommit(t *testing.T, worktree *git.Worktree, path string) {
	thirdFilePath := filepath.Join(path, "THIRD_FILE.md")
	err := os.WriteFile(thirdFilePath, []byte("# Third File"), 0644)
	require.NoError(t, err)
	_, err = worktree.Add("THIRD_FILE.md")
	require.NoError(t, err)
	_, err = worktree.Commit("Add third file", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
}

func addTag(t *testing.T, repo *git.Repository, tagName string) {
	head, err := repo.Head()
	require.NoError(t, err)

	_, err = repo.CreateTag(tagName, head.Hash(), &git.CreateTagOptions{
		Message: "Test tag",
		Tagger: &object.Signature{
			Name:  "Test Author",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)
}
