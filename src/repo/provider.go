package repo

import (
	"context"
	"fmt"

	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

type Provider struct {
	doToken string

	conf *config.CombinedConfig
}

func NewProvider(conf *config.CombinedConfig) *Provider {
	doToken := ""

	doConfig := config.Config{
		Token: doToken,
	}

	combinedConfig, err := doConfig.Client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return &Provider{
		doToken: doToken,

		conf: combinedConfig,
	}
}

const appKey = "digitalocean_app"

func errFromDiags(msg string, diags []diag.Diagnostic) error {
	for _, d := range diags {
		if d.Severity == diag.Error {
			return fmt.Errorf("%s: %s | %s", d.Summary, d.Detail)
		}
	}
	return nil
}

func (p *Provider) CreateAppResource(ctx context.Context, app tqsdk.App) error {
	provider := digitalocean.Provider()
	resource := provider.ResourcesMap[appKey]

	resourceData := tqsdk.MakeAppResourceData(app)
	diag := resource.CreateContext(ctx, resourceData, p.conf)
	if len(diag) != 0 {
		if diag.HasError() {
			return errFromDiags("failed to create app", diag)
		}
	}

	return nil55
}
