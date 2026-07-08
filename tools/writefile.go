package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/chat-cli/chat-cli/utils"
)

// maxWriteFileSummaryPreview is the content-length threshold past which the
// confirmation summary shows a truncated preview instead of the full
// content (BR4/NFR usability).
const maxWriteFileSummaryPreview = 4 * 1024

// WriteFileTool is a destructive built-in tool that creates or overwrites a
// file within the working directory. Every call must pass through a
// PermissionGate (RequiresConfirmation returns true) before Execute runs.
type WriteFileTool struct{}

// NewWriteFileTool creates a WriteFileTool.
func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Create or overwrite a text file within the current working directory."
}

func (t *WriteFileTool) InputSchema() document.Interface {
	return document.NewLazyDocument(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file, relative to the current working directory.",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The full content to write to the file.",
			},
		},
		"required": []interface{}{"path", "content"},
	})
}

type writeFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (t *WriteFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var params writeFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid tool input: %w", err)
	}

	fullPath, err := utils.ValidateLocalPathForWrite(params.Path)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(fullPath, []byte(params.Content), 0600); err != nil { // #nosec G306 - path is validated above; 0600 only applies when creating a new file
		return "", fmt.Errorf("unable to write file: %w", err)
	}

	return fmt.Sprintf("wrote %d bytes to %s", len(params.Content), params.Path), nil
}

func (t *WriteFileTool) RequiresConfirmation() bool {
	return true
}

func (t *WriteFileTool) ConfirmationSummary(input json.RawMessage) (string, string, error) {
	var params writeFileInput
	if err := json.Unmarshal(input, &params); err != nil {
		return "", "", fmt.Errorf("invalid tool input: %w", err)
	}

	fullPath, err := utils.ValidateLocalPathForWrite(params.Path)
	if err != nil {
		return "", "", err
	}

	content := params.Content
	if len(content) > maxWriteFileSummaryPreview {
		content = fmt.Sprintf("%s\n... (%d bytes total, shown truncated)", content[:maxWriteFileSummaryPreview], len(params.Content))
	}
	summary := fmt.Sprintf("Write to %s:\n%s", params.Path, content)

	patternKey := writeFilePatternKey(fullPath)

	return summary, patternKey, nil
}

// writeFilePatternKey implements BR5: the resolved path's containing
// directory, relative to the git repo root if inside one, else relative to
// cwd itself.
func writeFilePatternKey(fullPath string) string {
	dir := filepath.Dir(fullPath)

	root := ""
	if cwd, err := os.Getwd(); err == nil {
		root = utils.FindGitBoundary(cwd)
		if root == "" {
			root = cwd
		}
	}

	if root == "" {
		return dir
	}

	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return dir
	}
	return rel
}
