package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"crm-prospect-simulator/backend/internal/admin/model"
	"crm-prospect-simulator/backend/internal/admin/repository"
	"crm-prospect-simulator/backend/internal/admin/service"
	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	actor := actor(c)
	filter := model.ListFilter{
		Page:      queryInt(c, "page", 1),
		Limit:     queryInt(c, "limit", 10),
		Search:    c.Query("search"),
		Role:      c.Query("role"),
		Status:    c.Query("status"),
		ManagerID: c.Query("managerId"),
	}
	result, err := h.svc.ListUsers(c.UserContext(), actor, filter)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, result)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	actor := actor(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "USER_ID_INVALID", "User ID is invalid.")
	}
	user, err := h.svc.GetUserDetail(c.UserContext(), actor, id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, user)
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	actor := actor(c)
	var input model.CreateUserInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	if input.ManagerID != nil && *input.ManagerID == uuid.Nil {
		input.ManagerID = nil
	}
	user, err := h.svc.CreateUser(c.UserContext(), actor, input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, user)
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	actor := actor(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "USER_ID_INVALID", "User ID is invalid.")
	}
	var input model.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	user, err := h.svc.UpdateUser(c.UserContext(), actor, id, input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, user)
}

func (h *Handler) UpdateUserProfile(c *fiber.Ctx) error {
	actor := actor(c)
	id, err := uuid.Parse(c.Params("id")); if err != nil { return response.Error(c, fiber.StatusBadRequest, "USER_ID_INVALID", "User ID is invalid.") }
	var input model.ProfileUpdateInput
	if err := c.BodyParser(&input); err != nil { return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.") }
	user, err := h.svc.UpdateUserProfile(c.UserContext(), actor, id, input); if err != nil { return writeError(c, err) }
	return response.Data(c, fiber.StatusOK, user)
}

func (h *Handler) UpdateStatus(c *fiber.Ctx) error {
	actor := actor(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "USER_ID_INVALID", "User ID is invalid.")
	}
	var input model.UpdateStatusInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	if input.Status != authmodel.UserActive && input.Status != authmodel.UserInactive {
		return response.Error(c, fiber.StatusUnprocessableEntity, "VALIDATION_FAILED", "Status must be ACTIVE or INACTIVE.")
	}
	user, err := h.svc.UpdateStatus(c.UserContext(), actor, id, input.Status)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, user)
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	actor := actor(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "USER_ID_INVALID", "User ID is invalid.")
	}
	if err := h.svc.DeleteUser(c.UserContext(), actor, id); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListManagers(c *fiber.Ctx) error {
	actor := actor(c)
	managers, err := h.svc.ListManagers(c.UserContext(), actor)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, managers)
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	actor := actor(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "USER_ID_INVALID", "User ID is invalid.")
	}
	var input model.ResetPasswordInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	result, err := h.svc.ResetPassword(c.UserContext(), actor, id, input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, result)
}

func actor(c *fiber.Ctx) service.Actor {
	principal, _ := authmiddleware.Principal(c)
	return service.Actor{UserID: principal.UserID, Role: principal.Role}
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	val := c.Query(key, "")
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func writeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to perform this action.")
	case errors.Is(err, service.ErrProtectedSuperAdmin):
		return response.Error(c, fiber.StatusForbidden, "PROTECTED_SUPER_ADMIN", "Yummy Super Admin cannot be changed by this action.")
	case errors.Is(err, service.ErrInvalidResetMode):
		return response.Error(c, fiber.StatusUnprocessableEntity, "INVALID_RESET_MODE", err.Error())
	case errors.Is(err, service.ErrTemporaryPasswordRequired):
		return response.Error(c, fiber.StatusUnprocessableEntity, "TEMPORARY_PASSWORD_REQUIRED", err.Error())
	case errors.Is(err, service.ErrWeakTemporaryPassword):
		return response.Error(c, fiber.StatusUnprocessableEntity, "WEAK_TEMPORARY_PASSWORD", err.Error())
	case errors.Is(err, service.ErrInvalidOrganizationalRole):
		return response.Error(c, fiber.StatusUnprocessableEntity, "INVALID_ORGANIZATIONAL_ROLE", err.Error())
	case errors.Is(err, service.ErrValidation):
		return response.Error(c, fiber.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
	case errors.Is(err, service.ErrSelfDeactivate):
		return response.Error(c, fiber.StatusUnprocessableEntity, "SELF_DEACTIVATE", "You cannot deactivate your own account.")
	case errors.Is(err, service.ErrLastAdmin):
		return response.Error(c, fiber.StatusUnprocessableEntity, "LAST_ADMIN", "Cannot deactivate the last active administrator.")
	case errors.Is(err, service.ErrInvalidManager):
		return response.Error(c, fiber.StatusUnprocessableEntity, "INVALID_MANAGER", err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", "User not found.")
	case errors.Is(err, repository.ErrConflict):
		return response.Error(c, fiber.StatusConflict, "RECORD_CONFLICT", "Account cannot be deleted because it is still referenced by existing records.")
	default:
		slog.Error("admin handler error", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
	}
}
