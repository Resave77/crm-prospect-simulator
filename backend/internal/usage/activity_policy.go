package usage

import (
	"strings"
)

// ShouldPersistActivity keeps the request trace available to handlers while
// limiting persisted ActivityLog rows to useful audit and monitoring events.
func ShouldPersistActivity(method, endpoint string, responseStatus int, trace map[string]any, authenticated bool) bool {
	action, _ := trace["action"].(string)
	if meaningfulActivityAction(action) || traceProviderAttempted(trace) || trace["audit"] == true {
		return true
	}
	// Authenticated application failures are useful for audit; routine client
	// errors, health checks, static requests, and polling remain unpersisted.
	return authenticated && responseStatus >= 500 && strings.HasPrefix(endpoint, "/api/")
}

func meaningfulActivityAction(action string) bool {
	if action == "" {
		return false
	}
	return strings.HasPrefix(action, "LOGIN") || strings.HasPrefix(action, "LOGOUT") ||
		strings.HasPrefix(action, "PASSWORD_") || strings.HasPrefix(action, "CHANGE_PASSWORD") ||
		action == "RESET_PASSWORD" || action == "UPDATE_ACCOUNT" || action == "HIDE_MONITORING_HISTORY" ||
		action == "SEARCH_PROSPECT" || action == "VIEW_PROSPECT_DETAIL" || action == "LOAD_BUSINESS_INFO" ||
		action == "VIEW_PHOTOS" || action == "FIND_MENU" || action == "AI_SUMMARY" || action == "TANYA_AI" ||
		action == "MENU_PROFILING" || action == "ASSIGN_PROSPECT" || action == "CHECK_IN" || action == "CHECK_OUT"
}

func traceProviderAttempted(trace map[string]any) bool {
	if attempted, ok := trace["provider_attempted"].(bool); ok && attempted {
		return true
	}
	if count, ok := trace["provider_hit_count"].(int); ok && count > 0 {
		return true
	}
	_, hasProvider := trace["provider"]
	return hasProvider && trace["provider_hit_count"] != float64(0)
}
