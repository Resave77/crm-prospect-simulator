package usage

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityRecorder interface {
	RecordActivity(context.Context, Activity)
}

type Activity struct {
	ID, UserID, RequestID, Method, Endpoint, Domain         string
	RequestBody, QueryParams, ResponseBody, AdditionalTrace []byte
	ResponseStatus                                          int
	DurationMS                                              int64
	IP, UserAgent                                           string
}

type PostgresActivityRecorder struct{ pool *pgxpool.Pool }

func NewPostgresActivityRecorder(pool *pgxpool.Pool) *PostgresActivityRecorder {
	return &PostgresActivityRecorder{pool: pool}
}
func (r *PostgresActivityRecorder) RecordActivity(ctx context.Context, a Activity) {
	if r == nil || r.pool == nil {
		return
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO activity_logs (id,user_id,request_id,method,endpoint,request_body,query_params,response_body,response_status,domain,additional_trace,duration_ms,ip,user_agent) VALUES ($1,NULLIF($2,'')::uuid,$3,$4,$5,$6::jsonb,$7::jsonb,$8::jsonb,$9,$10,$11::jsonb,$12,$13,$14)`, uuid.New(), a.UserID, a.RequestID, a.Method, a.Endpoint, a.RequestBody, a.QueryParams, a.ResponseBody, a.ResponseStatus, a.Domain, a.AdditionalTrace, a.DurationMS, a.IP, a.UserAgent)
	if err != nil {
		slog.Warn("activity log write failed", "error", err)
	}
}

func ActivityMiddleware(recorder ActivityRecorder) fiber.Handler {
	return func(c *fiber.Ctx) error {
		started := time.Now()
		traceID := c.Get(fiber.HeaderXRequestID)
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.SetUserContext(WithTrace(c.UserContext(), traceID))
		err := c.Next()
		ctx := c.UserContext()
		uid, authenticated := UserID(ctx)
		userID := ""
		if authenticated {
			userID = uid.String()
		}
		query, _ := json.Marshal(c.Context().QueryArgs().String())
		trace, _ := json.Marshal(GetTrace(ctx))
		requestBody := SafeJSON(c.Request().Body())
		responseBody := SafeJSON(c.Response().Body())
		status := c.Response().StatusCode()
		if err != nil && status < 400 {
			status = 500
		}
		if ShouldPersistActivity(c.Method(), c.OriginalURL(), status, GetTrace(ctx), authenticated) {
			recorder.RecordActivity(ctx, Activity{UserID: userID, RequestID: RequestID(ctx), Method: c.Method(), Endpoint: c.OriginalURL(), RequestBody: requestBody, QueryParams: query, ResponseBody: responseBody, ResponseStatus: status, Domain: Domain(c.Path()), AdditionalTrace: trace, DurationMS: time.Since(started).Milliseconds(), IP: c.IP(), UserAgent: c.Get(fiber.HeaderUserAgent)})
		}
		return err
	}
}
