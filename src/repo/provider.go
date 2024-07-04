package repo

import (
	"context"
	"fmt"
	"io"

	"github.com/digitalocean/godo"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"
)

type Provider struct {
	client *godo.Client
}

func NewProvider(client *godo.Client) *Provider {
	return &Provider{
		client: client,
	}
}

func (p *Provider) CreateAppResource(ctx context.Context, image domain.Image, app tqsdk.Space) error {
	envs := make([]*godo.AppVariableDefinition, 0, len(app.Service.RuntimeEnvs)+len(app.Service.BuildEnvs)+len(app.Service.RuntimeSecrets)+len(app.Service.BuildSecrets))
	for k, v := range app.Service.RuntimeEnvs {
		envs = append(envs, &godo.AppVariableDefinition{
			Key:   k,
			Value: v,
			Scope: godo.AppVariableScope_RunTime,
			Type:  godo.AppVariableType_General,
		})
	}
	for k, v := range app.Service.BuildEnvs {
		envs = append(envs, &godo.AppVariableDefinition{
			Key:   k,
			Value: v,
			Scope: godo.AppVariableScope_BuildTime,
			Type:  godo.AppVariableType_General,
		})
	}
	for k, v := range app.Service.RuntimeSecrets {
		envs = append(envs, &godo.AppVariableDefinition{
			Key:   k,
			Value: v,
			Scope: godo.AppVariableScope_RunTime,
			Type:  godo.AppVariableType_Secret,
		})
	}
	for k, v := range app.Service.BuildSecrets {
		envs = append(envs, &godo.AppVariableDefinition{
			Key:   k,
			Value: v,
			Scope: godo.AppVariableScope_BuildTime,
			Type:  godo.AppVariableType_Secret,
		})
	}

	_, resp, err := p.client.Apps.Create(ctx, &godo.AppCreateRequest{
		Spec: &godo.AppSpec{
			Name:   app.Name,
			Region: app.Region,
			Services: []*godo.AppServiceSpec{
				{
					Name: app.Name,
					Image: &godo.ImageSourceSpec{
						RegistryType: godo.ImageSourceSpecRegistryType_DOCR,
						Repository:   image.Repository,
						Tag:          image.Tag,
						DeployOnPush: &godo.ImageSourceSpecDeployOnPush{
							Enabled: true,
						},
					},
					InstanceSizeSlug: string(app.Service.SizeSlug),
					InstanceCount:    int64(app.Service.InstanceCount),
					HTTPPort:         int64(app.Service.HttpPort),
					Envs:             envs,
				},
			},
		},
	})
	if resp != nil && resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create app resource: %s", string(body))
	}
	if err != nil {
		return fmt.Errorf("failed to create app resource: %w", err)
	}
	return nil
}
