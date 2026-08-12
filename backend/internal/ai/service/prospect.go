package service

import (
	"context"
	"encoding/json"
	"strings"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
)

type ProspectAI struct {
	client        *Client
	maxChatLength int
	maxHistory    int
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewProspectAI(client *Client, maxChatLength, maxHistory int) *ProspectAI {
	if maxChatLength <= 0 {
		maxChatLength = 1000
	}
	if maxHistory <= 0 {
		maxHistory = 6
	}
	return &ProspectAI{client: client, maxChatLength: maxChatLength, maxHistory: maxHistory}
}

func (s *ProspectAI) SummarizeProspect(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment) (TextResult, error) {
	return s.client.GenerateText(ctx, "Summarize this CRM prospect using only the provided facts. Do not invent missing data.", s.contextJSON(review, details, comments, nil))
}

func (s *ProspectAI) ProfileProspectMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, menuImages []prospectmodel.MenuImage) (TextResult, error) {
	payload := s.contextJSON(review, details, nil, map[string]any{"menuImages": menuImages})
	return s.client.GenerateText(ctx, "Profile likely menu positioning using only provided CRM and menu facts. Say when menu data is unavailable.", payload)
}

func (s *ProspectAI) AskProspectAI(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment, history []ChatMessage, question string) (TextResult, error) {
	question = boundString(question, s.maxChatLength)
	history = boundHistory(history, s.maxHistory, s.maxChatLength)
	extra := map[string]any{"history": history, "question": question}
	return s.client.GenerateText(ctx, "Answer the user's prospect question using only the provided CRM context. Be concise and state when data is missing.", s.contextJSON(review, details, comments, extra))
}

func (s *ProspectAI) contextJSON(review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment, extra map[string]any) string {
	prospect := review.Prospect
	payload := map[string]any{
		"prospect": map[string]any{
			"businessName":      prospect.PlaceName,
			"category":          prospect.PlaceCategory,
			"industryGroup":     prospect.IndustryGroup,
			"pipelineStatus":    prospect.Status,
			"address":           prospect.FormattedAddress,
			"phone":             prospect.PhoneNumber,
			"website":           prospect.WebsiteURL,
			"assignedSales":     prospect.AssignedSalesExecutive,
			"visitNotes":        prospect.VisitNotes,
			"followUpNotes":     prospect.FollowUpNotes,
			"placeTypes":        prospect.PlaceTypes,
			"deletionRequested": prospect.DeletionRequested,
		},
		"recentStatusHistory": review.History,
		"recentVisits":        review.Visits,
	}
	if details != nil {
		payload["googlePlaceDetails"] = map[string]any{
			"rating":           details.Rating,
			"reviewCount":      details.UserRatingCount,
			"businessStatus":   details.BusinessStatus,
			"openingHours":     details.OpeningHours,
			"delivery":         details.Delivery,
			"dineIn":           details.DineIn,
			"takeout":          details.Takeout,
			"priceLevel":       details.PriceLevel,
			"editorialSummary": details.EditorialSummary,
		}
	}
	if len(comments) > 0 {
		payload["recentComments"] = comments
	}
	for key, value := range extra {
		payload[key] = value
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func boundHistory(history []ChatMessage, maxHistory, maxLength int) []ChatMessage {
	if len(history) > maxHistory {
		history = history[len(history)-maxHistory:]
	}
	bounded := make([]ChatMessage, 0, len(history))
	for _, item := range history {
		role := strings.ToLower(strings.TrimSpace(item.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		bounded = append(bounded, ChatMessage{Role: role, Content: boundString(item.Content, maxLength)})
	}
	return bounded
}

func boundString(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	if maxLength <= 0 || len(value) <= maxLength {
		return value
	}
	return value[:maxLength]
}
