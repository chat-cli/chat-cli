package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/chat-cli/chat-cli/tools"
)

const (
	// RetryDelay is the delay between API calls to prevent throttling
	RetryDelay = 1000 * time.Millisecond
)

// FileEditAgent is an agent specialized in file editing tasks
type FileEditAgent struct {
	tools      []tools.Tool
	bedrockSvc *bedrockruntime.Client
	modelId    string
}

// NewFileEditAgent creates a new file editing agent
func NewFileEditAgent(region, modelId string) (*FileEditAgent, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %v", err)
	}

	return &FileEditAgent{
		tools:      tools.GetFileTools(),
		bedrockSvc: bedrockruntime.NewFromConfig(cfg),
		modelId:    modelId,
	}, nil
}

func (a *FileEditAgent) Name() string {
	return "file_edit_agent"
}

func (a *FileEditAgent) Description() string {
	return "An agent specialized in reading, writing, and modifying files in the current working directory"
}

func (a *FileEditAgent) Tools() []Tool {
	result := make([]Tool, len(a.tools))
	for i, tool := range a.tools {
		result[i] = tool
	}
	return result
}

func (a *FileEditAgent) CanHandle(task string) bool {
	taskLower := strings.ToLower(task)
	keywords := []string{"file", "edit", "read", "write", "modify", "create", "update", "save"}

	for _, keyword := range keywords {
		if strings.Contains(taskLower, keyword) {
			return true
		}
	}

	return false
}

func (a *FileEditAgent) Execute(ctx context.Context, task string, context map[string]interface{}) (*AgentResult, error) {
	// Create system prompt for the agent
	systemPrompt := `You are a file editing agent. You can read, write, and list files in the current working directory.

Available tools:
- read_file: Read the contents of a file (parameters: file_path)
- write_file: Write content to a file (parameters: file_path, content)
- list_files: List files and directories (parameters: dir_path [optional], extension [optional, e.g., ".md" for markdown files])

Your task is to understand the user's request and use the appropriate tools to complete it.

IMPORTANT: Always use tools when you need to interact with files. Never make assumptions about file contents.

CRITICAL: You MUST respond with ONLY a valid JSON object. Do not include any explanatory text before or after the JSON.

When you need to use a tool, respond with ONLY this JSON format:
{
  "action": "use_tool",
  "tool_name": "tool_name",
  "parameters": {
    "param1": "value1",
    "param2": "value2"
  }
}

When you're done with the task, respond with ONLY this JSON format:
{
  "action": "complete",
  "message": "Description of what was accomplished"
}

If you need to think or plan, you can respond with ONLY this JSON format:
{
  "action": "think",
  "message": "Your thoughts about the task"
}

Remember: ONLY JSON responses are accepted. No additional text allowed.
`

	userPrompt := fmt.Sprintf("Task: %s", task)
	if len(context) > 0 {
		contextStr, _ := json.Marshal(context)
		userPrompt += fmt.Sprintf("\n\nAdditional context: %s", string(contextStr))
	}

	result := &AgentResult{
		Success:     true,
		ToolResults: []ToolResult{},
	}

	// Execute the agent loop
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		response, err := a.callLLM(ctx, systemPrompt, userPrompt)
		if err != nil {
			return &AgentResult{
				Success: false,
				Error:   fmt.Sprintf("LLM call failed: %v", err),
			}, nil
		}

		// Parse the response
		var actionResult map[string]interface{}
		if err := json.Unmarshal([]byte(response), &actionResult); err != nil {
			// If JSON parsing fails, try to extract JSON from the response
			if extractedJSON := extractJSONFromResponse(response); extractedJSON != "" {
				if err := json.Unmarshal([]byte(extractedJSON), &actionResult); err == nil {
					// Successfully extracted and parsed JSON, continue with normal flow
				} else {
					// Extracted JSON is still invalid, ask for clarification
					userPrompt += fmt.Sprintf("\n\nYour response was not valid JSON. Please respond with a proper JSON object as specified in the instructions. Your response was: %s", response)
					// Add delay before retrying to prevent throttling
					time.Sleep(RetryDelay)
					continue
				}
			} else {
				// No JSON found, ask for clarification
				userPrompt += fmt.Sprintf("\n\nYour response was not valid JSON. Please respond with a proper JSON object as specified in the instructions. Your response was: %s", response)
				// Add delay before retrying to prevent throttling
				time.Sleep(RetryDelay)
				continue
			}
		}

		action, ok := actionResult["action"].(string)
		if !ok {
			result.Success = false
			result.Error = "Invalid response format: missing action"
			break
		}

		switch action {
		case "complete":
			message, _ := actionResult["message"].(string)
			result.Success = true
			result.Message = message
			return result, nil

		case "think":
			message, _ := actionResult["message"].(string)
			userPrompt += fmt.Sprintf("\n\nAgent thinking: %s", message)

		case "use_tool":
			toolName, _ := actionResult["tool_name"].(string)
			parameters, _ := actionResult["parameters"].(map[string]interface{})

			toolResult, err := a.executeTool(ctx, toolName, parameters)
			result.ToolResults = append(result.ToolResults, *toolResult)

			if err != nil {
				userPrompt += fmt.Sprintf("\n\nTool %s failed: %v", toolName, err)
			} else {
				resultStr, _ := json.Marshal(toolResult.Result)
				userPrompt += fmt.Sprintf("\n\nTool %s result: %s", toolName, string(resultStr))
			}

		default:
			result.Success = false
			result.Error = fmt.Sprintf("Unknown action: %s", action)
			return result, nil
		}

		// Add delay between bedrock calls to prevent throttling
		// Only add delay if we're going to continue to the next iteration
		// and we haven't already added a delay in this iteration
		if i < maxIterations-1 && action != "complete" {
			time.Sleep(RetryDelay)
		}
	}

	if result.Message == "" {
		result.Message = "Task completed after maximum iterations"
	}

	return result, nil
}

func (a *FileEditAgent) executeTool(ctx context.Context, toolName string, parameters map[string]interface{}) (*ToolResult, error) {
	for _, tool := range a.tools {
		if tool.Name() == toolName {
			result, err := tool.Execute(ctx, parameters)
			return &ToolResult{
				ToolName: toolName,
				Success:  err == nil,
				Result:   result,
				Error: func() string {
					if err != nil {
						return err.Error()
					}
					return ""
				}(),
			}, err
		}
	}

	return &ToolResult{
		ToolName: toolName,
		Success:  false,
		Error:    fmt.Sprintf("Tool %s not found", toolName),
	}, fmt.Errorf("tool %s not found", toolName)
}

func (a *FileEditAgent) callLLM(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []types.Message{
		{
			Role: types.ConversationRoleUser,
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: systemPrompt + "\n\n" + userPrompt,
				},
			},
		},
	}

	maxTokens := int32(1000)
	temperature := float32(0.1)
	topP := float32(0.9)

	input := &bedrockruntime.ConverseInput{
		ModelId:  &a.modelId,
		Messages: messages,
		InferenceConfig: &types.InferenceConfiguration{
			MaxTokens:   &maxTokens,
			Temperature: &temperature,
			TopP:        &topP,
		},
	}

	output, err := a.bedrockSvc.Converse(ctx, input)
	if err != nil {
		return "", err
	}

	response, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("unexpected response type")
	}

	contentBlock := response.Value.Content[0]
	text, ok := contentBlock.(*types.ContentBlockMemberText)
	if !ok {
		return "", fmt.Errorf("unexpected content type")
	}

	return text.Value, nil
}

// extractJSONFromResponse attempts to extract JSON from a text response
func extractJSONFromResponse(response string) string {
	// Look for JSON object in the response
	start := strings.Index(response, "{")
	if start == -1 {
		return ""
	}

	// Find the matching closing brace
	braceCount := 0
	for i := start; i < len(response); i++ {
		if response[i] == '{' {
			braceCount++
		} else if response[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return response[start : i+1]
			}
		}
	}

	return ""
}
