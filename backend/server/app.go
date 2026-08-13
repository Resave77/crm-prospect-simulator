package server

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"crm-prospect-simulator/backend/config"
	adminhandler "crm-prospect-simulator/backend/internal/admin/handler"
	adminservice "crm-prospect-simulator/backend/internal/admin/service"
	aiservice "crm-prospect-simulator/backend/internal/ai/service"
	authhandler "crm-prospect-simulator/backend/internal/auth/handler"
	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	"crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/auth/service"
	customerhandler "crm-prospect-simulator/backend/internal/customer/handler"
	customerservice "crm-prospect-simulator/backend/internal/customer/service"
	prospecthandler "crm-prospect-simulator/backend/internal/prospect/handler"
	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func New(cfg config.Config, authService *service.AuthService, prospectService *prospectservice.Service, customerService *customerservice.Service, adminService *adminservice.Service, initialAnalyzers ...*aiservice.InitialAnalyzer) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Yummy CRM API",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		BodyLimit:    26 * 1024 * 1024,
		ErrorHandler: jsonErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestid.New())

	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err == nil {
		app.Static("/uploads", uploadsDir)
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowMethods:     "GET,POST,PATCH,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
		MaxAge:           600,
	}))

	authHandler := authhandler.New(authService, cfg.CookieSecure)
	authMiddleware := authmiddleware.New(authService)
	prospectHandler := prospecthandler.New(prospectService, customerService)
	if len(initialAnalyzers) > 0 {
		prospectHandler.SetInitialAnalyzer(initialAnalyzers[0])
	}
	customerHandler := customerhandler.New(customerService, prospectService)
	adminHandler := adminhandler.New(adminService)

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
	auth.Post("/change-password", authMiddleware.Authenticate, authHandler.ChangePassword)

	ai := api.Group("/ai", authMiddleware.Authenticate, authMiddleware.RequirePasswordChanged)
	ai.Get("/status", func(c *fiber.Ctx) error {
		return response.Data(c, fiber.StatusOK, fiber.Map{
			"enabled":         cfg.AIEnabled,
			"configured":      cfg.AIConfigured(),
			"modelConfigured": cfg.OpenAIModel != "",
		})
	})
	ai.Post("/prospects/:id/chat", authMiddleware.RequirePermission("use_prospect_ai_chat"), prospectHandler.ChatAI)
	ai.Post("/prospects/:id/menu-profile", authMiddleware.RequirePermission("view_ai_menu_profiling"), prospectHandler.ProfileMenu)
	ai.Post("/prospects/:id/find-menu", authMiddleware.RequirePermission("view_ai_menu_profiling"), prospectHandler.FindMenu)
	ai.Post("/prospects/:id/summary", authMiddleware.RequirePermission("view_ai_summary"), prospectHandler.GenerateSummary)
	ai.Get("/prospects/:id/chat/history", authMiddleware.RequirePermission("use_prospect_ai_chat"), prospectHandler.ChatAIHistory)
	if len(initialAnalyzers) > 0 && initialAnalyzers[0] != nil {
		ai.Get("/prospects/:id/initial-analysis", authMiddleware.RequirePermission("view_ai_summary"), prospectHandler.InitialAnalysis)
	}

	dashboard := api.Group("/dashboard", authMiddleware.Authenticate, authMiddleware.RequirePasswordChanged)
	dashboard.Get("/admin", authMiddleware.RequireRole(model.RoleSuperAdmin, model.RoleAdministrator), func(c *fiber.Ctx) error {
		return response.Data(c, fiber.StatusOK, fiber.Map{"surface": "administrator"})
	})
	dashboard.Get("/sales", authMiddleware.RequirePermission("view_sales_dashboard"), func(c *fiber.Ctx) error {
		return response.Data(c, fiber.StatusOK, fiber.Map{"surface": "sales-executive"})
	})
	dashboard.Get("/sales/team", authMiddleware.RequirePermission("view_team_dashboard"), prospectHandler.TeamDashboard)

	places := api.Group("/places", authMiddleware.Authenticate, authMiddleware.RequirePasswordChanged)
	places.Get("/photo", prospectHandler.PlacePhoto)

	customers := api.Group("/customers", authMiddleware.Authenticate, authMiddleware.RequirePasswordChanged)
	customers.Get("/team", authMiddleware.RequirePermission("view_team_dashboard"), customerHandler.TeamCustomers)

	sales := api.Group("/sales", authMiddleware.Authenticate, authMiddleware.RequirePasswordChanged)
	sales.Get("/prospects", authMiddleware.RequirePermission("view_my_prospects"), prospectHandler.MyProspects)
	sales.Get("/prospects/:id", authMiddleware.RequirePermission("view_my_prospect_detail"), prospectHandler.MyProspect)
	sales.Patch("/prospects/:id/transition", authMiddleware.RequirePermission("update_prospect_pipeline"), prospectHandler.Decide)
	sales.Patch("/prospects/:id/decision", authMiddleware.RequirePermission("update_prospect_pipeline"), prospectHandler.Decide)
	sales.Post("/prospects/:id/visits/check-in", authMiddleware.RequirePermission("check_in_prospect"), prospectHandler.CheckIn)
	sales.Patch("/prospects/:id/visits/:visitId/check-out", authMiddleware.RequirePermission("check_out_prospect"), prospectHandler.CheckOut)
	sales.Get("/prospects/:id/comments", authMiddleware.RequirePermission("view_my_prospect_detail"), prospectHandler.ListComments)
	sales.Post("/prospects/:id/comments", authMiddleware.RequirePermission("manage_prospect_comments"), prospectHandler.CreateComment)
	sales.Delete("/prospects/:id/comments/:commentId", authMiddleware.RequirePermission("manage_prospect_comments"), prospectHandler.DeleteComment)
	sales.Get("/prospects/:id/comments/attachments/:attachmentId", authMiddleware.RequirePermission("view_my_prospect_detail"), prospectHandler.CommentAttachment)
	sales.Get("/prospects/:id/photo-tags", authMiddleware.RequirePermission("view_my_prospect_detail"), prospectHandler.ListPhotoTags)
	sales.Put("/prospects/:id/photo-tags", authMiddleware.RequirePermission("view_my_prospect_detail"), prospectHandler.SetPhotoTag)
	sales.Get("/mention-users", authMiddleware.RequirePermission("view_my_prospects"), prospectHandler.MentionUsers)
	sales.Get("/prospects/:id/place-details", authMiddleware.RequirePermission("view_my_prospect_detail"), prospectHandler.ProspectPlaceDetails)
	sales.Post("/prospects/:id/request-deletion", authMiddleware.RequirePermission("request_prospect_deletion"), prospectHandler.RequestDeletion)
	sales.Get("/visits", authMiddleware.RequirePermission("view_own_visits"), prospectHandler.ListMyVisits)
	sales.Post("/visits/:visitId/delete", authMiddleware.RequirePermission("delete_visit"), prospectHandler.DeleteVisit)
	sales.Get("/customers", authMiddleware.RequirePermission("view_my_customers"), customerHandler.MyCustomers)
	sales.Get("/customers/:id", authMiddleware.RequirePermission("view_my_customer_detail"), customerHandler.MyCustomer)
	sales.Get("/customers/:id/place-details", authMiddleware.RequirePermission("view_my_customer_detail"), customerHandler.MyCustomerPlaceDetails)

	admin := api.Group("/admin", authMiddleware.Authenticate, authMiddleware.RequirePasswordChanged, authMiddleware.RequireRole(model.RoleSuperAdmin, model.RoleAdministrator))
	admin.Get("/prospects/won", prospectHandler.WonQueue)
	admin.Get("/prospects/pipeline", prospectHandler.Pipeline)
	admin.Get("/sales-executives", prospectHandler.SalesExecutives)
	admin.Get("/prospect-finder/search", prospectHandler.SearchPlaces)
	admin.Get("/prospect-finder/customers", prospectHandler.CustomerMarkers)
	admin.Get("/prospect-finder/places/:placeId", prospectHandler.PlaceDetail)
	admin.Get("/prospect-finder/place-details/:googlePlaceId", prospectHandler.PlaceFinderPlaceDetails)
	admin.Get("/prospect-finder/menu-images", prospectHandler.PlaceFinderMenuImages)
	admin.Post("/prospects", prospectHandler.Save)
	admin.Delete("/prospects/:id", prospectHandler.DeleteProspect)
	admin.Get("/prospects/:id", prospectHandler.Review)
	admin.Get("/prospects/:id/comments", prospectHandler.ListComments)
	admin.Get("/mention-users", prospectHandler.MentionUsers)
	admin.Post("/prospects/:id/comments", prospectHandler.CreateComment)
	admin.Delete("/prospects/:id/comments/:commentId", prospectHandler.DeleteComment)
	admin.Get("/prospects/:id/comments/attachments/:attachmentId", prospectHandler.CommentAttachment)
	admin.Get("/prospects/:id/photo-tags", prospectHandler.ListPhotoTags)
	admin.Put("/prospects/:id/photo-tags", prospectHandler.SetPhotoTag)
	admin.Get("/prospects/:id/place-details", prospectHandler.ProspectPlaceDetails)
	admin.Get("/visits", prospectHandler.ListVisitMonitoring)
	admin.Get("/reports", prospectHandler.Report)
	admin.Get("/prospects/:prospectId/visits", prospectHandler.ListProspectVisits)
	admin.Post("/visits/:visitId/delete", prospectHandler.DeleteVisit)
	admin.Post("/prospects/:id/approve-deletion", prospectHandler.ApproveDeletion)
	admin.Post("/prospects/:id/reject-deletion", prospectHandler.RejectDeletion)
	admin.Get("/prospects/:id/conversion-form", customerHandler.ConversionForm)
	admin.Post("/prospects/:id/convert", customerHandler.Convert)
	admin.Get("/parent-companies", customerHandler.SearchParentCompanies)
	admin.Get("/customers", customerHandler.AdminCustomers)
	admin.Get("/customers/list", customerHandler.AdminCustomersList)
	admin.Get("/customers/filter-options", customerHandler.CustomerFilterOptions)
	admin.Get("/customers/:id", customerHandler.AdminCustomerDetail)
	admin.Get("/customers/:id/place-details", customerHandler.AdminCustomerPlaceDetails)
	admin.Delete("/customers/:id", customerHandler.DeleteCustomer)
	admin.Get("/companies/:id", customerHandler.GetParentCompanyByCode)
	admin.Patch("/companies/:id", customerHandler.UpdateParentCompany)

	admin.Get("/permissions", adminHandler.ListPermissions)
	admin.Get("/sales-roles", adminHandler.ListSalesRoles)
	admin.Post("/sales-roles", adminHandler.CreateSalesRole)
	admin.Get("/sales-roles/:id", adminHandler.GetSalesRole)
	admin.Patch("/sales-roles/:id", adminHandler.UpdateSalesRole)
	admin.Patch("/sales-roles/:id/status", adminHandler.UpdateSalesRoleStatus)
	admin.Delete("/sales-roles/:id", adminHandler.DeleteSalesRole)
	admin.Get("/sales-structure", adminHandler.ListSalesStructure)
	admin.Post("/sales-structure/assignments", adminHandler.CreateSalesAssignment)
	admin.Post("/sales-structure/assignments/:id/move", adminHandler.MoveSalesAssignment)
	admin.Patch("/sales-structure/assignments/:id/end", adminHandler.EndSalesAssignment)
	admin.Get("/sales-structure/users/:userId/history", adminHandler.SalesAssignmentHistory)

	admin.Get("/users", adminHandler.ListUsers)
	admin.Get("/users/options/managers", adminHandler.ListManagers)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Get("/users/:id", adminHandler.GetUser)
	admin.Patch("/users/:id", adminHandler.UpdateUser)
	admin.Patch("/users/:id/status", adminHandler.UpdateStatus)
	admin.Delete("/users/:id", adminHandler.DeleteUser)
	admin.Post("/users/:id/reset-password", adminHandler.ResetPassword)

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
	slog.Error("request failed", "path", c.Path(), "error", err)
	return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
}
