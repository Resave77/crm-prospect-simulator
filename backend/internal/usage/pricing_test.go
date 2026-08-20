package usage

import "testing"

func TestGoogleCostRequiresVerifiedConfiguration(t *testing.T) {
	cost := DefaultPricing().GoogleCost("PLACES_SEARCH", 7, 7)
	if cost.Configured || cost.BillingStatus != "UNCONFIGURED" || cost.EstimatedPayableCost != 0 {
		t.Fatalf("unexpected unconfigured cost: %+v", cost)
	}
}

func TestGoogleCostAppliesSharedProjectFreeTier(t *testing.T) {
	p := DefaultPricing()
	p.GoogleSKUs["NEARBY_SEARCH"] = GoogleSKU{PriceMicrosPer1000: 2500000, FreeMonthly: 10, Verified: true}
	below := p.GoogleCost("NEARBY_SEARCH", 3, 8)
	if !below.Configured || below.GrossCost != 0.0075 || below.EstimatedPayableCost != 0 {
		t.Fatalf("unexpected below-free-tier cost: %+v", below)
	}
	above := p.GoogleCost("NEARBY_SEARCH", 4, 12)
	if above.EstimatedPayableCost != 0.005 || above.FreeUsage != 10 {
		t.Fatalf("unexpected shared-tier cost: %+v", above)
	}
}

func TestGoogleCostCacheHitHasNoBillableUnits(t *testing.T) {
	p := DefaultPricing()
	p.GoogleSKUs["NEARBY_SEARCH"] = GoogleSKU{PriceMicrosPer1000: 2500000, FreeMonthly: 10, Verified: true}
	cost := p.GoogleCost("NEARBY_SEARCH", 0, 12)
	if cost.GrossCost != 0 || cost.EstimatedPayableCost != 0 {
		t.Fatalf("cache hit should add no cost: %+v", cost)
	}
}
