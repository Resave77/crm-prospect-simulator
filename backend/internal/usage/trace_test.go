package usage

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestTraceIsRequestScopedAndRedactsSecrets(t *testing.T) {
	ctx := WithTrace(context.Background(), "req-1")
	SetTrace(ctx, "action", "SEARCH_PROSPECT")
	SetTrace(ctx, "password", "do-not-store")
	AppendTrace(ctx, "events", map[string]any{"apiKey": "secret", "operation": "NEARBY_SEARCH"})
	values := GetTrace(ctx)
	if values["action"] != "SEARCH_PROSPECT" || values["password"] != "[REDACTED]" { t.Fatalf("trace=%v", values) }
	if strings.Contains(string(mustJSON(values)), "secret") || strings.Contains(string(mustJSON(values)), "do-not-store") { t.Fatal("trace leaked a secret") }
	other := WithTrace(context.Background(), "req-2")
	if len(GetTrace(other)) != 1 { t.Fatalf("trace state leaked across contexts: %v", GetTrace(other)) }
}

func mustJSON(value any) []byte { encoded, _ := json.Marshal(value); return encoded }
