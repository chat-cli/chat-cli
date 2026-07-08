package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/chat-cli/chat-cli/utils"
)

// ReadFileTool is the one built-in tool shipped with tool-use support: a
// read-only, working-directory-confined file reader.
type ReadFileTool struct{}

// NewReadFileTool creates a ReadFileTool.
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a text file within the current working directory."
}

func (t *ReadFileTool) InputSchema() document.Interface {
	return document.NewLazyDocument(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file, relative to the current working directory.",
			},
		},
		"required": []interface{}{"path"},
	})
}

type readFileInput struct {
	Path string `json:"path"`
}

func (t *ReadFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params readFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid tool input: %w", err)
	}

	fullPath, err := utils.ValidateLocalPath(params.Path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(fullPath) // #nosec G304 - path is validated above
	if err != nil {
		return "", fmt.Errorf("unable to read file: %w", err)
	}

	return string(data), nil
}
