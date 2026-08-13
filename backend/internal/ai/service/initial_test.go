package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	prospectmodel "crm-prospect-simulator/backend/internal/prospect/model"
)

const validSummaryFixture = `{"summary":"Ready","potential":"HIGH","customerProfile":"Profile","yoghurtOpportunity":"UNKNOWN","buyingSignals":[],"qualification":{"need":"UNKNOWN","interest":"UNKNOWN","authority":"UNKNOWN","timing":"UNKNOWN"},"dealRisks":[],"missingInformation":[],"nextBestAction":{"action":"Contact","reason":"Validate"},"recommendedQuestions":[]}`

func TestSummaryGenerationSucceedsOnFirstValidAttempt(t *testing.T) {
	calls := 0
	result, attempts, err := generateSummaryWithRecovery(context.Background(), func(context.Context, bool) (TextResult, error) {
		calls++
		return TextResult{Text: validSummaryFixture}, nil
	}, nil)
	if err != nil || calls != 1 || attempts != 1 || !validSummaryJSON(result.Text) {
		t.Fatalf("calls=%d attempts=%d err=%v", calls, attempts, err)
	}
}

func TestSummaryGenerationRecoversOnceFromInvalidJSON(t *testing.T) {
	calls := 0
	_, attempts, err := generateSummaryWithRecovery(context.Background(), func(_ context.Context, recovery bool) (TextResult, error) {
		calls++
		if !recovery {
			return TextResult{Text: "not-json"}, nil
		}
		return TextResult{Text: validSummaryFixture}, nil
	}, nil)
	if err != nil || calls != 2 || attempts != 2 {
		t.Fatalf("calls=%d attempts=%d err=%v", calls, attempts, err)
	}
}

func TestSummaryGenerationStopsAfterTwoInvalidAttempts(t *testing.T) {
	calls := 0
	_, attempts, err := generateSummaryWithRecovery(context.Background(), func(context.Context, bool) (TextResult, error) {
		calls++
		return TextResult{Text: `{}`}, nil
	}, nil)
	if !errors.Is(err, ErrAIInvalidResponse) || calls != 2 || attempts != 2 {
		t.Fatalf("calls=%d attempts=%d err=%v", calls, attempts, err)
	}
}

func TestSummaryGenerationDoesNotRetryNonRetryableFailure(t *testing.T) {
	calls := 0
	_, attempts, err := generateSummaryWithRecovery(context.Background(), func(context.Context, bool) (TextResult, error) {
		calls++
		return TextResult{}, ErrAIAuthentication
	}, nil)
	if !errors.Is(err, ErrAIAuthentication) || calls != 1 || attempts != 1 {
		t.Fatalf("calls=%d attempts=%d err=%v", calls, attempts, err)
	}
}

func TestSummarySchemaRequiresNestedFields(t *testing.T) {
	if validSummaryJSON(`{"summary":"partial"}`) {
		t.Fatal("partial summary must not validate")
	}
	if !validSummaryJSON(validSummaryFixture) {
		t.Fatal("complete summary must validate")
	}
}

func TestStructuredOperationsHaveBoundedLargerBudgets(t *testing.T) {
	if summaryMaxOutputTokens <= 800 {
		t.Fatalf("summary budget=%d", summaryMaxOutputTokens)
	}
	if menuFinderMaxOutputTokens <= summaryMaxOutputTokens {
		t.Fatalf("menu budget=%d summary=%d", menuFinderMaxOutputTokens, summaryMaxOutputTokens)
	}
}

func TestValidMenuProfilingRequiresCompactAggregateSchema(t *testing.T) {
	valid := json.RawMessage(`{"menuOpportunity":"HIGH","yoghurtFit":"MEDIUM","topOpportunity":"Dessert pairing","why":"Menu evidence supports a general opportunity.","recommendedAction":"Validate customer needs.","confidence":"MEDIUM"}`)
	if !validMenuProfiling(valid) {
		t.Fatal("valid aggregate profiling rejected")
	}
	if validMenuProfiling(json.RawMessage(`{"topOpportunity":"partial"}`)) {
		t.Fatal("partial profiling accepted")
	}
}

func TestMenuProfilingPromptRequiresIndonesianWithoutChangingContract(t *testing.T) {
	for _, required := range []string{
		"Bahasa Indonesia yang natural, profesional, ringkas",
		"Jangan menghasilkan kalimat atau penjelasan dalam Bahasa Inggris",
		"Jangan menerjemahkan nama property JSON atau enum",
		"menuOpportunity:HIGH|MEDIUM|LOW|UNKNOWN",
		"yoghurtFit:HIGH|MEDIUM|LOW|UNKNOWN",
		"topOpportunity:string",
		"why:string",
		"recommendedAction:string",
		"confidence:HIGH|MEDIUM|LOW|UNKNOWN",
	} {
		if !strings.Contains(menuProfilingInstructions, required) {
			t.Fatalf("profiling instructions missing %q", required)
		}
	}
}

func TestMenuProfilingCacheIsBypassedOnlyForExplicitForce(t *testing.T) {
	saved := json.RawMessage(`{"menuOpportunity":"HIGH","yoghurtFit":"MEDIUM","topOpportunity":"Saved","why":"Saved reason","recommendedAction":"Saved action","confidence":"MEDIUM"}`)
	if !shouldUseCachedMenuProfiling(saved, false) {
		t.Fatal("normal profiling request must reuse valid persisted profiling")
	}
	if shouldUseCachedMenuProfiling(saved, true) {
		t.Fatal("explicit re-analysis must bypass valid persisted profiling")
	}
}

func TestGeneratedMenuProfilingRequiresIndonesianDescriptions(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		valid bool
	}{
		{
			name:  "Indonesian output",
			raw:   `{"menuOpportunity":"MEDIUM","yoghurtFit":"MEDIUM","confidence":"MEDIUM","topOpportunity":"Evaluasi peluang menambahkan yoghurt sebagai pilihan menu tambahan.","why":"Menu prospect didominasi ayam dan minuman. Peluang ini masih perlu divalidasi dengan prospect.","recommendedAction":"Diskusikan kemungkinan uji coba yoghurt dalam skala kecil dengan prospect."}`,
			valid: true,
		},
		{
			name:  "English output",
			raw:   `{"menuOpportunity":"MEDIUM","yoghurtFit":"MEDIUM","confidence":"MEDIUM","topOpportunity":"Evaluate yoghurt as an additional menu option.","why":"The branch is operational and offers several food products.","recommendedAction":"Discuss a small yoghurt trial with the prospect."}`,
			valid: false,
		},
		{
			name:  "Indonesian with proper nouns",
			raw:   `{"menuOpportunity":"LOW","yoghurtFit":"MEDIUM","confidence":"HIGH","topOpportunity":"Validasi peluang yoghurt untuk Beef Cheese Burger.","why":"Menu dari GrabFood menunjukkan pilihan makanan berat, tetapi belum ada yoghurt yang terkonfirmasi.","recommendedAction":"Tanyakan kebutuhan minuman pendamping kepada prospect sebelum menawarkan uji coba."}`,
			valid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validIndonesianMenuProfiling(json.RawMessage(tt.raw)); got != tt.valid {
				t.Fatalf("validIndonesianMenuProfiling()=%v want=%v", got, tt.valid)
			}
		})
	}
}

func TestLegacyEnglishProfilingRemainsCacheableWithoutAutomaticRegeneration(t *testing.T) {
	legacy := json.RawMessage(`{"menuOpportunity":"MEDIUM","yoghurtFit":"MEDIUM","confidence":"MEDIUM","topOpportunity":"Evaluate yoghurt as an additional option.","why":"The branch is operational and offers food.","recommendedAction":"Discuss the opportunity with the prospect."}`)
	if !validMenuProfiling(legacy) || !shouldUseCachedMenuProfiling(legacy, false) {
		t.Fatal("legacy schema-valid profiling must remain readable and cacheable")
	}
	if validIndonesianMenuProfiling(legacy) {
		t.Fatal("newly generated English profiling must not pass persistence validation")
	}
}

func TestValidatedMenuFindingKeepsOnlyGroundedNonUncertainItems(t *testing.T) {
	raw := `{"status":"FOUND","place":{},"sources":[{"name":"Official","url":"https://example.com/menu","branchMatch":"exact_branch"},{"name":"Invented","url":"https://invented.invalid/menu","branchMatch":"exact_branch"}],"categories":[{"name":"Drinks","items":[{"name":"Tea","prices":[{"source":"Official dine-in","sourceUrl":"https://example.com/menu","price":12000,"currency":"IDR"},{"source":"Official delivery","sourceUrl":"https://example.com/menu","price":14000,"currency":"IDR"}],"availability":"available","branchMatch":"exact_branch","confidence":0.9},{"name":"Mystery","prices":[],"availability":"unknown","branchMatch":"uncertain","confidence":0.2}]}]}`
	result, err := validatedMenuFinding(raw, []WebSearchSource{{Title: "Official", URL: "https://example.com/menu"}}, prospectmodel.Review{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "FOUND" || len(result.Sources) != 1 || result.Summary.TotalItems != 1 || len(result.Categories[0].Items) != 1 {
		t.Fatalf("result=%+v", result)
	}
	if result.Categories[0].Items[0].Name != "Tea" {
		t.Fatalf("items=%+v", result.Categories[0].Items)
	}
	if len(result.Categories[0].Items[0].Prices) != 2 {
		t.Fatalf("prices=%+v", result.Categories[0].Items[0].Prices)
	}
}

func TestValidatedMenuFindingReturnsNotFoundWithoutGroundedEvidence(t *testing.T) {
	raw := `{"status":"FOUND","sources":[{"name":"Invented","url":"https://invented.invalid/menu","branchMatch":"exact_branch"}],"categories":[{"name":"Food","items":[{"name":"Invented item","prices":[],"branchMatch":"exact_branch","confidence":1}]}]}`
	result, err := validatedMenuFinding(raw, nil, prospectmodel.Review{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "NOT_FOUND" || len(result.Categories) != 0 || result.Summary.TotalItems != 0 {
		t.Fatalf("result=%+v", result)
	}
}

func TestValidatedMenuFindingEnforcesDiscoveryLimitsAndKeepsPartialValidSubset(t *testing.T) {
	if maxMenuCategories != 5 || maxMenuItems != 20 || maxMenuEvidencePerItem != 2 {
		t.Fatalf("limits=%d/%d/%d, want 5/20/2", maxMenuCategories, maxMenuItems, maxMenuEvidencePerItem)
	}
	sourceURL := "https://example.com/menu"
	finding := MenuFindingResult{Status: "FOUND", Sources: []MenuFindingSource{{Name: "Official", URL: sourceURL, BranchMatch: "exact_branch"}}}
	for categoryIndex := 0; categoryIndex < 10; categoryIndex++ {
		category := MenuFindingCategory{Name: fmt.Sprintf("Category %d", categoryIndex)}
		for itemIndex := 0; itemIndex < 7; itemIndex++ {
			item := MenuFindingItem{Name: fmt.Sprintf("Item %d-%d", categoryIndex, itemIndex), BranchMatch: "exact_branch", Confidence: .9}
			for priceIndex := 0; priceIndex < 5; priceIndex++ {
				item.Prices = append(item.Prices, MenuFindingPrice{Source: "Official", SourceURL: sourceURL, Price: float64(10000 + priceIndex), Currency: "IDR"})
			}
			category.Items = append(category.Items, item)
		}
		finding.Categories = append(finding.Categories, category)
	}
	raw, err := json.Marshal(finding)
	if err != nil {
		t.Fatal(err)
	}
	result, err := validatedMenuFinding(string(raw), []WebSearchSource{{Title: "Official", URL: sourceURL}}, prospectmodel.Review{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "FOUND" || len(result.Categories) > maxMenuCategories || result.Summary.TotalItems != maxMenuItems {
		t.Fatalf("categories=%d items=%d status=%s", len(result.Categories), result.Summary.TotalItems, result.Status)
	}
	for _, category := range result.Categories {
		for _, item := range category.Items {
			if len(item.Prices) > maxMenuEvidencePerItem {
				t.Fatalf("prices=%d", len(item.Prices))
			}
		}
	}
}

func TestValidatedMenuFindingTruncatesTwentyFirstItem(t *testing.T) {
	sourceURL := "https://example.com/menu"
	finding := MenuFindingResult{Status: "FOUND", Sources: []MenuFindingSource{{Name: "Official", URL: sourceURL, BranchMatch: "exact_branch"}}}
	category := MenuFindingCategory{Name: "Representative"}
	for index := 0; index < 21; index++ {
		category.Items = append(category.Items, MenuFindingItem{Name: fmt.Sprintf("Item %d", index), BranchMatch: "exact_branch", Confidence: .8})
	}
	finding.Categories = []MenuFindingCategory{category}
	raw, err := json.Marshal(finding)
	if err != nil {
		t.Fatal(err)
	}
	result, err := validatedMenuFinding(string(raw), []WebSearchSource{{Title: "Official", URL: sourceURL}}, prospectmodel.Review{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.TotalItems != 20 || len(result.Categories[0].Items) != 20 {
		t.Fatalf("items=%d category_items=%d, want 20", result.Summary.TotalItems, len(result.Categories[0].Items))
	}
}

func TestUsableMenuDataRejectsGeneralAndTaggedGooglePhotos(t *testing.T) {
	for _, details := range []*prospectmodel.PlaceDetails{
		nil,
		{Photos: []prospectmodel.PlacePhoto{{Name: "places/ChIJ123/photos/general", IsMenu: false}}},
		{Photos: []prospectmodel.PlacePhoto{{Name: "places/ChIJ123/photos/tagged", IsMenu: true}}},
	} {
		if usableMenuData(details) {
			t.Fatalf("usableMenuData(%+v)=true, want false", details)
		}
	}
}

func TestMenuDataUnavailablePayloadIsValidClearState(t *testing.T) {
	if !json.Valid(menuDataNotAvailableJSON) {
		t.Fatal("menuDataNotAvailableJSON must be valid JSON")
	}
	if string(menuDataNotAvailableJSON) == "" || !strings.Contains(string(menuDataNotAvailableJSON), menuDataNotAvailable) {
		t.Fatalf("menu unavailable payload=%s, want clear state", menuDataNotAvailableJSON)
	}
}
