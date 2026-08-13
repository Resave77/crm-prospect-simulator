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

const (
	summaryMaxOutputTokens      = 1600
	menuFinderMaxOutputTokens   = 4000
	menuRecoveryMaxOutputTokens = 4000
	menuProfileMaxOutputTokens  = 2000
)

const maxGoogleReviewsForAI = 5
const maxGoogleReviewTextLength = 500

const menuProfilingInstructions = `Anda adalah AI Sales Copilot di Yummy CRM. Analisis data menu prospect F&B yang sudah ditemukan dan tersimpan untuk menghasilkan insight penjualan yang akurat, ringkas, profesional, dan dapat digunakan Sales Executive.

Kembalikan HANYA satu objek JSON valid tanpa Markdown atau teks tambahan dengan struktur tepat: {menuOpportunity:HIGH|MEDIUM|LOW|UNKNOWN,yoghurtFit:HIGH|MEDIUM|LOW|UNKNOWN,confidence:HIGH|MEDIUM|LOW|UNKNOWN,topOpportunity:string,why:string,recommendedAction:string}.

ATURAN WAJIB:
- Semua isi topOpportunity, why, dan recommendedAction wajib menggunakan Bahasa Indonesia yang natural, profesional, ringkas, jelas, dan berorientasi pada kebutuhan penjualan. Jangan menghasilkan kalimat atau penjelasan dalam Bahasa Inggris.
- Sebelum menghasilkan JSON final, periksa ulang ketiga field deskriptif tersebut. Jika masih berbahasa Inggris, tulis ulang ke Bahasa Indonesia.
- Jangan menerjemahkan nama property JSON atau enum. Enum internal harus tetap LOW, MEDIUM, HIGH, atau UNKNOWN sesuai schema.
- Nama brand, outlet, produk, menu, platform, layanan, URL, tempat, dan proper noun boleh dipertahankan dalam bahasa aslinya. Jangan menerjemahkan nama menu secara paksa.
- Gunakan hanya stored menu discovery dan konteks prospect yang diberikan. Jangan melakukan web search atau pencarian menu ulang.
- Jangan mengarang menu, harga, promo, transaksi, pelanggan, preferensi, omzet, margin, stok, supplier, kebutuhan prospect, atau fakta lain yang tidak tersedia.
- Bedakan fakta dari hipotesis penjualan. Jangan menyatakan peluang sebagai kepastian. Turunkan confidence jika data terbatas atau kecocokan cabang/sumber kurang kuat.
- menuOpportunity dan yoghurtFit harus mencerminkan bukti secara bertanggung jawab; jangan selalu memilih HIGH dan jangan memaksakan kecocokan yoghurt.
- topOpportunity berisi satu peluang paling relevan dan realistis. why menjelaskan bukti menu dalam 2-4 kalimat ringkas. recommendedAction berisi langkah berikutnya yang konkret untuk Sales Executive.
- Jangan mengulang seluruh daftar menu dan jangan membuat rekomendasi harga tanpa data pendukung.`

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
	return s.client.GenerateTextWithMaxOutputTokens(ctx, "Return ONLY valid JSON matching: {summary:string,potential:HIGH|MEDIUM|LOW|UNKNOWN,customerProfile:string,yoghurtOpportunity:string,buyingSignals:string[],qualification:{need:string,interest:string,authority:string,timing:string},dealRisks:string[],missingInformation:string[],nextBestAction:{action:string,reason:string},recommendedQuestions:string[]}. Use only supplied CRM facts; unknown values must be UNKNOWN or Data belum tersedia. Do not invent products, SKUs, prices, flavours, or customer facts.", s.contextJSON(review, details, comments, nil), summaryMaxOutputTokens)
}

func (s *ProspectAI) RecoverProspectSummary(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, comments []prospectmodel.ProspectComment) (TextResult, error) {
	instructions := "Return ONLY one valid JSON object matching exactly: {summary:string,potential:HIGH|MEDIUM|LOW|UNKNOWN,customerProfile:string,yoghurtOpportunity:string,buyingSignals:string[],qualification:{need:string,interest:string,authority:string,timing:string},dealRisks:string[],missingInformation:string[],nextBestAction:{action:string,reason:string},recommendedQuestions:string[]}. No markdown, code fence, or prose before or after JSON. All required fields must be present. Use UNKNOWN, empty arrays, or Data belum tersedia for unavailable facts. Do not invent prospect facts, products, SKUs, prices, flavours, or customer facts."
	return s.client.GenerateTextWithMaxOutputTokens(ctx, instructions, s.contextJSON(review, details, comments, nil), summaryMaxOutputTokens)
}

func (s *ProspectAI) ProfileProspectMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, menuImages []prospectmodel.MenuImage) (TextResult, error) {
	payload := s.contextJSON(review, details, nil, map[string]any{"menuImages": menuImages})
	return s.client.GenerateText(ctx, "Return ONLY valid JSON matching: {menus:[{menuName:string,profile:string,yoghurtFit:HIGH|MEDIUM|LOW|UNKNOWN,opportunity:string,reason:string,recommendedSalesAction:string,confidence:string}],topOpportunity:string,recommendedAction:string}. Use only supplied menu/CRM facts. Do not invent brands, SKUs, flavours, or prices.", payload)
}

func (s *ProspectAI) ProfileProspectMenuVision(ctx context.Context, payload string, images []ImageInput, sourceNames []string) (TextResult, error) {
	instructions := "Return ONLY valid JSON matching {menus:[{menuName:string,profile:string,yoghurtFit:HIGH|MEDIUM|LOW|UNKNOWN,opportunity:string,reason:string,recommendedSalesAction:string,confidence:HIGH|MEDIUM|LOW}],topOpportunity:string,recommendedAction:string,sourcePhotoNames:string[]}. Extract only menu items clearly readable in the supplied images. Do not invent unreadable items, brands, SKUs, flavours, or prices. Treat yoghurt only as a general category. sourcePhotoNames must match the supplied source names."
	return s.client.GenerateMultimodal(ctx, instructions, payload, images)
}

func (s *ProspectAI) ProfileProspectStructuredMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails, finding json.RawMessage) (TextResult, error) {
	payload := s.contextJSON(review, details, nil, map[string]any{"menuFinding": json.RawMessage(finding)})
	return s.client.GenerateTextWithOptions(ctx, menuProfilingInstructions, payload, menuProfileMaxOutputTokens, s.client.menuProfileTimeout)
}

func (s *ProspectAI) FindProspectMenu(ctx context.Context, review prospectmodel.Review, details *prospectmodel.PlaceDetails) (WebSearchResult, error) {
	payload := s.contextJSON(review, details, nil, map[string]any{
		"task":           "Find publicly available menu evidence for this exact prospect branch.",
		"searchPriority": []string{"official website/menu", "official ordering page", "delivery platforms", "official social media", "other public menu pages"},
	})
	instructions := `STRICT OUTPUT LIMITS — apply these before doing detailed synthesis: return at most 5 source-defined categories, at most 20 menu items TOTAL, and at most 2 price/source evidence entries per item. Stop after maximum 20 menu items total. Do not attempt to return the full merchant catalog. Return only a representative subset sufficient for CRM sales analysis. Prefer representative best-known/current items across the most relevant merchant/source categories. Where practical, use only 4–5 items per category so one category does not consume the result. Never invent categories.

Use web search to find verifiable menu data for the supplied prospect. Prefer the exact branch and exact address; never merge an uncertain branch silently. Return ONLY valid JSON matching {status:"FOUND"|"NOT_FOUND",message:string,place:{name:string,branch:string,address:string,googlePlaceId:string,googleMapsUrl:string},sources:[{name:string,url:string,branchMatch:"exact_branch"|"likely_same_branch"|"brand_only"|"uncertain"}],categories:[{name:string,items:[{name:string,description:string|null,imageUrl:string|null,prices:[{source:string,sourceUrl:string,price:number,currency:string}],branchMatch:"exact_branch"|"likely_same_branch"|"brand_only"|"uncertain",confidence:number}]}],summary:{sourceCount:number,totalCategories:number,totalItems:number}}. Keep every item compact. Use a short useful description only when grounded; otherwise description must be null. Include imageUrl only when grounded and available; otherwise null. Do not repeat source narratives in item descriptions. Top-level sources are authoritative; per-price evidence must contain only the minimal existing schema reference. Every sourceUrl must be a URL actually opened by web search. Keep separate prices from separate sources. Exclude uncertain-branch items from categories. If reliable menu evidence is unavailable, return NOT_FOUND with empty categories. Do not invent menu items, prices, URLs, images, branches, or availability.

Returning fewer valid items is better than returning more items with an incomplete JSON response. Before finalizing, ensure the JSON object is complete and closed. If the output budget may be exceeded, reduce item count rather than truncating the JSON. Accuracy and valid JSON are more important than catalog completeness.`
	return s.client.GenerateWebSearchWithMaxOutputTokens(ctx, instructions, payload, menuFinderMaxOutputTokens)
}

// RecoverProspectMenuFormatting repairs only the already-grounded search output.
// It deliberately uses the text-only endpoint, so a Find Menu click can never
// trigger a second hosted web search.
func (s *ProspectAI) RecoverProspectMenuFormatting(ctx context.Context, review prospectmodel.Review, search WebSearchResult) (TextResult, error) {
	payload := s.contextJSON(review, nil, nil, map[string]any{
		"task":            "Repair the supplied grounded menu facts into valid JSON without adding facts.",
		"groundedOutput":  search.Text,
		"groundedSources": search.Sources,
	})
	instructions := `Return ONLY one complete, closed JSON object using exactly this schema: {status:"FOUND"|"NOT_FOUND",message:string,place:{name:string,branch:string,address:string,googlePlaceId:string,googleMapsUrl:string},sources:[{name:string,url:string,branchMatch:"exact_branch"|"likely_same_branch"|"brand_only"|"uncertain"}],categories:[{name:string,items:[{name:string,description:string|null,imageUrl:string|null,prices:[{source:string,sourceUrl:string,price:number,currency:string}],branchMatch:"exact_branch"|"likely_same_branch"|"brand_only"|"uncertain",confidence:number}]}],summary:{sourceCount:number,totalCategories:number,totalItems:number}}. This is formatting recovery only: reconstruct a smaller valid subset exclusively from the supplied prospect identity, grounded output, and groundedSources. Do not append blindly to truncated JSON. Do not search, infer, complete a catalog, or add any menu, price, URL, image, branch, or category fact. Use only URLs present in groundedSources. Return at most 5 categories, 20 items TOTAL, 4–5 items per category where practical, and 2 price evidence entries per item. Keep descriptions short and grounded or null. Preserve any valid partial subset and exclude uncertain items. If no valid grounded menu item remains, return NOT_FOUND with empty categories. Fewer grounded items and valid JSON are more important than quantity.`
	return s.client.GenerateTextWithOptions(ctx, instructions, payload, menuRecoveryMaxOutputTokens, s.client.findMenuTimeout)
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
			"googlePlaceId":     prospect.GooglePlaceID,
			"googleMapsUrl":     prospect.GoogleMapsURL,
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
