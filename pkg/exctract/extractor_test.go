package extract

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

//go:embed testdata/tq.go
var testBuildConfig []byte

func TestExtractor_ExtractConfig(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "test_repo")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(srcDir); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()

	tqDir := filepath.Join(srcDir, tqRelativePath)
	err = os.MkdirAll(tqDir, 0766)
	require.NoError(t, err)

	tmpFile := filepath.Join(tqDir, tqBuildLauncherFile)
	if err := os.WriteFile(tmpFile, testBuildConfig, 0766); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	builderDir := filepath.Join(filepath.Dir(currentDir), "builder")
	extractor, close, err := NewExtractor(builderDir)
	require.NoError(t, err)

	// check the builder directory is created
	_, err = os.Stat(builderDir)
	assert.ErrorIs(t, err, nil)

	resource, err := extractor.ExtractConfig(srcDir)
	assert.ErrorIs(t, err, nil)
	assert.Equal(t, resource, tqsdk.Resource{
		App: tqsdk.App{
			Name:         "name",
			Port:         ":8000",
			BuildCommand: "go build some/thing.go",
			RunCommand:   "./thing",
		},
	})

	// check the close removes the builder directory
	close()
	_, err = os.Stat(builderDir)
	assert.ErrorIs(t, err, os.ErrNotExist)
}
