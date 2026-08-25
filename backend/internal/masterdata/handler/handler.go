package handler

import (
	"errors"
	"log/slog"

	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	"crm-prospect-simulator/backend/internal/masterdata/model"
	"crm-prospect-simulator/backend/internal/masterdata/repository"
	"crm-prospect-simulator/backend/internal/masterdata/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) ListSegments(c *fiber.Ctx) error {
	items, err := h.service.ListSegments(c.UserContext(), actor(c), c.Query("search"), c.Query("status"))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) CreateSegment(c *fiber.Ctx) error {
	var input model.SegmentInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.CreateSegment(c.UserContext(), actor(c), input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, item)
}

func (h *Handler) UpdateSegment(c *fiber.Ctx) error {
	id, err := parseID(c, "SEGMENT_ID_INVALID")
	if err != nil {
		return err
	}
	var input model.SegmentInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.UpdateSegment(c.UserContext(), actor(c), id, input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) DeleteSegment(c *fiber.Ctx) error {
	id, err := parseID(c, "SEGMENT_ID_INVALID")
	if err != nil {
		return err
	}
	if err := h.service.DeleteSegment(c.UserContext(), actor(c), id); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListTrashedSegments(c *fiber.Ctx) error {
	items, err := h.service.ListTrashedSegments(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) RestoreSegment(c *fiber.Ctx) error {
	id, err := parseID(c, "SEGMENT_ID_INVALID")
	if err != nil {
		return err
	}
	if err := h.service.RestoreSegment(c.UserContext(), actor(c), id); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) SegmentCategories(c *fiber.Ctx) error {
	id, err := parseID(c, "SEGMENT_ID_INVALID")
	if err != nil {
		return err
	}
	items, err := h.service.ListCategoriesBySegment(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) ListCategories(c *fiber.Ctx) error {
	items, err := h.service.ListCategories(c.UserContext(), actor(c), c.Query("search"), c.Query("segmentId"), c.Query("status"))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) CreateCategory(c *fiber.Ctx) error {
	var input model.CategoryInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.CreateCategory(c.UserContext(), actor(c), input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, item)
}

func (h *Handler) UpdateCategory(c *fiber.Ctx) error {
	id, err := parseID(c, "CATEGORY_ID_INVALID")
	if err != nil {
		return err
	}
	var input model.CategoryInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.UpdateCategory(c.UserContext(), actor(c), id, input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) DeleteCategory(c *fiber.Ctx) error {
	id, err := parseID(c, "CATEGORY_ID_INVALID")
	if err != nil {
		return err
	}
	if err := h.service.DeleteCategory(c.UserContext(), actor(c), id); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListTrashedCategories(c *fiber.Ctx) error {
	items, err := h.service.ListTrashedCategories(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) RestoreCategory(c *fiber.Ctx) error {
	id, err := parseID(c, "CATEGORY_ID_INVALID")
	if err != nil {
		return err
	}
	if err := h.service.RestoreCategory(c.UserContext(), actor(c), id); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func actor(c *fiber.Ctx) service.Actor {
	principal, _ := authmiddleware.Principal(c)
	return service.Actor{UserID: principal.UserID, Role: principal.Role}
}

func parseID(c *fiber.Ctx, code string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, response.Error(c, fiber.StatusBadRequest, code, "The provided ID is invalid.")
	}
	return id, nil
}

func writeError(c *fiber.Ctx, err error) error {
	var validationErr *service.ValidationError
	switch {
	case errors.Is(err, service.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to perform this action.")
	case errors.As(err, &validationErr):
		return response.Error(c, fiber.StatusUnprocessableEntity, "VALIDATION_FAILED", validationErr.Error())
	case errors.Is(err, repository.ErrDuplicateName):
		return response.Error(c, fiber.StatusConflict, "MASTER_DATA_DUPLICATE", "A record with the same name already exists.")
	case errors.Is(err, repository.ErrSegmentInUse):
		return response.Error(c, fiber.StatusConflict, "SEGMENT_IN_USE", "This segment still has categories. Move or delete them first.")
	case errors.Is(err, repository.ErrParentInTrash):
		return response.Error(c, fiber.StatusConflict, "PARENT_IN_TRASH", "The parent segment is still in Trash. Restore it first.")
	case errors.Is(err, repository.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "RECORD_NOT_FOUND", "The requested record was not found.")
	default:
		slog.Error("unhandled master data error", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
	}
}
