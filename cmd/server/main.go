package main

import (
	"log"
	"net/http"

	"github.com/containers/buildah"
    "github.com/containers/storage/pkg/unshare"
	"github.com/treenq/treenq/src/api"
)

func main() {
	if buildah.InitReexec() {
		return
	}
	unshare.MaybeReexecUsingUserNamespace(false)

	conf, err := api.NewConfig()
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}
	m, err := api.New(conf)
	if err != nil {
		log.Fatalln("failed to build api:", err)
	}
	log.Println("service is running on:", conf.HttpPort)

	if err := http.ListenAndServe(":"+conf.HttpPort, m); err != nil {
		log.Println(err)
	}
}
