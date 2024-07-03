package artifacts

import (
	"context"
	"fmt"
	"os/exec"
)

type DockerArtifact struct {
	registry string
}

type Args struct {
	Name       string
	Path       string
	Dockerfile string
}

func NewDockerArtifactory(registry string) *DockerArtifact {
	return &DockerArtifact{
		registry: registry,
	}
}

func (a *DockerArtifact) Build(ctx context.Context, args Args) (string, error) {
	// build
	tag := fmt.Sprintf("%s:latest", args.Name)
	buildCmd := exec.Command("docker", "build", "-t", tag, args.Path)
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Println(buildOut)
		return "", fmt.Errorf("failed to build docker image: %w", err)
	}

	// tag
	registry := fmt.Sprintf("registry.digitalocean.com/%s/%s", a.registry, tag)
	if err := exec.Command("docker", "tag", tag, registry).Run(); err != nil {
		return "", fmt.Errorf("failed to tag docker image: %w", err)
	}

	// push
	if err := exec.Command("docker", "push", registry).Run(); err != nil {
		return "", fmt.Errorf("failed to push docker image: %w", err)
	}

	return registry, nil
}
