package artifacts

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/tonistiigi/fsutil"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"

	"github.com/moby/buildkit/client"

	gateway "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/progress/progresswriter"
	"golang.org/x/sync/errgroup"

	"github.com/treenq/treenq/src/domain"
)

var ErrUnknownDockerAuthType = errors.New("unknown docker auth type")

type DockerArtifact struct {
	buildkitHost string
	registry     string

	registryTLSVerify bool
	registryCert      string
	registryUsername  string
	registryPassword  string
}

func NewDockerArtifactory(
	buildkitHost string,
	registry string,
	registryTLSVerify bool,
	registryCert string,
	registryUsername string,
	registryPassword string,
) (*DockerArtifact, error) {
	return &DockerArtifact{
		buildkitHost:      buildkitHost,
		registry:          registry,
		registryTLSVerify: registryTLSVerify,
		registryCert:      registryCert,
		registryUsername:  registryUsername,
		registryPassword:  registryPassword,
	}, nil
}

func (a *DockerArtifact) Image(name, tag string) domain.Image {
	return domain.Image{
		Registry:   a.registry,
		Repository: name,
		Tag:        tag,
	}
}

func parseLocal(locals map[string]string) (map[string]fsutil.FS, error) {
	mounts := make(map[string]fsutil.FS, len(locals))

	var err error
	for k, v := range locals {
		mounts[k], err = fsutil.NewFS(v)
		if err != nil {
			return nil, fmt.Errorf("failed to mew docker local fs: %w", err)
		}
	}

	return mounts, nil
}

type fakeFile struct {
	io.Writer
}

func (fakeFile) Read(_ []byte) (int, error) {
	return 1, nil
}
func (fakeFile) Close() error { return nil }
func (fakeFile) Fd() uintptr  { return 1 }
func (fakeFile) Name() string { return "" }

func (a *DockerArtifact) Build(ctx context.Context, args domain.BuildArtifactRequest, progress *domain.ProgressBuf) (domain.Image, error) {
	image := a.Image(args.Name, args.Tag)
	out := progress.AsWriter(args.DeploymentID, slog.LevelInfo)

	c, err := client.New(ctx, a.buildkitHost, client.WithServerConfig(a.registry, a.registryCert))
	if err != nil {
		return image, err
	}

	dockerConfig := &configfile.ConfigFile{
		AuthConfigs: map[string]types.AuthConfig{
			a.registry: {
				Username: a.registryUsername,
				Password: a.registryPassword,
			},
		},
	}

	tlsConfig := map[string]*authprovider.AuthTLSConfig{
		a.registry: {
			Insecure: !a.registryTLSVerify,
			RootCAs: []string{
				a.registryCert,
			},
		},
	}

	attachable := []session.Attachable{authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
		ConfigFile: dockerConfig,
		TLSConfigs: tlsConfig,
	})}
	localMounts, err := parseLocal(map[string]string{
		"context":    args.DockerContext,
		"dockerfile": args.Dockerfile,
	})
	if err != nil {
		return image, errors.Wrap(err, "invalid local")
	}

	pw, err := progresswriter.NewPrinter(ctx, &fakeFile{out}, string(progressui.PlainMode))
	if err != nil {
		return image, err
	}

	eg, ctx := errgroup.WithContext(ctx)

	solveOpt := client.SolveOpt{
		Exports: []client.ExportEntry{{
			Type: "image",
			Attrs: map[string]string{
				"push": "true",
				"name": image.FullPath(),
			},
		}},
		Frontend:    "dockerfile.v1",
		Session:     attachable,
		Ref:         identity.NewID(),
		LocalMounts: localMounts,
	}

	eg.Go(func() error {
		sreq := gateway.SolveRequest{
			Frontend:    solveOpt.Frontend,
			FrontendOpt: solveOpt.FrontendAttrs,
		}

		_, err := c.Build(ctx, solveOpt, "buildctl", func(ctx context.Context, c gateway.Client) (*gateway.Result, error) {
			res, err := c.Solve(ctx, sreq)
			if err != nil {
				return nil, err
			}
			return res, err
		}, progresswriter.ResetTime(pw).Status())
		if err != nil {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		<-pw.Done()
		return pw.Err()
	})

	if err := eg.Wait(); err != nil {
		return image, err
	}

	return image, nil
}

func (a *DockerArtifact) Inspect(ctx context.Context, deployment domain.AppDeployment) (domain.Image, error) {
	image := a.Image(deployment.Space.Service.Name, deployment.BuildTag)
	return image, domain.ErrImageNotFound
}
