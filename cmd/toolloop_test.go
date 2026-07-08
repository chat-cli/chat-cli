/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"testing"
)

func TestFinalizeToolCall(t *testing.T) {
	t.Run("valid JSON input", func(t *testing.T) {
		call, err := finalizeToolCall("read_file", "tool-use-1", `{"path":"go.mod"}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if call.Name != "read_file" {
			t.Errorf("expected name 'read_file', got %q", call.Name)
		}
		if call.ToolUseID != "tool-use-1" {
			t.Errorf("expected tool use id 'tool-use-1', got %q", call.ToolUseID)
		}
		if string(call.Input) != `{"path":"go.mod"}` {
			t.Errorf("expected input to be preserved as raw JSON, got %q", string(call.Input))
		}
	})

	t.Run("malformed JSON input", func(t *testing.T) {
		_, err := finalizeToolCall("read_file", "tool-use-2", `{"path": not valid`)
		if err == nil {
			t.Error("expected an error for malformed JSON, got none")
		}
	})

	t.Run("empty input treated as empty object", func(t *testing.T) {
		call, err := finalizeToolCall("read_file", "tool-use-3", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(call.Input) != "{}" {
			t.Errorf("expected empty input to become '{}', got %q", string(call.Input))
		}
	})
}
