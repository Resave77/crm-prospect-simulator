package usage

import "testing"

func TestOpenAICostUsesVerifiedModelAndCachedSubset(t *testing.T) {
	p := DefaultPricing()
	cost, ok := p.OpenAICost("gpt-5.6-luna", 1000000, 250000, 1000000)
	if !ok || cost != 1355000 {
		t.Fatalf("cost=%d ok=%v, want 710000 micros", cost, ok)
	}
}

func TestOpenAICostUnknownModelAndZeroUsage(t *testing.T) {
	p := DefaultPricing()
	if _, ok := p.OpenAICost("unknown-model", 1, 0, 1); ok {
		t.Fatal("unknown model must not be priced")
	}
	if cost, ok := p.OpenAICost("gpt-5.6-luna", 0, 0, 0); !ok || cost != 0 {
		t.Fatalf("zero usage=%d ok=%v", cost, ok)
	}
}
