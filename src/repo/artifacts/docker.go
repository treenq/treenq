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

func (a *DockerArtifact) Build(ctx context.Context, args domain.BuildArtifactRequest) (domain.Image, error) {
	image := domain.Image{
		Registry:   a.registry,
		Repository: args.Name,
		Tag:        "latest",
	}

	// build
	buildCmd := exec.Command("docker", "build", "-t", image.Image(), args.Path)
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println(buildOut)
		return image, fmt.Errorf("failed to build docker image: %w", err)
	}

	registry := fmt.Sprintf("registry.digitalocean.com/%s", image.FullPath())
	// tag
	if err := exec.Command("docker", "tag", image.Image(), registry).Run(); err != nil {
		return image, fmt.Errorf("failed to tag docker image: %w", err)
	}

	// push
	if err := exec.Command("docker", "push", registry).Run(); err != nil {
		return image, fmt.Errorf("failed to push docker image: %w", err)
	}

	return image, nil
}
