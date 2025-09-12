package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	db "github.com/Vasya-lis/firstWorkWithgRPC/services/db"
)

func main() {

	app, err := db.NewAppDB()
	if err != nil {
		log.Fatalf("init app failed: %v", err)
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
