package providers

import (
	"testing"
	"time"

	_ "embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/prometheus.yml
var expectedPrometheusConfig string

func TestConfigBuilder(t *testing.T) {

	builder, err := NewConfigBuilder()
	require.NoError(t, err)

	data := Config{
		Interval: time.Second * 5,
		Jobs: []Job{
			{
				JobName: "job1",
				Targets: []string{"target1:9090", "target2:9090"},
			},
			{
				JobName: "job2",
				Targets: []string{"target3:9090"},
			},
		},
	}

	res, err := builder.BuildAsContent(data)

	require.NoError(t, err)
	assert.Equal(t, res, expectedPrometheusConfig)
}
