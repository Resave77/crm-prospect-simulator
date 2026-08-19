package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"crm-prospect-simulator/backend/config"
	usagepkg "crm-prospect-simulator/backend/internal/usage"
)

const responsesAPIURL = "https://api.openai.com/v1/responses"

type Client struct {
	apiKey             string
	model              string
	maxOutputTokens    int
	timeout            time.Duration
	findMenuTimeout    time.Duration
	menuProfileTimeout time.Duration
	httpClient         *http.Client
	endpoint           string
	recorder           usagepkg.Recorder
}

func (c *Client) SetUsageRecorder(r usagepkg.Recorder) { c.recorder = r }

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		if strings.TrimSpace(endpoint) != "" {
			c.endpoint = strings.TrimSpace(endpoint)
		}
	}
}

type TextResult struct {
	Text              string
	InputTokens       int
	OutputTokens      int
	TotalTokens       int
	UpstreamRequestID string
}

type WebSearchResult struct {
	TextResult
	Sources          []WebSearchSource
	ResponseStatus   string
	IncompleteReason string
}

type WebSearchSource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type ImageInput struct {
	Bytes       []byte
	ContentType string
}

func NewClient(cfg config.Config, opts ...Option) *Client {
	client := &Client{
		apiKey:             strings.TrimSpace(cfg.OpenAIAPIKey),
		model:              strings.TrimSpace(cfg.OpenAIModel),
		maxOutputTokens:    cfg.OpenAIMaxTokens,
		timeout:            cfg.OpenAITimeout,
		findMenuTimeout:    cfg.OpenAIFindMenuTimeout,
		menuProfileTimeout: cfg.OpenAIMenuProfileTimeout,
		httpClient:         &http.Client{},
		endpoint:           responsesAPIURL,
	}
	if !cfg.AIConfigured() {
		client.apiKey = ""
		client.model = ""
	}
	if client.timeout <= 0 {
		client.timeout = 20 * time.Second
	}
	if client.findMenuTimeout <= 0 {
		client.findMenuTimeout = 60 * time.Second
	}
	if client.menuProfileTimeout <= 0 {
		client.menuProfileTimeout = 40 * time.Second
	}
	if client.maxOutputTokens <= 0 {
		client.maxOutputTokens = 800
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *Client) Configured() bool {
	return c != nil && c.apiKey != "" && c.model != ""
}

func (c *Client) GenerateText(ctx context.Context, instructions, input string) (TextResult, error) {
	return c.GenerateTextWithMaxOutputTokens(ctx, instructions, input, c.maxOutputTokens)
}

func (c *Client) GenerateTextWithMaxOutputTokens(ctx context.Context, instructions, input string, maxOutputTokens int) (TextResult, error) {
	return c.GenerateTextWithOptions(ctx, instructions, input, maxOutputTokens, c.timeout)
}

func (c *Client) GenerateTextWithOptions(ctx context.Context, instructions, input string, maxOutputTokens int, requestTimeout time.Duration) (TextResult, error) {
	started := time.Now()
	if !c.Configured() {
		return TextResult{}, ErrAINotConfigured
	}
	if _, err := url.ParseRequestURI(c.endpoint); err != nil {
		return TextResult{}, ErrAIInvalidResponse
	}
	if maxOutputTokens <= 0 {
		maxOutputTokens = c.maxOutputTokens
	}
	body := requestBody{
		Model:           c.model,
		Store:           false,
		MaxOutputTokens: maxOutputTokens,
		Instructions:    strings.TrimSpace(instructions),
		Input:           strings.TrimSpace(input),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return TextResult{}, ErrAIInvalidResponse
	}
	if requestTimeout <= 0 {
		requestTimeout = c.timeout
	}
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return TextResult{}, ErrAIInvalidResponse
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.recordUsage(ctx, 0, 0, 0, 0, false, "request_error")
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return TextResult{}, ErrAITimeout
		}
		return TextResult{}, ErrAIUnavailable
	}
	defer resp.Body.Close()

	requestID := resp.Header.Get("x-request-id")
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "rate_limited")
		logUpstreamFailure(resp, requestID)
		return TextResult{UpstreamRequestID: requestID}, ErrAIRateLimited
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "authentication")
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, ErrAIAuthentication
	case resp.StatusCode >= 500:
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "upstream")
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, ErrAIUnavailable
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "rejected")
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, ErrAIRequestRejected
	}

	var parsed responseBody
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&parsed); err != nil {
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "invalid_response")
		return TextResult{UpstreamRequestID: requestID}, ErrAIInvalidResponse
	}
	text := strings.TrimSpace(parsed.OutputText)
	if text == "" {
		c.recordUsage(ctx, resp.StatusCode, parsed.Usage.InputTokens, parsed.Usage.OutputTokens, parsed.Usage.TotalTokens, false, "empty_response")
		text = strings.TrimSpace(parsed.firstText())
	}
	if text == "" {
		logResponseMetadata("text", c.model, requestID, resp.StatusCode, started, parsed, text, ErrAIInvalidResponse)
		return TextResult{UpstreamRequestID: requestID}, ErrAIInvalidResponse
	}
	if requestID == "" {
		requestID = parsed.ID
	}
	logResponseMetadata("text", c.model, requestID, resp.StatusCode, started, parsed, text, nil)
	c.recordUsage(ctx, resp.StatusCode, parsed.Usage.InputTokens, parsed.Usage.OutputTokens, parsed.Usage.TotalTokens, true, "")
	return TextResult{
		Text:              text,
		InputTokens:       parsed.Usage.InputTokens,
		OutputTokens:      parsed.Usage.OutputTokens,
		TotalTokens:       parsed.Usage.TotalTokens,
		UpstreamRequestID: requestID,
	}, nil
}

func (c *Client) recordUsage(ctx context.Context, status, input, output, total int, success bool, code string) {
	usagepkg.SetTrace(ctx, "provider", "OPENAI")
	usagepkg.SetTrace(ctx, "operation", "RESPONSES")
	usagepkg.SetTrace(ctx, "model", c.model)
	usagepkg.SetTrace(ctx, "provider_attempted", true)
	usagepkg.SetTrace(ctx, "provider_success", success)
	usagepkg.SetTrace(ctx, "provider_status", status)
	usagepkg.SetTrace(ctx, "provider_hit_count", 1)
	if c.recorder == nil {
		return
	}
	id, ok := usagepkg.UserID(ctx)
	if !ok {
		return
	}
	c.recorder.Record(ctx, usagepkg.Event{UserID: id, RequestID: usagepkg.RequestID(ctx), Provider: "OPENAI", Feature: usagepkg.Feature(ctx), Operation: "RESPONSES", APIOrModel: c.model, InputTokens: input, OutputTokens: output, TotalTokens: total, HTTPStatus: status, Success: success, ErrorCode: code})
}

// GenerateWebSearch performs one Responses API request with the hosted web
// search tool. Sources are taken from the tool output, not inferred from text.
func (c *Client) GenerateWebSearch(ctx context.Context, instructions, input string) (WebSearchResult, error) {
	return c.GenerateWebSearchWithMaxOutputTokens(ctx, instructions, input, c.maxOutputTokens)
}

func (c *Client) GenerateWebSearchWithMaxOutputTokens(ctx context.Context, instructions, input string, maxOutputTokens int) (WebSearchResult, error) {
	if !c.Configured() {
		return WebSearchResult{}, ErrAINotConfigured
	}
	if maxOutputTokens <= 0 {
		maxOutputTokens = c.maxOutputTokens
	}
	body := requestBody{
		Model: c.model, Store: false, MaxOutputTokens: maxOutputTokens,
		Instructions: strings.TrimSpace(instructions), Input: strings.TrimSpace(input),
		Tools:   []map[string]any{{"type": "web_search"}},
		Include: []string{"web_search_call.action.sources"},
	}
	result, parsed, err := c.doRequestParsed(ctx, body, c.findMenuTimeout)
	if err != nil {
		return WebSearchResult{TextResult: result}, err
	}
	return WebSearchResult{
		TextResult:       result,
		Sources:          parsed.webSources(),
		ResponseStatus:   parsed.Status,
		IncompleteReason: parsed.IncompleteDetails.Reason,
	}, nil
}

// GenerateMultimodal sends one Responses API request containing bounded image inputs.
// Images are supplied by the backend and encoded as data URLs; no API credential is exposed.
func (c *Client) GenerateMultimodal(ctx context.Context, instructions, input string, images []ImageInput) (TextResult, error) {
	if !c.Configured() || len(images) == 0 || len(images) > 3 {
		return TextResult{}, ErrAINotConfigured
	}
	content := []map[string]any{{"type": "input_text", "text": strings.TrimSpace(input)}}
	for _, image := range images {
		if len(image.Bytes) == 0 || strings.TrimSpace(image.ContentType) == "" {
			continue
		}
		content = append(content, map[string]any{"type": "input_image", "image_url": "data:" + image.ContentType + ";base64," + base64.StdEncoding.EncodeToString(image.Bytes)})
	}
	if len(content) == 1 {
		return TextResult{}, ErrAIInvalidResponse
	}
	body := requestBody{Model: c.model, Store: false, MaxOutputTokens: c.maxOutputTokens, Instructions: strings.TrimSpace(instructions), Input: []map[string]any{{"role": "user", "content": content}}}
	return c.doRequest(ctx, body)
}

func (c *Client) doRequest(ctx context.Context, body requestBody) (TextResult, error) {
	result, _, err := c.doRequestParsed(ctx, body, c.timeout)
	return result, err
}

func (c *Client) doRequestParsed(ctx context.Context, body requestBody, requestTimeout time.Duration) (TextResult, responseBody, error) {
	started := time.Now()
	if _, err := url.ParseRequestURI(c.endpoint); err != nil {
		return TextResult{}, responseBody{}, ErrAIInvalidResponse
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return TextResult{}, responseBody{}, ErrAIInvalidResponse
	}
	if requestTimeout <= 0 {
		requestTimeout = c.timeout
	}
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return TextResult{}, responseBody{}, ErrAIInvalidResponse
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.recordUsage(ctx, 0, 0, 0, 0, false, "request_error")
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return TextResult{}, responseBody{}, ErrAITimeout
		}
		return TextResult{}, responseBody{}, ErrAIUnavailable
	}
	defer resp.Body.Close()
	requestID := resp.Header.Get("x-request-id")
	if resp.StatusCode == http.StatusTooManyRequests {
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "rate_limited")
		logUpstreamFailure(resp, requestID)
		return TextResult{UpstreamRequestID: requestID}, responseBody{}, ErrAIRateLimited
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "authentication")
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, responseBody{}, ErrAIAuthentication
	}
	if resp.StatusCode >= 500 {
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "upstream")
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, responseBody{}, ErrAIUnavailable
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "rejected")
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, responseBody{}, ErrAIRequestRejected
	}
	var parsed responseBody
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&parsed); err != nil {
		c.recordUsage(ctx, resp.StatusCode, 0, 0, 0, false, "invalid_response")
		return TextResult{UpstreamRequestID: requestID}, responseBody{}, ErrAIInvalidResponse
	}
	text := strings.TrimSpace(parsed.OutputText)
	if text == "" {
		text = strings.TrimSpace(parsed.firstText())
	}
	if text == "" {
		logResponseMetadata("web_search_or_multimodal", c.model, requestID, resp.StatusCode, started, parsed, text, ErrAIInvalidResponse)
		c.recordUsage(ctx, resp.StatusCode, parsed.Usage.InputTokens, parsed.Usage.OutputTokens, parsed.Usage.TotalTokens, false, "empty_response")
		return TextResult{UpstreamRequestID: requestID}, parsed, ErrAIInvalidResponse
	}
	if requestID == "" {
		requestID = parsed.ID
	}
	logResponseMetadata("web_search_or_multimodal", c.model, requestID, resp.StatusCode, started, parsed, text, nil)
	c.recordUsage(ctx, resp.StatusCode, parsed.Usage.InputTokens, parsed.Usage.OutputTokens, parsed.Usage.TotalTokens, true, "")
	return TextResult{Text: text, InputTokens: parsed.Usage.InputTokens, OutputTokens: parsed.Usage.OutputTokens, TotalTokens: parsed.Usage.TotalTokens, UpstreamRequestID: requestID}, parsed, nil
}

type upstreamErrorBody struct {
	Error struct {
		Type string `json:"type"`
		Code string `json:"code"`
	} `json:"error"`
}

func logUpstreamFailure(resp *http.Response, requestID string) {
	var parsed upstreamErrorBody
	if resp.Body != nil {
		data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		if err == nil {
			_ = json.Unmarshal(data, &parsed)
		}
	}
	slog.Warn("OpenAI upstream request failed",
		"status", resp.StatusCode,
		"error_type", parsed.Error.Type,
		"error_code", parsed.Error.Code,
		"request_id", requestID,
		"ratelimit_limit_requests", resp.Header.Get("x-ratelimit-limit-requests"),
		"ratelimit_remaining_requests", resp.Header.Get("x-ratelimit-remaining-requests"),
		"ratelimit_reset_requests", resp.Header.Get("x-ratelimit-reset-requests"),
		"ratelimit_limit_tokens", resp.Header.Get("x-ratelimit-limit-tokens"),
		"ratelimit_remaining_tokens", resp.Header.Get("x-ratelimit-remaining-tokens"),
		"ratelimit_reset_tokens", resp.Header.Get("x-ratelimit-reset-tokens"),
	)
}

type requestBody struct {
	Model           string           `json:"model"`
	Store           bool             `json:"store"`
	MaxOutputTokens int              `json:"max_output_tokens"`
	Instructions    string           `json:"instructions"`
	Input           any              `json:"input"`
	Tools           []map[string]any `json:"tools,omitempty"`
	Include         []string         `json:"include,omitempty"`
}

type responseBody struct {
	ID                string       `json:"id"`
	Status            string       `json:"status"`
	OutputText        string       `json:"output_text"`
	Output            []outputItem `json:"output"`
	Usage             usage        `json:"usage"`
	IncompleteDetails struct {
		Reason string `json:"reason"`
	} `json:"incomplete_details"`
	Error struct {
		Code string `json:"code"`
		Type string `json:"type"`
	} `json:"error"`
}

type outputItem struct {
	Type    string        `json:"type"`
	Content []contentItem `json:"content"`
	Action  struct {
		Sources []WebSearchSource `json:"sources"`
	} `json:"action"`
}

func (r responseBody) webSources() []WebSearchSource {
	seen := map[string]bool{}
	result := make([]WebSearchSource, 0)
	for _, output := range r.Output {
		for _, source := range output.Action.Sources {
			source.URL = strings.TrimSpace(source.URL)
			if source.URL == "" || seen[source.URL] {
				continue
			}
			seen[source.URL] = true
			result = append(result, source)
		}
		for _, content := range output.Content {
			for _, annotation := range content.Annotations {
				if annotation.Type != "url_citation" {
					continue
				}
				u := strings.TrimSpace(annotation.URL)
				if u == "" || seen[u] {
					continue
				}
				seen[u] = true
				result = append(result, WebSearchSource{Title: annotation.Title, URL: u})
			}
		}
	}
	return result
}

func (r responseBody) toolCallCount() int {
	count := 0
	for _, output := range r.Output {
		if output.Type == "web_search_call" {
			count++
		}
	}
	return count
}

func logResponseMetadata(method, model, requestID string, statusCode int, started time.Time, parsed responseBody, text string, err error) {
	safeCode := ""
	if err != nil {
		safeCode = SafeErrorCode(err)
	}
	slog.Info("OpenAI response metadata",
		"operation", "openai_response", "openai_method", method, "model", model,
		"duration_ms", time.Since(started).Milliseconds(), "http_status", statusCode,
		"openai_request_id", requestID, "response_status", parsed.Status,
		"incomplete_reason", parsed.IncompleteDetails.Reason,
		"output_item_count", len(parsed.Output), "output_text_length", len(text),
		"tool_call_count", parsed.toolCallCount(), "input_tokens", parsed.Usage.InputTokens,
		"output_tokens", parsed.Usage.OutputTokens, "safe_error_code", safeCode,
	)
}

type contentItem struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []struct {
		Type  string `json:"type"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"annotations"`
}

type usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func (r responseBody) firstText() string {
	for _, output := range r.Output {
		for _, content := range output.Content {
			if content.Text != "" {
				return content.Text
			}
		}
	}
	return ""
}

func SafeErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrAINotConfigured):
		return "AI_NOT_CONFIGURED"
	case errors.Is(err, ErrAIRateLimited):
		return "AI_RATE_LIMITED"
	case errors.Is(err, ErrAITimeout):
		return "AI_TIMEOUT"
	case errors.Is(err, ErrAIInvalidResponse):
		return "AI_INVALID_RESPONSE"
	case errors.Is(err, ErrAIUnavailable):
		return "AI_UNAVAILABLE"
	case errors.Is(err, ErrAIAuthentication):
		return "AI_AUTHENTICATION_FAILED"
	case errors.Is(err, ErrAIRequestRejected):
		return "AI_REQUEST_REJECTED"
	default:
		return "AI_UNAVAILABLE"
	}
}

func SafeError(err error) error {
	return fmt.Errorf("%s", SafeErrorCode(err))
}
