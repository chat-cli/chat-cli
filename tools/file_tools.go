package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Tool represents a capability that an agent can use
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
	Schema() map[string]interface{}
}

// ReadFileTool allows reading file contents
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required and must be a string")
	}

	// Security check: ensure we're only working within current directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to get current directory: %v", err)
	}

	if !strings.HasPrefix(absPath, cwd) {
		return nil, fmt.Errorf("access denied: file must be within current working directory")
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	return map[string]interface{}{
		"content":   string(content),
		"file_path": absPath,
		"size":      len(content),
	}, nil
}

func (t *ReadFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read (relative to current working directory)",
			},
		},
		"required": []string{"file_path"},
	}
}

// WriteFileTool allows writing content to files
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file"
}

func (t *WriteFileTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required and must be a string")
	}

	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter is required and must be a string")
	}

	// Security check: ensure we're only working within current directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to get current directory: %v", err)
	}

	if !strings.HasPrefix(absPath, cwd) {
		return nil, fmt.Errorf("access denied: file must be within current working directory")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %v", err)
	}

	return map[string]interface{}{
		"file_path":    absPath,
		"bytes_written": len(content),
		"success":      true,
	}, nil
}

func (t *WriteFileTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write (relative to current working directory)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

// ListFilesTool allows listing files in a directory
type ListFilesTool struct{}

func (t *ListFilesTool) Name() string {
	return "list_files"
}

func (t *ListFilesTool) Description() string {
	return "List files and directories in a given path"
}

func (t *ListFilesTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dirPath := "."
	if path, ok := params["dir_path"].(string); ok {
		dirPath = path
	}

	// Check for file extension filter
	var extensionFilter string
	if ext, ok := params["extension"].(string); ok {
		extensionFilter = strings.ToLower(ext)
		if !strings.HasPrefix(extensionFilter, ".") {
			extensionFilter = "." + extensionFilter
		}
	}

	// Security check: ensure we're only working within current directory
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable to get current directory: %v", err)
	}

	if !strings.HasPrefix(absPath, cwd) {
		return nil, fmt.Errorf("access denied: directory must be within current working directory")
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	files := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Apply extension filter if specified
		if extensionFilter != "" && !entry.IsDir() {
			if !strings.HasSuffix(strings.ToLower(entry.Name()), extensionFilter) {
				continue
			}
		}

		files = append(files, map[string]interface{}{
			"name":         entry.Name(),
			"is_directory": entry.IsDir(),
			"size":         info.Size(),
			"modified":     info.ModTime(),
		})
	}

	return map[string]interface{}{
		"directory": absPath,
		"files":     files,
		"count":     len(files),
	}, nil
}

func (t *ListFilesTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"dir_path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to list (defaults to current directory)",
			},
			"extension": map[string]interface{}{
				"type":        "string",
				"description": "Filter files by extension (e.g., '.md', '.go', 'txt')",
			},
		},
	}
}

// GetFileTools returns all available file manipulation tools
func GetFileTools() []Tool {
	return []Tool{
		&ReadFileTool{},
		&WriteFileTool{},
		&ListFilesTool{},
	}
}