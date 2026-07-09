package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// ToolCall is the finalized, parsed form of a model-requested tool
// invocation, ready to dispatch to a registered Tool.
type ToolCall struct {
	Name      string
	ToolUseID string
	Input     json.RawMessage
}

// Registry holds the set of tools available to a chat session and mediates
// between Bedrock's tool-use protocol and concrete Tool implementations.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry, keyed by its Name().
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// ToolConfiguration builds the Bedrock ToolConfiguration for this registry's
// tools. Returns nil when no tools are registered, so a request's shape is
// unchanged when tool use isn't in play.
func (r *Registry) ToolConfiguration() *types.ToolConfiguration {
	if len(r.tools) == 0 {
		return nil
	}

	sdkTools := make([]types.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		sdkTools = append(sdkTools, &types.ToolMemberToolSpec{
			Value: types.ToolSpecification{
				Name:        aws.String(tool.Name()),
				Description: aws.String(tool.Description()),
				InputSchema: &types.ToolInputSchemaMemberJson{
					Value: tool.InputSchema(),
				},
			},
		})
	}

	return &types.ToolConfiguration{Tools: sdkTools}
}

// Dispatch executes the tool named by call.Name, if registered, and always
// returns a ToolResultBlock (success or error) - it never panics and never
// returns a Go error, so callers can send the result straight back to the
// model without special-casing failure.
//
// If the tool requires confirmation, gate.Check is consulted before Execute
// is called (BR3-BR6, BR11). gate may be nil only when no registered tool
// requires confirmation - Dispatch never dereferences a nil gate for a
// non-destructive tool.
func (r *Registry) Dispatch(ctx context.Context, call ToolCall, gate PermissionGate) types.ToolResultBlock {
	tool, ok := r.tools[call.Name]
	if !ok {
		return errorResult(call.ToolUseID, fmt.Sprintf("unknown tool: %s", call.Name))
	}

	if tool.RequiresConfirmation() {
		summary, patternKey, err := tool.ConfirmationSummary(call.Input)
		if err != nil {
			return errorResult(call.ToolUseID, fmt.Sprintf("invalid tool input: %s", err.Error()))
		}

		if gate.Check(tool.Name(), patternKey, summary) == DecisionDeny {
			return errorResult(call.ToolUseID, "user declined this action")
		}
	}

	output, err := tool.Execute(ctx, call.Input)
	if err != nil {
		return errorResult(call.ToolUseID, err.Error())
	}

	return types.ToolResultBlock{
		ToolUseId: aws.String(call.ToolUseID),
		Status:    types.ToolResultStatusSuccess,
		Content: []types.ToolResultContentBlock{
			&types.ToolResultContentBlockMemberText{Value: output},
		},
	}
}

func errorResult(toolUseID, message string) types.ToolResultBlock {
	return types.ToolResultBlock{
		ToolUseId: aws.String(toolUseID),
		Status:    types.ToolResultStatusError,
		Content: []types.ToolResultContentBlock{
			&types.ToolResultContentBlockMemberText{Value: message},
		},
	}
}
