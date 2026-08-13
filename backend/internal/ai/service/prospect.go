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

const maxGoogleReviewsForAI = 5
const maxGoogleReviewTextLength = 500

var skillInstructions = map[string]string{
	"PROSPECT_ANALYSIS":      "For PROSPECT_ANALYSIS, assess potential and fit from evidence in the CRM context, and explicitly call out missing information.",
	"QUALIFICATION_ANALYSIS": "For QUALIFICATION_ANALYSIS, evaluate need, interest, authority, timing, and budget/commercial sensitivity. If there is no evidence for any field, use UNKNOWN.",
	"NEXT_BEST_ACTION":       "For NEXT_BEST_ACTION, recommend the smallest realistic next step based on current status, status history, visits, and discussion.",
	"VISIT_PREPARATION":      "For VISIT_PREPARATION, provide visit objectives, talking points, and priority questions based on the prospect, previous visits, and discussion.",
	"DISCOVERY_QUESTIONS":    "For DISCOVERY_QUESTIONS, prioritize questions that close the most important missing information.",
	"OBJECTION_HANDLING":     "For OBJECTION_HANDLING, diagnose the likely objection or root cause from evidence first. Do not automatically recommend discounts.",
	"SALES_PITCH":            "For SALES_PITCH, personalize the pitch from actual prospect data. Avoid generic claims and do not invent product claims.",
	"FOLLOW_UP":              "For FOLLOW_UP, use the latest discussion, visit notes, visit outcome, and status history to propose a relevant follow-up.",
	"DEAL_RISK_ANALYSIS":     "For DEAL_RISK_ANALYSIS, identify blockers and risks based on evidence and missing information.",
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
	return s.client.GenerateText(ctx, "Return ONLY valid JSON matching: {summary:string,potential:HIGH|MEDIUM|LOW|UNKNOWN,customerProfile:string,yoghurtOpportunity:string,buyingSignals:string[],qualification:{need:string,interest:string,authority:string,timing:string},dealRisks:string[],missingInformation:string[],nextBestAction:{action:string,reason:string},recommendedQuestions:string[]}. Use only supplied CRM facts; unknown values must be UNKNOWN or Data belum tersedia. Do not invent products, SKUs, prices, flavours, or customer facts.", s.contextJSON(review, details, comments, nil))
}

func (s *ProspectAI) ProfileProspectMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, menuImages []prospectmodel.MenuImage) (TextResult, error) {
	payload := s.contextJSON(review, details, nil, map[string]any{"menuImages": menuImages})
	return s.client.GenerateText(ctx, "Profile likely menu positioning using only provided CRM and menu facts. Say when menu data is unavailable.", payload)
}

func (s *ProspectAI) AskProspectAI(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment, history []ChatMessage, question string) (TextResult, error) {
	return s.AskProspectAIWithSkill(ctx, review, details, comments, history, question, "AUTO", nil)
}

func (s *ProspectAI) AskProspectAIWithSkill(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment, history []ChatMessage, question, skill string, saved map[string]any) (TextResult, error) {
	question = boundString(question, s.maxChatLength)
	history = boundHistory(history, s.maxHistory, s.maxChatLength)
	extra := map[string]any{"history": history, "question": question, "skill": skill, "savedAIAnalysis": saved}
	return s.client.GenerateText(ctx, salesCopilotInstructions(skill), s.contextJSON(review, details, comments, extra))
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
		if reviews := boundedGoogleReviews(details.Reviews); len(reviews) > 0 {
			payload["googlePlaceReviews"] = reviews
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

func salesCopilotInstructions(skill string) string {
	base := "Return ONLY valid JSON matching {answer:string,skill:string,insight:string,why:string,recommendedAction:string}. Answer in Indonesian, concise and actionable for sales. Use only supplied CRM data. Missing data must be 'Data belum tersedia'. Do not assume buying intent, authority, budget, or timing. Yoghurt is only a general category; do not invent SKU, brand, flavour, specification, or price. Consider the current funnel stage and recommend a realistic next step."
	if instruction := skillInstructions[strings.ToUpper(strings.TrimSpace(skill))]; instruction != "" {
		return base + " " + instruction
	}
	return base
}

func boundedGoogleReviews(reviews []prospectmodel.PlaceReview) []map[string]any {
	bounded := make([]map[string]any, 0, maxGoogleReviewsForAI)
	for _, review := range reviews {
		text := boundString(review.Text, maxGoogleReviewTextLength)
		if text == "" {
			continue
		}
		bounded = append(bounded, map[string]any{
			"authorName":   review.AuthorName,
			"rating":       review.Rating,
			"text":         text,
			"time":         review.Time,
			"languageCode": review.LanguageCode,
		})
		if len(bounded) >= maxGoogleReviewsForAI {
			break
		}
	}
	return bounded
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
