package handler

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	customerservice "crm-prospect-simulator/backend/internal/customer/service"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"crm-prospect-simulator/backend/internal/prospect/repository"
	"crm-prospect-simulator/backend/internal/prospect/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service     *service.Service
	customerSvc *customerservice.Service
}

func New(prospectService *service.Service, customerSvc *customerservice.Service) *Handler {
	return &Handler{service: prospectService, customerSvc: customerSvc}
}

type transitionRequest struct {
	Status prospectmodel.Status `json:"status"`
	Notes  string               `json:"notes"`
}

func (h *Handler) MyProspects(c *fiber.Ctx) error {
	items, err := h.service.MyProspects(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) MyProspect(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	item, err := h.service.MyProspect(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) Decide(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	var request transitionRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.Transition(c.UserContext(), actor(c), id, request.Status, request.Notes)
	if err != nil {
		return writeError(c, err)
	}
	if request.Status == prospectmodel.StatusWon {
		go h.autoConvert(context.Background(), id)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) autoConvert(ctx context.Context, prospectID uuid.UUID) {
	if h.customerSvc == nil {
		return
	}
	if _, err := h.customerSvc.AutoConvert(ctx, prospectID); err != nil {
		slog.Error("auto convert failed", "prospect_id", prospectID, "error", err)
	}
}

func (h *Handler) CheckIn(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	var input prospectmodel.CheckInInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, 400, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.CheckIn(c.UserContext(), actor(c), id, input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, item)
}

func (h *Handler) CheckOut(c *fiber.Ctx) error {
	prospectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	visitID, err := uuid.Parse(c.Params("visitId"))
	if err != nil {
		return response.Error(c, 400, "VISIT_ID_INVALID", "Visit ID is invalid.")
	}
	var input prospectmodel.CheckOutInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, 400, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.CheckOut(c.UserContext(), actor(c), prospectID, visitID, input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) Pipeline(c *fiber.Ctx) error {
	items, err := h.service.Pipeline(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) SalesExecutives(c *fiber.Ctx) error {
	items, err := h.service.SalesExecutives(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) SearchPlaces(c *fiber.Ctx) error {
	lat, latErr := strconv.ParseFloat(c.Query("latitude"), 64)
	lng, lngErr := strconv.ParseFloat(c.Query("longitude"), 64)
	radius, radiusErr := strconv.ParseFloat(c.Query("radius"), 64)
	if latErr != nil || lngErr != nil || radiusErr != nil {
		return response.Error(c, 422, "FINDER_QUERY_INVALID", "Latitude, longitude, and radius are required numbers.")
	}
	categories := c.Context().QueryArgs().PeekMulti("categories")
	categoryValues := make([]string, 0)
	for _, raw := range categories {
		for _, value := range strings.Split(string(raw), ",") {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				categoryValues = append(categoryValues, trimmed)
			}
		}
	}
	items, err := h.service.SearchPlaces(c.UserContext(), actor(c), prospectmodel.PlaceSearchInput{Keyword: c.Query("keyword"), Categories: categoryValues, Radius: radius, Latitude: lat, Longitude: lng})
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) PlaceDetail(c *fiber.Ctx) error {
	item, err := h.service.PlaceDetail(c.UserContext(), actor(c), c.Params("placeId"))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) Save(c *fiber.Ctx) error {
	var input prospectmodel.SaveProspectInput
	if err := c.BodyParser(&input); err != nil {
		return response.Error(c, 400, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.Save(c.UserContext(), actor(c), input)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, item)
}

func (h *Handler) WonQueue(c *fiber.Ctx) error {
	items, err := h.service.WonQueue(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	if h.customerSvc != nil {
		for _, item := range items {
			go h.autoConvert(context.Background(), item.ID)
		}
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) Review(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	item, err := h.service.Review(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func actor(c *fiber.Ctx) service.Actor {
	principal, _ := authmiddleware.Principal(c)
	return service.Actor{UserID: principal.UserID, Role: principal.Role}
}

func writeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to perform this action.")
	case errors.Is(err, service.ErrTransition), errors.Is(err, service.ErrNotesRequired), errors.Is(err, service.ErrFinderInput), errors.Is(err, service.ErrVisitCoordinates):
		return response.Error(c, fiber.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
	case errors.Is(err, service.ErrPlacesDisabled):
		return response.Error(c, fiber.StatusServiceUnavailable, "PLACES_NOT_CONFIGURED", err.Error())
	case errors.Is(err, service.ErrProspectStatus), errors.Is(err, repository.ErrInvalidStatus):
		return response.Error(c, fiber.StatusConflict, "PROSPECT_STATUS_INVALID", "The prospect stage changed or this transition is not allowed.")
	case errors.Is(err, repository.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "PROSPECT_NOT_FOUND", "Prospect was not found.")
	case errors.Is(err, repository.ErrDuplicate):
		return response.Error(c, fiber.StatusConflict, "PROSPECT_DUPLICATE", "This Google Place is already saved as a prospect.")
	case errors.Is(err, repository.ErrVisitOpen):
		return response.Error(c, fiber.StatusConflict, "VISIT_ALREADY_OPEN", "Check out the open visit before starting another one.")
	case errors.Is(err, repository.ErrVisitClosed):
		return response.Error(c, fiber.StatusConflict, "VISIT_ALREADY_CLOSED", "This visit is already checked out.")
	default:
		return err
	}
}
