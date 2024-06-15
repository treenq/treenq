package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/treenq/treenq/pkg/builder/tq"
)

func main() {
	res, err := tq.Build()
	if err != nil {
		log.Fatalln(err)
	}

	data, err := json.Marshal(res)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Print(string(data))
}
