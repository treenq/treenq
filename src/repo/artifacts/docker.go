//go:build !no_buildah
// +build !no_buildah

package artifacts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/treenq/treenq/src/domain"
)

var ErrUnknownDockerAuthType = errors.New("unknown docker auth type")

type DockerArtifact struct {
	registry string

	store             storage.Store
	registryTLSVerify bool
	registryCertDir   string
	registryAuthType  string
	registryUsername  string
	registryPassword  string
	registryToken     string
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
	buildStoreOptions, err := storage.DefaultStoreOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build buildah storage option: %w", err)
	}

	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to build buildah store: %w", err)
	}

	return &DockerArtifact{
		registry:          registry,
		store:             buildStore,
		registryTLSVerify: registryTLSVerify,
		registryCertDir:   registryCertDir,
		registryAuthType:  registryAuthType,
		registryUsername:  registryUsername,
		registryPassword:  registryPassword,
		registryToken:     registryToken,
	}, nil
}

func (a *DockerArtifact) getAuth() (*types.DockerAuthConfig, error) {
	// TODO: reuse api.Oci constant values isntead of hardcoding auth type constants
	switch a.registryAuthType {
	case "basic":
		return &types.DockerAuthConfig{
			Username: a.registryUsername,
			Password: a.registryPassword,
		}, nil
	case "token":
		return &types.DockerAuthConfig{
			IdentityToken: a.registryToken,
		}, nil
	case "noauth":
		return nil, nil
	default:
		return nil, ErrUnknownDockerAuthType
	}
}

func (a *DockerArtifact) systemContext() (*types.SystemContext, error) {
	auth, err := a.getAuth()
	if err != nil {
		return nil, err
	}
	return &types.SystemContext{
		DockerCertPath:              a.registryCertDir,
		DockerInsecureSkipTLSVerify: types.NewOptionalBool(!a.registryTLSVerify),
		DockerAuthConfig:            auth,
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
	out := progress.AsWriter(args.DeploymentID, slog.LevelInfo)
	errOut := progress.AsWriter(args.DeploymentID, slog.LevelError)
	reportOut := progress.AsWriter(args.DeploymentID, slog.LevelDebug)

	id, _, err := imagebuildah.BuildDockerfiles(ctx, a.store, define.BuildOptions{
		ContextDirectory: args.Path,
		Registry:         a.registry,
		Output:           args.Name,
		Out:              out,
		Err:              errOut,
		ReportWriter:     reportOut,
		IgnoreFile:     args.Dockerignore,
		AdditionalTags: []string{image.FullPath()},
	}, args.Dockerfile)
	if err != nil {
		return image, fmt.Errorf("failed to build a docker container: %w", err)
	}

	storeRef, err := alltransports.ParseImageName("docker://" + image.FullPath())
	if err != nil {
		return image, fmt.Errorf("failed to parse store reference: %w", err)
	}

	systemContext, err := a.systemContext()
	if err != nil {
		return image, fmt.Errorf("failed to get auth type: %w", err)
	}

	_, _, err = buildah.Push(ctx, id, storeRef, buildah.PushOptions{
		Store:         a.store,
		SystemContext: systemContext,
	})
	if err != nil {
		return image, fmt.Errorf("failed to push image: %w", err)
	}

	return image, nil
}

func (a *DockerArtifact) Inspect(ctx context.Context, deployment domain.AppDeployment) (domain.Image, error) {
	image := a.Image(deployment.Space.Service.Name, deployment.BuildTag)

	storeRef, err := alltransports.ParseImageName("docker://" + image.FullPath())
	if err != nil {
		return image, fmt.Errorf("failed to parse store reference: %w", err)
	}
	systemContext, err := a.systemContext()
	if err != nil {
		return image, fmt.Errorf("failed to get auth type: %w", err)
	}

	src, err := storeRef.NewImageSource(ctx, systemContext)
	if err != nil {
		var e errcode.Error
		if errors.As(err, &e) && e.Message == "manifest unknown" {
			return image, domain.ErrImageNotFound
		}
		return image, fmt.Errorf("failed to get iamge source: %w", err)
	}
	defer src.Close()
	return image, nil
}
