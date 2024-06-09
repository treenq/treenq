package main

import (
	"fmt"

	"github.com/treenq/treenq/pkg/example/tq"
)

func main() {
	def, _ := tq.Build()
	conf := def.AsConfig()

	fmt.Println("ConnStr")
	fmt.Println(conf)
}
