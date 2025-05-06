package cmd

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
)

func TestPromptCommand(t *testing.T) {
	// Test that the prompt command has the expected properties
	assert.Equal(t, "prompt", promptCmd.Use)
	assert.Equal(t, "Send a one-time prompt to an LLM", promptCmd.Short)
	assert.Contains(t, promptCmd.Long, "Send a one-time prompt to an LLM")
	
	// Test that the prompt command has the expected flags
	modelFlag := promptCmd.Flags().Lookup("model")
	assert.NotNil(t, modelFlag)
	assert.Equal(t, "model", modelFlag.Name)
	assert.Equal(t, "m", modelFlag.Shorthand)
	assert.Equal(t, "anthropic.claude-v2", modelFlag.DefValue)
	
	promptFlag := promptCmd.Flags().Lookup("prompt")
	assert.NotNil(t, promptFlag)
	assert.Equal(t, "prompt", promptFlag.Name)
	assert.Equal(t, "p", promptFlag.Shorthand)
	assert.Equal(t, "", promptFlag.DefValue)
	
	streamFlag := promptCmd.Flags().Lookup("stream")
	assert.NotNil(t, streamFlag)
	assert.Equal(t, "stream", streamFlag.Name)
	assert.Equal(t, "s", streamFlag.Shorthand)
	assert.Equal(t, "false", streamFlag.DefValue)
}

func TestSendPrompt(t *testing.T) {
	// Mock BedrockRuntimeClient
	mockClient := &mockBedrockRuntimeClient{
		ConverseOutput: &bedrockruntime.ConverseOutput{
			Output: &bedrockruntime.ConverseOutput_Output{
				Message: &types.Message{
					Role: "assistant",
					Content: []types.ContentBlock{
						&types.ContentBlockMemberText{
							Value: "This is a test response",
						},
					},
				},
			},
		},
		ConverseError: nil,
		
		ConverseStreamOutput: &bedrockruntime.ConverseStreamOutput{
			Output: &mockEventStream{
				events: []types.ConverseStreamOutputMember{
					&types.ConverseStreamOutputMemberMessageStart{
						Value: types.MessageStart{
							Role: "assistant",
						},
					},
					&types.ConverseStreamOutputMemberContentBlockDelta{
						Value: types.ContentBlockDelta{
							Delta: &types.ContentBlockDeltaMemberText{
								Value: "This is a streaming response",
							},
						},
					},
				},
			},
		},
		ConverseStreamError: nil,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Test non-streaming prompt
	err := sendPrompt(context.Background(), mockClient, "test-model", "test prompt", false)
	
	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Read the output
	out, _ := ioutil.ReadAll(r)
	
	// Assert the results
	assert.NoError(t, err)
	assert.Contains(t, string(out), "This is a test response")
	
	// Test streaming prompt
	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	
	err = sendPrompt(context.Background(), mockClient, "test-model", "test prompt", true)
	
	w.Close()
	os.Stdout = oldStdout
	
	out, _ = ioutil.ReadAll(r)
	
	assert.NoError(t, err)
	assert.Contains(t, string(out), "This is a streaming response")
	
	// Test error cases
	mockClient.ConverseError = errors.New("converse error")
	mockClient.ConverseStreamError = errors.New("stream error")
	
	err = sendPrompt(context.Background(), mockClient, "test-model", "test prompt", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "converse error")
	
	err = sendPrompt(context.Background(), mockClient, "test-model", "test prompt", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream error")
}