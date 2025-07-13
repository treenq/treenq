package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dennypenta/vel"
	"github.com/dennypenta/vel/gen"
	"github.com/treenq/treenq/src/resources"
)

func main() {
	router := resources.NewRouter(nil, vel.NoopMiddleware, vel.NoopMiddleware)
	gener, err := gen.New(gen.ClientDesc{
		TypeName:    "Client",
		PackageName: "client",
	}, router.Meta())
	if err != nil {
		log.Fatalln("failed to create codegen", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("failed to getwd", err)
	}
	log.Println("wd is", wd)
	clientDir := filepath.Join(wd, "client")
	err = os.MkdirAll(clientDir, 0665)
	if err != nil {
		log.Fatalln("failed to mkdir", clientDir, err)
	}
	filePath := filepath.Join(clientDir, "client.go")
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0665)
	if err != nil {
		log.Fatalln("failed to open file", filePath, err)
	}
	log.Println("filepath:", filePath)
	defer f.Close()
	err = gener.Generate(f, "go:default", "goimports")
	if err != nil {
		log.Fatalln(err)
	}
}
