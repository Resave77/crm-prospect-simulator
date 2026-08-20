package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment                      string
	Port                             string
	DatabaseURL                      string
	JWTSecret                        string
	JWTIssuer                        string
	JWTAudience                      string
	AccessTokenTTL                   time.Duration
	RefreshTokenTTL                  time.Duration
	AllowedOrigins                   string
	CookieSecure                     bool
	GoogleMapsAPIKey                 string
	GooglePlacesCredentialAlias      string
	GoogleCSEID                      string
	GoogleCSEAPIKey                  string
	GooglePlacesSearchCacheTTL       time.Duration
	GooglePlacesCoreDetailCacheTTL   time.Duration
	GooglePlacesBusinessInfoCacheTTL time.Duration
	AIEnabled                        bool
	OpenAIAPIKey                     string
	OpenAICredentialAlias            string
	OpenAIModel                      string
	OpenAITimeout                    time.Duration
	OpenAIFindMenuTimeout            time.Duration
	OpenAIMenuProfileTimeout         time.Duration
	OpenAICacheTTL                   time.Duration
	OpenAIMaxTokens                  int
	AIChatMaxLength                  int
	AIChatMaxHistory                 int
}

func Load() (Config, error) {
	loadDotenv()

	accessTTL, err := duration("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := duration("REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	secure, err := strconv.ParseBool(value("COOKIE_SECURE", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("COOKIE_SECURE must be true or false: %w", err)
	}
	aiEnabled, err := strconv.ParseBool(value("AI_ENABLED", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("AI_ENABLED must be true or false: %w", err)
	}
	openAITimeout, err := duration("OPENAI_TIMEOUT", 20*time.Second)
	if err != nil {
		return Config{}, err
	}
	openAIFindMenuTimeout, err := duration("OPENAI_FIND_MENU_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	openAIMenuProfileTimeout, err := duration("OPENAI_MENU_PROFILE_TIMEOUT", 40*time.Second)
	if err != nil {
		return Config{}, err
	}
	openAICacheTTL := safeDuration("OPENAI_CACHE_TTL", 15*time.Minute)
	openAIMaxTokens, err := positiveInt("OPENAI_MAX_OUTPUT_TOKENS", 800)
	if err != nil {
		return Config{}, err
	}
	chatMaxLength, err := positiveInt("AI_CHAT_MAX_MESSAGE_LENGTH", 1000)
	if err != nil {
		return Config{}, err
	}
	chatMaxHistory, err := positiveInt("AI_CHAT_MAX_HISTORY_MESSAGES", 6)
	if err != nil {
		return Config{}, err
	}
	googleSearchCacheTTL := safeDuration("GOOGLE_PLACES_SEARCH_CACHE_TTL", 12*time.Hour)
	googleCoreCacheTTL := safeDuration("GOOGLE_PLACES_CORE_DETAIL_CACHE_TTL", 12*time.Hour)
	googleBusinessCacheTTL := safeDuration("GOOGLE_PLACES_BUSINESS_INFO_CACHE_TTL", 12*time.Hour)

	cfg := Config{
		Environment:                      value("APP_ENV", "development"),
		Port:                             value("PORT", "8080"),
		DatabaseURL:                      strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JWTSecret:                        os.Getenv("JWT_SECRET"),
		JWTIssuer:                        value("JWT_ISSUER", "yummy-crm"),
		JWTAudience:                      value("JWT_AUDIENCE", "yummy-crm-api"),
		AccessTokenTTL:                   accessTTL,
		RefreshTokenTTL:                  refreshTTL,
		AllowedOrigins:                   value("ALLOWED_ORIGINS", "http://localhost:5173"),
		CookieSecure:                     secure,
		GoogleMapsAPIKey:                 strings.TrimSpace(os.Getenv("GOOGLE_MAPS_API_KEY")),
		GooglePlacesCredentialAlias:      value("GOOGLE_PLACES_CREDENTIAL_ALIAS", ""),
		GoogleCSEID:                      strings.TrimSpace(os.Getenv("GOOGLE_CSE_ID")),
		GoogleCSEAPIKey:                  strings.TrimSpace(os.Getenv("GOOGLE_CSE_API_KEY")),
		GooglePlacesSearchCacheTTL:       googleSearchCacheTTL,
		GooglePlacesCoreDetailCacheTTL:   googleCoreCacheTTL,
		GooglePlacesBusinessInfoCacheTTL: googleBusinessCacheTTL,
		AIEnabled:                        aiEnabled,
		OpenAIAPIKey:                     strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		OpenAICredentialAlias:            value("OPENAI_CREDENTIAL_ALIAS", ""),
		OpenAIModel:                      strings.TrimSpace(os.Getenv("OPENAI_MODEL")),
		OpenAITimeout:                    openAITimeout,
		OpenAIFindMenuTimeout:            openAIFindMenuTimeout,
		OpenAIMenuProfileTimeout:         openAIMenuProfileTimeout,
		OpenAICacheTTL:                   openAICacheTTL,
		OpenAIMaxTokens:                  openAIMaxTokens,
		AIChatMaxLength:                  chatMaxLength,
		AIChatMaxHistory:                 chatMaxHistory,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	return cfg, nil
}

func (cfg Config) AIConfigured() bool {
	return cfg.AIEnabled && strings.TrimSpace(cfg.OpenAIAPIKey) != "" && strings.TrimSpace(cfg.OpenAIModel) != ""
}

func safeDuration(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func loadDotenv() {
	candidates := []string{".env"}
	if cwd, err := os.Getwd(); err == nil {
		switch filepath.Base(cwd) {
		case "backend":
			candidates = append(candidates, filepath.Join("..", ".env"))
		default:
			candidates = append(candidates, filepath.Join("backend", ".env"))
		}
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			_ = godotenv.Load(candidate)
		}
	}
}

func duration(name string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}
	return parsed, nil
}

func value(name, fallback string) string {
	if current := strings.TrimSpace(os.Getenv(name)); current != "" {
		return current
	}
	return fallback
}

func positiveInt(name string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}
