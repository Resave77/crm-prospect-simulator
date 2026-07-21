package handler

import (
	"io"
	"net/http"
	"sync"

	"google-places-playground/config"
	"google-places-playground/internal/server"

	"github.com/gofiber/fiber/v2"
)

var (
	appOnce sync.Once
	app     *fiber.App
)

// Handler is the Vercel Go Function entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	appOnce.Do(func() {
		config.LoadEnv()
		app = server.New()
	})
	// Vercel rewrites all public routes to this single function. Restore the
	// original path so Fiber can keep owning the application's routing table.
	query := r.URL.Query()
	if originalPath := query.Get("__path"); originalPath != "" {
		r.URL.Path = originalPath
		query.Del("__path")
		r.URL.RawQuery = query.Encode()
		r.RequestURI = r.URL.RequestURI()
	}
	response, err := app.Test(r, -1)
	if err != nil {
		http.Error(w, "CRM service is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	defer response.Body.Close()
	for name, values := range response.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}
