package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	fileName := filepath.Join(wd, "../../pkg/sdk/sizes.go")

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	err = GenerateSizeSlugs(f)
	if err != nil {
		log.Fatalln(err)
	}
}
