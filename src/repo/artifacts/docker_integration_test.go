package artifacts

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	tqsdk "github.com/treenq/treenq/pkg/sdk"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"

	"github.com/treenq/treenq/src/domain"
)

type testCase struct {
	name              string
	registryPort      string
	buildkitTLSCA     string
	registryTLSVerify bool
	registryCert      string
	registryUsername  string
	registryPassword  string
}

func TestDockerArtifact_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	// Disable testcontainers reaper which has issues with Colima or buildkit
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	testCases := []testCase{
		{
			name:              "no_tls_no_auth",
			registryPort:      "15000",
			buildkitTLSCA:     "",
			registryTLSVerify: false,
			registryCert:      "",
			registryUsername:  "",
			registryPassword:  "",
		},
		{
			name:              "tls_no_auth",
			registryPort:      "15001",
			buildkitTLSCA:     "",
			registryTLSVerify: true,
			registryCert:      filepath.Join("testdata", "certs", "ca.crt"),
			registryUsername:  "",
			registryPassword:  "",
		},
		{
			name:              "tls_with_auth",
			registryPort:      "15002",
			buildkitTLSCA:     "",
			registryTLSVerify: true,
			registryCert:      filepath.Join("testdata", "certs", "ca.crt"),
			registryUsername:  "testuser",
			registryPassword:  "testpass",
		},
		{
			name:              "no_tls_with_auth",
			registryPort:      "15003",
			buildkitTLSCA:     "",
			registryTLSVerify: false,
			registryCert:      "",
			registryUsername:  "testuser",
			registryPassword:  "testpass",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runDockerArtifactTest(t, tc)
		})
	}
}

func runDockerArtifactTest(t *testing.T, tc testCase) {
	ctx := context.Background()
	composeFilePaths := []string{
		filepath.Join("testdata", "docker-compose.yaml"),
		filepath.Join("testdata", fmt.Sprintf("docker-compose.%s.yaml", tc.name)),
	}
	composeStack, err := compose.NewDockerCompose(composeFilePaths...)
	require.NoError(t, err, "Failed to create docker compose")

	t.Cleanup(func() {
		err := composeStack.Down(ctx, compose.RemoveOrphans(true))
		require.NoError(t, err, "Failed to tear down compose stack")
	})

	err = composeStack.Up(ctx, compose.Wait(true))
	require.NoError(t, err, "Failed to start docker compose")

	// BuildKit uses host networking, so it's on localhost:1234
	dockerArtifact, err := NewDockerArtifactory(
		"tcp://localhost:1234",
		tc.buildkitTLSCA,
		fmt.Sprintf("localhost:%s", tc.registryPort),
		tc.registryTLSVerify,
		tc.registryCert,
		tc.registryUsername,
		tc.registryPassword,
	)
	require.NoError(t, err, "Failed to create docker artifact")

	tag := fmt.Sprintf("test-tag-%s", tc.name)
	deployment := domain.AppDeployment{
		Space: tqsdk.Space{
			Service: tqsdk.Service{
				Name: "test-app",
			},
		},
		BuildTag: tag,
	}

	// Test Inspect operation - should return ErrImageNotFound for non-existent image
	_, err = dockerArtifact.Inspect(ctx, deployment)
	require.ErrorIs(t, err, domain.ErrImageNotFound, "Expected ErrImageNotFound for non-existent image")

	testDataDir, err := filepath.Abs("testdata")
	require.NoError(t, err, "Failed to get absolute path for testdata")

	dockerfilePath := filepath.Join(testDataDir, "Dockerfile.busybox")

	buildArgs := domain.BuildArtifactRequest{
		Name:          "test-app",
		Tag:           tag,
		DockerContext: testDataDir,
		Dockerfile:    dockerfilePath,
		DeploymentID:  fmt.Sprintf("test-deployment-%s", tc.name),
	}

	// Test Build operation
	progress := domain.NewProgressBuf()
	builtImage, err := dockerArtifact.Build(ctx, buildArgs, progress)
	require.NoError(t, err, "Failed to build image")

	expectedImage := dockerArtifact.Image("test-app", tag)
	require.Equal(t, expectedImage.FullPath(), builtImage.FullPath(), "Built image path mismatch")

	// Test Inspect operation - should now find the built image
	inspectedImage, err := dockerArtifact.Inspect(ctx, deployment)
	require.NoError(t, err, "Failed to inspect built image")
	require.Equal(t, expectedImage.FullPath(), inspectedImage.FullPath(), "Inspected image path mismatch")
}
