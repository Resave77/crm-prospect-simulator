package handler

import (
	"io"
	"net/http"
	"sync"

	"crm-prospect-prototype/config"
	"crm-prospect-prototype/server"

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
	// Vercel rewrites all public routes to /api?__path=/:path*.
	// The destination template preserves the leading slash, so __path
	// already contains the original path (e.g. "/" or "/foo/bar").
	// Strip the query parameter and restore the path for Fiber routing.
	query := r.URL.Query()
	originalPath := query.Get("__path")
	if originalPath != "" {
		r.URL.Path = originalPath
	} else {
		r.URL.Path = "/"
	}
	query.Del("__path")
	r.URL.RawQuery = query.Encode()
	r.RequestURI = r.URL.RequestURI()
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
