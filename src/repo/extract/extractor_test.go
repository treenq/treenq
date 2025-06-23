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

//go:embed testdata/tq.json
var testBuildConfig []byte

func TestExtractor_ExtractConfig(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "test_repo")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(srcDir); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()

	tqConfigFile := filepath.Join(srcDir, tqRelativePath)
	require.NoError(t, err)

	if err := os.WriteFile(tqConfigFile, testBuildConfig, 0766); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	extractor := NewExtractor()

	resource, err := extractor.ExtractConfig(srcDir)
	assert.ErrorIs(t, err, nil)
	assert.Equal(t, resource, tqsdk.Space{
		Service: tqsdk.Service{
			DockerfilePath: "Dockerfile",
			Name:           "treenq-e2e-sample",
			HttpPort:       8000,
			Replicas:       1,
		},
	})
}
