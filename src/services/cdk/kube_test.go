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
	res, err := k.DefineApp(ctx, "id-1234", "space", tqsdk.Space{
		Service: tqsdk.Service{
			Name: "simple-app",
			RuntimeEnvs: map[string]string{
				"DO_TOKEN": "111",
			},
			HttpPort: 8000,
			Replicas: 1,
			ComputationResource: tqsdk.ComputationResource{
				CpuUnits:   250,
				MemoryMibs: 512,
				DiskGibs:   1,
			},
		},
	}, domain.Image{
		Registry:   "registry:5000",
		Repository: "treenq",
		Tag:        "0.0.1",
	}, secretKeys)

	assert.Equal(t, appYaml, res)
	assert.NoError(t, err)
}
