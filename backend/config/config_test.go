package config

import (
	"testing"
	"time"
)

func TestAIConfiguredRequiresEnabledKeyAndModel(t *testing.T) {
	cfg := Config{AIEnabled: true, OpenAIAPIKey: "key", OpenAIModel: "model"}
	if !cfg.AIConfigured() {
		t.Fatal("expected configured AI")
	}
	for _, candidate := range []Config{
		{AIEnabled: false, OpenAIAPIKey: "key", OpenAIModel: "model"},
		{AIEnabled: true, OpenAIAPIKey: "", OpenAIModel: "model"},
		{AIEnabled: true, OpenAIAPIKey: "key", OpenAIModel: ""},
	} {
		if candidate.AIConfigured() {
			t.Fatalf("expected unconfigured for %+v", candidate)
		}
	}
}

func TestPositiveIntRejectsInvalidValues(t *testing.T) {
	t.Setenv("OPENAI_MAX_OUTPUT_TOKENS", "0")
	if _, err := positiveInt("OPENAI_MAX_OUTPUT_TOKENS", 800); err == nil {
		t.Fatal("expected invalid positive integer error")
	}
	t.Setenv("OPENAI_MAX_OUTPUT_TOKENS", "123")
	value, err := positiveInt("OPENAI_MAX_OUTPUT_TOKENS", 800)
	if err != nil || value != 123 {
		t.Fatalf("value=%d err=%v", value, err)
	}
}

func TestDurationRejectsInvalidAIValues(t *testing.T) {
	t.Setenv("OPENAI_TIMEOUT", "-1s")
	if _, err := duration("OPENAI_TIMEOUT", 20*time.Second); err == nil {
		t.Fatal("expected invalid duration error")
	}
}
