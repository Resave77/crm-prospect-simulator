package main

import (
	"log"

	"crm-prospect-prototype/config"
	"crm-prospect-prototype/internal/server"
)

func main() {

	config.LoadEnv()

	app := server.New()
	log.Fatal(app.Listen(":" + config.ServerPort()))

}
