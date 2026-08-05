package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"crm-prospect-simulator/backend/config"
	authmodel "crm-prospect-simulator/backend/internal/auth/model"
	authservice "crm-prospect-simulator/backend/internal/auth/service"
	adminservice "crm-prospect-simulator/backend/internal/admin/service"
	prospectrepo "crm-prospect-simulator/backend/internal/prospect/repository"
	prospectservice "crm-prospect-simulator/backend/internal/prospect/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type stubProspectRepo struct {
	prospectrepo.Repository
	deleteErr   error
	approveErr  error
	deleteCalls int
	approveCalls int
}

func (r *stubProspectRepo) DeleteProspect(_ context.Context, _ uuid.UUID) ([]string, error) {
	r.deleteCalls++
	return nil, r.deleteErr
}

func (r *stubProspectRepo) ApproveDeletion(_ context.Context, _ uuid.UUID) ([]string, error) {
	r.approveCalls++
	return nil, r.approveErr
}

func buildTestAppWithProspects(user authmodel.User, prospectRepo prospectrepo.Repository) (*fiber.App, string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	user.PasswordHash = string(hash)
	if user.Status == "" {
		user.Status = authmodel.UserActive
	}
	users := &stubUserRepo{user: user}
	sessions := &stubSessionRepo{}
	jwtSecret := "01234567890123456789012345678901"
	tokens := authservice.NewTokenManager(jwtSecret, "test", "test-api", time.Minute)
	authSvc := authservice.NewAuthService(users, sessions, tokens, time.Hour)
	prospectSvc := prospectservice.New(prospectRepo)
	adminSvc := adminservice.New(&adminRepoStub{})
	app := New(config.Config{AllowedOrigins: "http://localhost:5173"}, authSvc, prospectSvc, nil, adminSvc)
	access, _, _ := tokens.IssueAccess(user, uuid.New(), time.Now())
	return app, access
}

func prospectDeleteRequest(t *testing.T, app *fiber.App, token, method, path string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestAdminDirectDeleteProspectSucceeds(t *testing.T) {
	repo := &stubProspectRepo{}
	app, token := buildTestAppWithProspects(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, TokenVersion: 1}, repo)

	status, body := prospectDeleteRequest(t, app, token, http.MethodDelete, "/api/v1/admin/prospects/"+uuid.New().String())
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s, want 200", status, body)
	}
	if !strings.Contains(body, `"deleted":true`) {
		t.Fatalf("body=%s, want deleted:true", body)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("repo delete calls=%d, want 1", repo.deleteCalls)
	}
}

func TestDeleteProspectUnknownReturnsNotFound(t *testing.T) {
	repo := &stubProspectRepo{deleteErr: prospectrepo.ErrNotFound}
	app, token := buildTestAppWithProspects(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, TokenVersion: 1}, repo)

	status, body := prospectDeleteRequest(t, app, token, http.MethodDelete, "/api/v1/admin/prospects/"+uuid.New().String())
	if status != http.StatusNotFound {
		t.Fatalf("status=%d body=%s, want 404", status, body)
	}
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if envelope.Error.Code != "PROSPECT_NOT_FOUND" {
		t.Fatalf("code=%q, want PROSPECT_NOT_FOUND", envelope.Error.Code)
	}
}

func TestDeleteProspectWithRelatedDataReturnsConflict(t *testing.T) {
	repo := &stubProspectRepo{deleteErr: prospectrepo.ErrConflict}
	app, token := buildTestAppWithProspects(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, TokenVersion: 1}, repo)

	status, body := prospectDeleteRequest(t, app, token, http.MethodDelete, "/api/v1/admin/prospects/"+uuid.New().String())
	if status != http.StatusConflict {
		t.Fatalf("status=%d body=%s, want 409", status, body)
	}
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if envelope.Error.Code != "RECORD_CONFLICT" {
		t.Fatalf("code=%q, want RECORD_CONFLICT", envelope.Error.Code)
	}
	if !strings.Contains(envelope.Error.Message, "cannot be deleted") {
		t.Fatalf("message=%q, want a clear cannot-be-deleted message", envelope.Error.Message)
	}
}

func TestApproveProspectDeletionStillWorks(t *testing.T) {
	repo := &stubProspectRepo{}
	app, token := buildTestAppWithProspects(authmodel.User{ID: uuid.New(), Email: "admin@test.test", Role: authmodel.RoleAdministrator, TokenVersion: 1}, repo)

	status, body := prospectDeleteRequest(t, app, token, http.MethodPost, "/api/v1/admin/prospects/"+uuid.New().String()+"/approve-deletion")
	if status != http.StatusOK {
		t.Fatalf("status=%d body=%s, want 200", status, body)
	}
	if !strings.Contains(body, `"deleted":true`) {
		t.Fatalf("body=%s, want deleted:true", body)
	}
	if repo.approveCalls != 1 {
		t.Fatalf("repo approve calls=%d, want 1", repo.approveCalls)
	}
}
