package config

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	// Local development may use .env; hosted environments provide variables
	// directly and should not fail when the file is absent.
	if GoogleAPIKey() != "" {
		return
	}
	for _, candidate := range []string{".env", "../.env", "../../.env"} {
		if _, err := os.Stat(candidate); err == nil {
			_ = godotenv.Load(candidate)
			return
		}
	}
}

func GoogleAPIKey() string {
	return os.Getenv("GOOGLE_MAPS_API_KEY")
}

func GoogleMapsBrowserAPIKey() string {
	if key := os.Getenv("GOOGLE_MAPS_BROWSER_API_KEY"); key != "" {
		return key
	}
	return GoogleAPIKey()
}

func GoogleBaseURL() string {
	return "https://places.googleapis.com/v1"
}

func ServerPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}
