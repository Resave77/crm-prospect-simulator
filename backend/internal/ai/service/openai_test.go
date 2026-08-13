package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"crm-prospect-simulator/backend/config"
	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
)

func TestGenerateTextSendsSecureResponsesRequestAndParsesUsage(t *testing.T) {
	var captured struct {
		Authorization string
		ContentType   string
		Body          requestBody
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Authorization = r.Header.Get("Authorization")
		captured.ContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&captured.Body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("x-request-id", "req_mock")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"output_text":"hello","usage":{"input_tokens":11,"output_tokens":7,"total_tokens":18}}`))
	}))
	defer server.Close()

	client := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client()))
	result, err := client.GenerateText(context.Background(), "instructions", "input")
	if err != nil {
		t.Fatal(err)
	}
	if captured.Authorization != "Bearer test-secret-key" {
		t.Fatalf("authorization header=%q", captured.Authorization)
	}
	if captured.ContentType != "application/json" {
		t.Fatalf("content-type=%q", captured.ContentType)
	}
	if captured.Body.Store {
		t.Fatal("store must be false")
	}
	if captured.Body.Model != "gpt-test" {
		t.Fatalf("model=%q", captured.Body.Model)
	}
	if captured.Body.MaxOutputTokens != 321 {
		t.Fatalf("max_output_tokens=%d", captured.Body.MaxOutputTokens)
	}
	if result.Text != "hello" || result.InputTokens != 11 || result.OutputTokens != 7 || result.TotalTokens != 18 || result.UpstreamRequestID != "req_mock" {
		t.Fatalf("unexpected result=%+v", result)
	}
}

func TestGenerateMultimodalSendsBackendImageAsInputImage(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output_text":"{}"}`))
	}))
	defer server.Close()
	client := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client()))
	if _, err := client.GenerateMultimodal(context.Background(), "instructions", "context", []ImageInput{{Bytes: []byte("image"), ContentType: "image/jpeg"}}); err != nil {
		t.Fatal(err)
	}
	input, ok := body["input"].([]any)
	if !ok || len(input) != 1 {
		t.Fatalf("input=%#v", body["input"])
	}
	message := input[0].(map[string]any)
	content := message["content"].([]any)
	if len(content) != 2 {
		t.Fatalf("content=%#v", content)
	}
	image := content[1].(map[string]any)
	if image["type"] != "input_image" || !strings.HasPrefix(image["image_url"].(string), "data:image/jpeg;base64,") {
		t.Fatalf("image=%#v", image)
	}
}

func TestGenerateWebSearchUsesOneHostedToolRequestAndReturnsToolSources(t *testing.T) {
	calls := 0
	var body requestBody
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"output_text":"{\"status\":\"NOT_FOUND\"}","output":[{"type":"web_search_call","action":{"sources":[{"title":"Official menu","url":"https://example.com/menu"}]}}]}`))
	}))
	defer server.Close()
	result, err := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateWebSearch(context.Background(), "find", "prospect")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls=%d, want 1", calls)
	}
	if len(body.Tools) != 1 || body.Tools[0]["type"] != "web_search" {
		t.Fatalf("tools=%#v", body.Tools)
	}
	if len(body.Include) != 1 || body.Include[0] != "web_search_call.action.sources" {
		t.Fatalf("include=%#v", body.Include)
	}
	if len(result.Sources) != 1 || result.Sources[0].URL != "https://example.com/menu" {
		t.Fatalf("sources=%#v", result.Sources)
	}
}

func TestGenerateWebSearchReturnsIncompleteMetadataWithoutRetry(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output_text":"{truncated","output":[{"type":"web_search_call","action":{"sources":[{"title":"Official","url":"https://example.com/menu"}]}}]}`))
	}))
	defer server.Close()
	result, err := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateWebSearch(context.Background(), "find", "prospect")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || result.ResponseStatus != "incomplete" || result.IncompleteReason != "max_output_tokens" {
		t.Fatalf("calls=%d result=%+v", calls, result)
	}
}

func TestMenuFormattingRecoveryMakesOneTextOnlyRequest(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body requestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Tools) != 0 || len(body.Include) != 0 {
			t.Fatalf("recovery must not include tools: %+v", body)
		}
		_, _ = w.Write([]byte(`{"output_text":"{\"status\":\"NOT_FOUND\"}"}`))
	}))
	defer server.Close()
	client := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client()))
	prospectAI := NewProspectAI(client, 1000, 6)
	_, err := prospectAI.RecoverProspectMenuFormatting(context.Background(), prospectmodel.Review{}, WebSearchResult{TextResult: TextResult{Text: "{truncated"}, Sources: []WebSearchSource{{URL: "https://example.com/menu"}}})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls=%d, want 1", calls)
	}
}

func TestFindMenuPromptUsesHardCompactRepresentativeLimits(t *testing.T) {
	var body requestBody
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"output_text":"{\"status\":\"NOT_FOUND\"}"}`))
	}))
	defer server.Close()
	client := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client()))
	_, err := NewProspectAI(client, 1000, 6).FindProspectMenu(context.Background(), prospectmodel.Review{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"at most 5 source-defined categories",
		"at most 20 menu items TOTAL",
		"at most 2 price/source evidence entries per item",
		"Stop after maximum 20 menu items total",
		"Do not attempt to return the full merchant catalog",
		"representative subset sufficient for CRM sales analysis",
		"ensure the JSON object is complete and closed",
	} {
		if !strings.Contains(body.Instructions, required) {
			t.Fatalf("prompt missing %q", required)
		}
	}
	if body.MaxOutputTokens != 4000 {
		t.Fatalf("max_output_tokens=%d, want 4000", body.MaxOutputTokens)
	}
	if len(body.Tools) != 1 || body.Tools[0]["type"] != "web_search" {
		t.Fatalf("tools=%#v", body.Tools)
	}
}

func TestWebSearchSourcesIncludeGroundedCitationAnnotations(t *testing.T) {
	var parsed responseBody
	raw := `{"output":[{"type":"web_search_call","action":{"sources":[{"title":"Tool","url":"https://example.com/tool"}]}},{"type":"message","content":[{"type":"output_text","text":"result","annotations":[{"type":"url_citation","title":"Citation","url":"https://example.com/citation"}]}]}]}`
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatal(err)
	}
	sources := parsed.webSources()
	if len(sources) != 2 || sources[0].URL != "https://example.com/tool" || sources[1].URL != "https://example.com/citation" {
		t.Fatalf("sources=%+v", sources)
	}
}

func TestOperationSpecificOutputBudgetDoesNotChangeClientDefault(t *testing.T) {
	budgets := make([]int, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body requestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		budgets = append(budgets, body.MaxOutputTokens)
		_, _ = w.Write([]byte(`{"output_text":"{}"}`))
	}))
	defer server.Close()
	client := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client()))
	if _, err := client.GenerateTextWithMaxOutputTokens(context.Background(), "", "", 1600); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GenerateText(context.Background(), "", ""); err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 2 || budgets[0] != 1600 || budgets[1] != 321 {
		t.Fatalf("budgets=%v", budgets)
	}
}

func TestClientKeepsDedicatedFindMenuTimeout(t *testing.T) {
	cfg := testConfig()
	cfg.OpenAITimeout = 20 * time.Second
	cfg.OpenAIFindMenuTimeout = 60 * time.Second
	client := NewClient(cfg)
	if client.timeout != 20*time.Second || client.findMenuTimeout != 60*time.Second {
		t.Fatalf("normal=%s finder=%s", client.timeout, client.findMenuTimeout)
	}
}

func TestClientKeepsDedicatedMenuProfileTimeout(t *testing.T) {
	cfg := testConfig()
	cfg.OpenAITimeout = 20 * time.Second
	cfg.OpenAIMenuProfileTimeout = 40 * time.Second
	client := NewClient(cfg)
	if client.timeout != 20*time.Second || client.menuProfileTimeout != 40*time.Second {
		t.Fatalf("normal=%s profiling=%s", client.timeout, client.menuProfileTimeout)
	}
}

func TestGenerateTextParsesNestedOutputText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"resp_mock","output":[{"content":[{"type":"output_text","text":"nested"}]}],"usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}`))
	}))
	defer server.Close()

	result, err := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateText(context.Background(), "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "nested" || result.UpstreamRequestID != "resp_mock" {
		t.Fatalf("unexpected result=%+v", result)
	}
}

func TestGenerateTextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	cfg := testConfig()
	cfg.OpenAITimeout = time.Millisecond
	_, err := NewClient(cfg, WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateText(context.Background(), "", "")
	if !errors.Is(err, ErrAITimeout) {
		t.Fatalf("err=%v, want timeout", err)
	}
}

func TestGenerateTextRateLimited(t *testing.T) {
	err := errorFromStatus(t, http.StatusTooManyRequests)
	if !errors.Is(err, ErrAIRateLimited) {
		t.Fatalf("err=%v, want rate limited", err)
	}
}

func TestGenerateTextRateLimitedKeepsSafeMappingWithDiagnosticsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req_rate")
		w.Header().Set("x-ratelimit-remaining-requests", "0")
		w.Header().Set("x-ratelimit-reset-requests", "1s")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"type":"rate_limit_exceeded","code":"rate_limit_exceeded","message":"sensitive upstream detail"}}`))
	}))
	defer server.Close()

	_, err := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateText(context.Background(), "", "")
	if !errors.Is(err, ErrAIRateLimited) || SafeErrorCode(err) != "AI_RATE_LIMITED" {
		t.Fatalf("err=%v, code=%s", err, SafeErrorCode(err))
	}
}

func TestGenerateTextUnavailableStatuses(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadGateway} {
		err := errorFromStatus(t, status)
		if !errors.Is(err, ErrAIUnavailable) {
			t.Fatalf("status=%d err=%v, want unavailable", status, err)
		}
	}
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		if err := errorFromStatus(t, status); !errors.Is(err, ErrAIAuthentication) {
			t.Fatalf("status=%d err=%v, want authentication", status, err)
		}
	}
}

func TestGenerateTextInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"output":[]}`))
	}))
	defer server.Close()

	_, err := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateText(context.Background(), "", "")
	if !errors.Is(err, ErrAIInvalidResponse) {
		t.Fatalf("err=%v, want invalid response", err)
	}
}

func TestUnconfiguredDoesNotMakeOutboundRequest(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer server.Close()

	for _, cfg := range []config.Config{
		{AIEnabled: false, OpenAIAPIKey: "test-secret-key", OpenAIModel: "gpt-test", OpenAITimeout: time.Second, OpenAIMaxTokens: 1},
		{AIEnabled: true, OpenAIAPIKey: "", OpenAIModel: "gpt-test", OpenAITimeout: time.Second, OpenAIMaxTokens: 1},
		{AIEnabled: true, OpenAIAPIKey: "test-secret-key", OpenAIModel: "", OpenAITimeout: time.Second, OpenAIMaxTokens: 1},
	} {
		_, err := NewClient(cfg, WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateText(context.Background(), "", "")
		if !errors.Is(err, ErrAINotConfigured) {
			t.Fatalf("err=%v, want not configured", err)
		}
	}
	if calls != 0 {
		t.Fatalf("outbound calls=%d, want zero", calls)
	}
}

func TestAPIKeyNeverAppearsInReturnedErrors(t *testing.T) {
	for _, err := range []error{
		errorFromStatus(t, http.StatusInternalServerError),
		errorFromStatus(t, http.StatusUnauthorized),
		ErrAINotConfigured,
	} {
		if strings.Contains(err.Error(), "test-secret-key") {
			t.Fatalf("error leaked API key: %v", err)
		}
		if strings.Contains(SafeError(err).Error(), "test-secret-key") {
			t.Fatalf("safe error leaked API key: %v", err)
		}
	}
}

func errorFromStatus(t *testing.T, status int) error {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream body must not be returned"}}`))
	}))
	defer server.Close()
	_, err := NewClient(testConfig(), WithEndpoint(server.URL), WithHTTPClient(server.Client())).GenerateText(context.Background(), "", "")
	return err
}

func testConfig() config.Config {
	return config.Config{
		AIEnabled:       true,
		OpenAIAPIKey:    "test-secret-key",
		OpenAIModel:     "gpt-test",
		OpenAITimeout:   time.Second,
		OpenAIMaxTokens: 321,
	}
}
