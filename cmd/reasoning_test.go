/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"encoding/json"
	"testing"
)

func TestBuildReasoningConfig(t *testing.T) {
	t.Run("disabled yields nil", func(t *testing.T) {
		if got := buildReasoningConfig(false, 1024); got != nil {
			t.Errorf("expected nil when disabled, got %v", got)
		}
	})

	t.Run("enabled yields the expected reasoning_config shape", func(t *testing.T) {
		doc := buildReasoningConfig(true, 2048)
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

		reasoningConfig, ok := decoded["reasoning_config"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected a 'reasoning_config' object, got %v", decoded)
		}
		if reasoningConfig["type"] != "enabled" {
			t.Errorf("expected type 'enabled', got %v", reasoningConfig["type"])
		}
		// JSON numbers decode as float64 through the generic document unmarshaler.
		if budget, ok := reasoningConfig["budget_tokens"].(float64); !ok || budget != 2048 {
			t.Errorf("expected budget_tokens 2048, got %v", reasoningConfig["budget_tokens"])
		}
	})
}
