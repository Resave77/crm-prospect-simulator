package main

import (
	"context"
	"log"

	"crm-prospect-simulator/backend/config"
	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/repository"
	"crm-prospect-simulator/backend/platform/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewPostgresRepository(pool)
	users := []model.User{
		{ID: uuid.New(), Email: "admin@yummy.test", PasswordHash: string(hash), FullName: "Yummy Administrator", Role: model.RoleAdministrator, Status: model.UserActive},
		{ID: uuid.New(), Email: "sales@yummy.test", PasswordHash: string(hash), FullName: "Yummy Sales Executive", Role: model.RoleSalesExecutive, Status: model.UserActive},
	}
	for _, user := range users {
		if err := repo.UpsertSeed(ctx, user); err != nil {
			log.Fatal(err)
		}
		log.Printf("seeded %s (%s)", user.Email, user.Role)
	}
}
