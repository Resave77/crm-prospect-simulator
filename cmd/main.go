package main

import (
	"log"

	"crm-prospect-prototype/config"
	"crm-prospect-prototype/server"
)

func main() {

	config.LoadEnv()

	app := server.New()
	log.Fatal(app.Listen(":" + config.ServerPort()))

}
