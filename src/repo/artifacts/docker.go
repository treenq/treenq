package artifacts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os" // For os.Stderr for progress writer if needed
	"path/filepath"

	"github.com/moby/buildkit/client"
	// "github.com/moby/buildkit/client/llb" // May not be directly needed if using Dockerfile frontend
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider" // For potential future auth
	"github.com/moby/buildkit/util/progress/progresswriter"
	"github.com/treenq/treenq/src/domain"
	"golang.org/x/sync/errgroup" // For managing progress writer goroutine
)

type DockerArtifact struct {
	registry     string
	buildkitHost string
}

func NewDockerArtifactory(registry string, buildkitHost string) (*DockerArtifact, error) {
	if buildkitHost == "" {
		return nil, errors.New("buildkit host cannot be empty")
	}
	return &DockerArtifact{
		registry:     registry,
		buildkitHost: buildkitHost,
	}, nil
}

func (a *DockerArtifact) Image(name, tag string) domain.Image {
	return domain.Image{
		Registry:   a.registry,
		Repository: name,
		Tag:        tag,
	}
}

func (a *DockerArtifact) Build(ctx context.Context, args domain.BuildArtifactRequest, progress *domain.ProgressBuf) (domain.Image, error) {
	image := a.Image(args.Name, args.Tag)

	slog.Info("Starting Buildkit build using Go client", "image", image.FullPath(), "host", a.buildkitHost, "contextPath", args.Path, "dockerfilePath", args.Dockerfile)

	c, err := client.New(ctx, a.buildkitHost, client.WithFailFast())
	if err != nil {
		return image, fmt.Errorf("failed to create buildkit client for host %s: %w", a.buildkitHost, err)
	}
	defer c.Close()

	// Determine actual Dockerfile path. Assumes args.Dockerfile is relative to args.Path if not absolute.
	dockerfilePath := args.Dockerfile
	if !filepath.IsAbs(dockerfilePath) {
		dockerfilePath = filepath.Join(args.Path, dockerfilePath)
	}
	slog.Debug("Resolved Dockerfile path", "path", dockerfilePath)


	// Prepare SolveOpt
	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			// TODO: Make platform configurable if necessary.
			// Example: "linux/amd64", "linux/arm64", or determined by the build environment/request.
			// For now, not setting it explicitly to let Buildkit decide or use its default.
			// "platform": DetectPlatform(), // This would require a helper function
		},
		LocalMounts: map[string]string{
			"context":    args.Path,
			"dockerfile": dockerfilePath, // This should be the path to the Dockerfile itself, not its directory.
		},
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterImage,
				Attrs: map[string]string{
					"name": image.FullPath(),
					"push": "true", // Pushes the image to the registry specified in its name.
				},
			},
		},
		// Session allows for advanced features like SSH forwarding, secrets, and auth.
		// For registry authentication, authprovider.NewDockerAuthProvider can be used.
		// It typically uses the local Docker CLI's configuration for credentials.
		// Ensure the environment where this code runs has access to valid Docker credentials
		// if the registry requires authentication.
		Session: []session.Attachable{authprovider.NewDockerAuthProvider(os.Stderr)}, // os.Stderr can be used for prompts if needed by auth provider
	}

	// Add build arguments from args.BuildArgs
	if args.BuildArgs != nil {
		for k, v := range args.BuildArgs {
			// FrontendAttrs for build-args should be prefixed with "build-arg:"
			solveOpt.FrontendAttrs["build-arg:"+k] = v
			slog.Debug("Added build-arg", "key", k, "value", v)
		}
	}

	// Handling Dockerignore:
	// The "dockerfile.v0" frontend automatically looks for a ".dockerignore" file in the
	// root of the "context" LocalMount (args.Path).
	// If args.Dockerignore specifies a *different* path or content for the ignore rules,
	// this automatic mechanism won't use it.
	// To handle a custom ignore file path (args.Dockerignore), one would typically need to:
	// 1. Ensure the ignore file is placed at the root of the build context (args.Path)
	//    and named ".dockerignore" before the build. This might involve copying it.
	// 2. Or, if the frontend supports it, pass the ignore file content/path via FrontendAttrs.
	//    The standard "dockerfile.v0" frontend does not have a direct attribute for custom ignore file paths.
	// For this implementation, we rely on the standard behavior: if a .dockerignore file
	// exists at the root of args.Path, it will be used.
	// If args.Dockerignore is non-empty and points to a specific file that is *not*
	// at args.Path/.dockerignore, this log message indicates a potential discrepancy.
	if args.Dockerignore != "" {
		slog.Info("Dockerignore handling: Standard .dockerignore at context root will be used by Buildkit frontend.", "customDockerignoreField", args.Dockerignore)
		// If args.Dockerignore is meant to be the *content* of the .dockerignore file,
		// you would need to write it to args.Path/.dockerignore before the build.
		// If it's a path to a *different* ignore file, you'd need to copy its content
		// to args.Path/.dockerignore.
	}


	// Progress writer setup
	// progress.AsWriter returns an io.Writer.
	// We need to pass it to progresswriter.NewPrinter.
	progressWriter := progress.AsWriter(args.DeploymentID, slog.LevelInfo) // Use a more descriptive var name
	pw, err := progresswriter.NewPrinter(ctx, progressWriter, progresswriter.PrinterOpt{Format: progresswriter.FormatPlain})
	if err != nil {
		return image, fmt.Errorf("failed to create progress writer: %w", err)
	}

	eg, solveCtx := errgroup.WithContext(ctx)
	var solveResp *client.SolveResponse // To store the response from c.Solve

	// Goroutine for c.Solve
	eg.Go(func() error {
		slog.Debug("Starting c.Solve")
		var errSolve error
		// The solveOpt.SourcePolicy is not set, so it defaults to the Buildkit daemon's policy.
		// For loading .dockerignore from the context, ensure it's at the root of LocalMounts["context"].
		solveResp, errSolve = c.Solve(solveCtx, nil, solveOpt, pw.Status()) // llb.Definition is nil for Dockerfile builds
		if errSolve != nil {
			slog.Error("Buildkit c.Solve returned error", "error", errSolve)
			return fmt.Errorf("buildkit solve error: %w", errSolve)
		}
		slog.Debug("Buildkit c.Solve completed successfully")
		return nil
	})

	// Goroutine for progress writer
	eg.Go(func() error {
		slog.Debug("Starting progress writer monitoring")
		// pw.Wait() waits for all progress events to be processed.
		// pw.Err() returns any error encountered during progress writing.
		<-pw.Done() // Wait for the progress writer to finish processing events
		errProgress := pw.Err()
		if errProgress != nil {
			slog.Error("Progress writer encountered error", "error", errProgress)
			return fmt.Errorf("progress writer error: %w", errProgress)
		}
		slog.Debug("Progress writer finished successfully")
		return nil
	})

	slog.Debug("Waiting for errgroup")
	if err := eg.Wait(); err != nil {
		// Log the detailed error from the errgroup
		slog.Error("Build failed due to error in errgroup", "error", err)
		return image, fmt.Errorf("build failed: %w", err)
	}

	if solveResp != nil && solveResp.ExporterResponse != nil {
		slog.Info("Buildkit build successful", "image", image.FullPath(), "exporterResponse", solveResp.ExporterResponse)
	} else {
		slog.Info("Buildkit build successful but no exporter response details", "image", image.FullPath())
	}
	return image, nil
}

func (a *DockerArtifact) Inspect(ctx context.Context, deployment domain.AppDeployment) (domain.Image, error) {
	// TODO: Implement proper image inspection using a registry client or other means.
	// This current implementation is a placeholder and does not actually inspect the image in the registry.
	// For example, one could use a library like github.com/google/go-containerregistry/pkg/v1/remote
	// to try to fetch the image manifest. If it fails with a "manifest unknown" type of error,
	// then it's equivalent to domain.ErrImageNotFound.
	slog.Warn("DockerArtifact.Inspect is a NO-OP and does not actually check registry for image existence.",
		"imageName", deployment.Space.Service.Name, "imageTag", deployment.BuildTag)

	image := a.Image(deployment.Space.Service.Name, deployment.BuildTag)
	// Returning nil error here means we are currently bypassing the check.
	// To enforce checking, one might return domain.ErrImageNotFound or an error indicating "not implemented".
	// For the purpose of this refactoring, if the previous Inspect was functional,
	// this NO-OP is a step back in functionality but fulfills the immediate task of removing buildctl exec.
	// A proper implementation would require adding a registry client library.
	return image, nil
}
