/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/chat-cli/chat-cli/tools"
	"github.com/chat-cli/chat-cli/utils"
)

// maxToolRoundTrips caps consecutive tool-use round trips within a single
// user turn (Rule 5 / REL-1), guarding against a runaway model that keeps
// requesting tools without end.
const maxToolRoundTrips = 10

// converseStreamFunc abstracts the Bedrock ConverseStream call so
// runChatTurnWithTools is unit-testable without needing to mock the AWS
// SDK's unexported stream internals - only the resulting event channel
// matters to the loop.
type converseStreamFunc func(ctx context.Context, input *bedrockruntime.ConverseStreamInput) (<-chan types.ConverseStreamOutput, error)

type blockKind int

const (
	blockKindText blockKind = iota
	blockKindToolUse
	blockKindReasoning
)

// blockAccumulator tracks one in-progress content block by its stream index,
// per functional-design/domain-entities.md.
type blockAccumulator struct {
	kind               blockKind
	text               strings.Builder
	toolName           string
	toolUseID          string
	toolInput          strings.Builder
	reasoningText      strings.Builder
	reasoningSignature string
}

// accumulateStream drains a Bedrock ConverseStream event channel, invoking
// onText for each text delta as it arrives (same behavior as
// utils.ProcessStreamingOutput), and finalizes any tool-use blocks
// encountered. Returns the finalized assistant Message (ready to append to
// conversation history), any finalized tool calls (in stream order), the
// stop reason, and an error only for malformed tool-input JSON (Rule 4) -
// never for an unknown tool, which is Registry.Dispatch's job (Rule 2).
func accumulateStream(events <-chan types.ConverseStreamOutput, onText, onReasoning utils.StreamingOutputHandler) (types.Message, []tools.ToolCall, types.StopReason, error) {
	blocks := make(map[int32]*blockAccumulator)
	var order []int32
	var stopReason types.StopReason

	for event := range events {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberContentBlockStart:
			idx := aws.ToInt32(v.Value.ContentBlockIndex)
			if toolStart, ok := v.Value.Start.(*types.ContentBlockStartMemberToolUse); ok {
				blocks[idx] = &blockAccumulator{
					kind:      blockKindToolUse,
					toolName:  aws.ToString(toolStart.Value.Name),
					toolUseID: aws.ToString(toolStart.Value.ToolUseId),
				}
				order = append(order, idx)
			}

		case *types.ConverseStreamOutputMemberContentBlockDelta:
			idx := aws.ToInt32(v.Value.ContentBlockIndex)
			acc, ok := blocks[idx]
			if !ok {
				// Text and reasoning blocks have no explicit
				// ContentBlockStart event (only tool-use does) - the kind
				// is determined by whichever delta type arrives first.
				kind := blockKindText
				if _, isReasoning := v.Value.Delta.(*types.ContentBlockDeltaMemberReasoningContent); isReasoning {
					kind = blockKindReasoning
				}
				acc = &blockAccumulator{kind: kind}
				blocks[idx] = acc
				order = append(order, idx)
			}

			switch delta := v.Value.Delta.(type) {
			case *types.ContentBlockDeltaMemberText:
				acc.text.WriteString(delta.Value)
				if err := onText(context.Background(), delta.Value); err != nil {
					return types.Message{}, nil, "", fmt.Errorf("handler error: %w", err)
				}
			case *types.ContentBlockDeltaMemberToolUse:
				acc.toolInput.WriteString(aws.ToString(delta.Value.Input))
			case *types.ContentBlockDeltaMemberReasoningContent:
				switch reasoningDelta := delta.Value.(type) {
				case *types.ReasoningContentBlockDeltaMemberText:
					acc.reasoningText.WriteString(reasoningDelta.Value)
					if err := onReasoning(context.Background(), reasoningDelta.Value); err != nil {
						return types.Message{}, nil, "", fmt.Errorf("handler error: %w", err)
					}
				case *types.ReasoningContentBlockDeltaMemberSignature:
					acc.reasoningSignature = reasoningDelta.Value
				// ReasoningContentBlockDeltaMemberRedactedContent is preserved
				// implicitly (nothing to accumulate into visible text) - it's
				// encrypted bytes, not rendered per Rule 5.
				default:
				}
			}

		case *types.ConverseStreamOutputMemberMessageStop:
			stopReason = v.Value.StopReason
		}
	}

	var content []types.ContentBlock
	var toolCalls []tools.ToolCall

	for _, idx := range order {
		acc := blocks[idx]
		switch acc.kind {
		case blockKindText:
			content = append(content, &types.ContentBlockMemberText{Value: acc.text.String()})
		case blockKindReasoning:
			reasoningTextBlock := types.ReasoningTextBlock{Text: aws.String(acc.reasoningText.String())}
			if acc.reasoningSignature != "" {
				reasoningTextBlock.Signature = aws.String(acc.reasoningSignature)
			}
			content = append(content, &types.ContentBlockMemberReasoningContent{
				Value: &types.ReasoningContentBlockMemberReasoningText{Value: reasoningTextBlock},
			})
		case blockKindToolUse:
			call, err := finalizeToolCall(acc.toolName, acc.toolUseID, acc.toolInput.String())
			if err != nil {
				return types.Message{}, nil, "", err
			}
			toolCalls = append(toolCalls, call)

			content = append(content, &types.ContentBlockMemberToolUse{
				Value: types.ToolUseBlock{
					Name:      aws.String(acc.toolName),
					ToolUseId: aws.String(acc.toolUseID),
					Input:     toolInputDocument(call.Input),
				},
			})
		}
	}

	return types.Message{Role: types.ConversationRoleAssistant, Content: content}, toolCalls, stopReason, nil
}

// runChatTurnWithTools sends input via send, processes the response (text
// via onText, tool-use via registry), and - per the algorithm in
// functional-design/business-logic-model.md - loops on StopReasonToolUse
// until the model produces a final text response or maxToolRoundTrips is
// exceeded. input.Messages is mutated in place to build up the full
// conversation, including any intermediate tool-use/tool-result exchanges,
// exactly as Bedrock requires for context continuity within the turn.
func runChatTurnWithTools(
	ctx context.Context,
	send converseStreamFunc,
	input *bedrockruntime.ConverseStreamInput,
	registry *tools.Registry,
	gate tools.PermissionGate,
	onText utils.StreamingOutputHandler,
	onReasoning utils.StreamingOutputHandler,
) (string, error) {
	input.ToolConfig = registry.ToolConfiguration()

	roundTrips := 0
	for {
		events, err := send(ctx, input)
		if err != nil {
			return "", err
		}

		assistantMsg, toolCalls, stopReason, err := accumulateStream(events, onText, onReasoning)
		if err != nil {
			return "", err
		}

		input.Messages = append(input.Messages, assistantMsg)

		if stopReason != types.StopReasonToolUse {
			finalText := ""
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.ContentBlockMemberText); ok {
					finalText += textBlock.Value
				}
			}
			return finalText, nil
		}

		roundTrips++
		if roundTrips > maxToolRoundTrips {
			return "", fmt.Errorf("stopped after %d tool calls in a single turn to avoid a runaway loop - you can ask a follow-up to continue", maxToolRoundTrips)
		}

		var resultContent []types.ContentBlock
		for _, call := range toolCalls {
			result := registry.Dispatch(ctx, call, gate)
			resultContent = append(resultContent, &types.ContentBlockMemberToolResult{Value: result})
		}
		input.Messages = append(input.Messages, types.Message{
			Role:    types.ConversationRoleUser,
			Content: resultContent,
		})
	}
}

// finalizeToolCall parses a tool call's accumulated raw JSON input fragments
// into a tools.ToolCall, validating the JSON is well-formed (Rule 4 in
// functional-design/business-rules.md) before it's ever dispatched.
func finalizeToolCall(name, toolUseID, rawInput string) (tools.ToolCall, error) {
	if rawInput == "" {
		rawInput = "{}"
	}

	if !json.Valid([]byte(rawInput)) {
		return tools.ToolCall{}, fmt.Errorf("invalid tool input for %s: not valid JSON", name)
	}

	return tools.ToolCall{
		Name:      name,
		ToolUseID: toolUseID,
		Input:     json.RawMessage(rawInput),
	}, nil
}

// toolInputDocument converts validated tool-call JSON into the document
// shape Bedrock expects when echoing a ToolUseBlock back into conversation
// history. Passing json.RawMessage directly to NewLazyDocument serializes as
// a JSON string, not an object, which Bedrock rejects on the follow-up turn.
func toolInputDocument(raw json.RawMessage) document.Interface {
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil || parsed == nil {
		parsed = map[string]interface{}{}
	}

	return document.NewLazyDocument(parsed)
}
