package gen

import "github.com/treenq/treenq/src/api"

type ClientGen struct {
	template     string
	handlersmeta []api.HandlerMeta
}

func New(template string, handlersmeta []api.HandlerMeta) *ClientGen {
	return &ClientGen{
		template:     template,
		handlersmeta: handlersmeta,
	}
}

func (g *ClientGen) Generate() ([]byte, error) {

}
