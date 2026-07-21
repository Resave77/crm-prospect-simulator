package main

import (
	"log"

	"google-places-playground/config"
	"google-places-playground/internal/server"
)

func main() {

	config.LoadEnv()

	app := server.New()
	log.Fatal(app.Listen(":" + config.ServerPort()))

}
