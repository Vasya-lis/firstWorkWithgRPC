package main

import (
	"log"

	"github.com/Vasya-lis/firstWorkWithgRPC/services/api"
)

func main() {
	app, err := api.NewAppApi()
	if err != nil {
		log.Fatalf("failed to init API: %v", err)
	}

	app.Start()
	app.Stop()
}
