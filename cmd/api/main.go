package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Vasya-lis/firstWorkWithgRPC/services/api"
)

func main() {
	app, err := api.NewAppApi()
	if err != nil {
		log.Fatalf("failed to init API: %v", err)
	}

	app.Start()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	for killSignal := range interrupt {
		fmt.Println("Got signal:", killSignal)
		if killSignal == os.Interrupt {
			fmt.Println("Daemon was interrupted by system signal")
		}
		fmt.Println("Daemon was killed")
		app.Stop()
		break
	}
}
