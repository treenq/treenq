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

type Image struct {
	// Registry is a registry name in the cloud provider
	Registry string
	// Repository is a facto name of the image
	Repository string
	// Tag is a version of the image
	Tag string
}

func (i Image) Image() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
}

func (i Image) FullPath() string {
	return fmt.Sprintf("%s/%s:%s", i.Registry, i.Repository, i.Tag)
}

func (a *DockerArtifact) Build(ctx context.Context, args Args) (Image, error) {
	image := Image{
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
