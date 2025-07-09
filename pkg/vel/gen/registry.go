package gen

import (
	"strings"
	"text/template"

	_ "embed"
)

//go:embed templates/go.tpl
var goTemplate string

//go:embed templates/ts.tpl
var tsTemplate string

var templateRegistry map[string]*template.Template

func init() {
	templateRegistry = make(map[string]*template.Template)

	tplGo, err := template.New("goTemplate").Parse(goTemplate)
	if err != nil {
		panic("failed to registry go template: " + err.Error())
	}
	templateRegistry["go:default"] = tplGo

	tplTs, err := template.New("tsTemplate").Parse(tsTemplate)
	if err != nil {
		panic("failed to registry ts template: " + err.Error())
	}
	templateRegistry["ts:default"] = tplTs
}

func RegisterTemplate(name string, tpl *template.Template) {
	if _, ok := templateRegistry[name]; ok {
		panic(name + " template already exists, consider adding a prefix")
	}
	parts := strings.Split(name, ":")
	if len(parts) != 2 {
		panic("template name must consists of {0}:{1}, where {0} is a language code, {1} is a template name")
	}
	templateRegistry[name] = tpl
}
