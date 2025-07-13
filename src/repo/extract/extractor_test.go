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

//go:embed testdata/tq.yaml
var testBuildConfigYaml []byte

func TestExtractor_ExtractConfig(t *testing.T) {
	type testCase struct {
		name string
		path string
	}

	expectedSpace := tqsdk.Space{
		Service: tqsdk.Service{
			DockerfilePath: "Dockerfile",
			DockerContext:  ".",
			Name:           "treenq-e2e-sample",
			HttpPort:       8000,
			Replicas:       1,
			ComputationResource: tqsdk.ComputationResource{
				CpuUnits:   1000,
				MemoryMibs: 2048,
				DiskGibs:   20,
			},
		},
	}

	for _, tt := range []testCase{
		{
			name: "extract json",
			path: tqJsonPath,
		},
		{
			name: "extract yaml",
			path: tqYamlPath,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srcDir, err := os.MkdirTemp("", "test_repo")
			require.NoError(t, err)
			defer func() {
				if err := os.RemoveAll(srcDir); err != nil {
					t.Fatalf("Failed to remove temp directory: %v", err)
				}
			}()

			tqConfigFile := filepath.Join(srcDir, tt.path)
			require.NoError(t, err)

			if err := os.WriteFile(tqConfigFile, testBuildConfig, 0766); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			extractor := NewExtractor()

			resource, err := extractor.ExtractConfig(srcDir)
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, expectedSpace, resource)
		})
	}
}

func TestExtractor_ExtractConfigFallback(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "test_repo")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(srcDir); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()

	// Only create tq.yaml, not tq.json
	tqConfigFile := filepath.Join(srcDir, tqYamlPath)
	require.NoError(t, err)

	if err := os.WriteFile(tqConfigFile, testBuildConfigYaml, 0766); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	extractor := NewExtractor()

	resource, err := extractor.ExtractConfig(srcDir)
	assert.ErrorIs(t, err, nil)
	assert.Equal(t, resource, tqsdk.Space{
		Service: tqsdk.Service{
			DockerfilePath: "Dockerfile",
			DockerContext:  ".",
			Name:           "treenq-e2e-sample",
			HttpPort:       8000,
			Replicas:       1,
			ComputationResource: tqsdk.ComputationResource{
				CpuUnits:   1000,
				MemoryMibs: 2048,
				DiskGibs:   20,
			},
		},
	})
}

func TestExtractor_ExtractConfigJsonPriority(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "test_repo")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(srcDir); err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create both files, JSON should take priority
	tqJsonFile := filepath.Join(srcDir, tqJsonPath)
	tqYamlFile := filepath.Join(srcDir, tqYamlPath)

	if err := os.WriteFile(tqJsonFile, testBuildConfig, 0766); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Create YAML with different name to verify JSON is used
	yamlWithDifferentName := []byte(`service:
  dockerfilePath: Dockerfile
  sizeSlug: 250-512-1
  name: different-name
  httpPort: 8000
  replicas: 1`)

	if err := os.WriteFile(tqYamlFile, yamlWithDifferentName, 0766); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	extractor := NewExtractor()

	resource, err := extractor.ExtractConfig(srcDir)
	assert.ErrorIs(t, err, nil)
	// Should use JSON file (with name "treenq-e2e-sample"), not YAML (with name "different-name")
	assert.Equal(t, "treenq-e2e-sample", resource.Service.Name)
}
