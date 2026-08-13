package service

import (
	"bytes"
	"context"
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
)

const responsesAPIURL = "https://api.openai.com/v1/responses"

type Client struct {
	apiKey          string
	model           string
	maxOutputTokens int
	timeout         time.Duration
	httpClient      *http.Client
	endpoint        string
}

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

func NewClient(cfg config.Config, opts ...Option) *Client {
	client := &Client{
		apiKey:          strings.TrimSpace(cfg.OpenAIAPIKey),
		model:           strings.TrimSpace(cfg.OpenAIModel),
		maxOutputTokens: cfg.OpenAIMaxTokens,
		timeout:         cfg.OpenAITimeout,
		httpClient:      &http.Client{},
		endpoint:        responsesAPIURL,
	}
	if !cfg.AIConfigured() {
		client.apiKey = ""
		client.model = ""
	}
	if client.timeout <= 0 {
		client.timeout = 20 * time.Second
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
	if !c.Configured() {
		return TextResult{}, ErrAINotConfigured
	}
	if _, err := url.ParseRequestURI(c.endpoint); err != nil {
		return TextResult{}, ErrAIInvalidResponse
	}
	body := requestBody{
		Model:           c.model,
		Store:           false,
		MaxOutputTokens: c.maxOutputTokens,
		Instructions:    strings.TrimSpace(instructions),
		Input:           strings.TrimSpace(input),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return TextResult{}, ErrAIInvalidResponse
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return TextResult{}, ErrAIInvalidResponse
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return TextResult{}, ErrAITimeout
		}
		return TextResult{}, ErrAIUnavailable
	}
	defer resp.Body.Close()

	requestID := resp.Header.Get("x-request-id")
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		logUpstreamFailure(resp, requestID)
		return TextResult{UpstreamRequestID: requestID}, ErrAIRateLimited
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode >= 500:
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, ErrAIUnavailable
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		logUpstreamFailure(resp, requestID)
		io.Copy(io.Discard, resp.Body)
		return TextResult{UpstreamRequestID: requestID}, ErrAIInvalidResponse
	}

	var parsed responseBody
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&parsed); err != nil {
		return TextResult{UpstreamRequestID: requestID}, ErrAIInvalidResponse
	}
	text := strings.TrimSpace(parsed.OutputText)
	if text == "" {
		text = strings.TrimSpace(parsed.firstText())
	}
	if text == "" {
		return TextResult{UpstreamRequestID: requestID}, ErrAIInvalidResponse
	}
	if requestID == "" {
		requestID = parsed.ID
	}
	return TextResult{
		Text:              text,
		InputTokens:       parsed.Usage.InputTokens,
		OutputTokens:      parsed.Usage.OutputTokens,
		TotalTokens:       parsed.Usage.TotalTokens,
		UpstreamRequestID: requestID,
	}, nil
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
	Model           string `json:"model"`
	Store           bool   `json:"store"`
	MaxOutputTokens int    `json:"max_output_tokens"`
	Instructions    string `json:"instructions"`
	Input           string `json:"input"`
}

type responseBody struct {
	ID         string       `json:"id"`
	OutputText string       `json:"output_text"`
	Output     []outputItem `json:"output"`
	Usage      usage        `json:"usage"`
}

type outputItem struct {
	Content []contentItem `json:"content"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
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
	default:
		return "AI_UNAVAILABLE"
	}
}

func SafeError(err error) error {
	return fmt.Errorf("%s", SafeErrorCode(err))
}
