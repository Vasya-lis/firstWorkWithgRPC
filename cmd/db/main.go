package db

import (
	"log"

	db "github.com/Vasya-lis/firstWorkWithgRPC/services/db"
)

func main() {

	app, err := db.NewAppDB()
	if err != nil {
		log.Fatalf("init app failed: %v", err)
	}
	app.Start()
	app.Stop()
}
