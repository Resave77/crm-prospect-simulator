package server

import (
	"errors"
	"time"

	"crm-prospect-simulator/backend/config"
	authhandler "crm-prospect-simulator/backend/internal/auth/handler"
	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func New(cfg config.Config, authService *service.AuthService) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Yummy CRM API",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    2 * 1024 * 1024,
		ErrorHandler: jsonErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowMethods:     "GET,POST,PATCH,DELETE,OPTIONS",
		AllowCredentials: true,
		MaxAge:           600,
	}))

	authHandler := authhandler.New(authService, cfg.CookieSecure)
	authMiddleware := authmiddleware.New(authService)

	health := func(c *fiber.Ctx) error {
		return response.Data(c, fiber.StatusOK, fiber.Map{"status": "ok"})
	}
	app.Get("/api/health", health)

	api := app.Group("/api/v1")
	api.Get("/health", health)

	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/me", authMiddleware.Authenticate, authHandler.Me)
	auth.Post("/logout-all", authMiddleware.Authenticate, authHandler.LogoutAll)

	dashboard := api.Group("/dashboard", authMiddleware.Authenticate)
	dashboard.Get("/admin", authMiddleware.RequireRole(model.RoleAdministrator), func(c *fiber.Ctx) error {
		return response.Data(c, fiber.StatusOK, fiber.Map{"surface": "administrator"})
	})
	dashboard.Get("/sales", authMiddleware.RequireRole(model.RoleSalesExecutive), func(c *fiber.Ctx) error {
		return response.Data(c, fiber.StatusOK, fiber.Map{"surface": "sales-executive"})
	})

	app.Use(func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusNotFound, "ROUTE_NOT_FOUND", "The requested API route does not exist.")
	})
	return app
}

func jsonErrorHandler(c *fiber.Ctx, err error) error {
	var fiberError *fiber.Error
	if errors.As(err, &fiberError) {
		return response.Error(c, fiberError.Code, "HTTP_ERROR", fiberError.Message)
	}
	return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
}
