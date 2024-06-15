package main

import (
	"log"
	"net/http"

	"github.com/treenq/treenq/src/api"
)

func main() {
	m, err := api.New()
	if err != nil {
		log.Fatalln("failed to build api", err)
	}
	port := ":8000"
	log.Println("service is running on", port)

	if err := http.ListenAndServe(port, m); err != nil {
		log.Println(err)
	}
}
