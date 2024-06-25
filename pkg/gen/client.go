package gen

import (
	_ "embed"
	"fmt"
	"github.com/treenq/treenq/src/handlers"
	"io"
	"reflect"
	"strings"
	"text/template"
	"unicode"

	"github.com/treenq/treenq/src/api"
)

type Client struct {
	UpperName  string
	LowerName  string
	Url        string
	RespFields string
	ReqFields  string
}

type ClientGen struct {
	template     string
	handlersmeta []api.HandlerMeta
}

type InfoRequest struct {
	Try string `json:"try"`
}

//go:embed templates/client.txt
var clientTemplate string

func New(template string, handlersmeta []api.HandlerMeta) *ClientGen {
	return &ClientGen{
		template:     template,
		handlersmeta: handlersmeta,
	}
}

func (g *ClientGen) Generate(w io.Writer) error {
	if len(g.handlersmeta) == 0 {
		return io.EOF
	}
	name := g.handlersmeta[0].OperationID
	endUrl := "/" + name

	tmpl, err := template.New("goTemplate").Parse(clientTemplate)
	if err != nil {
		return err
	}
	var client Client
	client.UpperName = Capitalize(name)
	client.LowerName = name
	client.Url = endUrl
	client.RespFields = GenerateFields(handlers.InfoResponse{}, 1)
	client.ReqFields = GenerateFields(struct{}{}, 1)

	return tmpl.Execute(w, client)
}

func Capitalize(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	s = string(r)
	return s
}

func GenerateFields(st any, depth int) string {
	var s, start string
	var end = "\n"
	fmt.Println(depth)
	start = strings.Repeat("\t", depth)
	v := reflect.ValueOf(st)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if i == v.NumField()-1 {
			end = ""
		}
		TypeField := t.Field(i)
		ValField := v.Field(i)
		if v.Field(i).Kind() == reflect.Struct {
			s += start + TypeField.Name + " struct {\n" + GenerateFields(ValField.Interface(), depth+1) + start + "} `json:\"" + TypeField.Tag.Get("json") + "\"`" + end
		} else {
			s += start + TypeField.Name + " " + ValField.Type().String() + " `json:\"" + TypeField.Tag.Get("json") + "\"`" + end
		}
	}
	return s

}
