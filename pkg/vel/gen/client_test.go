package gen

import (
	"bytes"
	_ "embed"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/src/api"
	"github.com/treenq/treenq/src/domain"
)

//go:embed testdata/test.go
var infoClientOutput string

type TestTypeNoJsonTags struct {
	Value string
}

type TestTypeNestedTypes struct {
	Data  TestStruct `json:"data"`
	Chunk []byte     `json:"chunk"`
	// High level elements
	NextLevelSlice   []HighElem          `json:"slice"`
	Map              map[int]HighMapElem `json:"map"`
	NextLevelNestedP *HighPointer        `json:"nextP"`
}

type TestStruct struct {
	Row              int                   `json:"row"`
	Line             string                `json:"line"`
	NextLevelNested  TestNextLevelStruct   `json:"next"`
	NextLevelSlice   []TestNextLevelElem   `json:"slice"`
	Map              map[int]MapValue      `json:"map"`
	NextLevelNestedP *TestNextLevelStructP `json:"nextP"`
	// TODO: Highlight as not supported (who ever might need them??? )
	// NextLevelSliceP  []*TestNextLevelElemP `json:"sliceP"`
	// MapP             map[int]*MapValueP    `json:"mapP"`
}

type TestNextLevelStruct struct {
	Extra string `json:"extra"`
}

type TestNextLevelElem struct {
	Int int `json:"int"`
}
type MapValue struct {
	Value string
}
type TestNextLevelStructP struct {
	Extra string `json:"extra"`
}
type HighElem struct {
	Int int `json:"int"`
}
type HighMapElem struct {
	Value string
}
type HighPointer struct {
	Extra string `json:"extra"`
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

func TestJo(t *testing.T) {
	buf := &bytes.Buffer{}

	router := api.NewRouter(&domain.Handler{}, vel.NoopMiddleware, vel.NoopMiddleware)
	gener, err := New(ClientDesc{
		TypeName:    "Client",
		PackageName: "client",
	}, router.Meta())
	if err != nil {
		log.Fatalln(err)
	}

	err = gener.Generate(buf)
	t.Log(buf.String())
	require.NoError(t, err)
}
