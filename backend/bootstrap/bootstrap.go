package bootstrap

import (
	"context"
	"fmt"

	"crm-prospect-simulator/backend/config"
	"crm-prospect-simulator/backend/internal/auth/repository"
	"crm-prospect-simulator/backend/internal/auth/service"
	customerrepository "crm-prospect-simulator/backend/internal/customer/repository"
	customerservice "crm-prospect-simulator/backend/internal/customer/service"
	prospectrepository "crm-prospect-simulator/backend/internal/prospect/repository"
	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"crm-prospect-simulator/backend/platform/database"
	"crm-prospect-simulator/backend/server"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Fiber *fiber.App
	Pool  *pgxpool.Pool
}

func Build(ctx context.Context) (*Application, config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, config.Config{}, fmt.Errorf("load configuration: %w", err)
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, config.Config{}, err
	}
	repo := repository.NewPostgresRepository(pool)
	tokens := service.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTokenTTL)
	authService := service.NewAuthService(repo, repo, tokens, cfg.RefreshTokenTTL)
	prospectRepo := prospectrepository.NewPostgresRepository(pool)
	placesClient := prospectservice.NewGooglePlacesClient(cfg.GoogleMapsAPIKey)
	prospectService := prospectservice.New(prospectRepo, placesClient)
	customerRepo := customerrepository.NewPostgresRepository(pool)
	customerService := customerservice.New(customerRepo, prospectService)
	return &Application{Fiber: server.New(cfg, authService, prospectService, customerService), Pool: pool}, cfg, nil
}
