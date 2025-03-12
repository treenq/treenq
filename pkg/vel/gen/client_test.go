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

type Empty struct{}

type GetQuery struct {
	Value string `schema:"value"`
	Field int    `schema:"field"`
}

type GetResp struct {
	Getting int
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
		{Input: TestTypeNoJsonTags{}, Output: TestTypeNoJsonTags{}, OperationID: "test1", Method: "POST"},
		{Input: TestTypeNestedTypes{}, Output: TestTypeNestedTypes{}, OperationID: "test2", Method: "POST"},
		{Input: struct{}{}, Output: Empty{}, OperationID: "testEmpty", Method: "POST"},
		{Input: GetQuery{}, Output: GetResp{}, OperationID: "testGet", Method: "GET"},
	})
	require.NoError(t, err)
	err = gener.Generate(buf)
	require.NoError(t, err)
	assert.Equal(t, infoClientOutput, buf.String())

	// optional step to visualize the diff 
	// f, err := os.OpenFile("./testdata/out.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	// require.NoError(t, err)
	// defer f.Close()

	// _, err = f.Write(buf.Bytes())
	// require.NoError(t, err)
}
