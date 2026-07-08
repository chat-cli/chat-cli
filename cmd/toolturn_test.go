/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/chat-cli/chat-cli/tools"
)

func textOnlyChannel(text string) <-chan types.ConverseStreamOutput {
	ch := make(chan types.ConverseStreamOutput, 3)
	ch <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta:             &types.ContentBlockDeltaMemberText{Value: text},
		},
	}
	ch <- &types.ConverseStreamOutputMemberMessageStop{
		Value: types.MessageStopEvent{StopReason: types.StopReasonEndTurn},
	}
	close(ch)
	return ch
}

func toolUseChannel(toolUseID string) <-chan types.ConverseStreamOutput {
	ch := make(chan types.ConverseStreamOutput, 4)
	ch <- &types.ConverseStreamOutputMemberContentBlockStart{
		Value: types.ContentBlockStartEvent{
			ContentBlockIndex: aws.Int32(0),
			Start: &types.ContentBlockStartMemberToolUse{
				Value: types.ToolUseBlockStart{Name: aws.String("fake_tool"), ToolUseId: aws.String(toolUseID)},
			},
		},
	}
	ch <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta:             &types.ContentBlockDeltaMemberToolUse{Value: types.ToolUseBlockDelta{Input: aws.String(`{}`)}},
		},
	}
	ch <- &types.ConverseStreamOutputMemberMessageStop{
		Value: types.MessageStopEvent{StopReason: types.StopReasonToolUse},
	}
	close(ch)
	return ch
}

// fakeTurnTool is a minimal Tool used only to satisfy Registry.Register in
// these tests; its actual output doesn't matter for the cap/no-tool-use
// behavior being tested.
type fakeTurnTool struct{}

func (f *fakeTurnTool) Name() string        { return "fake_tool" }
func (f *fakeTurnTool) Description() string { return "test tool" }
func (f *fakeTurnTool) InputSchema() document.Interface {
	return document.NewLazyDocument(map[string]interface{}{"type": "object"})
}
func (f *fakeTurnTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	return "ok", nil
}

func TestRunChatTurnWithTools_NoToolUse(t *testing.T) {
	callCount := 0
	send := func(_ context.Context, _ *bedrockruntime.ConverseStreamInput) (<-chan types.ConverseStreamOutput, error) {
		callCount++
		return textOnlyChannel("final answer"), nil
	}

	registry := tools.NewRegistry()
	input := &bedrockruntime.ConverseStreamInput{}
	onText := func(_ context.Context, _ string) error { return nil }

	onReasoning := func(_ context.Context, _ string) error { return nil }

	result, err := runChatTurnWithTools(context.Background(), send, input, registry, onText, onReasoning)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "final answer" {
		t.Errorf("expected 'final answer', got %q", result)
	}
	if callCount != 1 {
		t.Errorf("expected send to be called exactly once, got %d", callCount)
	}
}

func TestRunChatTurnWithTools_RoundTripCap(t *testing.T) {
	callCount := 0
	send := func(_ context.Context, _ *bedrockruntime.ConverseStreamInput) (<-chan types.ConverseStreamOutput, error) {
		callCount++
		return toolUseChannel("call-id"), nil
	}

	registry := tools.NewRegistry()
	registry.Register(&fakeTurnTool{})
	input := &bedrockruntime.ConverseStreamInput{}
	onText := func(_ context.Context, _ string) error { return nil }

	onReasoning := func(_ context.Context, _ string) error { return nil }

	_, err := runChatTurnWithTools(context.Background(), send, input, registry, onText, onReasoning)
	if err == nil {
		t.Fatal("expected an error when the round-trip cap is exceeded, got none")
	}
	if callCount > maxToolRoundTrips+1 {
		t.Errorf("expected at most %d calls to send before stopping, got %d", maxToolRoundTrips+1, callCount)
	}
	if callCount < maxToolRoundTrips {
		t.Errorf("expected at least %d calls to send before the cap kicks in, got %d", maxToolRoundTrips, callCount)
	}
}
