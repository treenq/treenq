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
	Name string
	Path string
}

func (a *DockerArtifact) Build(ctx context.Context, args Args) error {
	tag := fmt.Sprintf("%s:latest", args.Name)
	if err := exec.Command("docker", "build", "-t", tag, args.Path).Run(); err != nil {
		return fmt.Errorf("failed to build docker image: %w", err)
	}

	return nil
}
func (a *DockerArtifact) Deliver(ctx context.Context, args Args) error {
	// tag
	tag := fmt.Sprintf("%s:latest", args.Name)
	registry := fmt.Sprintf("registry.digitalocean.com/%s/%s", a.registry, tag)
	if err := exec.Command("docker", "tag", tag, registry).Run(); err != nil {
		return fmt.Errorf("failed to tag docker image: %w", err)
	}

	// push
	if err := exec.Command("docker", "push", registry).Run(); err != nil {
		return fmt.Errorf("failed to push docker image: %w", err)
	}

	return nil
}
