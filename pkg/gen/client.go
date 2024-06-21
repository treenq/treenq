package gen

import (
	_ "embed"
	"html/template"
	"log"
	"os"
	"path/filepath"
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

func (g *ClientGen) Generate(name string, endUrl string) error {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	fileName := filepath.Join(wd, "../../pkg/sdk/"+name+"client.go")

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		log.Fatalln(err)
	}
	tmpl, err := template.New("goTemplate").Parse(clientTemplate)
	if err != nil {
		return err
	}
	var client Client
	client.UpperName = strings.ToUpper(name)
	client.LowerName = name
	client.Url = endUrl

	return tmpl.Execute(file, client)
}
