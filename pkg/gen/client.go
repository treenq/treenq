package gen

import (
	"bytes"
	_ "embed"
	"io"
	"reflect"
	"strings"
	"text/template"
	"unicode"

	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/src/domain"
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
	handlersmeta []vel.HandlerMeta
}

//go:embed templates/client.txt
var clientTemplate string

func New(template string, handlersmeta []vel.HandlerMeta) *ClientGen {
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
	client.RespFields = GenerateFields(domain.InfoResponse{}, 1)
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
	s := bytes.Buffer{}
	var start string
	var end = "\n"
	start = strings.Repeat("\t", depth)
	v := reflect.ValueOf(st)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if i == v.NumField()-1 {
			end = ""
		}
		typeField := t.Field(i)
		valField := v.Field(i)
		s.WriteString(start)
		s.WriteString(typeField.Name)
		if v.Field(i).Kind() == reflect.Struct {
			s.WriteString(" struct {\n")
			s.WriteString(GenerateFields(valField.Interface(), depth+1))
			s.WriteString(start)
			s.WriteString("}")
		} else {
			s.WriteString(" ")
			s.WriteString(valField.Type().String())
		}
		s.WriteString(" `json:\"")
		s.WriteString(typeField.Tag.Get("json"))
		s.WriteString("\"`")
		s.WriteString(end)
	}
	return s.String()
}
