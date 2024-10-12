package artifacts

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/treenq/treenq/src/domain"
)

type DockerArtifact struct {
	registry string
}

func NewDockerArtifactory(registry string) *DockerArtifact {
	return &DockerArtifact{
		registry: registry,
	}
}

func (a *DockerArtifact) Image(args domain.BuildArtifactRequest) domain.Image {
	return domain.Image{
		Registry:   a.registry,
		Repository: args.Name,
		Tag:        args.Tag,
	}
}

func (a *DockerArtifact) Build(ctx context.Context, args domain.BuildArtifactRequest) (domain.Image, error) {
	image := a.Image(args)

	buildCmd := exec.Command("docker", "build", "-t", image.Image(), "-f", args.Dockerfile, args.Path)
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return image, fmt.Errorf("failed to build docker image: %s: %w", string(buildOut), err)
	}

	if buildOut, err := exec.Command("docker", "tag", image.Image(), image.FullPath()).CombinedOutput(); err != nil {
		return image, fmt.Errorf("failed to tag docker image: %s: %w", string(buildOut), err)
	}

	if buildOut, err := exec.Command("docker", "push", image.FullPath()).CombinedOutput(); err != nil {
		return image, fmt.Errorf("failed to push docker image: %s: %w", string(buildOut), err)
	}

	return image, nil
}
