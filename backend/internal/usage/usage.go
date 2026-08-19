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
	_, err := r.pool.Exec(ctx, `INSERT INTO provider_usage_events (id,user_id,request_id,provider,feature,operation,api_or_model,sku_category,field_mask,input_tokens,output_tokens,total_tokens,request_count,estimated_cost_usd,http_status,success,error_code) VALUES ($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`, uuid.New(), e.UserID, e.RequestID, e.Provider, e.Feature, e.Operation, e.APIOrModel, nullable(e.SKUCategory), nullable(e.FieldMask), nullableInt(e.InputTokens), nullableInt(e.OutputTokens), nullableInt(e.TotalTokens), e.RequestCount, nullableCost(e.EstimatedCost), nullableInt(e.HTTPStatus), e.Success, nullable(e.ErrorCode))
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
	GooglePerRequest                              map[string]float64
	OpenAIInputPerMillion, OpenAIOutputPerMillion map[string]float64
	FreeTierPerMonth                              map[string]int
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
	return Pricing{GooglePerRequest: map[string]float64{}, OpenAIInputPerMillion: map[string]float64{}, OpenAIOutputPerMillion: map[string]float64{}, FreeTierPerMonth: map[string]int{}}
}
func (p Pricing) Google(operation string) float64 { return p.GooglePerRequest[operation] }
func (p Pricing) FreeTier(provider, operation string) (int, bool) {
	value, ok := p.FreeTierPerMonth[provider+":"+operation]
	return value, ok && value > 0
}
func (p Pricing) OpenAI(model string, in, out int) float64 {
	return float64(in)/1e6*p.OpenAIInputPerMillion[model] + float64(out)/1e6*p.OpenAIOutputPerMillion[model]
}
