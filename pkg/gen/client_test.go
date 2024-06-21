package gen

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/treenq/treenq/src/api"
)

//go:embed testdata/info.go
var infoClientOutput string

func TestGenClient(t *testing.T) {
	buf := &bytes.Buffer{}

	g := New(clientTemplate, []api.HandlerMeta{
		{
			OperationID: "info",
		},
	})
	g.Generate(buf)

	assert.Equal(t, string(infoClientOutput), buf.String())
}
