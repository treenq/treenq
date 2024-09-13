package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/pkg/vel/gen"
	"github.com/treenq/treenq/src/api"
	"github.com/treenq/treenq/src/domain"
)

func main() {
	router := api.NewRouter(&domain.Handler{}, vel.NoopMiddleware, vel.NoopMiddleware)
	gener, err := gen.New(gen.ClientDesc{
		TypeName:    "Client",
		PackageName: "client",
	}, router.Meta())
	if err != nil {
		log.Fatalln(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	clientDir := filepath.Join(wd, "client")
	err = os.MkdirAll(clientDir, 0665)
	if err != nil {
		log.Fatalln(err)
	}
	filePath := filepath.Join(clientDir, "client.go")
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0665)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	err = gener.Generate(f)
	if err != nil {
		log.Fatalln(err)
	}
}
