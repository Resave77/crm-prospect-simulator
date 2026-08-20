package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"

	"crm-prospect-simulator/backend/config"
	adminrepository "crm-prospect-simulator/backend/internal/admin/repository"
	adminservice "crm-prospect-simulator/backend/internal/admin/service"
	aiservice "crm-prospect-simulator/backend/internal/ai/service"
	"crm-prospect-simulator/backend/internal/auth/repository"
	"crm-prospect-simulator/backend/internal/auth/service"
	customerrepository "crm-prospect-simulator/backend/internal/customer/repository"
	customerservice "crm-prospect-simulator/backend/internal/customer/service"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	prospectrepository "crm-prospect-simulator/backend/internal/prospect/repository"
	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"crm-prospect-simulator/backend/internal/usage"
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
	placesClient := prospectservice.NewGooglePlacesClient(cfg.GoogleMapsAPIKey, cfg.GoogleCSEID, cfg.GoogleCSEAPIKey)
	placesClient.SetCacheTTLs(cfg.GooglePlacesSearchCacheTTL, cfg.GooglePlacesCoreDetailCacheTTL, cfg.GooglePlacesBusinessInfoCacheTTL)
	usageRecorder := usage.NewPostgresRecorder(pool)
	placesClient.SetUsageRecorder(usageRecorder)
	placesClient.SetUsageMetadata(cfg.GooglePlacesCredentialAlias, cfg.Environment)
	aiClient := aiservice.NewClient(cfg)
	aiClient.SetUsageRecorder(usageRecorder)
	aiClient.SetUsageMetadata(cfg.OpenAICredentialAlias, cfg.Environment)
	prospectAI := aiservice.NewProspectAI(aiClient, cfg.AIChatMaxLength, cfg.AIChatMaxHistory)
	initialAnalyzer := aiservice.NewInitialAnalyzer(pool, prospectAI)
	prospectService := prospectservice.New(prospectRepo, placesClient)
	prospectService.SetSummaryAI(initialAnalyzer.GenerateSummary)
	prospectService.SetMenuAI(func(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, images []prospectservice.MenuImageInput) (json.RawMessage, error) {
		inputs := make([]aiservice.MenuImageInput, 0, len(images))
		for _, image := range images {
			inputs = append(inputs, aiservice.MenuImageInput{Name: image.Name, Bytes: image.Bytes, ContentType: image.ContentType})
		}
		return initialAnalyzer.ProfileMenu(ctx, review, details, inputs)
	})
	prospectService.SetFindMenu(initialAnalyzer.FindMenu)
	prospectService.SetStructuredMenuAI(initialAnalyzer.ProfileStructuredMenu)
	prospectService.SetChatAI(func(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment, history []prospectservice.ChatTurn, message, skill string) (string, error) {
		turns := make([]aiservice.ChatMessage, 0, len(history))
		for _, turn := range history {
			turns = append(turns, aiservice.ChatMessage{Role: turn.Role, Content: turn.Content})
		}
		return initialAnalyzer.ChatWithHistory(ctx, review, details, comments, turns, message, skill)
	})
	customerRepo := customerrepository.NewPostgresRepository(pool)
	customerService := customerservice.New(customerRepo, prospectService)
	adminRepo := adminrepository.NewPostgresRepository(pool)
	adminService := adminservice.New(adminRepo)
	return &Application{Fiber: server.New(cfg, authService, prospectService, customerService, adminService, initialAnalyzer, pool), Pool: pool}, cfg, nil
}
