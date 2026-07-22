package main

import (
	"context"
	"log"

	"crm-prospect-simulator/backend/bootstrap"
)

func main() {
	application, cfg, err := bootstrap.Build(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer application.Pool.Close()
	log.Fatal(application.Fiber.Listen(":" + cfg.Port))
}
