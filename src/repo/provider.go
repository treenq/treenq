package repo

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

type Provider struct {
	client *godo.Client
}

func NewProvider(client *godo.Client) *Provider {
	return &Provider{
		client: client,
	}
}

func (p *Provider) CreateAppResource(ctx context.Context, projectID, image, sizeSlug string, app tqsdk.App) error {
	services := make([]*godo.AppServiceSpec, len(app.Services))
	for i, s := range app.Services {
		envs := make([]*godo.AppVariableDefinition, 0, len(s.RuntimeEnvs)+len(s.BuildEnvs)+len(s.RuntimeSecrets)+len(s.BuildSecrets))
		for k, v := range s.RuntimeEnvs {
			envs = append(envs, &godo.AppVariableDefinition{
				Key:   k,
				Value: v,
				Scope: godo.AppVariableScope_RunTime,
				Type:  godo.AppVariableType_General,
			})
		}
		for k, v := range s.BuildEnvs {
			envs = append(envs, &godo.AppVariableDefinition{
				Key:   k,
				Value: v,
				Scope: godo.AppVariableScope_BuildTime,
				Type:  godo.AppVariableType_General,
			})
		}
		for k, v := range s.RuntimeSecrets {
			envs = append(envs, &godo.AppVariableDefinition{
				Key:   k,
				Value: v,
				Scope: godo.AppVariableScope_RunTime,
				Type:  godo.AppVariableType_Secret,
			})
		}
		for k, v := range s.BuildSecrets {
			envs = append(envs, &godo.AppVariableDefinition{
				Key:   k,
				Value: v,
				Scope: godo.AppVariableScope_BuildTime,
				Type:  godo.AppVariableType_Secret,
			})
		}

		services[i] = &godo.AppServiceSpec{
			Name: s.Name,
			Image: &godo.ImageSourceSpec{
				RegistryType: godo.ImageSourceSpecRegistryType_DOCR,
				Repository:   image,
				DeployOnPush: &godo.ImageSourceSpecDeployOnPush{
					Enabled: true,
				},
			},
			InstanceSizeSlug: sizeSlug,
			InstanceCount:    int64(s.InstanceCount),
			HTTPPort:         int64(s.HttpPort),
			Envs:             envs,
		}
	}
	_, _, err := p.client.Apps.Create(ctx, &godo.AppCreateRequest{
		ProjectID: app.ProjectID,
		Spec: &godo.AppSpec{
			Name:     app.Name,
			Region:   app.Region,
			Services: services,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create app resource: %w", err)
	}
	return nil
}
