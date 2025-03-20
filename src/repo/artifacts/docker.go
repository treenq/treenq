package artifacts

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/image/v5/docker"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
	"github.com/treenq/treenq/src/domain"
)

func init() {
	if buildah.InitReexec() {
		return
	}
	unshare.MaybeReexecUsingUserNamespace(false)
}

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

	id, ref, err := imagebuildah.BuildDockerfiles(context.Background(), a.store, define.BuildOptions{
		Registry:       "jopa",
		Output:         "jopa2",
		Out:            os.Stdout,
		Err:            os.Stderr,
		ReportWriter:   os.Stdout,
		IgnoreFile:     "./.dockerignore",
		AdditionalTags: []string{image.FullPath()},
	}, "./Dockerfile")
	if err != nil {
		return image, fmt.Errorf("failed to build a docker container: %w", err)
	}

	// TODO: do something with that
	_, _ = id, ref

	imageRef, err := docker.ParseReference(ref.String())
	if err != nil {
		return image, fmt.Errorf("failed to parse image ref: %w", err)
	}

	buildah.Push(ctx, image.FullPath(), imageRef, buildah.PushOptions{
		Store: a.store,
	})

	return image, nil
}
