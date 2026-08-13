package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	aiservice "crm-prospect-simulator/backend/internal/ai/service"
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
	service         *service.Service
	customerSvc     *customerservice.Service
	initialAnalyzer *aiservice.InitialAnalyzer
}

type createCommentRequest struct {
	Content string `json:"content"`
}

type chatRequest struct {
	Message string `json:"message"`
	Skill   string `json:"skill"`
}

type setPhotoTagRequest struct {
	PhotoName  string `json:"photoName"`
	PhotoIndex *int   `json:"photoIndex"`
	Category   string `json:"category"`
}

func New(prospectService *service.Service, customerSvc *customerservice.Service) *Handler {
	return &Handler{service: prospectService, customerSvc: customerSvc}
}

func (h *Handler) SetInitialAnalyzer(analyzer *aiservice.InitialAnalyzer) {
	h.initialAnalyzer = analyzer
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

func (h *Handler) TeamDashboard(c *fiber.Ctx) error {
	item, err := h.service.TeamDashboard(c.UserContext(), actor(c))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
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

func (h *Handler) ChatAI(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	var input chatRequest
	if err := c.BodyParser(&input); err != nil || strings.TrimSpace(input.Message) == "" {
		return response.Error(c, fiber.StatusBadRequest, "AI_MESSAGE_INVALID", "A question is required.")
	}
	result, err := h.service.ChatAI(c.UserContext(), actor(c), id, input.Message, input.Skill)
	if err != nil {
		return writeError(c, err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		return writeError(c, aiservice.ErrAIInvalidResponse)
	}
	return response.Data(c, fiber.StatusOK, payload)
}

func (h *Handler) ChatAIHistory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	items, err := h.service.AIChatHistory(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) InitialAnalysis(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	if h.initialAnalyzer == nil {
		return response.Error(c, fiber.StatusNotFound, "AI_ANALYSIS_NOT_AVAILABLE", "AI analysis is not available.")
	}
	if err := h.service.AuthorizeProspectAccess(c.UserContext(), actor(c), id); err != nil {
		return writeError(c, err)
	}
	item, err := h.initialAnalyzer.Get(c.UserContext(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "AI_ANALYSIS_NOT_AVAILABLE", "AI analysis is not available.")
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
	custActor := customerservice.Actor{UserID: act.UserID, Role: act.Role, PermissionKeys: act.PermissionKeys}
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

func (h *Handler) CustomerMarkers(c *fiber.Ctx) error {
	items, err := h.service.CustomerMarkers(c.UserContext(), actor(c))
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

func (h *Handler) PlacePhoto(c *fiber.Ctx) error {
	data, contentType, err := h.service.PlacePhoto(c.UserContext(), c.Query("name"))
	if err != nil {
		return writeError(c, err)
	}
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderCacheControl, "private, max-age=86400")
	return c.Send(data)
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

func (h *Handler) MentionUsers(c *fiber.Ctx) error {
	items, err := h.service.MentionUsers(c.UserContext(), actor(c))
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
	request := createCommentRequest{Content: c.FormValue("content")}
	attachments := make([]prospectmodel.CommentAttachment, 0)
	if strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		if err := c.BodyParser(&request); err != nil {
			return response.Error(c, 400, "REQUEST_INVALID", "The request body is invalid.")
		}
	} else if form, err := c.MultipartForm(); err == nil {
		files := form.File["files"]
		if len(files) > 5 {
			return response.Error(c, 422, "TOO_MANY_FILES", "A comment can contain up to 5 files.")
		}
		workingDir, err := os.Getwd()
		if err != nil {
			return response.Error(c, 500, "UPLOAD_FAILED", "Unable to prepare attachment storage.")
		}
		if filepath.Base(workingDir) != "backend" {
			workingDir = filepath.Join(workingDir, "backend")
		}
		dir := filepath.Join(workingDir, "private_uploads", "ticketing", id.String())
		if len(files) > 0 {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return response.Error(c, 500, "UPLOAD_FAILED", "Unable to prepare attachment storage.")
			}
		}
		for _, file := range files {
			if file.Size > 5*1024*1024 {
				removeCommentFiles(attachments)
				return response.Error(c, 422, "FILE_TOO_LARGE", "Each attachment must be 5 MB or smaller.")
			}
			src, err := file.Open()
			if err != nil {
				removeCommentFiles(attachments)
				return response.Error(c, 422, "FILE_INVALID", "Unable to read attachment.")
			}
			head := make([]byte, 512)
			n, readErr := src.Read(head)
			_ = src.Close()
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				removeCommentFiles(attachments)
				return response.Error(c, 422, "FILE_INVALID", "Unable to read attachment.")
			}
			contentType := http.DetectContentType(head[:n])
			ext := strings.ToLower(filepath.Ext(filepath.Base(file.Filename)))
			allowed := (contentType == "image/jpeg" && (ext == ".jpg" || ext == ".jpeg")) || (contentType == "image/png" && ext == ".png") || (contentType == "application/pdf" && ext == ".pdf") || ((contentType == "application/zip" || contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document") && ext == ".docx") || ((contentType == "application/zip" || contentType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") && ext == ".xlsx")
			if !allowed {
				removeCommentFiles(attachments)
				return response.Error(c, 422, "FILE_TYPE_INVALID", "Only JPG, PNG, PDF, DOCX, and XLSX files are allowed.")
			}
			attachmentID := uuid.New()
			path := filepath.Join(dir, attachmentID.String()+ext)
			if err := c.SaveFile(file, path); err != nil {
				removeCommentFiles(attachments)
				return response.Error(c, 500, "UPLOAD_FAILED", "Unable to save attachment.")
			}
			if _, err := os.Stat(path); err != nil {
				removeCommentFiles(attachments)
				return response.Error(c, 500, "UPLOAD_FAILED", "Attachment was not persisted.")
			}
			attachments = append(attachments, prospectmodel.CommentAttachment{ID: attachmentID, Name: filepath.Base(file.Filename), ContentType: contentType, Size: file.Size, Path: filepath.ToSlash(path)})
		}
	}
	item, err := h.service.CreateComment(c.UserContext(), actor(c), id, request.Content, attachments)
	if err != nil {
		removeCommentFiles(attachments)
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusCreated, item)
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	prospectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	commentID, err := uuid.Parse(c.Params("commentId"))
	if err != nil {
		return response.Error(c, 400, "COMMENT_ID_INVALID", "Comment ID is invalid.")
	}
	attachments, err := h.service.DeleteComment(c.UserContext(), actor(c), prospectID, commentID)
	if err != nil {
		return writeError(c, err)
	}
	removeCommentFiles(attachments)
	return response.Data(c, fiber.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) CommentAttachment(c *fiber.Ctx) error {
	prospectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	attachmentID, err := uuid.Parse(c.Params("attachmentId"))
	if err != nil {
		return response.Error(c, 400, "ATTACHMENT_ID_INVALID", "Attachment ID is invalid.")
	}
	item, err := h.service.CommentAttachment(c.UserContext(), actor(c), prospectID, attachmentID)
	if err != nil {
		return writeError(c, err)
	}
	path := item.Path
	if !filepath.IsAbs(path) {
		for _, candidate := range []string{path, filepath.Join("backend", path)} {
			if _, statErr := os.Stat(candidate); statErr == nil {
				path = candidate
				break
			}
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return response.Error(c, fiber.StatusNotFound, "ATTACHMENT_NOT_FOUND", "The attachment file is no longer available.")
		}
		return response.Error(c, fiber.StatusInternalServerError, "ATTACHMENT_READ_FAILED", "Unable to open the attachment.")
	}
	disposition := "attachment"
	if strings.HasPrefix(item.ContentType, "image/") {
		disposition = "inline"
	}
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`%s; filename="%s"`, disposition, strings.ReplaceAll(item.Name, `"`, "")))
	c.Set(fiber.HeaderContentType, item.ContentType)
	return c.Send(data)
}

func removeCommentFiles(items []prospectmodel.CommentAttachment) {
	for _, item := range items {
		_ = os.Remove(item.Path)
	}
}

func (h *Handler) ListPhotoTags(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	items, err := h.service.ListPhotoTags(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func (h *Handler) ProfileMenu(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	var request struct {
		Force bool `json:"force"`
	}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&request); err != nil {
			return response.Error(c, 400, "REQUEST_INVALID", "The request body is invalid.")
		}
	}
	result, err := h.service.ProfileMenu(c.UserContext(), actor(c), id, request.Force)
	if err != nil {
		return writeError(c, err)
	}
	var data any
	if json.Unmarshal(result, &data) != nil {
		return response.Error(c, 503, "AI_INVALID_RESPONSE", "AI menu profiling returned invalid data.")
	}
	return response.Data(c, fiber.StatusOK, data)
}

func (h *Handler) FindMenu(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	result, err := h.service.FindMenu(c.UserContext(), actor(c), id)
	if err != nil {
		switch {
		case errors.Is(err, aiservice.ErrAITimeout):
			return response.Error(c, fiber.StatusServiceUnavailable, aiservice.SafeErrorCode(err), "Pencarian menu membutuhkan waktu lebih lama dari biasanya.")
		case errors.Is(err, aiservice.ErrAIRateLimited), errors.Is(err, aiservice.ErrAIUnavailable):
			return response.Error(c, fiber.StatusServiceUnavailable, aiservice.SafeErrorCode(err), "Layanan pencarian menu sedang sibuk. Silakan coba lagi.")
		case errors.Is(err, aiservice.ErrAINotConfigured), errors.Is(err, aiservice.ErrAIInvalidResponse), errors.Is(err, aiservice.ErrAIAuthentication), errors.Is(err, aiservice.ErrAIRequestRejected):
			return response.Error(c, fiber.StatusServiceUnavailable, aiservice.SafeErrorCode(err), "Pencarian menu belum dapat diselesaikan.")
		default:
			return writeError(c, err)
		}
	}
	var data any
	if json.Unmarshal(result, &data) != nil {
		return response.Error(c, 503, "MENU_FINDER_INVALID_RESPONSE", "Pencarian menu belum dapat diselesaikan.")
	}
	return response.Data(c, fiber.StatusOK, data)
}

func (h *Handler) GenerateSummary(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	summary, err := h.service.GenerateSummary(c.UserContext(), actor(c), id)
	if err != nil {
		return writeError(c, err)
	}
	var data any
	if json.Unmarshal(summary, &data) != nil {
		return response.Error(c, 503, "AI_INVALID_RESPONSE", "AI summary returned invalid data.")
	}
	return response.Data(c, fiber.StatusOK, data)
}

func (h *Handler) SetPhotoTag(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	var request setPhotoTagRequest
	if err := c.BodyParser(&request); err != nil {
		return response.Error(c, 400, "REQUEST_INVALID", "The request body is invalid.")
	}
	item, err := h.service.SetPhotoTag(c.UserContext(), actor(c), id, request.PhotoName, request.PhotoIndex, prospectmodel.PhotoCategory(strings.ToUpper(request.Category)))
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, item)
}

func (h *Handler) ProspectPlaceDetails(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, 400, "PROSPECT_ID_INVALID", "Prospect ID is invalid.")
	}
	currentActor := actor(c)
	var item prospectmodel.Review
	if currentActor.Role.IsAdminRole() {
		item, err = h.service.Review(c.UserContext(), currentActor, id)
	} else {
		item, err = h.service.MyProspect(c.UserContext(), currentActor, id)
	}
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

func (h *Handler) PlaceFinderMenuImages(c *fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		return response.Error(c, 400, "QUERY_REQUIRED", "Search query is required.")
	}
	limit, _ := strconv.Atoi(c.Query("limit", "8"))
	items, err := h.service.MenuImages(c.UserContext(), query, limit)
	if err != nil {
		return writeError(c, err)
	}
	return response.Data(c, fiber.StatusOK, items)
}

func actor(c *fiber.Ctx) service.Actor {
	principal, _ := authmiddleware.Principal(c)
	var permissionKeys []string
	if principal.SalesRole != nil {
		permissionKeys = principal.SalesRole.PermissionKeys
	}
	return service.Actor{UserID: principal.UserID, Role: principal.Role, PermissionKeys: permissionKeys}
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
	case errors.Is(err, aiservice.ErrAINotConfigured), errors.Is(err, aiservice.ErrAIUnavailable), errors.Is(err, aiservice.ErrAITimeout), errors.Is(err, aiservice.ErrAIRateLimited), errors.Is(err, aiservice.ErrAIInvalidResponse), errors.Is(err, aiservice.ErrAIAuthentication), errors.Is(err, aiservice.ErrAIRequestRejected):
		return response.Error(c, fiber.StatusServiceUnavailable, aiservice.SafeErrorCode(err), "AI is temporarily unavailable.")
	case errors.Is(err, service.ErrTransition), errors.Is(err, service.ErrNotesRequired), errors.Is(err, service.ErrFinderInput), errors.Is(err, service.ErrVisitCoordinates), errors.Is(err, service.ErrPhotoTagInvalid), errors.Is(err, service.ErrPlacePhotoInvalid):
		return response.Error(c, fiber.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
	case errors.Is(err, service.ErrPlacesDisabled):
		return response.Error(c, fiber.StatusServiceUnavailable, "PLACES_NOT_CONFIGURED", err.Error())
	case errors.Is(err, service.ErrMenuDataNotAvailable):
		return response.Error(c, fiber.StatusUnprocessableEntity, "MENU_DATA_NOT_AVAILABLE", "No usable MENU-tagged photo is available.")
	case errors.Is(err, service.ErrMenuImagesDisabled):
		return response.Error(c, fiber.StatusServiceUnavailable, "MENU_IMAGES_NOT_CONFIGURED", err.Error())
	case errors.Is(err, service.ErrPlacePhotoUnavailable):
		return response.Error(c, fiber.StatusBadGateway, "PLACE_PHOTO_UNAVAILABLE", "The place photo is temporarily unavailable.")
	case errors.Is(err, service.ErrAlreadyCustomer):
		return response.Error(c, fiber.StatusConflict, "ALREADY_CUSTOMER", "This place is already an existing customer and cannot be assigned to sales.")
	case errors.Is(err, service.ErrProspectStatus), errors.Is(err, repository.ErrInvalidStatus):
		return response.Error(c, fiber.StatusConflict, "PROSPECT_STATUS_INVALID", "The prospect stage changed or this transition is not allowed.")
	case errors.Is(err, repository.ErrNotFound):
		return response.Error(c, fiber.StatusNotFound, "PROSPECT_NOT_FOUND", "Prospect was not found.")
	case errors.Is(err, repository.ErrDuplicate):
		return response.Error(c, fiber.StatusConflict, "PROSPECT_DUPLICATE", "This Google Place is already saved as a prospect.")
	case errors.Is(err, repository.ErrConflict):
		return response.Error(c, fiber.StatusConflict, "RECORD_CONFLICT", "Prospect cannot be deleted because it is still referenced by existing records.")
	case errors.Is(err, repository.ErrVisitOpen):
		return response.Error(c, fiber.StatusConflict, "VISIT_ALREADY_OPEN", "Check out the open visit before starting another one.")
	case errors.Is(err, repository.ErrVisitClosed):
		return response.Error(c, fiber.StatusConflict, "VISIT_ALREADY_CLOSED", "This visit is already checked out.")
	default:
		slog.Error("unhandled error", "error", err)
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
	}
}
