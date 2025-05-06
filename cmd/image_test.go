package cmd

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
)

func TestImageCommand(t *testing.T) {
	// Test that the image command has the expected properties
	assert.Equal(t, "image", imageCmd.Use)
	assert.Equal(t, "Generate an image with LLMs", imageCmd.Short)
	assert.Contains(t, imageCmd.Long, "Generate an image using an LLM")
	
	// Test that the image command has the expected flags
	modelFlag := imageCmd.Flags().Lookup("model")
	assert.NotNil(t, modelFlag)
	assert.Equal(t, "model", modelFlag.Name)
	assert.Equal(t, "m", modelFlag.Shorthand)
	assert.Equal(t, "amazon.titan-image-generator-v1", modelFlag.DefValue)
	
	promptFlag := imageCmd.Flags().Lookup("prompt")
	assert.NotNil(t, promptFlag)
	assert.Equal(t, "prompt", promptFlag.Name)
	assert.Equal(t, "p", promptFlag.Shorthand)
	assert.Equal(t, "", promptFlag.DefValue)
	
	negativePromptFlag := imageCmd.Flags().Lookup("negative-prompt")
	assert.NotNil(t, negativePromptFlag)
	assert.Equal(t, "negative-prompt", negativePromptFlag.Name)
	assert.Equal(t, "n", negativePromptFlag.Shorthand)
	assert.Equal(t, "", negativePromptFlag.DefValue)
}

func TestCreateImage(t *testing.T) {
	// Mock BedrockRuntimeClient
	mockClient := &mockBedrockRuntimeClient{
		InvokeModelWithResponseStreamOutput: &bedrockruntime.InvokeModelWithResponseStreamOutput{
			Body: ioutil.NopCloser(bytes.NewReader([]byte("fake-image-data"))),
		},
		InvokeModelWithResponseStreamError: nil,
	}

	// Test successful image creation
	imageData, err := createImage(context.Background(), mockClient, "test-model", "test prompt", "negative prompt")
	assert.NoError(t, err)
	assert.Equal(t, []byte("fake-image-data"), imageData)
	
	// Test error case
	mockClient.InvokeModelWithResponseStreamError = errors.New("api error")
	imageData, err = createImage(context.Background(), mockClient, "test-model", "test prompt", "negative prompt")
	assert.Error(t, err)
	assert.Nil(t, imageData)
	assert.Contains(t, err.Error(), "api error")
}

// Extend the mock BedrockRuntimeClient for image generation
func (m *mockBedrockRuntimeClient) InvokeModelWithResponseStream(
	ctx context.Context, 
	params *bedrockruntime.InvokeModelWithResponseStreamInput, 
	optFns ...func(*bedrockruntime.Options),
) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error) {
	return m.InvokeModelWithResponseStreamOutput, m.InvokeModelWithResponseStreamError
}