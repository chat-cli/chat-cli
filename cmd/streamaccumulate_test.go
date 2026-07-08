/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

func TestAccumulateStream_TextOnly(t *testing.T) {
	events := make(chan types.ConverseStreamOutput, 10)
	events <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta:             &types.ContentBlockDeltaMemberText{Value: "Hello"},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta:             &types.ContentBlockDeltaMemberText{Value: " world"},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockStop{
		Value: types.ContentBlockStopEvent{ContentBlockIndex: aws.Int32(0)},
	}
	events <- &types.ConverseStreamOutputMemberMessageStop{
		Value: types.MessageStopEvent{StopReason: types.StopReasonEndTurn},
	}
	close(events)

	var received string
	onText := func(_ context.Context, part string) error {
		received += part
		return nil
	}

	onReasoning := func(_ context.Context, _ string) error { return nil }

	msg, toolCalls, stopReason, err := accumulateStream(events, onText, onReasoning)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stopReason != types.StopReasonEndTurn {
		t.Errorf("expected StopReasonEndTurn, got %v", stopReason)
	}
	if len(toolCalls) != 0 {
		t.Errorf("expected no tool calls, got %d", len(toolCalls))
	}
	if received != "Hello world" {
		t.Errorf("expected onText to receive 'Hello world', got %q", received)
	}
	if len(msg.Content) != 1 {
		t.Fatalf("expected 1 content block in the finalized message, got %d", len(msg.Content))
	}
	textBlock, ok := msg.Content[0].(*types.ContentBlockMemberText)
	if !ok {
		t.Fatalf("expected text content block, got %T", msg.Content[0])
	}
	if textBlock.Value != "Hello world" {
		t.Errorf("expected finalized text 'Hello world', got %q", textBlock.Value)
	}
}

func TestAccumulateStream_ToolUse(t *testing.T) {
	events := make(chan types.ConverseStreamOutput, 10)
	events <- &types.ConverseStreamOutputMemberContentBlockStart{
		Value: types.ContentBlockStartEvent{
			ContentBlockIndex: aws.Int32(0),
			Start: &types.ContentBlockStartMemberToolUse{
				Value: types.ToolUseBlockStart{
					Name:      aws.String("read_file"),
					ToolUseId: aws.String("tool-use-1"),
				},
			},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta:             &types.ContentBlockDeltaMemberToolUse{Value: types.ToolUseBlockDelta{Input: aws.String(`{"path":`)}},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta:             &types.ContentBlockDeltaMemberToolUse{Value: types.ToolUseBlockDelta{Input: aws.String(`"go.mod"}`)}},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockStop{
		Value: types.ContentBlockStopEvent{ContentBlockIndex: aws.Int32(0)},
	}
	events <- &types.ConverseStreamOutputMemberMessageStop{
		Value: types.MessageStopEvent{StopReason: types.StopReasonToolUse},
	}
	close(events)

	onText := func(_ context.Context, _ string) error { return nil }

	onReasoning := func(_ context.Context, _ string) error { return nil }

	msg, toolCalls, stopReason, err := accumulateStream(events, onText, onReasoning)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stopReason != types.StopReasonToolUse {
		t.Errorf("expected StopReasonToolUse, got %v", stopReason)
	}
	if len(toolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(toolCalls))
	}
	if toolCalls[0].Name != "read_file" || toolCalls[0].ToolUseID != "tool-use-1" {
		t.Errorf("unexpected tool call: %+v", toolCalls[0])
	}
	if string(toolCalls[0].Input) != `{"path":"go.mod"}` {
		t.Errorf("expected accumulated input %q, got %q", `{"path":"go.mod"}`, string(toolCalls[0].Input))
	}
	if len(msg.Content) != 1 {
		t.Fatalf("expected 1 content block in the finalized message, got %d", len(msg.Content))
	}
	if _, ok := msg.Content[0].(*types.ContentBlockMemberToolUse); !ok {
		t.Fatalf("expected tool use content block, got %T", msg.Content[0])
	}
}

func TestAccumulateStream_ReasoningContent(t *testing.T) {
	events := make(chan types.ConverseStreamOutput, 10)
	events <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta: &types.ContentBlockDeltaMemberReasoningContent{
				Value: &types.ReasoningContentBlockDeltaMemberText{Value: "Let me think"},
			},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(0),
			Delta: &types.ContentBlockDeltaMemberReasoningContent{
				Value: &types.ReasoningContentBlockDeltaMemberSignature{Value: "sig-123"},
			},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockStop{
		Value: types.ContentBlockStopEvent{ContentBlockIndex: aws.Int32(0)},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockDelta{
		Value: types.ContentBlockDeltaEvent{
			ContentBlockIndex: aws.Int32(1),
			Delta:             &types.ContentBlockDeltaMemberText{Value: "final answer"},
		},
	}
	events <- &types.ConverseStreamOutputMemberContentBlockStop{
		Value: types.ContentBlockStopEvent{ContentBlockIndex: aws.Int32(1)},
	}
	events <- &types.ConverseStreamOutputMemberMessageStop{
		Value: types.MessageStopEvent{StopReason: types.StopReasonEndTurn},
	}
	close(events)

	onText := func(_ context.Context, _ string) error { return nil }

	var receivedReasoning string
	onReasoning := func(_ context.Context, part string) error {
		receivedReasoning += part
		return nil
	}

	msg, _, _, err := accumulateStream(events, onText, onReasoning)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedReasoning != "Let me think" {
		t.Errorf("expected onReasoning to receive 'Let me think', got %q", receivedReasoning)
	}
	if len(msg.Content) != 2 {
		t.Fatalf("expected 2 content blocks (reasoning + text), got %d", len(msg.Content))
	}

	reasoningBlock, ok := msg.Content[0].(*types.ContentBlockMemberReasoningContent)
	if !ok {
		t.Fatalf("expected reasoning content block first, got %T", msg.Content[0])
	}
	reasoningText, ok := reasoningBlock.Value.(*types.ReasoningContentBlockMemberReasoningText)
	if !ok {
		t.Fatalf("expected reasoning text member, got %T", reasoningBlock.Value)
	}
	if aws.ToString(reasoningText.Value.Text) != "Let me think" {
		t.Errorf("expected finalized reasoning text 'Let me think', got %q", aws.ToString(reasoningText.Value.Text))
	}
	if aws.ToString(reasoningText.Value.Signature) != "sig-123" {
		t.Errorf("expected finalized signature 'sig-123', got %q", aws.ToString(reasoningText.Value.Signature))
	}

	if _, ok := msg.Content[1].(*types.ContentBlockMemberText); !ok {
		t.Fatalf("expected text content block second, got %T", msg.Content[1])
	}
}
