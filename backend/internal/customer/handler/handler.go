package handler

import (
	"errors"
	"strconv"

	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	customermodel "crm-prospect-simulator/backend/internal/customer/model"
	"crm-prospect-simulator/backend/internal/customer/repository"
	"crm-prospect-simulator/backend/internal/customer/service"
	prospectrepository "crm-prospect-simulator/backend/internal/prospect/repository"
	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.Service
}

func New(customerService *service.Service) *Handler {
	return &Handler{service: customerService}
}

func (h *Handler) ConversionForm(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	form, err := h.service.ConversionForm(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, form)
}

func (h *Handler) SearchParentCompanies(c *fiber.Ctx) error {
	items, err := h.service.SearchParentCompanies(c.UserContext(), actor(c), c.Query("search"))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) Convert(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var request customermodel.ConversionInput
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.Convert(c.UserContext(), actor(c), id, request)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, item)
}

func (h *Handler) AdminCustomers(c *fiber.Ctx) error {
	items, err := h.service.AdminCustomers(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) AdminCustomersList(c *fiber.Ctx) error {
	params := customermodel.CustomerListParams{
		Page:     queryInt(c, "page", 1),
		Limit:    queryInt(c, "limit", 20),
		Keyword:  c.Query("keyword", ""),
		Segment:  c.Query("segment", ""),
		Category: c.Query("category", ""),
		Sales:    c.Query("sales", ""),
		Region:   c.Query("region", ""),
		Sort:     c.Query("sort", ""),
	}
	result, err := h.service.AdminCustomersList(c.UserContext(), actor(c), params)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, result)
}

func (h *Handler) CustomerFilterOptions(c *fiber.Ctx) error {
	opts, err := h.service.CustomerFilterOptions(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, opts)
}

func (h *Handler) MyCustomers(c *fiber.Ctx) error {
	items, err := h.service.MyCustomers(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) MyCustomer(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "CUSTOMER_ID_INVALID", "Customer ID is invalid.")
	}
	item, err := h.service.MyCustomer(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) AdminCustomerDetail(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "CUSTOMER_ID_INVALID", "Customer ID is invalid.")
	}
	item, err := h.service.AdminCustomer(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func parseID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, response.Error(c, fiber.StatusBadRequest, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	return id, nil
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

func actor(c *fiber.Ctx) service.Actor {
	principal, _ := authmiddleware.Principal(c)
	return service.Actor{UserID: principal.UserID, Role: principal.Role}
}

func writeError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrForbidden), errors.Is(err, prospectservice.ErrForbidden):
		return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to perform this action.")
	case errors.Is(err, service.ErrValidation), errors.Is(err, service.ErrAssignmentOverlap):
		return response.Error(c, fiber.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
	case errors.Is(err, repository.ErrProspectNotWon):
		return response.Error(c, fiber.StatusConflict, "PROSPECT_NOT_WON", "Only a WON prospect can be converted.")
	case errors.Is(err, repository.ErrAlreadyConverted):
		return response.Error(c, fiber.StatusConflict, "PROSPECT_ALREADY_CONVERTED", "This prospect has already been converted.")
	case errors.Is(err, repository.ErrDuplicatePlace):
		return response.Error(c, fiber.StatusConflict, "GOOGLE_PLACE_ALREADY_CUSTOMER", "This Google Place is already linked to a Customer Existing record.")
	case errors.Is(err, repository.ErrCodeConflict):
		return response.Error(c, fiber.StatusConflict, "GENERATED_CODE_CONFLICT", "A generated code conflicts with an existing record.")
	case errors.Is(err, repository.ErrParentUnavailable):
		return response.Error(c, fiber.StatusUnprocessableEntity, "PARENT_COMPANY_INVALID", "Parent Company cannot be created or linked.")
	case errors.Is(err, repository.ErrSalesUnavailable):
		return response.Error(c, fiber.StatusUnprocessableEntity, "SALES_EXECUTIVE_INVALID", "Sales Executive must reference an active Sales Executive user.")
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, prospectrepository.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "RECORD_NOT_FOUND", "The requested record was not found.")
	default:
		return err
	}
}
