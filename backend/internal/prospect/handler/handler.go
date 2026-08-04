package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	customerrepo "crm-prospect-simulator/backend/internal/customer/repository"
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

type createCommentRequest struct {
	Content string `json:"content"`
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

func (h *Handler) resolveProspectID(ctx context.Context, id uuid.UUID, act service.Actor) (uuid.UUID, error) {
	if h.customerSvc == nil {
		return id, nil
	}
	custActor := customerservice.Actor{UserID: act.UserID, Role: act.Role}
	detail, err := h.customerSvc.MyCustomer(ctx, custActor, id)
	if err != nil {
		if errors.Is(err, customerrepo.ErrNotFound) || errors.Is(err, customerservice.ErrForbidden) {
			return id, nil
		}
		return uuid.UUID{}, err
	}
	return detail.Customer.SourceProspectID, nil
}

func (h *Handler) CheckIn(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}

	latitude, err := strconv.ParseFloat(c.FormValue("latitude"), 64)
	if err != nil {
		return response.Error(c, 400, "REQUEST_INVALID", "Latitude is required.")
	}
	longitude, err := strconv.ParseFloat(c.FormValue("longitude"), 64)
	if err != nil {
		return response.Error(c, 400, "REQUEST_INVALID", "Longitude is required.")
	}
	visitNotes := c.FormValue("visitNotes", "")

	input := prospectmodel.CheckInInput{
		Latitude:   latitude,
		Longitude:  longitude,
		VisitNotes: visitNotes,
	}

	file, err := c.FormFile("selfie")
	if err == nil && file != nil {
		uploadDir := filepath.Join("uploads", "selfies")
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return response.Error(c, 500, "UPLOAD_FAILED", "Unable to save selfie.")
		}
		filename := fmt.Sprintf("%s.jpg", uuid.New().String())
		dst := filepath.Join(uploadDir, filename)
		if err := c.SaveFile(file, dst); err != nil {
			return response.Error(c, 500, "UPLOAD_FAILED", "Unable to save selfie file.")
		}
		input.SelfieReference = "/uploads/selfies/" + filename
	}

	prospectID, err := h.resolveProspectID(c.UserContext(), id, actor(c))
	if err != nil {
		return writeError(c, err)
	}

	item, err := h.service.CheckIn(c.UserContext(), actor(c), prospectID, input)
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
	resolvedID, err := h.resolveProspectID(c.UserContext(), prospectID, actor(c))
	if err != nil {
		return writeError(c, err)
	}

	item, err := h.service.CheckOut(c.UserContext(), actor(c), resolvedID, visitID, input)
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

func (h *Handler) ListVisitMonitoring(c *fiber.Ctx) error {
	filter := prospectmodel.VisitMonitoringFilter{
		DateFrom:         c.Query("dateFrom"),
		DateTo:           c.Query("dateTo"),
		SalesExecutiveID: c.Query("salesExecutiveId"),
		CustomerName:     c.Query("customerName"),
		RadiusStatus:     c.Query("radiusStatus"),
	}
	items, err := h.service.ListVisitMonitoring(c.UserContext(), actor(c), filter)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) Report(c *fiber.Ctx) error {
	item, err := h.service.Report(c.UserContext(), actor(c), prospectmodel.ReportFilter{DateFrom: c.Query("dateFrom"), DateTo: c.Query("dateTo"), SalesExecutiveID: c.Query("salesExecutiveId"), Territory: c.Query("territory")})
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) ListMyVisits(c *fiber.Ctx) error {
	filter := prospectmodel.VisitMonitoringFilter{
		DateFrom:     c.Query("dateFrom"),
		DateTo:       c.Query("dateTo"),
		CustomerName: c.Query("customerName"),
	}
	items, err := h.service.ListMyVisits(c.UserContext(), actor(c), filter)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) ListProspectVisits(c *fiber.Ctx) error {
	prospectID, err := uuid.Parse(c.Params("prospectId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	items, err := h.service.ListProspectVisits(c.UserContext(), prospectID)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) DeleteVisit(c *fiber.Ctx) error {
	visitID, err := uuid.Parse(c.Params("visitId"))
	if err != nil {
		return response.Error(c, 400, "VISIT_ID_INVALID", "Visit ID is invalid.")
	}
	visit, err := h.service.DeleteVisit(c.UserContext(), actor(c), visitID)
	if err != nil {
		return writeError(c, err)
	}
	removeSelfieFiles(visit.SelfieReference)
	return response.Data(c, fiber.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) DeleteProspect(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	selfies, err := h.service.DeleteProspect(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	removeSelfieFiles(selfies...)
	return response.Data(c, fiber.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) RequestDeletion(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	if err := h.service.RequestDeletion(c.UserContext(), actor(c), id); err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, fiber.Map{"deletionRequested": true})
}

func (h *Handler) ApproveDeletion(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	selfies, err := h.service.ApproveDeletion(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	removeSelfieFiles(selfies...)
	return response.Data(c, fiber.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) RejectDeletion(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	if err := h.service.RejectDeletion(c.UserContext(), actor(c), id); err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, fiber.Map{"deletionRejected": true})
}

func (h *Handler) ListComments(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	items, err := h.service.ListComments(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) CreateComment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	var request createCommentRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.CreateComment(c.UserContext(), actor(c), id, request.Content)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, item)
}

func (h *Handler) ProspectPlaceDetails(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	item, err := h.service.Review(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	if item.Prospect.GooglePlaceID == "" {
		return response.Data(c, fiber.StatusOK, nil)
	}
	place, err := h.service.PlaceDetailFull(c.UserContext(), item.Prospect.GooglePlaceID)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, place)
}

func (h *Handler) PlaceFinderPlaceDetails(c *fiber.Ctx) error {
	placeID := c.Params("googlePlaceId")
	if strings.TrimSpace(placeID) == "" {
		return response.Error(c, 400, "PLACE_ID_REQUIRED", "Google Place ID is required.")
	}
	place, err := h.service.PlaceDetailFull(c.UserContext(), placeID)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, place)
}

func actor(c *fiber.Ctx) service.Actor {
	principal, _ := authmiddleware.Principal(c)
	return service.Actor{UserID: principal.UserID, Role: principal.Role}
}

func removeSelfieFiles(references ...string) {
	for _, reference := range references {
		if reference == "" || strings.HasPrefix(reference, "SIMULATED_") {
			continue
		}
		rel := filepath.FromSlash(strings.TrimPrefix(reference, "/"))
		if !strings.HasPrefix(filepath.Clean(rel), "uploads"+string(os.PathSeparator)) {
			continue
		}
		if err := os.Remove(rel); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("failed to remove selfie file", "reference", reference, "error", err)
		}
	}
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
		slog.Error("unhandled error", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
	}
}
