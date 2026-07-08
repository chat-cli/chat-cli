/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"encoding/json"
	"testing"
)

func TestUsesAdaptiveThinking(t *testing.T) {
	tests := []struct {
		modelID string
		want    bool
	}{
		{"us.anthropic.claude-sonnet-5", true},
		{"us.anthropic.claude-sonnet-4-6", true},
		{"us.anthropic.claude-opus-4-6-v1", true},
		{"us.anthropic.claude-sonnet-4-20250514-v1:0", false},
		{"anthropic.claude-3-7-sonnet-20250219-v1:0", false},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			if got := usesAdaptiveThinking(tt.modelID); got != tt.want {
				t.Fatalf("usesAdaptiveThinking(%q) = %v, want %v", tt.modelID, got, tt.want)
			}
		})
	}
}

func TestNormalizeThinkingEffort(t *testing.T) {
	t.Run("empty defaults to medium", func(t *testing.T) {
		got, err := normalizeThinkingEffort("")
		if err != nil || got != "medium" {
			t.Fatalf("expected medium, got %q err=%v", got, err)
		}
	})

	t.Run("invalid effort is rejected", func(t *testing.T) {
		_, err := normalizeThinkingEffort("turbo")
		if err == nil {
			t.Fatal("expected error for invalid effort")
		}
	})
}

func TestBuildReasoningConfig(t *testing.T) {
	t.Run("disabled yields nil", func(t *testing.T) {
		if got := buildReasoningConfig("us.anthropic.claude-sonnet-5", false, 1024, "medium"); got != nil {
			t.Errorf("expected nil when disabled, got %v", got)
		}
	})

	t.Run("legacy model uses enabled thinking with budget", func(t *testing.T) {
		doc := buildReasoningConfig("us.anthropic.claude-sonnet-4-20250514-v1:0", true, 2048, "medium")
		if doc == nil {
			t.Fatal("expected a non-nil document when enabled")
		}

		raw, err := doc.MarshalSmithyDocument()
		if err != nil {
			t.Fatalf("failed to marshal document: %v", err)
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("failed to unmarshal document JSON: %v", err)
		}

		thinking, ok := decoded["thinking"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected a 'thinking' object, got %v", decoded)
		}
		if thinking["type"] != "enabled" {
			t.Errorf("expected type 'enabled', got %v", thinking["type"])
		}
		if budget, ok := thinking["budget_tokens"].(float64); !ok || budget != 2048 {
			t.Errorf("expected budget_tokens 2048, got %v", thinking["budget_tokens"])
		}
		if _, ok := decoded["output_config"]; ok {
			t.Errorf("did not expect output_config for legacy model, got %v", decoded["output_config"])
		}
	})

	t.Run("adaptive model uses adaptive thinking with effort", func(t *testing.T) {
		doc := buildReasoningConfig("us.anthropic.claude-sonnet-5", true, 2048, "high")
		if doc == nil {
			t.Fatal("expected a non-nil document when enabled")
		}

		raw, err := doc.MarshalSmithyDocument()
		if err != nil {
			t.Fatalf("failed to marshal document: %v", err)
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("failed to unmarshal document JSON: %v", err)
		}

		thinking, ok := decoded["thinking"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected a 'thinking' object, got %v", decoded)
		}
		if thinking["type"] != "adaptive" {
			t.Errorf("expected type 'adaptive', got %v", thinking["type"])
		}
		if _, ok := thinking["budget_tokens"]; ok {
			t.Errorf("did not expect budget_tokens for adaptive model, got %v", thinking["budget_tokens"])
		}

		outputConfig, ok := decoded["output_config"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected output_config object, got %v", decoded["output_config"])
		}
		if outputConfig["effort"] != "high" {
			t.Errorf("expected effort high, got %v", outputConfig["effort"])
		}
	})
}
