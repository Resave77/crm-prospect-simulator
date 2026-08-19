package usage

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type traceKey struct{}

type TraceState struct {
	mu     sync.RWMutex
	values map[string]any
}

func NewTrace() *TraceState { return &TraceState{values: make(map[string]any)} }

func WithTrace(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey{}, &TraceState{values: map[string]any{"trace_id": traceID}})
}

func Trace(ctx context.Context) *TraceState {
	if t, ok := ctx.Value(traceKey{}).(*TraceState); ok && t != nil {
		return t
	}
	return nil
}

func SetTrace(ctx context.Context, key string, value any) {
	if t := Trace(ctx); t != nil {
		t.mu.Lock()
		defer t.mu.Unlock()
		if secretKey.MatchString(key) {
			value = "[REDACTED]"
		}
		t.values[key] = RedactValue(value)
	}
}

func AppendTrace(ctx context.Context, key string, value any) {
	if t := Trace(ctx); t != nil {
		t.mu.Lock()
		defer t.mu.Unlock()
		current, _ := t.values[key].([]any)
		t.values[key] = append(current, RedactValue(value))
	}
}

func GetTrace(ctx context.Context) map[string]any {
	if t := Trace(ctx); t != nil {
		t.mu.RLock()
		defer t.mu.RUnlock()
		out := make(map[string]any, len(t.values))
		for k, v := range t.values {
			out[k] = v
		}
		return out
	}
	return map[string]any{}
}

func RequestID(ctx context.Context) string {
	value := GetTrace(ctx)["trace_id"]
	if id, ok := value.(string); ok {
		return id
	}
	return uuid.NewString()
}

var secretKey = regexp.MustCompile(`(?i)(password|token|authorization|cookie|api[_-]?key|secret|refresh)`)

func RedactValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			if secretKey.MatchString(k) {
				out[k] = "[REDACTED]"
			} else {
				out[k] = RedactValue(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = RedactValue(item)
		}
		return out
	case string:
		if len(v) > 4096 {
			return v[:4096] + "…"
		}
		return v
	default:
		return value
	}
}

func SafeJSON(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return []byte(`{"value":"[UNPARSEABLE]"}`)
	}
	encoded, _ := json.Marshal(RedactValue(value))
	return encoded
}

func Domain(path string) string {
	switch {
	case strings.Contains(path, "/places") || strings.Contains(path, "prospect-finder"):
		return "gmaps"
	case strings.Contains(path, "/ai"):
		return "openai"
	case strings.Contains(path, "/auth"):
		return "auth"
	case strings.Contains(path, "/employees") || strings.Contains(path, "/users"):
		return "employee"
	case strings.Contains(path, "/prospects"):
		return "prospect"
	case strings.Contains(path, "/customers"):
		return "customer"
	default:
		return "crm"
	}
}
