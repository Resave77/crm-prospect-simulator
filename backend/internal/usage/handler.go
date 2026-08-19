package usage

import (
	"crm-prospect-simulator/backend/internal/shared/response"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func (h *Handler) Summary(c *fiber.Ctx) error {
	rows, err := h.pool.Query(c.UserContext(), `SELECT provider, operation, COUNT(*), COALESCE(SUM(request_count),0), COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COALESCE(SUM(total_tokens),0), COALESCE(SUM(estimated_cost_usd),0), MAX(created_at) FROM provider_usage_events WHERE ($1='' OR user_id::text=$1) AND ($2='' OR provider=$2) AND ($3='' OR feature=$3) AND ($4::timestamptz IS NULL OR created_at >= $4) AND ($5::timestamptz IS NULL OR created_at < $5) GROUP BY provider, operation ORDER BY provider, operation`, c.Query("userId"), c.Query("provider"), c.Query("feature"), nullableDate(c.Query("dateFrom")), nullableDate(c.Query("dateTo")))
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage summary is unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var provider, operation string
		var events, requests, in, outTokens, total int
		var cost float64
		var last time.Time
		if err := rows.Scan(&provider, &operation, &events, &requests, &in, &outTokens, &total, &cost, &last); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage summary is unavailable.")
		}
		freeTier, verified := DefaultPricing().FreeTier(provider, operation)
		var usagePercent any
		if verified { usagePercent = float64(requests) / float64(freeTier) * 100 }
		out = append(out, fiber.Map{"provider": provider, "operation": operation, "events": events, "requests": requests, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "estimatedCost": cost, "freeTier": freeTier, "freeTierVerified": verified, "usagePercent": usagePercent, "lastUsed": last})
	}
	return response.Data(c, 200, out)
}
func (h *Handler) History(c *fiber.Ctx) error {
	rows, err := h.pool.Query(c.UserContext(), `SELECT created_at,user_id,provider,feature,operation,api_or_model,COALESCE(field_mask,''),COALESCE(request_count,0),COALESCE(input_tokens,0),COALESCE(output_tokens,0),COALESCE(total_tokens,0),COALESCE(estimated_cost_usd,0),COALESCE(http_status,0),success FROM provider_usage_events WHERE ($1='' OR user_id::text=$1) AND ($2='' OR provider=$2) AND ($3='' OR feature=$3) AND ($4='' OR operation=$4) AND ($5::timestamptz IS NULL OR created_at >= $5) AND ($6::timestamptz IS NULL OR created_at < $6) ORDER BY created_at DESC LIMIT 200`, c.Query("userId"), c.Query("provider"), c.Query("feature"), c.Query("operation"), nullableDate(c.Query("dateFrom")), nullableDate(c.Query("dateTo")))
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage history is unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var t time.Time
		var uid, provider, feature, operation, api, fieldMask string
		var calls, in, outTokens, total, status int
		var cost float64
		var ok bool
		if err := rows.Scan(&t, &uid, &provider, &feature, &operation, &api, &fieldMask, &calls, &in, &outTokens, &total, &cost, &status, &ok); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage history is unavailable.")
		}
		out = append(out, fiber.Map{"createdAt": t, "userId": uid, "provider": provider, "feature": feature, "operation": operation, "apiOrModel": api, "fieldMask": fieldMask, "requestCount": calls, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "estimatedCost": cost, "httpStatus": status, "success": ok})
	}
	return response.Data(c, 200, out)
}

func (h *Handler) ActivityHistory(c *fiber.Ctx) error {
	rows, err := h.pool.Query(c.UserContext(), `SELECT created_at,request_id,method,endpoint,domain,response_status,additional_trace FROM activity_logs WHERE ($1='' OR user_id::text=$1) ORDER BY created_at DESC LIMIT 200`, c.Query("userId"))
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Activity history is unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var created time.Time
		var requestID, method, endpoint, domain string
		var status int
		var trace []byte
		if err := rows.Scan(&created, &requestID, &method, &endpoint, &domain, &status, &trace); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Activity history is unavailable.")
		}
		var details any
		_ = json.Unmarshal(trace, &details)
		out = append(out, fiber.Map{"createdAt": created, "requestId": requestID, "method": method, "endpoint": endpoint, "domain": domain, "responseStatus": status, "trace": details})
	}
	return response.Data(c, 200, out)
}

func (h *Handler) Daily(c *fiber.Ctx) error {
	date := c.Query("date")
	if date == "" {
		return response.Error(c, 400, "DATE_REQUIRED", "A calendar date is required.")
	}
	rows, err := h.pool.Query(c.UserContext(), `SELECT created_at,user_id,provider,operation,COUNT(*),COALESCE(SUM(request_count),0),COALESCE(SUM(input_tokens),0),COALESCE(SUM(output_tokens),0),COALESCE(SUM(total_tokens),0),COALESCE(SUM(estimated_cost_usd),0) FROM provider_usage_events WHERE ($1='' OR user_id::text=$1) AND created_at >= $2::date AND created_at < ($2::date + INTERVAL '1 day') GROUP BY created_at::date,user_id,provider,operation ORDER BY user_id,provider,operation`, c.Query("userId"), date)
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Daily usage is unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var uid, provider, operation string
		var day time.Time
		var events, requests, in, outTokens, total int
		var cost float64
		if err := rows.Scan(&day, &uid, &provider, &operation, &events, &requests, &in, &outTokens, &total, &cost); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Daily usage is unavailable.")
		}
		out = append(out, fiber.Map{"date": day, "userId": uid, "provider": provider, "operation": operation, "events": events, "requests": requests, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "estimatedCost": cost})
	}
	return response.Data(c, 200, out)
}

// Monthly returns the current calendar-month rollup; it is database-only and
// deliberately does not contact either provider.
func (h *Handler) Monthly(c *fiber.Ctx) error {
	rows, err := h.pool.Query(c.UserContext(), `SELECT provider, COUNT(*), COALESCE(SUM(request_count),0), COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COALESCE(SUM(total_tokens),0), COALESCE(SUM(estimated_cost_usd),0), MAX(created_at) FROM provider_usage_events WHERE created_at >= date_trunc('month', now()) GROUP BY provider ORDER BY provider`)
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Monthly usage is unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var provider string
		var events, requests, in, outTokens, total int
		var cost float64
		var last time.Time
		if err := rows.Scan(&provider, &events, &requests, &in, &outTokens, &total, &cost, &last); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Monthly usage is unavailable.")
		}
		out = append(out, fiber.Map{"provider": provider, "events": events, "requests": requests, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "estimatedCost": cost, "lastUsed": last})
	}
	return response.Data(c, 200, out)
}

func (h *Handler) TopUsers(c *fiber.Ctx) error {
	rows, err := h.pool.Query(c.UserContext(), `SELECT user_id, COUNT(*), COALESCE(SUM(request_count),0), COALESCE(SUM(estimated_cost_usd),0) FROM provider_usage_events WHERE created_at >= date_trunc('month', now()) GROUP BY user_id ORDER BY SUM(request_count) DESC LIMIT 10`)
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Top usage users are unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var id string
		var events, requests int
		var cost float64
		if err := rows.Scan(&id, &events, &requests, &cost); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Top usage users are unavailable.")
		}
		out = append(out, fiber.Map{"userId": id, "events": events, "requests": requests, "estimatedCost": cost})
	}
	return response.Data(c, 200, out)
}
func nullableDate(v string) any {
	if v == "" {
		return nil
	}
	return v
}
