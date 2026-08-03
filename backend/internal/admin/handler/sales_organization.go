package handler

import (
	"errors"
	"time"

	"crm-prospect-simulator/backend/internal/admin/model"
	"crm-prospect-simulator/backend/internal/admin/repository"
	"crm-prospect-simulator/backend/internal/admin/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (h *Handler) ListSalesRoles(c *fiber.Ctx) error {
	roles, err := h.svc.ListSalesRoles(c.UserContext(), actor(c))
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusOK, roles)
}

func (h *Handler) CreateSalesRole(c *fiber.Ctx) error {
	var input model.CreateSalesRoleInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	role, err := h.svc.CreateSalesRole(c.UserContext(), actor(c), input)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, role)
}

func (h *Handler) GetSalesRole(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "SALES_ROLE_ID_INVALID", "Sales role ID is invalid.")
	}
	role, err := h.svc.GetSalesRole(c.UserContext(), actor(c), id)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusOK, role)
}

func (h *Handler) UpdateSalesRole(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "SALES_ROLE_ID_INVALID", "Sales role ID is invalid.")
	}
	var input model.UpdateSalesRoleInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	role, err := h.svc.UpdateSalesRole(c.UserContext(), actor(c), id, input)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusOK, role)
}

func (h *Handler) UpdateSalesRoleStatus(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "SALES_ROLE_ID_INVALID", "Sales role ID is invalid.")
	}
	var input model.UpdateSalesRoleStatusInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	role, err := h.svc.UpdateSalesRoleStatus(c.UserContext(), actor(c), id, input.IsActive)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusOK, role)
}

func (h *Handler) DeleteSalesRole(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "SALES_ROLE_ID_INVALID", "Sales role ID is invalid.")
	}
	if err := h.svc.DeleteSalesRole(c.UserContext(), actor(c), id); err != nil {
		return writeSalesError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListSalesStructure(c *fiber.Ctx) error {
	date := time.Now()
	if raw := c.Query("effectiveDate"); raw != "" {
		parsed, err := time.Parse(model.DateLayout, raw)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "EFFECTIVE_DATE_INVALID", "Effective date must be YYYY-MM-DD.")
		}
		date = parsed
	}
	items, err := h.svc.ListSalesStructure(c.UserContext(), actor(c), date)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) CreateSalesAssignment(c *fiber.Ctx) error {
	var input model.CreateAssignmentInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	assignment, err := h.svc.CreateSalesAssignment(c.UserContext(), actor(c), input)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, assignment)
}

func (h *Handler) MoveSalesAssignment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "ASSIGNMENT_ID_INVALID", "Assignment ID is invalid.")
	}
	var input model.MoveAssignmentInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	assignment, err := h.svc.MoveSalesAssignment(c.UserContext(), actor(c), id, input)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, assignment)
}

func (h *Handler) SalesAssignmentHistory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "USER_ID_INVALID", "User ID is invalid.")
	}
	items, err := h.svc.ListSalesAssignmentHistory(c.UserContext(), actor(c), id)
	if err != nil {
		return writeSalesError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func writeSalesError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to perform this action.")
	case errors.Is(err, service.ErrSalesRoleNameRequired):
		return response.Error(c, fiber.StatusUnprocessableEntity, "SALES_ROLE_NAME_REQUIRED", err.Error())
	case errors.Is(err, service.ErrInvalidSalesRoleLevel):
		return response.Error(c, fiber.StatusUnprocessableEntity, "INVALID_SALES_ROLE_LEVEL", err.Error())
	case errors.Is(err, service.ErrSalesRoleNameExists):
		return response.Error(c, fiber.StatusConflict, "SALES_ROLE_NAME_EXISTS", err.Error())
	case errors.Is(err, service.ErrSalesRoleInUse):
		return response.Error(c, fiber.StatusUnprocessableEntity, "SALES_ROLE_IN_USE", err.Error())
	case errors.Is(err, service.ErrSalesRoleInactive):
		return response.Error(c, fiber.StatusUnprocessableEntity, "SALES_ROLE_INACTIVE", err.Error())
	case errors.Is(err, service.ErrInvalidHierarchy):
		return response.Error(c, fiber.StatusUnprocessableEntity, "INVALID_SALES_HIERARCHY", err.Error())
	case errors.Is(err, service.ErrInvalidEffectiveDate):
		return response.Error(c, fiber.StatusUnprocessableEntity, "INVALID_EFFECTIVE_DATE", err.Error())
	case errors.Is(err, service.ErrAssignmentOverlap):
		return response.Error(c, fiber.StatusConflict, "SALES_ASSIGNMENT_OVERLAP", err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "SALES_RESOURCE_NOT_FOUND", "Sales organization resource not found.")
	case errors.Is(err, repository.ErrConflict):
		return response.Error(c, fiber.StatusConflict, "RECORD_CONFLICT", "A sales organization record conflicts with existing data.")
	default:
		return writeError(c, err)
	}
}
