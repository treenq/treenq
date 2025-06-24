package artifacts

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"os" // For os.Getenv
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// imageNameBase is the base name for the image, the registry prefix will be added.
	imageNameBase   = "simple-app"
	imageTag        = "latest"
	buildContextDir = "." // Assuming Dockerfile.simple, index.html are in the same directory as the test
)

func TestDockerIntegrationWithTestcontainers(t *testing.T) {
	ctx := context.Background()
	dockerHost := os.Getenv("DOCKER_HOST")
	t.Logf("DOCKER_HOST environment variable: %s", dockerHost)

	// Check for Docker availability and permissions early
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("Failed to create Docker client for pre-check: %v", err)
	}
	defer dockerCli.Close()
	if _, err := dockerCli.Ping(ctx); err != nil {
		t.Skipf("Docker daemon is not accessible or responding, skipping integration test: %v", err)
	}
	t.Log("Docker daemon ping successful, proceeding with test.")


	// 1. Setup BuildKit container (optional, if not relying on host's default BuildKit)
	// For this example, we'll assume the host Docker daemon has BuildKit enabled or is sufficient.
	// Setting up a separate BuildKit container via Testcontainers can be complex due to client targeting.
	// If a separate BuildKit instance is strictly needed, this part would need more configuration
	// (e.g., custom Docker client pointing to the Testcontainers BuildKit).
	// For now, we proceed assuming host Docker's build capabilities are used.
	t.Log("Skipping dedicated BuildKit container setup, relying on host Docker's build capabilities.")

	// 2. Setup Docker Registry container
	registryContainer, err := registry.Run(ctx, "registry:2")
	if err != nil {
		t.Fatalf("Failed to start registry container: %v", err)
	}
	defer func() {
		if err := registryContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate registry container: %v", err)
		}
	}()

	registryHost, err := registryContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get registry host: %v", err)
	}
	registryPort, err := registryContainer.MappedPort(ctx, "5000/tcp")
	if err != nil {
		t.Fatalf("Failed to get registry mapped port: %v", err)
	}
	dynamicImageName := fmt.Sprintf("%s:%s/%s:%s", registryHost, registryPort.Port(), imageNameBase, imageTag)
	t.Logf("Using dynamic image name: %s", dynamicImageName)

	// Allow time for the registry to be fully up and ready.
	// A more robust wait strategy could involve trying to push/pull a tiny image.
	time.Sleep(5 * time.Second)


	// 3. Initialize Docker client (connects to host Docker daemon)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// 4. Check image does not exist initially
	_, _, err = cli.ImageInspectWithRaw(ctx, dynamicImageName)
	if err == nil {
		t.Logf("Image %s unexpectedly found. Attempting to remove it before test.", dynamicImageName)
		_, removeErr := cli.ImageRemove(ctx, dynamicImageName, image.RemoveOptions{Force: true, PruneChildren: true})
		if removeErr != nil {
			t.Fatalf("Image %s found and failed to remove it: %v", dynamicImageName, removeErr)
		}
		_, _, err = cli.ImageInspectWithRaw(ctx, dynamicImageName)
	}
	if !client.IsErrNotFound(err) {
		t.Fatalf("Expected ImageNotFound error for %s before build, but got: %v", dynamicImageName, err)
	}
	t.Logf("Image %s correctly not found before build.", dynamicImageName)

	// 5. Build the image
	t.Logf("Building image %s from context %s...", dynamicImageName, buildContextDir)

	// Define absolute path for build context
	absBuildContextDir, err := filepath.Abs(buildContextDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path for build context: %v", err)
	}
	t.Logf("Absolute build context directory: %s", absBuildContextDir)


	buildCtxTar, err := archive.TarWithOptions(absBuildContextDir, &archive.TarOptions{
		ExcludePatterns: []string{
			"docker_integration_test.go", // This file itself
			"go.mod",
			"go.sum",
			"*.md",
			// Ensure only Dockerfile.simple and index.html are included from the context
		},
	})
	if err != nil {
		t.Fatalf("Failed to create build context tar: %v", err)
	}

	buildOptions := types.ImageBuildOptions{
		Dockerfile: "Dockerfile.simple", // Relative to the root of the tarred context
		Tags:       []string{dynamicImageName},
		Remove:     true,
		PullParent: true,
		// BuildKit specific options could be added here if needed, e.g., platform
	}

	resp, err := cli.ImageBuild(ctx, buildCtxTar, buildOptions)
	if err != nil {
		t.Fatalf("Failed to build image: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), "error") {
			t.Logf("Build output (potential error): %s", line)
		}
		// t.Logf("Build log: %s", line) // Uncomment for full build log
	}
	if err := scanner.Err(); err != nil {
		t.Logf("Error reading build response: %v", err)
	}

	inspectAfterBuild, _, err := cli.ImageInspectWithRaw(ctx, dynamicImageName)
	if err != nil {
		// Read full build log if inspect fails, to aid debugging
		// This requires re-opening the build context if resp.Body was already fully read by scanner
		// For simplicity, we're not re-reading here.
		t.Fatalf("Failed to inspect image %s after build: %v. Check build logs.", dynamicImageName, err)
	}
	t.Logf("Image %s built successfully. ID: %s", dynamicImageName, inspectAfterBuild.ID)

	// 6. Push the image to the Testcontainers-managed local registry
	t.Logf("Pushing image %s to registry %s...", dynamicImageName, dynamicImageName)
	// For insecure local registries (like the one Testcontainers sets up by default),
	// the Docker daemon needs to be configured to trust it.
	// Testcontainers usually handles this if the Docker client is configured correctly or if it's a known pattern.
	// If pushes fail due to HTTPS errors, daemon configuration for "insecure-registries" might be needed
	// on the host where tests run, or the registry Testcontainer needs to be setup with TLS.
	// The `registry` module for Testcontainers might handle some of this.
	pushResp, err := cli.ImagePush(ctx, dynamicImageName, image.PushOptions{RegistryAuth: "dummy"}) // Dummy auth for local registry
	if err != nil {
		t.Fatalf("Failed to push image %s: %v", dynamicImageName, err)
	}
	defer pushResp.Close()

	pushScanner := bufio.NewScanner(pushResp)
	for pushScanner.Scan() {
		// t.Logf("Push log: %s", pushScanner.Text()) // Uncomment for full push log
	}
	if err := pushScanner.Err(); err != nil {
		t.Logf("Error reading push response: %v", err)
	}
	t.Logf("Image %s push initiated.", dynamicImageName)

	// 7. Verify the image exists in the registry by pulling it
	t.Logf("Verifying image presence in registry by removing local and pulling: %s", dynamicImageName)
	_, err = cli.ImageRemove(ctx, dynamicImageName, image.RemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		t.Fatalf("Failed to remove local image %s for pull test: %v", dynamicImageName, err)
	}
	t.Logf("Removed local copy of %s to test pull from registry.", dynamicImageName)

	pullResp, err := cli.ImagePull(ctx, dynamicImageName, image.PullOptions{RegistryAuth: "dummy"})
	if err != nil {
		t.Fatalf("Failed to pull image %s from registry: %v", dynamicImageName, err)
	}
	defer pullResp.Close()

	pullScan := bufio.NewScanner(pullResp)
	for pullScan.Scan() {
		// t.Logf("Pull log: %s", pullScan.Text()) // Uncomment for full pull log
	}
	if err := pullScan.Err(); err != nil {
		t.Fatalf("Error reading pull response: %v", err)
	}

	_, _, err = cli.ImageInspectWithRaw(ctx, dynamicImageName)
	if err != nil {
		t.Fatalf("Failed to inspect image %s after pulling from registry: %v", dynamicImageName, err)
	}
	t.Logf("Image %s successfully pulled from registry and inspected.", dynamicImageName)

	t.Log("Docker integration test with Testcontainers completed successfully.")
}

// Helper function to set up a generic BuildKit container if needed in the future.
// This is more involved due to Docker client configuration ( DOCKER_HOST )
// and ensuring the main Docker client used for builds targets this BuildKit.
// For now, this is unused.
func setupBuildkitContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "moby/buildkit:stable",
		ExposedPorts: []string{"1234/tcp"}, // Default BuildKit port
		Privileged:   true,                 // BuildKit needs privileged mode
		WaitingFor:   wait.ForLog("serving grpc traffic on"),
		// Potentially needs custom entrypoint or command if the default doesn't expose the port correctly for TCP.
		// Also, mounting /var/lib/buildkit for persistence if desired across test runs (not typical for tests).
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start BuildKit container: %v", err)
	}

	// Get host and port for DOCKER_HOST
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get BuildKit container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "1234/tcp")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get BuildKit container mapped port: %v", err)
	}
	buildkitdAddress := fmt.Sprintf("tcp://%s:%s", host, port.Port())
	t.Logf("BuildKit container started at %s", buildkitdAddress)

	// Important: The Docker client used for `cli.ImageBuild` would need to be configured
	// to use this DOCKER_HOST (e.g., by setting os.Setenv("DOCKER_HOST", buildkitdAddress)
	// before client initialization, or by creating a new client with this host).
	// This adds complexity and is why relying on host Docker's BuildKit is simpler initially.

	return container, buildkitdAddress
}
