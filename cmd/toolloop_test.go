/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/chat-cli/chat-cli/tools"
)

func malformedWriteFileChannel(toolUseID string) <-chan types.ConverseStreamOutput {
	ch := make(chan types.ConverseStreamOutput, 5)
	ch <- &types.ConverseStreamOutputMemberContentBlockStart{
		Value: types.ContentBlockStartEvent{
			ContentBlockIndex: aws.Int32(0),
			Start: &types.ContentBlockStartMemberToolUse{
				Value: types.ToolUseBlockStart{Name: aws.String("write_file"), ToolUseId: aws.String(toolUseID)},
			},
		},
	}
	// Truncated JSON, as happens when max-tokens cuts off mid tool-input stream.
	ch <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta: &types.ContentBlockDeltaMemberToolUse{
				Value: types.ToolUseBlockDelta{Input: aws.String(`{"path":"hello.py","content":"def main(`)},
			},
		},
	}
	ch <- &types.ConverseStreamOutputMemberMessageStop{
		Value: types.MessageStopEvent{StopReason: types.StopReasonToolUse},
	}
	close(ch)
	return ch
}

func TestAccumulateStream_MalformedToolInputDoesNotError(t *testing.T) {
	_, toolCalls, stopReason, err := accumulateStream(
		malformedWriteFileChannel("write-1"),
		func(context.Context, string) error { return nil },
		func(context.Context, string) error { return nil },
	)
	if err != nil {
		t.Fatalf("expected malformed tool JSON to be recoverable, got error: %v", err)
	}
	if stopReason != types.StopReasonToolUse {
		t.Fatalf("expected tool_use stop reason, got %v", stopReason)
	}
	if len(toolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(toolCalls))
	}
	if toolCalls[0].InputParseErr == nil {
		t.Fatal("expected InputParseErr to be set for truncated JSON")
	}
}

func TestRunChatTurnWithTools_MalformedToolInputRecovers(t *testing.T) {
	callCount := 0
	send := func(_ context.Context, _ *bedrockruntime.ConverseStreamInput) (<-chan types.ConverseStreamOutput, error) {
		callCount++
		if callCount == 1 {
			return malformedWriteFileChannel("write-1"), nil
		}
		return textOnlyChannel("recovered after tool input error"), nil
	}

	registry := tools.NewRegistry()
	registry.Register(tools.NewWriteFileTool())
	input := &bedrockruntime.ConverseStreamInput{}

	result, err := runChatTurnWithTools(
		context.Background(),
		send,
		input,
		registry,
		nil,
		func(context.Context, string) error { return nil },
		func(context.Context, string) error { return nil },
	)
	if err != nil {
		t.Fatalf("expected recovery from malformed tool input, got: %v", err)
	}
	if !strings.Contains(result, "recovered after tool input error") {
		t.Fatalf("expected model to continue after tool error, got %q", result)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 round trips (error result + retry), got %d", callCount)
	}
}

func TestFinalizeToolCall(t *testing.T) {
	t.Run("valid JSON input", func(t *testing.T) {
		call := finalizeToolCall("read_file", "tool-use-1", `{"path":"go.mod"}`)
		if call.InputParseErr != nil {
			t.Fatalf("unexpected parse error: %v", call.InputParseErr)
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

	t.Run("malformed JSON input sets parse error", func(t *testing.T) {
		call := finalizeToolCall("read_file", "tool-use-2", `{"path": not valid`)
		if call.InputParseErr == nil {
			t.Error("expected InputParseErr for malformed JSON, got none")
		}
	})

	t.Run("empty input treated as empty object", func(t *testing.T) {
		call := finalizeToolCall("read_file", "tool-use-3", "")
		if call.InputParseErr != nil {
			t.Fatalf("unexpected parse error: %v", call.InputParseErr)
		}
		if string(call.Input) != "{}" {
			t.Errorf("expected empty input to become '{}', got %q", string(call.Input))
		}
	})
}

func TestToolInputDocument(t *testing.T) {
	t.Run("parses object fields from raw JSON", func(t *testing.T) {
		doc := toolInputDocument(json.RawMessage(`{"path":"go.mod"}`))
		if doc == nil {
			t.Fatal("expected non-nil document")
		}
	})

	t.Run("invalid object JSON falls back to empty object", func(t *testing.T) {
		doc := toolInputDocument(json.RawMessage(`[]`))
		if doc == nil {
			t.Fatal("expected non-nil document")
		}
	})
}

func TestFinalizeToolCall_TruncatedJSONSetsParseErr(t *testing.T) {
	call := finalizeToolCall("write_file", "id-1", `{"path":"a.py","content":"unfinished`)
	if call.InputParseErr == nil {
		t.Fatal("expected InputParseErr for truncated JSON")
	}
	if string(call.Input) != "{}" {
		t.Fatalf("expected placeholder input {}, got %s", call.Input)
	}
}
