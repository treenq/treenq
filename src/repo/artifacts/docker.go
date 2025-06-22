package artifacts

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"

	"github.com/pkg/errors"
	"github.com/tonistiigi/fsutil"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"

	"github.com/moby/buildkit/client"

	gateway "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/progress/progresswriter"
	"golang.org/x/sync/errgroup"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	"github.com/treenq/treenq/src/domain"
)

var ErrUnknownDockerAuthType = errors.New("unknown docker auth type")

type DockerArtifact struct {
	buildkitHost  string
	buildkitTLSCA string

	registry          string
	registryTLSVerify bool
	registryCert      string
	registryUsername  string
	registryPassword  string
}

func NewDockerArtifactory(
	buildkitHost string,
	buildkitTLSCA string,
	registry string,
	registryTLSVerify bool,
	registryCert string,
	registryUsername string,
	registryPassword string,
) (*DockerArtifact, error) {
	return &DockerArtifact{
		buildkitHost:      buildkitHost,
		buildkitTLSCA:     buildkitTLSCA,
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
	return 0, nil
}
func (fakeFile) Close() error { return nil }
func (fakeFile) Fd() uintptr  { return 0 }
func (fakeFile) Name() string { return "" }

func (a *DockerArtifact) Build(ctx context.Context, args domain.BuildArtifactRequest, progress *domain.ProgressBuf) (domain.Image, error) {
	image := a.Image(args.Name, args.Tag)
	out := progress.AsWriter(args.DeploymentID, slog.LevelInfo)

	var clientOpts []client.ClientOpt
	if a.buildkitTLSCA != "" {
		u, err := url.Parse(a.buildkitHost)
		if err != nil {
			return image, fmt.Errorf("given invalid buildkit host: %w", err)
		}
		host := u.Hostname()
		clientOpts = append(clientOpts, client.WithServerConfig(host, a.buildkitTLSCA))
	}

	c, err := client.New(ctx, a.buildkitHost, clientOpts...)
	if err != nil {
		return image, err
	}

	dockerConfig := &configfile.ConfigFile{
		AuthConfigs: map[string]types.AuthConfig{
			a.registry: {
				Username:      a.registryUsername,
				Password:      a.registryPassword,
				Auth:          base64.StdEncoding.EncodeToString([]byte(a.registryUsername + ":" + a.registryPassword)),
				ServerAddress: a.registry,
			},
		},
	}

	tlsConfig := map[string]*authprovider.AuthTLSConfig{
		a.registry: {
			Insecure: !a.registryTLSVerify,
		},
	}
	if a.registryTLSVerify {
		tlsConfig[a.registry].RootCAs = []string{a.registryCert}
	}

	attachable := []session.Attachable{authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
		ConfigFile: dockerConfig,
		TLSConfigs: tlsConfig,
	})}
	localMounts, err := parseLocal(map[string]string{
		"context":    args.DockerContext,
		"dockerfile": filepath.Dir(args.Dockerfile),
	})
	if err != nil {
		return image, errors.Wrap(err, "invalid local")
	}

	frontendAttrs := make(map[string]string)
	if args.Dockerfile != "" {
		frontendAttrs["filename"] = filepath.Base(args.Dockerfile)
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
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		Session:       attachable,
		LocalMounts:   localMounts,
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

	ref := fmt.Sprintf("%s/%s", a.registry, image.Repository)
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return image, fmt.Errorf("failed to create repository: %w", err)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !a.registryTLSVerify,
	}
	if a.registryTLSVerify && a.registryCert != "" {
		certPEM, err := os.ReadFile(a.registryCert)
		if err != nil {
			return image, fmt.Errorf("failed to read registry certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(certPEM) {
			return image, fmt.Errorf("failed to parse registry certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	repo.Client = &auth.Client{
		Client: httpClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(repo.Reference.Registry, auth.Credential{
			Username: a.registryUsername,
			Password: a.registryPassword,
		}),
	}

	exists := false
	err = repo.Tags(ctx, "", func(tags []string) error {
		if slices.Contains(tags, image.Tag) {
			exists = true
		}
		return nil
	})
	if err != nil {
		return image, fmt.Errorf("failed to list tags: %w", err)
	}

	if !exists {
		return image, domain.ErrImageNotFound
	}

	return image, nil
}
