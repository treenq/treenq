package artifacts

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/storage"
	is "github.com/containers/image/v5/storage"
	"github.com/treenq/treenq/src/domain"
)

type DockerArtifact struct {
	registry string

	store storage.Store
}

func NewDockerArtifactory(registry string) (*DockerArtifact, error) {
	buildStoreOptions, err := storage.DefaultStoreOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build buildah storage option: %w", err)
	}

	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to build buildah store: %w", err)
	}

	return &DockerArtifact{
		registry: registry,
		store:    buildStore,
	}, nil
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

	id, ref, err := imagebuildah.BuildDockerfiles(ctx, a.store, define.BuildOptions{
		ContextDirectory: args.Path,
		Registry:         a.registry,
		Output:           args.Name,
		Out:              os.Stdout,
		Err:              os.Stderr,
		ReportWriter:     os.Stdout,
		IgnoreFile:       "./.dockerignore",
		AdditionalTags:   []string{image.FullPath()},
	}, args.Dockerfile)
	if err != nil {
		return image, fmt.Errorf("failed to build a docker container: %w", err)
	}

	// TODO: do something with that
	_, _ = id, ref

	storeRef, err := is.Transport.ParseStoreReference(a.store, image.FullPath())
	if err != nil {
		return image, fmt.Errorf("failed to parse store reference: %w", err)
	}

	_, _, err = buildah.Push(ctx, image.FullPath(), storeRef, buildah.PushOptions{
		Store: a.store,
	})
	if err != nil {
		return image, fmt.Errorf("failed to push image: %w", err)
	}

	return image, nil
}
