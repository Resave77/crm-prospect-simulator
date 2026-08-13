package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"time"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const menuDataNotAvailable = "MENU_DATA_NOT_AVAILABLE"

const (
	maxMenuCategories      = 5
	maxMenuItems           = 20
	maxMenuEvidencePerItem = 2
)

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

type MenuImageInput struct {
	Name        string
	Bytes       []byte
	ContentType string
}

type MenuFindingSource struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	BranchMatch string `json:"branchMatch"`
}
type MenuFindingPrice struct {
	Source    string  `json:"source"`
	SourceURL string  `json:"sourceUrl"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
}
type MenuFindingItem struct {
	Name         string             `json:"name"`
	Description  *string            `json:"description"`
	ImageURL     *string            `json:"imageUrl"`
	Prices       []MenuFindingPrice `json:"prices"`
	Availability string             `json:"availability"`
	BranchMatch  string             `json:"branchMatch"`
	Confidence   float64            `json:"confidence"`
}
type MenuFindingCategory struct {
	Name  string            `json:"name"`
	Items []MenuFindingItem `json:"items"`
}
type MenuFindingResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Place   struct {
		Name          string `json:"name"`
		Branch        string `json:"branch"`
		Address       string `json:"address"`
		GooglePlaceID string `json:"googlePlaceId"`
		GoogleMapsURL string `json:"googleMapsUrl"`
	} `json:"place"`
	Sources    []MenuFindingSource   `json:"sources"`
	Categories []MenuFindingCategory `json:"categories"`
	Summary    struct {
		SourceCount     int `json:"sourceCount"`
		TotalCategories int `json:"totalCategories"`
		TotalItems      int `json:"totalItems"`
	} `json:"summary"`
}
type menuDocument struct {
	Discovery *MenuFindingResult `json:"discovery,omitempty"`
	Profiling json.RawMessage    `json:"profiling,omitempty"`
}

func NewInitialAnalyzer(pool *pgxpool.Pool, client *ProspectAI) *InitialAnalyzer {
	return &InitialAnalyzer{pool: pool, client: client}
}

func (a *InitialAnalyzer) loadMenuDocument(ctx context.Context, id uuid.UUID) (menuDocument, error) {
	var raw []byte
	err := a.pool.QueryRow(ctx, `SELECT menu_json FROM prospect_ai_analyses WHERE prospect_id=$1`, id).Scan(&raw)
	if err != nil && err != pgx.ErrNoRows {
		return menuDocument{}, err
	}
	if len(raw) == 0 {
		return menuDocument{}, nil
	}
	var doc menuDocument
	if json.Unmarshal(raw, &doc) == nil && (doc.Discovery != nil || len(doc.Profiling) > 0) {
		return doc, nil
	}
	var legacy struct {
		Finding *MenuFindingResult `json:"finding"`
		Profile json.RawMessage    `json:"profile"`
	}
	if json.Unmarshal(raw, &legacy) == nil && (legacy.Finding != nil || len(legacy.Profile) > 0) {
		return menuDocument{Discovery: legacy.Finding, Profiling: legacy.Profile}, nil
	}
	if json.Valid(raw) {
		doc.Profiling = append(json.RawMessage(nil), raw...)
	}
	return doc, nil
}

func (a *InitialAnalyzer) saveMenuDocument(ctx context.Context, id uuid.UUID, doc menuDocument) (json.RawMessage, error) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	_, err = a.pool.Exec(ctx, `INSERT INTO prospect_ai_analyses (prospect_id,menu_json,status,error_code,updated_at) VALUES ($1,$2,'SUCCESS','',now()) ON CONFLICT (prospect_id) DO UPDATE SET menu_json=$2,updated_at=now()`, id, raw)
	return raw, err
}

func validatedMenuFinding(raw string, toolSources []WebSearchSource, review prospectmodel.Review) (*MenuFindingResult, error) {
	var finding MenuFindingResult
	if err := json.Unmarshal([]byte(raw), &finding); err != nil {
		return nil, ErrAIInvalidResponse
	}
	allowedURLs := map[string]WebSearchSource{}
	for _, source := range toolSources {
		u := strings.TrimSpace(source.URL)
		parsed, err := url.Parse(u)
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
			allowedURLs[u] = source
		}
	}
	finding.Place.Name = strings.TrimSpace(finding.Place.Name)
	if finding.Place.Name == "" {
		finding.Place.Name = review.Prospect.PlaceName
	}
	if finding.Place.Address == "" {
		finding.Place.Address = review.Prospect.FormattedAddress
	}
	if finding.Place.GooglePlaceID == "" {
		finding.Place.GooglePlaceID = review.Prospect.GooglePlaceID
	}
	if finding.Place.GoogleMapsURL == "" {
		finding.Place.GoogleMapsURL = review.Prospect.GoogleMapsURL
	}
	validMatch := func(value string) bool {
		switch value {
		case "exact_branch", "likely_same_branch", "brand_only", "uncertain":
			return true
		default:
			return false
		}
	}
	sources := make([]MenuFindingSource, 0, len(finding.Sources))
	seen := map[string]bool{}
	for _, source := range finding.Sources {
		source.URL = strings.TrimSpace(source.URL)
		if _, ok := allowedURLs[source.URL]; !ok || seen[source.URL] || !validMatch(source.BranchMatch) {
			continue
		}
		seen[source.URL] = true
		if strings.TrimSpace(source.Name) == "" {
			source.Name = allowedURLs[source.URL].Title
		}
		sources = append(sources, source)
	}
	finding.Sources = sources
	itemCount := 0
	categories := make([]MenuFindingCategory, 0, len(finding.Categories))
	for _, category := range finding.Categories {
		if len(categories) >= maxMenuCategories || itemCount >= maxMenuItems {
			break
		}
		items := make([]MenuFindingItem, 0, len(category.Items))
		for _, item := range category.Items {
			if itemCount >= maxMenuItems {
				break
			}
			item.Name = strings.TrimSpace(item.Name)
			if item.Name == "" || !validMatch(item.BranchMatch) || item.BranchMatch == "uncertain" {
				continue
			}
			if item.Confidence < 0 {
				item.Confidence = 0
			}
			if item.Confidence > 1 {
				item.Confidence = 1
			}
			if item.Availability != "available" {
				item.Availability = "unknown"
			}
			if item.ImageURL != nil {
				if _, ok := allowedURLs[strings.TrimSpace(*item.ImageURL)]; !ok {
					item.ImageURL = nil
				}
			}
			prices := make([]MenuFindingPrice, 0, len(item.Prices))
			for _, price := range item.Prices {
				if len(prices) >= maxMenuEvidencePerItem {
					break
				}
				if _, ok := allowedURLs[strings.TrimSpace(price.SourceURL)]; !ok || price.Price < 0 {
					continue
				}
				prices = append(prices, price)
			}
			item.Prices = prices
			items = append(items, item)
			itemCount++
		}
		if len(items) > 0 {
			category.Items = items
			categories = append(categories, category)
		}
	}
	finding.Categories = categories
	finding.Summary.SourceCount, finding.Summary.TotalCategories, finding.Summary.TotalItems = len(sources), len(categories), itemCount
	if finding.Status != "FOUND" || len(sources) == 0 || itemCount == 0 {
		finding.Status = "NOT_FOUND"
		finding.Categories = []MenuFindingCategory{}
		finding.Summary.TotalCategories, finding.Summary.TotalItems = 0, 0
		if strings.TrimSpace(finding.Message) == "" {
			finding.Message = "Menu belum ditemukan dari sumber yang dapat diverifikasi."
		}
	} else {
		finding.Status = "FOUND"
	}
	return &finding, nil
}

func (a *InitialAnalyzer) FindMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails) (json.RawMessage, error) {
	started := time.Now()
	doc, err := a.loadMenuDocument(ctx, review.Prospect.ID)
	if err != nil {
		return nil, err
	}
	search, err := a.client.FindProspectMenu(ctx, review, details)
	if err != nil {
		slog.Warn("Find Menu failed", "operation", "find_menu", "prospect_id", review.Prospect.ID, "timeout_ms", a.client.client.findMenuTimeout.Milliseconds(), "duration_ms", time.Since(started).Milliseconds(), "openai_request_id", search.UpstreamRequestID, "safe_error_code", SafeErrorCode(err))
		return nil, err
	}
	finding, err := validatedMenuFinding(search.Text, search.Sources, review)
	if err != nil {
		failureCategory := "invalid_json"
		if search.IncompleteReason == "max_output_tokens" {
			failureCategory = "output_token_limit"
		} else if search.ResponseStatus == "incomplete" {
			failureCategory = "responses_incomplete"
		}
		slog.Warn("Find Menu validation failed", "operation", "find_menu", "prospect_id", review.Prospect.ID, "timeout_ms", a.client.client.findMenuTimeout.Milliseconds(), "duration_ms", time.Since(started).Milliseconds(), "openai_request_id", search.UpstreamRequestID, "source_count", len(search.Sources), "output_tokens", search.OutputTokens, "output_chars", len(search.Text), "response_status", search.ResponseStatus, "incomplete_reason", search.IncompleteReason, "failure_category", failureCategory, "validation_status", "invalid", "safe_error_code", SafeErrorCode(err))
		recovery, recoveryErr := a.client.RecoverProspectMenuFormatting(ctx, review, search)
		if recoveryErr != nil {
			return nil, recoveryErr
		}
		finding, err = validatedMenuFinding(recovery.Text, search.Sources, review)
		if err != nil {
			return nil, err
		}
		slog.Info("Find Menu formatting recovery completed", "operation", "find_menu_recovery", "prospect_id", review.Prospect.ID, "web_search_calls", 0, "openai_request_id", recovery.UpstreamRequestID, "output_tokens", recovery.OutputTokens, "output_chars", len(recovery.Text), "category_count", len(finding.Categories), "item_count", finding.Summary.TotalItems, "validation_status", finding.Status)
	}
	slog.Info("Find Menu completed", "operation", "find_menu", "prospect_id", review.Prospect.ID, "timeout_ms", a.client.client.findMenuTimeout.Milliseconds(), "duration_ms", time.Since(started).Milliseconds(), "openai_request_id", search.UpstreamRequestID, "source_count", len(search.Sources), "accepted_source_count", len(finding.Sources), "output_tokens", search.OutputTokens, "output_chars", len(search.Text), "category_count", len(finding.Categories), "item_count", finding.Summary.TotalItems, "validation_status", finding.Status)
	doc.Discovery = finding
	if _, err := a.saveMenuDocument(ctx, review.Prospect.ID, doc); err != nil {
		return nil, err
	}
	return json.Marshal(doc.Discovery)
}

func (a *InitialAnalyzer) ProfileStructuredMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, force bool) (json.RawMessage, bool, error) {
	started := time.Now()
	doc, err := a.loadMenuDocument(ctx, review.Prospect.ID)
	if err != nil {
		return nil, false, err
	}
	if doc.Discovery == nil || doc.Discovery.Status != "FOUND" || len(doc.Discovery.Categories) == 0 {
		return nil, false, nil
	}
	if shouldUseCachedMenuProfiling(doc.Profiling, force) {
		return doc.Profiling, true, nil
	}
	finding, _ := json.Marshal(doc.Discovery)
	result, err := a.client.ProfileProspectStructuredMenu(ctx, review, details, finding)
	if err != nil || !validIndonesianMenuProfiling(json.RawMessage(result.Text)) {
		if err == nil {
			err = ErrAIInvalidResponse
		}
		slog.Warn("Menu profiling failed", "operation", "menu_profiling", "prospect_id", review.Prospect.ID, "timeout_ms", a.client.client.menuProfileTimeout.Milliseconds(), "duration_ms", time.Since(started).Milliseconds(), "openai_request_id", result.UpstreamRequestID, "output_tokens", result.OutputTokens, "output_chars", len(result.Text), "validation_status", "invalid", "safe_error_code", SafeErrorCode(err))
		return nil, true, err
	}
	doc.Profiling = json.RawMessage(result.Text)
	if _, err := a.saveMenuDocument(ctx, review.Prospect.ID, doc); err != nil {
		return nil, true, err
	}
	slog.Info("Menu profiling completed", "operation", "menu_profiling", "prospect_id", review.Prospect.ID, "timeout_ms", a.client.client.menuProfileTimeout.Milliseconds(), "duration_ms", time.Since(started).Milliseconds(), "openai_request_id", result.UpstreamRequestID, "output_tokens", result.OutputTokens, "output_chars", len(result.Text), "validation_status", "valid")
	return doc.Profiling, true, nil
}

func shouldUseCachedMenuProfiling(raw json.RawMessage, force bool) bool {
	return !force && validMenuProfiling(raw)
}

func validMenuProfiling(raw json.RawMessage) bool {
	var result struct {
		MenuOpportunity   string `json:"menuOpportunity"`
		YoghurtFit        string `json:"yoghurtFit"`
		TopOpportunity    string `json:"topOpportunity"`
		Why               string `json:"why"`
		RecommendedAction string `json:"recommendedAction"`
		Confidence        string `json:"confidence"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &result) != nil {
		return false
	}
	validOpportunity := func(value string) bool {
		switch value {
		case "HIGH", "MEDIUM", "LOW", "UNKNOWN":
			return true
		}
		return false
	}
	validConfidence := result.Confidence == "HIGH" || result.Confidence == "MEDIUM" || result.Confidence == "LOW" || result.Confidence == "UNKNOWN"
	return validOpportunity(result.MenuOpportunity) && validOpportunity(result.YoghurtFit) && validConfidence && strings.TrimSpace(result.TopOpportunity) != "" && strings.TrimSpace(result.Why) != "" && strings.TrimSpace(result.RecommendedAction) != ""
}

func validIndonesianMenuProfiling(raw json.RawMessage) bool {
	if !validMenuProfiling(raw) {
		return false
	}
	var result struct {
		TopOpportunity    string `json:"topOpportunity"`
		Why               string `json:"why"`
		RecommendedAction string `json:"recommendedAction"`
	}
	if json.Unmarshal(raw, &result) != nil {
		return false
	}
	return !looksLikeEnglishSentence(result.TopOpportunity) &&
		!looksLikeEnglishSentence(result.Why) &&
		!looksLikeEnglishSentence(result.RecommendedAction)
}

func looksLikeEnglishSentence(value string) bool {
	words := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	englishSignals := map[string]struct{}{
		"the": {}, "this": {}, "that": {}, "with": {}, "without": {}, "and": {}, "or": {}, "for": {}, "from": {}, "into": {},
		"is": {}, "are": {}, "was": {}, "were": {}, "has": {}, "have": {}, "offers": {}, "evaluate": {}, "discuss": {}, "consider": {},
		"additional": {}, "option": {}, "opportunity": {}, "recommended": {}, "action": {}, "branch": {}, "customer": {}, "needs": {},
	}
	signals := 0
	for _, word := range words {
		if _, found := englishSignals[word]; found {
			signals++
		}
	}
	return signals >= 2
}

// ProfileMenu profiles only explicitly supplied MENU-tagged images and updates menu_json only.
func (a *InitialAnalyzer) ProfileMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, images []MenuImageInput) (json.RawMessage, error) {
	if len(images) == 0 || len(images) > 3 {
		return nil, ErrAINotConfigured
	}
	names := make([]string, 0, len(images))
	inputs := make([]ImageInput, 0, len(images))
	for _, image := range images {
		names = append(names, image.Name)
		inputs = append(inputs, ImageInput{Bytes: image.Bytes, ContentType: image.ContentType})
	}
	sort.Strings(names)
	doc, err := a.loadMenuDocument(ctx, review.Prospect.ID)
	if err != nil {
		return nil, err
	}
	var saved struct {
		SourcePhotoNames []string `json:"sourcePhotoNames"`
	}
	if json.Unmarshal(doc.Profiling, &saved) == nil {
		sort.Strings(saved.SourcePhotoNames)
		if len(saved.SourcePhotoNames) == len(names) {
			same := true
			for i := range names {
				if names[i] != saved.SourcePhotoNames[i] {
					same = false
					break
				}
			}
			if same && len(names) > 0 {
				return doc.Profiling, nil
			}
		}
	}
	payload := a.client.contextJSON(review, details, nil, map[string]any{"sourcePhotoNames": names})
	result, err := a.client.ProfileProspectMenuVision(ctx, payload, inputs, names)
	if err != nil || !json.Valid([]byte(result.Text)) {
		if err == nil {
			err = ErrAIInvalidResponse
		}
		return nil, err
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
		return nil, ErrAIInvalidResponse
	}
	parsed["sourcePhotoNames"] = names
	menu, err := json.Marshal(parsed)
	if err != nil {
		return nil, ErrAIInvalidResponse
	}
	doc.Profiling = menu
	if _, err = a.saveMenuDocument(ctx, review.Prospect.ID, doc); err != nil {
		return nil, err
	}
	return menu, nil
}

func (a *InitialAnalyzer) Analyze(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment) {
	_, _ = a.GenerateSummary(ctx, review, details, comments)
}

// GenerateSummary lazily creates one persisted summary. A transaction-scoped
// advisory lock makes concurrent first-open requests reuse the same result.
func (a *InitialAnalyzer) GenerateSummary(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment) (json.RawMessage, error) {
	if a == nil || a.pool == nil || a.client == nil || !a.client.client.Configured() {
		return nil, ErrAINotConfigured
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	id := review.Prospect.ID
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, id.String()); err != nil {
		return nil, err
	}
	var existing []byte
	err = tx.QueryRow(ctx, `SELECT summary_json FROM prospect_ai_analyses WHERE prospect_id=$1`, id).Scan(&existing)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if len(existing) > 0 && validSummaryJSON(string(existing)) {
		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}
		return json.RawMessage(existing), nil
	}
	_, err = tx.Exec(ctx, `INSERT INTO prospect_ai_analyses (prospect_id,status,error_code,updated_at) VALUES ($1,'PENDING','',now()) ON CONFLICT (prospect_id) DO UPDATE SET status='PENDING',error_code='',updated_at=now()`, id)
	if err != nil {
		return nil, err
	}
	var attemptDuration time.Duration
	summary, attempts, aiErr := generateSummaryWithRecovery(ctx, func(ctx context.Context, recovery bool) (TextResult, error) {
		started := time.Now()
		defer func() { attemptDuration = time.Since(started) }()
		if recovery {
			return a.client.RecoverProspectSummary(ctx, review, details, comments)
		}
		return a.client.SummarizeProspect(ctx, review, details, comments)
	}, func(attempt int, stage string, result TextResult, attemptErr error) {
		slog.Warn("AI summary generation attempt failed",
			"operation", "summary_generation", "prospect_id", id, "attempt", attempt,
			"failure_stage", stage, "safe_error_code", SafeErrorCode(attemptErr),
			"duration_ms", attemptDuration.Milliseconds(),
			"openai_request_id", result.UpstreamRequestID,
		)
	})
	if aiErr != nil {
		_, _ = tx.Exec(ctx, `UPDATE prospect_ai_analyses SET status='FAILED',error_code=$2,updated_at=now() WHERE prospect_id=$1`, id, SafeErrorCode(aiErr))
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return nil, commitErr
		}
		return nil, aiErr
	}
	slog.Info("AI summary generation succeeded", "operation", "summary_generation", "prospect_id", id, "attempt_used", attempts, "duration_ms", attemptDuration.Milliseconds(), "openai_request_id", summary.UpstreamRequestID)
	raw := []byte(summary.Text)
	_, err = tx.Exec(ctx, `INSERT INTO prospect_ai_analyses (prospect_id,summary_json,status,error_code,updated_at) VALUES ($1,$2,'SUCCESS','',now()) ON CONFLICT (prospect_id) DO UPDATE SET summary_json=$2,status='SUCCESS',error_code='',updated_at=now()`, id, raw)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

type summaryAttempt func(context.Context, bool) (TextResult, error)

func generateSummaryWithRecovery(ctx context.Context, generate summaryAttempt, onFailure func(int, string, TextResult, error)) (TextResult, int, error) {
	for attempt := 1; attempt <= 2; attempt++ {
		result, err := generate(ctx, attempt == 2)
		stage := "openai_request_or_response"
		if err == nil && !validSummaryJSON(result.Text) {
			err, stage = ErrAIInvalidResponse, "summary_json_validation"
		}
		if err == nil {
			return result, attempt, nil
		}
		if onFailure != nil {
			onFailure(attempt, stage, result, err)
		}
		if attempt == 2 || !retryableSummaryError(err) {
			return TextResult{}, attempt, err
		}
	}
	return TextResult{}, 2, ErrAIInvalidResponse
}

func retryableSummaryError(err error) bool {
	return errors.Is(err, ErrAIInvalidResponse) || errors.Is(err, ErrAIUnavailable) || errors.Is(err, ErrAITimeout) || errors.Is(err, ErrAIRateLimited)
}

func validSummaryJSON(value string) bool {
	var root map[string]json.RawMessage
	if json.Unmarshal([]byte(strings.TrimSpace(value)), &root) != nil {
		return false
	}
	required := []string{"summary", "potential", "customerProfile", "yoghurtOpportunity", "buyingSignals", "qualification", "dealRisks", "missingInformation", "nextBestAction", "recommendedQuestions"}
	for _, key := range required {
		if _, ok := root[key]; !ok {
			return false
		}
	}
	var qualification map[string]json.RawMessage
	if json.Unmarshal(root["qualification"], &qualification) != nil {
		return false
	}
	for _, key := range []string{"need", "interest", "authority", "timing"} {
		if _, ok := qualification[key]; !ok {
			return false
		}
	}
	var next map[string]json.RawMessage
	if json.Unmarshal(root["nextBestAction"], &next) != nil {
		return false
	}
	for _, key := range []string{"action", "reason"} {
		if _, ok := next[key]; !ok {
			return false
		}
	}
	var typed struct {
		Summary              string                                             `json:"summary"`
		Potential            string                                             `json:"potential"`
		CustomerProfile      string                                             `json:"customerProfile"`
		YoghurtOpportunity   string                                             `json:"yoghurtOpportunity"`
		BuyingSignals        []string                                           `json:"buyingSignals"`
		Qualification        struct{ Need, Interest, Authority, Timing string } `json:"qualification"`
		DealRisks            []string                                           `json:"dealRisks"`
		MissingInformation   []string                                           `json:"missingInformation"`
		NextBestAction       struct{ Action, Reason string }                    `json:"nextBestAction"`
		RecommendedQuestions []string                                           `json:"recommendedQuestions"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(value)), &typed) != nil {
		return false
	}
	switch typed.Potential {
	case "HIGH", "MEDIUM", "LOW", "UNKNOWN":
	default:
		return false
	}
	return typed.BuyingSignals != nil && typed.DealRisks != nil && typed.MissingInformation != nil && typed.RecommendedQuestions != nil
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
	return len(summary) > 0 && validSummaryJSON(string(summary)), nil
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
