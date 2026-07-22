package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"crm-prospect-simulator/backend/bootstrap"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

var (
	initialize sync.Once
	apiHandler http.HandlerFunc
	startupErr error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	restoreRewrittenAPIPath(r)
	initialize.Do(func() {
		application, _, err := bootstrap.Build(context.Background())
		if err != nil {
			startupErr = err
			return
		}
		apiHandler = adaptor.FiberApp(application.Fiber)
	})
	if startupErr != nil || apiHandler == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{"code": "SERVICE_UNAVAILABLE", "message": "The API is temporarily unavailable."},
		})
		return
	}
	apiHandler(w, r)
}

// restoreRewrittenAPIPath restores the public API path captured by the Vercel
// rewrite. The internal query parameter is removed before Fiber receives the
// request, while all public query parameters are preserved.
func restoreRewrittenAPIPath(r *http.Request) {
	query := r.URL.Query()
	route := strings.Trim(query.Get("__api_path"), "/")
	if route == "" {
		return
	}
	for _, segment := range strings.Split(route, "/") {
		if segment == "" || segment == "." || segment == ".." || strings.Contains(segment, "\\") {
			return
		}
	}
	r.URL.Path = "/api/" + route
	query.Del("__api_path")
	r.URL.RawQuery = query.Encode()
	r.RequestURI = r.URL.RequestURI()
}
