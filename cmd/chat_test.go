package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Mock BedrockRuntime client for testing
type mockBedrockRuntimeClient struct {
	ConverseOutput *bedrockruntime.ConverseOutput
	ConverseError  error
	
	ConverseStreamOutput *bedrockruntime.ConverseStreamOutput
	ConverseStreamError  error
}

func (m *mockBedrockRuntimeClient) Converse(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
	return m.ConverseOutput, m.ConverseError
}

func (m *mockBedrockRuntimeClient) ConverseStream(ctx context.Context, params *bedrockruntime.ConverseStreamInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseStreamOutput, error) {
	return m.ConverseStreamOutput, m.ConverseStreamError
}

func TestChatCommand(t *testing.T) {
	// Test that the chat command has the expected properties
	assert.Equal(t, "chat", chatCmd.Use)
	assert.Equal(t, "Start an interactive chat session", chatCmd.Short)
	assert.Contains(t, chatCmd.Long, "Begin an interactive chat session with an LLM")
}

func TestCreateContextWithRegion(t *testing.T) {
	// Test creating context with a specified region
	ctx, err := createContextWithRegion("us-west-2")
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
	
	// Cannot easily test the AWS config directly, but we can verify the context is not nil
	config := aws.BackgroundContext()
	assert.NotNil(t, config)
}

func TestConvertToConverseMessages(t *testing.T) {
	// Test converting repository chats to converse messages
	messages := convertToConverseMessages([]string{"Hello", "How are you?"})
	assert.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Hello", *messages[0].Content[0].Text)
	assert.Equal(t, "user", messages[1].Role)
	assert.Equal(t, "How are you?", *messages[1].Content[0].Text)
}