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

func TestGenerateTextUnavailableStatuses(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusUnauthorized, http.StatusForbidden} {
		err := errorFromStatus(t, status)
		if !errors.Is(err, ErrAIUnavailable) {
			t.Fatalf("status=%d err=%v, want unavailable", status, err)
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
