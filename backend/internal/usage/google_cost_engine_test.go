package usage

import (
	"testing"
	"time"
)

func TestAllocateGoogleCostSharesMonthlyFreeTierAndPeriodBoundary(t *testing.T) {
	p := DefaultPricing()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	events := make([]GoogleCostEvent, 0, 5001)
	for i := 0; i < 5001; i++ {
		events = append(events, GoogleCostEvent{ID: string(rune(i)), UserID: "a", Operation: "NEARBY_SEARCH", CreatedAt: start.Add(time.Duration(i) * time.Minute), Success: true})
	}
	rows := AllocateGoogleCost(events, p, start, start.Add(24*time.Hour*31))
	row := rows["NEARBY_SEARCH"]
	if row.SuccessfulRequests != 5001 || row.PaidRequestCount != 1 || row.FreeUsageConsumed != 5000 || row.EstimatedPayableMicros != 32000 {
		t.Fatalf("unexpected allocation: %+v", row)
	}
}

func TestAllocateGoogleCostFailedAndUnknownEvents(t *testing.T) {
	p := DefaultPricing()
	now := time.Now()
	rows := AllocateGoogleCost([]GoogleCostEvent{{ID: "1", UserID: "a", Operation: "NEARBY_SEARCH", CreatedAt: now, Success: false}}, p, now.Add(-time.Hour), now.Add(time.Hour))
	row := rows["NEARBY_SEARCH"]
	if row.TotalRequests != 1 || row.FailedRequests != 1 || row.BillableRequests != 0 || row.EstimatedPayableMicros != 0 {
		t.Fatalf("failed event was billed: %+v", row)
	}
	if len(AllocateGoogleCost([]GoogleCostEvent{{Operation: "UNKNOWN", Success: true}}, p, time.Time{}, time.Time{})) != 0 {
		t.Fatal("unknown SKU was classified")
	}
}
