package usage

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectAggregateAndContributionsAcrossUsers(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	events := make([]Event, 0, 23)
	for i := 0; i < 20; i++ { events = append(events, Event{UserID: a, Provider: "GOOGLE_MAPS", Operation: "NEARBY_SEARCH", APIOrModel: "Places API (New)", SKUCategory: "NEARBY_SEARCH", RequestCount: 1}) }
	for i := 0; i < 30; i++ { events = append(events, Event{UserID: b, Provider: "GOOGLE_MAPS", Operation: "NEARBY_SEARCH", APIOrModel: "Places API (New)", SKUCategory: "NEARBY_SEARCH", RequestCount: 1}) }
	groups, users := aggregateMemory(events)
	if len(groups) != 1 || groups[0].Requests != 50 { t.Fatalf("project aggregate=%+v, want one group with 50 requests", groups) }
	for _, user := range users { want := 40.0; if user.UserID == b.String() { want = 60 }; if user.ContributionPercent != want { t.Fatalf("user %s contribution=%v, want %v", user.UserID, user.ContributionPercent, want) } }
}

func TestProjectAggregateOpenAITokensAndHistoricalNulls(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	groups, _ := aggregateMemory([]Event{
		{UserID: a, Provider: "OPENAI", Operation: "RESPONSES", APIOrModel: "gpt-test", RequestCount: 2, InputTokens: 100, OutputTokens: 50, TotalTokens: 150, CachedTokens: 40},
		{UserID: b, Provider: "OPENAI", Operation: "RESPONSES", APIOrModel: "gpt-test", RequestCount: 1, InputTokens: 50, OutputTokens: 25, TotalTokens: 75, CachedTokens: 10},
		{UserID: a, Provider: "OPENAI", Operation: "RESPONSES", APIOrModel: "gpt-legacy", RequestCount: 1},
	})
	var found bool
	for _, group := range groups { if group.APIOrModel == "gpt-test" { found = true; if group.Requests != 3 || group.InputTokens != 150 || group.OutputTokens != 75 || group.CachedTokens != 50 { t.Fatalf("OpenAI aggregate=%+v", group) } } }
	if !found { t.Fatal("OpenAI aggregate group missing") }
}
