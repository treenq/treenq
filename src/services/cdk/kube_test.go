package cdk

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"
)

//go:embed testdata/app.yaml
var appYaml string

func TestAppDefinition(t *testing.T) {
	secretKeys := []string{"SECRET"}
	k := NewKube("treenq.com", "registry:5000", "testuser", "testpassword")
	ctx := context.Background()
	res := k.DefineApp(ctx, "id-1234", tqsdk.Space{
		Key: "space",
		Service: tqsdk.Service{
			Name: "simple-app",
			RuntimeEnvs: map[string]string{
				"DO_TOKEN":                     "111",
				"DOCKER_REGISTRY":              "registry:5000",
				"GITHUB_WEBHOOK_SECRET_ENABLE": "false",
			},
			HttpPort: 8000,
			Replicas: 1,
			SizeSlug: tqsdk.SizeSlugS,
		},
	}, domain.Image{
		Registry:   "registry:5000",
		Repository: "treenq",
		Tag:        "0.0.1",
	}, secretKeys)

	assert.Equal(t, appYaml, res)
}
