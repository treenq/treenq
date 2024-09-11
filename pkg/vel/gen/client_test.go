package gen

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treenq/treenq/pkg/vel"
)

//go:embed testdata/test.go
var infoClientOutput string

type TestTypeNoJsonTags struct {
	Value string
}

type TestTypeNestedTypes struct {
	Data  TestStruct `json:"data"`
	Chunk []byte     `json:"chunk"`
}

type TestStruct struct {
	Row  int    `json:"row"`
	Line string `json:"line"`
}

type Empty struct {
}

func TestGenClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skip: requires goimports installation")
	}

	buf := &bytes.Buffer{}

	gener, err := New(ClientDesc{
		TypeName:    "Client",
		PackageName: "client",
	}, []vel.HandlerMeta{
		{Input: TestTypeNoJsonTags{}, Output: TestTypeNoJsonTags{}, OperationID: "test1"},
		{Input: TestTypeNestedTypes{}, Output: TestTypeNestedTypes{}, OperationID: "test2"},
		{Input: struct{}{}, Output: Empty{}, OperationID: "testEmpty"},
	})
	require.NoError(t, err)
	err = gener.Generate(buf)
	require.NoError(t, err)
	assert.Equal(t, infoClientOutput, buf.String())
}
