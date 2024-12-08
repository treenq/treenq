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

//go:embed testdata/kubeconfig.yaml
var conf string

func TestAppDefinition(t *testing.T) {
	k := NewKube()
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
			Replicas:  1,
			Host:     "treenq.local",
			SizeSlug: tqsdk.SizeSlugS,
		},
	}, domain.Image{
		Registry:   "registry:5000",
		Repository: "treenq",
		Tag:        "0.0.1",
	})

	assert.Equal(t, appYaml, res)

	err := k.Apply(ctx, conf, res)
	assert.NoError(t, err)
}

func TestInvalidNamespaceName(t *testing.T) {

}

func TestRunAsNonNumbericNonRootUser(t *testing.T) {

}
