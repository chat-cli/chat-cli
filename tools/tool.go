// Package tools provides the tool-use registry and built-in tools that
// chat-cli's chat command can offer to Bedrock models supporting tool use.
package tools

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
)

// Tool is the contract any tool-callable capability must implement to be
// registered with a Registry.
type Tool interface {
	// Name is the tool's identifier, as advertised to the model and used to
	// route a ToolUseBlock back to this Tool.
	Name() string
	// Description explains what the tool does, shown to the model.
	Description() string
	// InputSchema is the JSON schema describing the tool's expected input.
	InputSchema() document.Interface
	// Execute runs the tool with the model-supplied input and returns a
	// human/model-readable result, or an error if execution failed.
	Execute(ctx context.Context, input json.RawMessage) (string, error)

	// RequiresConfirmation reports whether calls to this tool must pass
	// through a PermissionGate before Execute is called. Read-only tools
	// return false and never have ConfirmationSummary called.
	RequiresConfirmation() bool

	// ConfirmationSummary returns human-readable text describing what a
	// specific call will do (shown at the confirmation prompt), and a
	// coarse pattern key used for sticky-approval matching. Only called
	// when RequiresConfirmation returns true; an error here is treated as
	// a denial - the call never reaches the gate or Execute.
	ConfirmationSummary(input json.RawMessage) (summary string, patternKey string, err error)
}
