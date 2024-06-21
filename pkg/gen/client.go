package gen

import (
	_ "embed"
	"html/template"
	"io"
	"strings"

	"github.com/treenq/treenq/src/api"
)

type Client struct {
	UpperName string
	LowerName string
	Url       string
}

type ClientGen struct {
	template     string
	handlersmeta []api.HandlerMeta
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
	client.UpperName = strings.ToUpper(name)
	client.LowerName = name
	client.Url = endUrl

	return tmpl.Execute(w, client)
}
