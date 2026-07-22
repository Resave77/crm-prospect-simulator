package handler

import (
	"errors"
	"strings"
	"time"

	authmiddleware "crm-prospect-simulator/backend/internal/auth/middleware"
	"crm-prospect-simulator/backend/internal/auth/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
)

const refreshCookieName = "crm_refresh"

type Handler struct {
	auth         *service.AuthService
	cookieSecure bool
}

func New(auth *service.AuthService, cookieSecure bool) *Handler {
	return &Handler{auth: auth, cookieSecure: cookieSecure}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var request loginRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REQUEST_INVALID", "The request body is invalid.")
	}
	if !strings.Contains(request.Email, "@") || request.Password == "" {
		return response.Error(c, fiber.StatusUnprocessableEntity, "VALIDATION_FAILED", "Email and password are required.")
	}
	result, err := h.auth.Login(c.UserContext(), request.Email, request.Password, clientContext(c))
	if err != nil {
		if service.IsClientAuthError(err) {
			return response.Error(c, fiber.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect.")
		}
		return err
	}
	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	return response.Data(c, fiber.StatusOK, result)
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	raw := c.Cookies(refreshCookieName)
	result, err := h.auth.Refresh(c.UserContext(), raw, clientContext(c))
	if err != nil {
		h.clearRefreshCookie(c)
		if service.IsClientAuthError(err) || errors.Is(err, service.ErrInvalidToken) {
			return response.Error(c, fiber.StatusUnauthorized, "REFRESH_SESSION_INVALID", "The session is invalid or expired.")
		}
		return err
	}
	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	return response.Data(c, fiber.StatusOK, result)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	if err := h.auth.Logout(c.UserContext(), c.Cookies(refreshCookieName)); err != nil {
		return err
	}
	h.clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) LogoutAll(c *fiber.Ctx) error {
	principal, ok := authmiddleware.Principal(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication is required.")
	}
	if err := h.auth.LogoutAll(c.UserContext(), principal); err != nil {
		return err
	}
	h.clearRefreshCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	principal, ok := authmiddleware.Principal(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication is required.")
	}
	user, err := h.auth.Me(c.UserContext(), principal)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "ACCESS_TOKEN_INVALID", "The access token is invalid or expired.")
	}
	return response.Data(c, fiber.StatusOK, user)
}

func (h *Handler) setRefreshCookie(c *fiber.Ctx, value string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name: refreshCookieName, Value: value, Path: "/api/v1/auth", HTTPOnly: true,
		Secure: h.cookieSecure, SameSite: fiber.CookieSameSiteStrictMode, Expires: expiresAt,
	})
}

func (h *Handler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name: refreshCookieName, Value: "", Path: "/api/v1/auth", HTTPOnly: true,
		Secure: h.cookieSecure, SameSite: fiber.CookieSameSiteStrictMode, Expires: time.Unix(0, 0),
	})
}

func clientContext(c *fiber.Ctx) service.ClientContext {
	return service.ClientContext{UserAgent: c.Get(fiber.HeaderUserAgent), IPAddress: c.IP()}
}
