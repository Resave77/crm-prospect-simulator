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

func TestFindMenuTimeoutDefaultsLongerWithoutChangingDefault(t *testing.T) {
	t.Setenv("OPENAI_TIMEOUT", "20s")
	t.Setenv("OPENAI_FIND_MENU_TIMEOUT", "60s")
	normal, err := duration("OPENAI_TIMEOUT", 20*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	finder, err := duration("OPENAI_FIND_MENU_TIMEOUT", 60*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if normal != 20*time.Second || finder != 60*time.Second {
		t.Fatalf("normal=%s finder=%s", normal, finder)
	}
}

func TestMenuProfileTimeoutIsOperationSpecific(t *testing.T) {
	t.Setenv("OPENAI_TIMEOUT", "20s")
	t.Setenv("OPENAI_MENU_PROFILE_TIMEOUT", "40s")
	normal, _ := duration("OPENAI_TIMEOUT", 20*time.Second)
	profiling, err := duration("OPENAI_MENU_PROFILE_TIMEOUT", 40*time.Second)
	if err != nil || normal != 20*time.Second || profiling != 40*time.Second {
		t.Fatalf("normal=%s profiling=%s err=%v", normal, profiling, err)
	}
}
