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

func TestDockerArtifact_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	composeFilePaths := []string{filepath.Join("testdata", "docker-compose-test.yaml")}
	composeStack, err := compose.NewDockerCompose(composeFilePaths...)
	require.NoError(t, err, "Failed to create docker compose")

	t.Cleanup(func() {
		err := composeStack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesAll)
		require.NoError(t, err, "Failed to tear down compose stack")
	})

	err = composeStack.Up(ctx, compose.Wait(true))
	require.NoError(t, err, "Failed to start docker compose")

	registryContainer, err := composeStack.ServiceContainer(ctx, "registry")
	require.NoError(t, err, "Failed to get registry container")

	buildkitContainer, err := composeStack.ServiceContainer(ctx, "buildkit")
	require.NoError(t, err, "Failed to get buildkit container")

	registryPort, err := registryContainer.MappedPort(ctx, "5005")
	require.NoError(t, err, "Failed to get registry port")

	buildkitPort, err := buildkitContainer.MappedPort(ctx, "1234")
	require.NoError(t, err, "Failed to get buildkit port")

	dockerArtifact, err := NewDockerArtifactory(
		fmt.Sprintf("tcp://localhost:%s", buildkitPort.Port()),
		"", // no TLS CA for test
		fmt.Sprintf("localhost:%s", registryPort.Port()),
		false, // TLS verify disabled for test
		"",    // no registry cert
		"testuser",
		"testpassword",
	)
	require.NoError(t, err, "Failed to create docker artifact")

	tag := "test-tag"
	deployment := domain.AppDeployment{
		Space: tqsdk.Space{
			Service: tqsdk.Service{
				Name: "test-app",
			},
		},
		BuildTag: tag,
	}

	_, err = dockerArtifact.Inspect(ctx, deployment)
	require.ErrorIs(t, err, domain.ErrImageNotFound, "Expected ErrImageNotFound for non-existent image, got")

	testDataDir, err := filepath.Abs("testdata")
	require.NoError(t, err, "Failed to get absolute path for testdata")

	dockerfilePath := filepath.Join(testDataDir, "Dockerfile.busybox")
	require.NoError(t, err, "Dockerfile not found")

	buildArgs := domain.BuildArtifactRequest{
		Name:          "test-app",
		Tag:           tag,
		DockerContext: testDataDir,
		Dockerfile:    dockerfilePath,
		DeploymentID:  "test-deployment-123",
	}

	progress := domain.NewProgressBuf()
	builtImage, err := dockerArtifact.Build(ctx, buildArgs, progress)
	require.NoError(t, err, "Failed to build image")

	expectedImage := dockerArtifact.Image("test-app", "test-tag")
	require.Equal(t, expectedImage.FullPath(), builtImage.FullPath(), "Built image path mismatch")

	inspectedImage, err := dockerArtifact.Inspect(ctx, deployment)
	require.NoError(t, err, "Failed to inspect built image")
	require.Equal(t, expectedImage.FullPath(), inspectedImage.FullPath(), "Inspected image path mismatch")
}
