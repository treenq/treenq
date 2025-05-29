//go:build no_buildah
// +build no_buildah

package artifacts

import (
	"context"
	"errors"

	"github.com/treenq/treenq/src/domain"
)

var ErrUnknownDockerAuthType = errors.New("unknown docker auth type")

type DockerArtifact struct {
	registry string
}

func NewDockerArtifactory(
	registry string,
	registryTLSVerify bool,
	registryCertDir,
	registryAuthType,
	registryUsername,
	registryPassword,
	registryToken string,
) (*DockerArtifact, error) {
	return &DockerArtifact{
		registry: registry,
	}, nil
}

func (a *DockerArtifact) Image(args domain.BuildArtifactRequest) domain.Image {
	return domain.Image{
		Registry:   a.registry,
		Repository: args.Name,
		Tag:        args.Tag,
	}
}

func (a *DockerArtifact) Build(ctx context.Context, args domain.BuildArtifactRequest, progress *domain.ProgressBuf) (domain.Image, error) {
	image := a.Image(args)
	return image, nil
}

func (a *DockerArtifact) Inspect() (domain.Image, error) {
	return domain.Image{}, domain.ErrImageNotFound
}
