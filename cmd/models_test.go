package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestModelsCommand(t *testing.T) {
	// Test that the models command has the expected properties
	assert.Equal(t, "models", modelsCmd.Use)
	assert.Equal(t, "List available LLMs on Amazon Bedrock", modelsCmd.Short)
	assert.Contains(t, modelsCmd.Long, "List all available LLMs on Amazon Bedrock")
}

func TestModelsListCommand(t *testing.T) {
	// Test that the models list command has the expected properties
	assert.Equal(t, "list", modelsListCmd.Use)
	assert.Equal(t, "List available LLMs on Amazon Bedrock", modelsListCmd.Short)
	assert.Contains(t, modelsListCmd.Long, "List all available LLMs on Amazon Bedrock")
}

func TestGetModels(t *testing.T) {
	// Mock BedrockClient
	mockClient := &mockBedrockClient{
		ListFoundationModelsOutput: &bedrock.ListFoundationModelsOutput{
			ModelSummaries: []types.FoundationModelSummary{
				{
					ModelId: aws.String("anthropic.claude-v2"),
					ModelName: aws.String("Claude V2"),
				},
				{
					ModelId: aws.String("amazon.titan-text"),
					ModelName: aws.String("Titan Text"),
				},
			},
		},
		ListFoundationModelsErr: nil,
	}

	// Mock stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Get models
	models, err := getModels(context.Background(), mockClient)
	
	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Read the output
	out, _ := ioutil.ReadAll(r)
	
	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, models, 2)
	assert.Equal(t, "anthropic.claude-v2", *models[0].ModelId)
	assert.Equal(t, "Claude V2", *models[0].ModelName)
	assert.Equal(t, "amazon.titan-text", *models[1].ModelId)
	assert.Equal(t, "Titan Text", *models[1].ModelName)
}

// Mock BedrockClient for testing
type mockBedrockClient struct {
	ListFoundationModelsOutput *bedrock.ListFoundationModelsOutput
	ListFoundationModelsErr    error
}

func (m *mockBedrockClient) ListFoundationModels(ctx context.Context, params *bedrock.ListFoundationModelsInput, optFns ...func(*bedrock.Options)) (*bedrock.ListFoundationModelsOutput, error) {
	return m.ListFoundationModelsOutput, m.ListFoundationModelsErr
}