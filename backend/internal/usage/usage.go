package usage

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const userKey contextKey = "usage.user_id"
const featureKey contextKey = "usage.feature"

func WithUser(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userKey, id)
}
func WithFeature(ctx context.Context, feature string) context.Context {
	return context.WithValue(ctx, featureKey, feature)
}
func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userKey).(uuid.UUID)
	return id, ok && id != uuid.Nil
}
func Feature(ctx context.Context) string {
	if v, ok := ctx.Value(featureKey).(string); ok {
		return v
	}
	return "UNKNOWN"
}

type Event struct {
	Provider, Feature, Operation, APIOrModel, SKUCategory, FieldMask string
	RequestID                                                        string
	UserID                                                           uuid.UUID
	InputTokens, OutputTokens, TotalTokens, RequestCount             int
	CachedTokens                                                     int
	CredentialAlias, Environment                                     string
	EstimatedCost                                                    float64
	HTTPStatus                                                       int
	Success                                                          bool
	ErrorCode                                                        string
}

type Recorder interface{ Record(context.Context, Event) }

// MemoryRecorder is intentionally small and only used by local tests. It makes
// request/event cardinality and user attribution observable without a database.
type MemoryRecorder struct {
	mu     sync.Mutex
	events []Event
}

func (r *MemoryRecorder) Record(_ context.Context, e Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}
func (r *MemoryRecorder) Events() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Event(nil), r.events...)
}

type PostgresRecorder struct{ pool *pgxpool.Pool }

func NewPostgresRecorder(pool *pgxpool.Pool) *PostgresRecorder { return &PostgresRecorder{pool: pool} }
func (r *PostgresRecorder) Record(ctx context.Context, e Event) {
	if r == nil || r.pool == nil || e.UserID == uuid.Nil {
		return
	}
	if e.RequestCount == 0 {
		e.RequestCount = 1
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO provider_usage_events (id,user_id,request_id,provider,feature,operation,api_or_model,sku_category,field_mask,input_tokens,output_tokens,total_tokens,cached_tokens,request_count,estimated_cost_usd,http_status,success,error_code,credential_alias,environment) VALUES ($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`, uuid.New(), e.UserID, e.RequestID, e.Provider, e.Feature, e.Operation, e.APIOrModel, nullable(e.SKUCategory), nullable(e.FieldMask), nullableInt(e.InputTokens), nullableInt(e.OutputTokens), nullableInt(e.TotalTokens), nullableInt(e.CachedTokens), e.RequestCount, nullableCost(e.EstimatedCost), nullableInt(e.HTTPStatus), e.Success, nullable(e.ErrorCode), nullable(e.CredentialAlias), nullable(e.Environment))
	if err != nil {
		slog.Warn("provider usage event write failed", "provider", e.Provider, "operation", e.Operation, "error", err)
	}
}
func nullable(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func nullableInt(v int) any {
	if v == 0 {
		return nil
	}
	return v
}
func nullableCost(v float64) any {
	if v == 0 {
		return nil
	}
	return v
}

type Pricing struct {
	GoogleSKUs                                    map[string]GoogleSKU
	OpenAIInputPerMillion, OpenAIOutputPerMillion map[string]float64
	OpenAIModels                                  map[string]OpenAIModelPricing
}

type OpenAIModelPricing struct {
	Model, Currency, Unit, Source, VerifiedAt                             string
	InputMicrosPerMillion, CachedMicrosPerMillion, OutputMicrosPerMillion int64
	Verified                                                              bool
}

type GoogleSKU struct {
	Operation, Name, ID, Tier, Currency, Source string
	PriceMicrosPer1000                          int64
	FreeMonthly                                 int
	Verified                                    bool
	VerifiedAt                                  string
}

type CostEstimate struct {
	Configured           bool
	GrossCost            float64
	FreeUsage            int
	EstimatedPayableCost float64
	BillingStatus        string
}

// GuardConfig is the policy foundation only. Enforcement remains opt-in so a
// local audit cannot unexpectedly block provider operations.
type GuardConfig struct {
	Enabled            bool
	WarningPercent     float64
	HighWarningPercent float64
	RestrictPercent    float64
	BlockPercent       float64
}

func DefaultGuardConfig() GuardConfig {
	return GuardConfig{Enabled: false, WarningPercent: 70, HighWarningPercent: 85, RestrictPercent: 95, BlockPercent: 100}
}

func DefaultPricing() Pricing {
	return Pricing{GoogleSKUs: map[string]GoogleSKU{
		"NEARBY_SEARCH": {Operation: "NEARBY_SEARCH", Name: "Places API Nearby Search Pro", ID: "99F9-A108-83A6", Tier: "Pro", Currency: "USD", PriceMicrosPer1000: 32000000, FreeMonthly: 5000, Verified: true, VerifiedAt: "2026-08-20", Source: "Google Maps Platform core services pricing list; Places API data-field/SKU documentation"},
		"TEXT_SEARCH":   {Operation: "TEXT_SEARCH", Name: "Places API Text Search Pro", ID: "4FDA-34B1-A910", Tier: "Pro", Currency: "USD", PriceMicrosPer1000: 32000000, FreeMonthly: 5000, Verified: true, VerifiedAt: "2026-08-20", Source: "Google Maps Platform core services pricing list; Places API data-field/SKU documentation"},
		"CORE_DETAIL":   {Operation: "CORE_DETAIL", Name: "Places API Place Details Pro", ID: "4ED6-464A-2AFC", Tier: "Pro", Currency: "USD", PriceMicrosPer1000: 17000000, FreeMonthly: 5000, Verified: true, VerifiedAt: "2026-08-20", Source: "Google Maps Platform core services pricing list; Places API data-field/SKU documentation"},
		"BUSINESS_INFO": {Operation: "BUSINESS_INFO", Name: "Places API Place Details Enterprise", ID: "2D9A-3DE0-3766", Tier: "Enterprise", Currency: "USD", PriceMicrosPer1000: 20000000, FreeMonthly: 1000, Verified: true, VerifiedAt: "2026-08-20", Source: "Google Maps Platform core services pricing list; Places API data-field/SKU documentation"},
		"PLACE_PHOTO":   {Operation: "PLACE_PHOTO", Name: "Places API Place Details Photos", ID: "DCD1-FE97-8C71", Tier: "Photos", Currency: "USD", PriceMicrosPer1000: 7000000, FreeMonthly: 1000, Verified: true, VerifiedAt: "2026-08-20", Source: "Google Maps Platform core services pricing list; Places API data-field/SKU documentation"},
	}, OpenAIInputPerMillion: map[string]float64{}, OpenAIOutputPerMillion: map[string]float64{}, OpenAIModels: map[string]OpenAIModelPricing{
		"gpt-5.6-luna": {Model: "gpt-5.6-luna", InputMicrosPerMillion: 200000, CachedMicrosPerMillion: 20000, OutputMicrosPerMillion: 1200000, Currency: "USD", Unit: "per 1,000,000 tokens", Verified: true, VerifiedAt: "2026-08-20", Source: "Official OpenAI API pricing/model documentation"},
	}}
}

func (p Pricing) OpenAICost(model string, input, cached, output int) (int64, bool) {
	price, ok := p.OpenAIModels[model]
	if !ok || !price.Verified || input < 0 || cached < 0 || output < 0 {
		return 0, false
	}
	if cached > input {
		cached = input
	}
	uncached := input - cached
	return (int64(uncached)*price.InputMicrosPerMillion + int64(cached)*price.CachedMicrosPerMillion + int64(output)*price.OutputMicrosPerMillion) / 1000000, true
}
func (p Pricing) Google(operation string) float64 {
	sku, ok := p.GoogleSKUs[operation]
	if !ok {
		return 0
	}
	return float64(sku.PriceMicrosPer1000) / 1e9
}
func (p Pricing) FreeTier(provider, operation string) (int, bool) {
	if provider != "GOOGLE_MAPS" {
		return 0, false
	}
	sku, ok := p.GoogleSKUs[operation]
	return sku.FreeMonthly, ok && sku.Verified
}

// GoogleCost applies the shared project/SKU free tier once. It is deliberately
// unconfigured until verified pricing is supplied; no provider price is
// guessed from memory or inferred from an outbound attempt.
func (p Pricing) GoogleCost(operation string, billableRequests, projectMonthlyRequests int) CostEstimate {
	sku, configured := p.GoogleSKUs[operation]
	if !configured || !sku.Verified || sku.PriceMicrosPer1000 <= 0 {
		return CostEstimate{BillingStatus: "UNCONFIGURED"}
	}
	if billableRequests < 0 {
		billableRequests = 0
	}
	if projectMonthlyRequests < 0 {
		projectMonthlyRequests = 0
	}
	projectOverage := projectMonthlyRequests - sku.FreeMonthly
	if projectOverage < 0 {
		projectOverage = 0
	}
	payable := projectOverage
	if billableRequests < payable {
		payable = billableRequests
	}
	return CostEstimate{Configured: true, GrossCost: float64(int64(billableRequests)*sku.PriceMicrosPer1000) / 1e6 / 1000, FreeUsage: sku.FreeMonthly, EstimatedPayableCost: float64(int64(payable)*sku.PriceMicrosPer1000) / 1e6 / 1000, BillingStatus: "ESTIMATE"}
}
func (p Pricing) OpenAI(model string, in, out int) float64 {
	return float64(in)/1e6*p.OpenAIInputPerMillion[model] + float64(out)/1e6*p.OpenAIOutputPerMillion[model]
}
