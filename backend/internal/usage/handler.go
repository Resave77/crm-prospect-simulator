package usage

import (
	"context"
	"crm-prospect-simulator/backend/internal/auth/service"
	"crm-prospect-simulator/backend/internal/shared/response"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"time"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func (h *Handler) googleRuntimeCosts(ctx context.Context, from, to time.Time) (map[string]GoogleCostRow, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	anchor := from
	if anchor.IsZero() {
		anchor = time.Now().In(loc)
	}
	monthStart := time.Date(anchor.In(loc).Year(), anchor.In(loc).Month(), 1, 0, 0, 0, 0, loc)
	rows, err := h.pool.Query(ctx, `SELECT id::text,created_at,user_id::text,operation,COALESCE(field_mask,''),COALESCE(request_count,1),success FROM provider_usage_events WHERE provider='GOOGLE_MAPS' AND created_at >= $1 AND ($2::timestamptz IS NULL OR created_at < $2) ORDER BY created_at,id`, monthStart, nullableTime(to))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]GoogleCostEvent, 0)
	for rows.Next() {
		var id, uid, operation, mask string
		var created time.Time
		var count int
		var success bool
		if err := rows.Scan(&id, &created, &uid, &operation, &mask, &count, &success); err != nil {
			return nil, err
		}
		resolved, _, ok := ResolveGoogleSKU(operation, "", mask)
		if !ok {
			continue
		}
		events = append(events, GoogleCostEvent{ID: id, UserID: uid, Operation: resolved, CreatedAt: created, Success: success, RequestCount: count})
	}
	return AllocateGoogleCost(events, DefaultPricing(), from, to), rows.Err()
}

func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func parseUsageTimes(from, to string) (time.Time, time.Time) {
	var f, t time.Time
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if from != "" {
		f, _ = time.ParseInLocation("2006-01-02", from, loc)
		if f.IsZero() {
			f, _ = time.Parse(time.RFC3339, from)
		}
	}
	if to != "" {
		t, _ = time.ParseInLocation("2006-01-02", to, loc)
		if !t.IsZero() {
			t = t.AddDate(0, 0, 1)
		} else {
			t, _ = time.Parse(time.RFC3339, to)
		}
	}
	return f, t
}
func payableMicrosRequests(micros int64) int {
	if micros <= 0 {
		return 0
	}
	return int(micros / 1000)
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	fromTime, toTime := parseUsageTimes(c.Query("dateFrom"), c.Query("dateTo"))
	runtimeCosts, _ := h.googleRuntimeCosts(c.UserContext(), fromTime, toTime)
	rows, err := h.pool.Query(c.UserContext(), `SELECT provider, operation, COALESCE(api_or_model,'') , COUNT(*), COALESCE(SUM(request_count),0), COALESCE(SUM(request_count) FILTER (WHERE success),0), COALESCE(SUM(request_count) FILTER (WHERE NOT success),0), COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COALESCE(SUM(total_tokens),0), COALESCE(SUM(cached_tokens),0), COALESCE(SUM(estimated_cost_usd),0), MAX(created_at) FROM provider_usage_events WHERE ($1='' OR user_id::text=$1) AND ($2='' OR provider=$2) AND ($3='' OR feature=$3) AND ($4::timestamptz IS NULL OR created_at >= $4) AND ($5::timestamptz IS NULL OR created_at < $5) GROUP BY provider, operation, api_or_model ORDER BY provider, operation, api_or_model`, c.Query("userId"), c.Query("provider"), c.Query("feature"), nullableDate(c.Query("dateFrom")), nullableDate(c.Query("dateTo")))
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage summary is unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var provider, operation, model string
		var events, requests, success, failed, in, outTokens, total int
		var cached int
		var cost float64
		var last time.Time
		if err := rows.Scan(&provider, &operation, &model, &events, &requests, &success, &failed, &in, &outTokens, &total, &cached, &cost, &last); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage summary is unavailable.")
		}
		freeTier, verified := DefaultPricing().FreeTier(provider, operation)
		costStatus := "UNCONFIGURED"
		if cost > 0 {
			costStatus = "ESTIMATE"
		}
		var usagePercent any
		if verified && c.Query("userId") == "" {
			usagePercent = float64(requests) / float64(freeTier) * 100
		}
		row := fiber.Map{"provider": provider, "operation": operation, "apiOrModel": model, "events": events, "requests": requests, "success": success, "failed": failed, "billableRequests": success, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "cachedTokens": cached, "estimatedCost": cost, "costStatus": costStatus, "lastUsed": last}
		if provider == "OPENAI" {
			if calculated, ok := DefaultPricing().OpenAICost(model, in, cached, outTokens); ok && total > 0 {
				row["estimatedCost"], row["estimatedCostMicros"], row["costState"] = float64(calculated)/1e6, calculated, "VERIFIED_ESTIMATE"
			} else {
				row["costState"] = "USAGE_UNKNOWN"
			}
		}
		if provider == "GOOGLE_MAPS" {
			row["estimatedCost"], row["costState"], row["currency"] = 0, "SKU_UNKNOWN", "USD"
			if runtime, ok := runtimeCosts[operation]; ok {
				payable := runtime.UserPayableMicros[c.Query("userId")]
				row["grossCostMicros"], row["estimatedPayableCostMicros"], row["costState"], row["currency"] = runtime.GrossCostMicros, payable, "VERIFIED_ESTIMATE", "USD"
				row["estimatedCost"] = float64(payable) / 1e6
			}
		}
		if c.Query("userId") == "" {
			row["freeTier"], row["freeTierVerified"], row["usagePercent"] = freeTier, verified, usagePercent
		}
		out = append(out, row)
	}
	return response.Data(c, 200, out)
}

type AggregateGroup struct {
	Provider, Operation, SKUCategory, APIOrModel                           string
	Events, Requests, InputTokens, OutputTokens, TotalTokens, CachedTokens int
	EstimatedCost                                                          float64
	FreeTier                                                               int
	FreeTierVerified                                                       bool
	UsagePercent                                                           *float64
}

type UserContribution struct {
	UserID, UserName, EmployeeID, Provider, Operation, SKUCategory, APIOrModel string
	Events, Requests, InputTokens, OutputTokens, TotalTokens, CachedTokens     int
	ContributionPercent                                                        float64
}

type aggregateKey struct{ Provider, Operation, SKUCategory, APIOrModel string }

func contributionPercent(requests, total int) float64 {
	if total <= 0 || requests <= 0 {
		return 0
	}
	return float64(requests) / float64(total) * 100
}

func aggregateMemory(events []Event) ([]AggregateGroup, []UserContribution) {
	groups := map[aggregateKey]*AggregateGroup{}
	users := map[aggregateKey]map[string]*UserContribution{}
	for _, event := range events {
		key := aggregateKey{event.Provider, event.Operation, event.SKUCategory, event.APIOrModel}
		group := groups[key]
		if group == nil {
			group = &AggregateGroup{Provider: key.Provider, Operation: key.Operation, SKUCategory: key.SKUCategory, APIOrModel: key.APIOrModel}
			groups[key] = group
		}
		group.Events++
		group.Requests += event.RequestCount
		group.InputTokens += event.InputTokens
		group.OutputTokens += event.OutputTokens
		group.TotalTokens += event.TotalTokens
		group.CachedTokens += event.CachedTokens
		group.EstimatedCost += event.EstimatedCost
		byUser := users[key]
		if byUser == nil {
			byUser = map[string]*UserContribution{}
			users[key] = byUser
		}
		uid := event.UserID.String()
		user := byUser[uid]
		if user == nil {
			user = &UserContribution{UserID: uid, Provider: key.Provider, Operation: key.Operation, SKUCategory: key.SKUCategory, APIOrModel: key.APIOrModel}
			byUser[uid] = user
		}
		user.Events++
		user.Requests += event.RequestCount
		user.InputTokens += event.InputTokens
		user.OutputTokens += event.OutputTokens
		user.TotalTokens += event.TotalTokens
		user.CachedTokens += event.CachedTokens
	}
	result := make([]AggregateGroup, 0, len(groups))
	contributions := make([]UserContribution, 0)
	for key, group := range groups {
		_ = key
		result = append(result, *group)
		for _, user := range users[key] {
			user.ContributionPercent = contributionPercent(user.Requests, group.Requests)
			contributions = append(contributions, *user)
		}
	}
	return result, contributions
}

// ProjectSummary is a database-only rollup across every user. It intentionally
// contains no user_id predicate; attribution is returned separately.
func (h *Handler) ProjectSummary(c *fiber.Ctx) error {
	from, to := nullableDate(c.Query("dateFrom")), nullableDate(c.Query("dateTo"))
	fromTime, toTime := parseUsageTimes(c.Query("dateFrom"), c.Query("dateTo"))
	runtimeCosts, _ := h.googleRuntimeCosts(c.UserContext(), fromTime, toTime)
	where := `WHERE ($1='' OR provider=$1) AND ($2='' OR operation=$2) AND ($3='' OR api_or_model ILIKE '%' || $3 || '%' OR sku_category ILIKE '%' || $3 || '%') AND ($4::timestamptz IS NULL OR created_at >= $4) AND ($5::timestamptz IS NULL OR created_at < $5)`
	resolvedSKU := `CASE WHEN provider='GOOGLE_MAPS' AND operation='PLACE_DETAILS' AND field_mask ILIKE '%nationalPhoneNumber%' AND field_mask ILIKE '%regularOpeningHours%' THEN 'BUSINESS_INFO' WHEN provider='GOOGLE_MAPS' AND operation='PLACE_DETAILS' AND field_mask ILIKE '%displayName%' AND field_mask ILIKE '%googleMapsUri%' THEN 'CORE_DETAIL' ELSE COALESCE(sku_category,'') END`
	userWhere := strings.ReplaceAll(where, "created_at", "e.created_at")
	rows, err := h.pool.Query(c.UserContext(), fmt.Sprintf(`SELECT provider,operation,%s,COALESCE(api_or_model,''),COUNT(*),COALESCE(SUM(request_count),0),COALESCE(SUM(request_count) FILTER (WHERE success),0),COALESCE(SUM(request_count) FILTER (WHERE NOT success),0),COALESCE(SUM(input_tokens),0),COALESCE(SUM(output_tokens),0),COALESCE(SUM(total_tokens),0),COALESCE(SUM(cached_tokens),0),COALESCE(SUM(estimated_cost_usd),0) FROM provider_usage_events %s GROUP BY provider,operation,%s,api_or_model ORDER BY provider,operation,%s,api_or_model`, resolvedSKU, where, resolvedSKU, resolvedSKU), c.Query("provider"), c.Query("operation"), c.Query("q"), from, to)
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Project usage is unavailable.")
	}
	defer rows.Close()
	groups := make([]fiber.Map, 0)
	groupTotals := map[aggregateKey]int{}
	groupTokenTotals := map[aggregateKey]int{}
	for rows.Next() {
		var provider, operation, sku, model string
		var events, requests, success, failed, in, outTokens, total, cached int
		var cost float64
		if err := rows.Scan(&provider, &operation, &sku, &model, &events, &requests, &success, &failed, &in, &outTokens, &total, &cached, &cost); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Project usage is unavailable.")
		}
		key := aggregateKey{provider, operation, sku, model}
		groupTotals[key] = requests
		groupTokenTotals[key] = total
		free, verified := DefaultPricing().FreeTier(provider, operation)
		var percent any
		if verified {
			percent = float64(requests) / float64(free) * 100
		}
		groups = append(groups, fiber.Map{"provider": provider, "operation": operation, "skuCategory": sku, "apiOrModel": model, "events": events, "requests": requests, "success": success, "failed": failed, "billableRequests": success, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "cachedTokens": cached, "estimatedCost": cost, "costStatus": costStatus(provider, cost), "costState": projectCostState(provider, cost), "freeTier": free, "freeTierVerified": verified, "usagePercent": percent})
		if provider == "OPENAI" {
			if calculated, ok := DefaultPricing().OpenAICost(model, in, cached, outTokens); ok && total > 0 {
				groups[len(groups)-1]["estimatedCost"] = float64(calculated) / 1e6
				groups[len(groups)-1]["estimatedCostMicros"] = calculated
				groups[len(groups)-1]["costState"] = "VERIFIED_ESTIMATE"
			} else {
				groups[len(groups)-1]["costState"] = "USAGE_UNKNOWN"
			}
		}
		if provider == "GOOGLE_MAPS" {
			groups[len(groups)-1]["estimatedCost"], groups[len(groups)-1]["costState"], groups[len(groups)-1]["currency"] = 0, "SKU_UNKNOWN", "USD"
			runtimeOperation := operation
			if operation == "PLACE_DETAILS" && (sku == "CORE_DETAIL" || sku == "BUSINESS_INFO") {
				runtimeOperation = sku
			}
			if runtime, ok := runtimeCosts[runtimeOperation]; ok {
				groups[len(groups)-1]["billableRequests"] = runtime.BillableRequests
				groups[len(groups)-1]["freeUsageCap"] = runtime.FreeUsageCap
				groups[len(groups)-1]["freeUsageConsumed"] = runtime.FreeUsageConsumed
				groups[len(groups)-1]["paidRequests"] = runtime.PaidRequestCount
				groups[len(groups)-1]["grossCostMicros"] = runtime.GrossCostMicros
				groups[len(groups)-1]["estimatedPayableCostMicros"] = runtime.EstimatedPayableMicros
				groups[len(groups)-1]["estimatedCost"] = float64(runtime.EstimatedPayableMicros) / 1e6
				groups[len(groups)-1]["costState"] = "VERIFIED_ESTIMATE"
				groups[len(groups)-1]["currency"] = "USD"
			}
		}
	}
	userResolvedSKU := strings.ReplaceAll(resolvedSKU, "provider", "e.provider")
	userResolvedSKU = strings.ReplaceAll(userResolvedSKU, "operation", "e.operation")
	userResolvedSKU = strings.ReplaceAll(userResolvedSKU, "field_mask", "e.field_mask")
	userResolvedSKU = strings.ReplaceAll(userResolvedSKU, "sku_category", "e.sku_category")
	userRows, err := h.pool.Query(c.UserContext(), fmt.Sprintf(`SELECT e.user_id,COALESCE(u.full_name,''),COALESCE(u.employee_id,''),e.provider,e.operation,%s,COALESCE(e.api_or_model,''),COUNT(*),COALESCE(SUM(e.request_count),0),COALESCE(SUM(e.input_tokens),0),COALESCE(SUM(e.output_tokens),0),COALESCE(SUM(e.total_tokens),0),COALESCE(SUM(e.cached_tokens),0) FROM provider_usage_events e LEFT JOIN users u ON u.id=e.user_id %s GROUP BY e.user_id,u.full_name,u.employee_id,e.provider,e.operation,%s,e.api_or_model ORDER BY SUM(e.request_count) DESC`, userResolvedSKU, userWhere, userResolvedSKU), c.Query("provider"), c.Query("operation"), c.Query("q"), from, to)
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Project contributors are unavailable.")
	}
	defer userRows.Close()
	contributors := make([]fiber.Map, 0)
	for userRows.Next() {
		var uid, name, employee, provider, operation, sku, model string
		var events, requests, in, outTokens, total, cached int
		if err := userRows.Scan(&uid, &name, &employee, &provider, &operation, &sku, &model, &events, &requests, &in, &outTokens, &total, &cached); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Project contributors are unavailable.")
		}
		key := aggregateKey{provider, operation, sku, model}
		item := fiber.Map{"userId": uid, "userName": name, "employeeId": employee, "provider": provider, "operation": operation, "skuCategory": sku, "apiOrModel": model, "events": events, "requests": requests, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "cachedTokens": cached, "contributionPercent": contributionPercent(requests, groupTotals[key]), "tokenContributionPercent": contributionPercent(total, groupTokenTotals[key])}
		if provider == "OPENAI" {
			if calculated, ok := DefaultPricing().OpenAICost(model, in, cached, outTokens); ok && total > 0 {
				item["estimatedCostMicros"], item["estimatedCost"], item["costState"], item["currency"] = calculated, float64(calculated)/1e6, "VERIFIED_ESTIMATE", "USD"
			} else {
				item["costState"] = "USAGE_UNKNOWN"
			}
		}
		contributors = append(contributors, item)
	}
	return response.Data(c, 200, fiber.Map{"groups": groups, "contributors": contributors})
}
func (h *Handler) History(c *fiber.Ctx) error {
	status := historyVisibility(c.Query("historyStatus"))
	rows, err := h.pool.Query(c.UserContext(), `SELECT e.created_at,e.user_id,COALESCE(e.request_id::text,''),e.provider,e.feature,e.operation,e.api_or_model,COALESCE(e.field_mask,''),COALESCE(e.request_count,0),COALESCE(e.input_tokens,0),COALESCE(e.output_tokens,0),COALESCE(e.total_tokens,0),e.cached_tokens,COALESCE(e.estimated_cost_usd,0),COALESCE(e.http_status,0),e.success,COALESCE(e.credential_alias,''),COALESCE(e.environment,''),e.hidden_at,COALESCE(e.hidden_by::text,''),COALESCE(e.hide_reason,''),COALESCE(u.full_name,'') FROM provider_usage_events e LEFT JOIN users u ON u.id=e.hidden_by WHERE ($1='' OR e.user_id::text=$1) AND ($2='' OR e.provider=$2) AND ($3='' OR e.feature=$3) AND ($4='' OR e.operation=$4) AND ($5::timestamptz IS NULL OR e.created_at >= $5) AND ($6::timestamptz IS NULL OR e.created_at < $6) AND `+visibilityPredicate("e.hidden_at", status)+` ORDER BY e.created_at DESC LIMIT 200`, c.Query("userId"), c.Query("provider"), c.Query("feature"), c.Query("operation"), nullableDate(c.Query("dateFrom")), nullableDate(c.Query("dateTo")))
	if err != nil {
		return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage history is unavailable.")
	}
	defer rows.Close()
	out := make([]fiber.Map, 0)
	for rows.Next() {
		var t time.Time
		var uid, requestID, provider, feature, operation, api, fieldMask string
		var calls, in, outTokens, total, status int
		var cached *int
		var cost float64
		var ok bool
		var credentialAlias, environment, hiddenBy, hideReason, hiddenByName string
		var hiddenAt *time.Time
		if err := rows.Scan(&t, &uid, &requestID, &provider, &feature, &operation, &api, &fieldMask, &calls, &in, &outTokens, &total, &cached, &cost, &status, &ok, &credentialAlias, &environment, &hiddenAt, &hiddenBy, &hideReason, &hiddenByName); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Usage history is unavailable.")
		}
		out = append(out, fiber.Map{"createdAt": t, "userId": uid, "requestId": requestID, "provider": provider, "feature": feature, "operation": operation, "apiOrModel": api, "fieldMask": fieldMask, "requestCount": calls, "inputTokens": in, "outputTokens": outTokens, "totalTokens": total, "cachedTokens": cached, "estimatedCost": cost, "httpStatus": status, "success": ok, "credentialAlias": credentialAlias, "environment": environment, "hiddenAt": hiddenAt, "hiddenBy": hiddenBy, "hiddenByName": hiddenByName, "hideReason": hideReason})
	}
	return response.Data(c, 200, out)
}

func (h *Handler) ActivityHistory(c *fiber.Ctx) error {
	visibility := historyVisibility(c.Query("historyStatus"))
	rows, err := h.pool.Query(c.UserContext(), `SELECT a.created_at,a.request_id,a.method,a.endpoint,a.domain,a.response_status,a.additional_trace,a.hidden_at,COALESCE(a.hidden_by::text,''),COALESCE(a.hide_reason,''),COALESCE(u.full_name,'') FROM activity_logs a LEFT JOIN users u ON u.id=a.hidden_by WHERE ($1='' OR a.user_id::text=$1) AND (`+
		`a.request_id IN (SELECT request_id FROM provider_usage_events WHERE request_id IS NOT NULL) OR `+
		`a.additional_trace->>'provider' IN ('GOOGLE_MAPS','OPENAI') OR `+
		`COALESCE((a.additional_trace->>'provider_attempted')::boolean,FALSE) OR `+
		`COALESCE((a.additional_trace->>'provider_hit_count')::int,0)>0 OR `+
		`a.additional_trace->>'action' IN ('SEARCH_PROSPECT','VIEW_PROSPECT_DETAIL','LOAD_BUSINESS_INFO','VIEW_PHOTOS','FIND_MENU','AI_SUMMARY','MENU_PROFILING','TANYA_AI')) AND `+
		visibilityPredicate("a.hidden_at", visibility)+` ORDER BY a.created_at DESC LIMIT 200`, c.Query("userId"))
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
		var hiddenAt *time.Time
		var hiddenBy, hideReason, hiddenByName string
		if err := rows.Scan(&created, &requestID, &method, &endpoint, &domain, &status, &trace, &hiddenAt, &hiddenBy, &hideReason, &hiddenByName); err != nil {
			return response.Error(c, 500, "USAGE_QUERY_FAILED", "Activity history is unavailable.")
		}
		var details any
		_ = json.Unmarshal(trace, &details)
		out = append(out, fiber.Map{"createdAt": created, "requestId": requestID, "method": method, "endpoint": endpoint, "domain": domain, "responseStatus": status, "trace": details, "hiddenAt": hiddenAt, "hiddenBy": hiddenBy, "hiddenByName": hiddenByName, "hideReason": hideReason})
	}
	return response.Data(c, 200, out)
}

type hideHistoryRequest struct {
	UserID    string `json:"userId"`
	RequestID string `json:"requestId"`
	Reason    string `json:"reason"`
}

func (h *Handler) HideHistory(c *fiber.Ctx) error {
	SetTrace(c.UserContext(), "action", "HIDE_MONITORING_HISTORY")
	var input hideHistoryRequest
	if err := c.BodyParser(&input); err != nil || input.UserID == "" || input.RequestID == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_HISTORY_REQUEST", "A user and request ID are required.")
	}
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_HISTORY_REQUEST", "The user ID is invalid.")
	}
	reason := strings.TrimSpace(input.Reason)
	if len(reason) > 500 {
		reason = reason[:500]
	}
	tx, err := h.pool.Begin(c.UserContext())
	if err != nil {
		return response.Error(c, 500, "USAGE_UPDATE_FAILED", "Usage history could not be hidden.")
	}
	defer tx.Rollback(c.UserContext())
	actorID := c.Locals("auth.principal")
	principal, ok := actorID.(service.Principal)
	if !ok {
		return response.Error(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "You do not have permission to hide usage history.")
	}
	hiddenBy := principal.UserID
	if _, err = tx.Exec(c.UserContext(), `UPDATE activity_logs SET hidden_at=now(), hidden_by=$3, hide_reason=$4 WHERE user_id=$1 AND request_id=$2`, userID, input.RequestID, hiddenBy, nullableReason(reason)); err != nil {
		return response.Error(c, 500, "USAGE_UPDATE_FAILED", "Usage history could not be hidden.")
	}
	if _, err = tx.Exec(c.UserContext(), `UPDATE provider_usage_events SET hidden_at=now(), hidden_by=$3, hide_reason=$4 WHERE user_id=$1 AND request_id=$2`, userID, input.RequestID, hiddenBy, nullableReason(reason)); err != nil {
		return response.Error(c, 500, "USAGE_UPDATE_FAILED", "Usage history could not be hidden.")
	}
	if err = tx.Commit(c.UserContext()); err != nil {
		return response.Error(c, 500, "USAGE_UPDATE_FAILED", "Usage history could not be hidden.")
	}
	return response.Data(c, fiber.StatusOK, fiber.Map{"hidden": true, "requestId": input.RequestID})
}

func nullableReason(reason string) any {
	if reason == "" {
		return nil
	}
	return reason
}

func historyVisibility(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "HIDDEN":
		return "HIDDEN"
	case "ALL":
		return "ALL"
	default:
		return "ACTIVE"
	}
}

func visibilityPredicate(column, status string) string {
	switch status {
	case "HIDDEN":
		return column + " IS NOT NULL"
	case "ALL":
		return "TRUE"
	default:
		return column + " IS NULL"
	}
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

func costStatus(provider string, cost float64) string {
	if provider == "GOOGLE_MAPS" {
		return "UNCONFIGURED"
	}
	if cost > 0 {
		return "ESTIMATE"
	}
	return "UNCONFIGURED"
}

func projectCostState(provider string, cost float64) string {
	if provider == "OPENAI" {
		return "PRICING_UNVERIFIED"
	}
	if provider == "GOOGLE_MAPS" && cost == 0 {
		return "VERIFIED_ESTIMATE"
	}
	return costStatus(provider, cost)
}

func nullableDate(v string) any {
	if v == "" {
		return nil
	}
	return v
}
