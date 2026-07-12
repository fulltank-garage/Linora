package config

import (
	"strings"
	"testing"
)

func TestAIConfigValidateRequiresEveryEnvironmentValue(t *testing.T) {
	err := (AIConfig{}).Validate()
	if err == nil {
		t.Fatal("expected a validation error")
	}
	for _, name := range []string{"AI_PROVIDER", "AI_BASE_URL", "AI_MODEL", "AI_API_KEY"} {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("expected %q in %q", name, err)
		}
	}
}

func TestAIConfigValidateAllowsOpenAICompatibleProvider(t *testing.T) {
	err := (AIConfig{APIKey: "key", BaseURL: "https://example.test/v1", Model: "model", Provider: "openai-compatible"}).Validate()
	if err != nil {
		t.Fatalf("expected valid configuration, got %v", err)
	}
}
