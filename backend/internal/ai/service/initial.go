package service

import (
	"context"
	"encoding/json"
	"time"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const menuDataNotAvailable = "MENU_DATA_NOT_AVAILABLE"

var menuDataNotAvailableJSON = json.RawMessage(`{"state":"MENU_DATA_NOT_AVAILABLE","message":"Menu data not available yet. Add structured menu data or explicit menu inputs before running menu profiling."}`)

type InitialAnalysis struct {
	ProspectID uuid.UUID       `json:"prospectId"`
	Summary    json.RawMessage `json:"summary"`
	Menu       json.RawMessage `json:"menu"`
	Status     string          `json:"status"`
	ErrorCode  string          `json:"errorCode,omitempty"`
	CreatedAt  *time.Time      `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time      `json:"updatedAt,omitempty"`
}

type InitialAnalyzer struct {
	pool   *pgxpool.Pool
	client *ProspectAI
}

func NewInitialAnalyzer(pool *pgxpool.Pool, client *ProspectAI) *InitialAnalyzer {
	return &InitialAnalyzer{pool: pool, client: client}
}

func (a *InitialAnalyzer) Analyze(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment) {
	if a == nil || a.pool == nil || a.client == nil || !a.client.client.Configured() {
		return
	}
	id := review.Prospect.ID
	if reusable, err := a.hasSuccessfulSummary(ctx, id); err != nil || reusable {
		return
	}
	_ = a.setStatus(ctx, id, "PENDING", "")
	summary, err := a.client.SummarizeProspect(ctx, review, details, comments)
	if err != nil || !json.Valid([]byte(summary.Text)) {
		code := SafeErrorCode(err)
		if err == nil {
			code = SafeErrorCode(ErrAIInvalidResponse)
		}
		_ = a.setStatus(ctx, id, "FAILED", code)
		return
	}
	menu := []byte(menuDataNotAvailableJSON)
	if details != nil && usableMenuData(details) {
		menuResult, menuErr := a.client.ProfileProspectMenu(ctx, review, details, nil)
		if menuErr != nil || !json.Valid([]byte(menuResult.Text)) {
			code := SafeErrorCode(menuErr)
			if menuErr == nil {
				code = SafeErrorCode(ErrAIInvalidResponse)
			}
			_ = a.save(ctx, id, []byte(summary.Text), nil, "FAILED", code)
			return
		}
		menu = []byte(menuResult.Text)
	}
	_ = a.save(ctx, id, []byte(summary.Text), menu, "SUCCESS", "")
}

func (a *InitialAnalyzer) hasSuccessfulSummary(ctx context.Context, id uuid.UUID) (bool, error) {
	var summary []byte
	err := a.pool.QueryRow(ctx, `SELECT summary_json FROM prospect_ai_analyses WHERE prospect_id=$1 AND status='SUCCESS'`, id).Scan(&summary)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return len(summary) > 0 && json.Valid(summary), nil
}

func usableMenuData(details *prospectmodel.PlaceDetails) bool {
	if details == nil {
		return false
	}
	// Google Places photos do not expose menu item names or photo categories.
	// A MENU-tagged photo resource name alone is not meaningful menu input for
	// text-only profiling, and this flow intentionally does not run OCR/vision.
	return false
}

func (a *InitialAnalyzer) setStatus(ctx context.Context, id uuid.UUID, status, code string) error {
	_, err := a.pool.Exec(ctx, `INSERT INTO prospect_ai_analyses (prospect_id,status,error_code,updated_at) VALUES ($1,$2,$3,now()) ON CONFLICT (prospect_id) DO UPDATE SET status=$2,error_code=$3,updated_at=now()`, id, status, code)
	return err
}

func (a *InitialAnalyzer) save(ctx context.Context, id uuid.UUID, summary, menu []byte, status, code string) error {
	_, err := a.pool.Exec(ctx, `INSERT INTO prospect_ai_analyses (prospect_id,summary_json,menu_json,status,error_code,updated_at) VALUES ($1,$2,$3,$4,$5,now()) ON CONFLICT (prospect_id) DO UPDATE SET summary_json=$2,menu_json=$3,status=$4,error_code=$5,updated_at=now()`, id, summary, menu, status, code)
	return err
}

func (a *InitialAnalyzer) Get(ctx context.Context, id uuid.UUID) (InitialAnalysis, error) {
	var r InitialAnalysis
	var summary, menu []byte
	var created, updated time.Time
	err := a.pool.QueryRow(ctx, `SELECT prospect_id,summary_json,menu_json,status,error_code,created_at,updated_at FROM prospect_ai_analyses WHERE prospect_id=$1`, id).Scan(&r.ProspectID, &summary, &menu, &r.Status, &r.ErrorCode, &created, &updated)
	if err != nil {
		return r, err
	}
	r.Summary, r.Menu, r.CreatedAt, r.UpdatedAt = summary, menu, &created, &updated
	return r, nil
}

func (a *InitialAnalyzer) Chat(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment, message, skill string) (string, error) {
	return a.ChatWithHistory(ctx, review, details, comments, nil, message, skill)
}

func (a *InitialAnalyzer) ChatWithHistory(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment, history []ChatMessage, message, skill string) (string, error) {
	analysis, err := a.Get(ctx, review.Prospect.ID)
	if err != nil {
		return "", err
	}
	saved := map[string]any{"summary": analysis.Summary, "menu": analysis.Menu, "status": analysis.Status}
	result, err := a.client.AskProspectAIWithSkill(ctx, review, details, comments, history, message, skill, saved)
	if err != nil {
		return "", err
	}
	return result.Text, nil
}
