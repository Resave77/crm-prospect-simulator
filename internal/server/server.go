package server

import (
	"google-places-playground/internal/handler"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

// New creates the CRM application for both local development and Vercel.
func New() *fiber.App {
	templateDirectory := "./views"
	for _, candidate := range []string{"./views", "../views", "../../views"} {
		if _, err := os.Stat(candidate); err == nil {
			templateDirectory = candidate
			break
		}
	}
	engine := html.New(templateDirectory, ".html")
	engine.Reload(true)
	app := fiber.New(fiber.Config{Views: engine, AppName: "Atlas CRM"})

	app.Get("/", handler.NearbyProspectFinder)
	app.Get("/prospects/nearby", handler.NearbyProspectFinder)
	app.Get("/debug/place", func(c *fiber.Ctx) error { return c.Redirect("/", fiber.StatusMovedPermanently) })
	app.Get("/api/places/:id", handler.ProspectDetail)
	app.Get("/api/nearby-search", handler.NearbySearch)
	app.Post("/api/nearby-search", handler.NearbySearch)
	app.Post("/api/assignments", handler.AssignExistingCustomer)

	return app
}
