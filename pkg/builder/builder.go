package main

import (
	"encoding/json"
	"os"

	"github.com/treenq/treenq/pkg/builder/tq"
)

func main() {
	res, _ := tq.Build()
	// if res.Size != 1 {
	// 	panic("not 1")
	// }

	f, err := os.Create("tq.json")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if err := json.NewEncoder(f).Encode(res); err != nil {
		panic(err)
	}
}
