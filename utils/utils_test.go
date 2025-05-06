package utils

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeImage(t *testing.T) {
	// Test case for valid base64 string
	validBase64 := base64.StdEncoding.EncodeToString([]byte("test image data"))
	data, err := DecodeImage(validBase64)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test image data"), data)

	// Test case for invalid base64 string
	invalidBase64 := "not a valid base64 string"
	_, err = DecodeImage(invalidBase64)
	assert.Error(t, err)
}

func TestReadImage(t *testing.T) {
	// Create a temporary image file
	tmpDir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	imgPath := filepath.Join(tmpDir, "test.jpg")
	err = ioutil.WriteFile(imgPath, []byte("fake image data"), 0644)
	assert.NoError(t, err)

	// We need to be in the same directory as the image for the test
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(currentDir)
	
	err = os.Chdir(tmpDir)
	assert.NoError(t, err)

	// Test reading a valid image
	data, contentType, err := ReadImage("test.jpg")
	assert.NoError(t, err)
	assert.Equal(t, []byte("fake image data"), data)
	assert.Equal(t, "image/jpeg", contentType)

	// Test reading a non-existent image
	_, _, err = ReadImage("non-existent.jpg")
	assert.Error(t, err)
	
	// Test reading an image with invalid extension
	invalidPath := filepath.Join(".", "test.invalid")
	err = ioutil.WriteFile(invalidPath, []byte("fake image data"), 0644)
	assert.NoError(t, err)
	
	_, _, err = ReadImage("test.invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
	
	// Test reading a file outside the allowed directory
	_, _, err = ReadImage("../outside.jpg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside of the allowed directory")
}

// mockEventStream implements types.ConverseStreamOutput_stream
type mockEventStream struct {
	events []types.ConverseStreamOutputMember
}

func (m *mockEventStream) Events() <-chan types.ConverseStreamOutputMember {
	ch := make(chan types.ConverseStreamOutputMember, len(m.events))
	for _, event := range m.events {
		ch <- event
	}
	close(ch)
	return ch
}

func TestProcessStreamingOutput(t *testing.T) {
	// Create a mock event stream
	mockEventStream := &mockEventStream{
		events: []types.ConverseStreamOutputMember{
			&types.ConverseStreamOutputMemberMessageStart{
				Value: types.MessageStart{
					Role: "assistant",
				},
			},
			&types.ConverseStreamOutputMemberContentBlockDelta{
				Value: types.ContentBlockDelta{
					Delta: &types.ContentBlockDeltaMemberText{
						Value: "Hello, ",
					},
				},
			},
			&types.ConverseStreamOutputMemberContentBlockDelta{
				Value: types.ContentBlockDelta{
					Delta: &types.ContentBlockDeltaMemberText{
						Value: "world!",
					},
				},
			},
		},
	}

	// Create the mock stream output
	output := &bedrockruntime.ConverseStreamOutput{
		Output: mockEventStream,
	}

	var collectedText string
	handler := func(ctx context.Context, part string) error {
		collectedText += part
		return nil
	}

	message, err := ProcessStreamingOutput(output, handler)
	assert.NoError(t, err)
	assert.Equal(t, "assistant", message.Role)
	assert.Equal(t, "Hello, world!", *message.Content[0].Text)
	assert.Equal(t, "Hello, world!", collectedText)
}

func TestStringPrompt(t *testing.T) {
	// Save original stdin and restore it after the test
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate user input
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdin = r

	// Write test input to the pipe
	go func() {
		w.Write([]byte("test input\n"))
		w.Close()
	}()

	// Capture stderr to test label output
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	
	r2, w2, err := os.Pipe()
	assert.NoError(t, err)
	os.Stderr = w2
	
	// Test the prompt
	result := StringPrompt("Enter:")
	
	// Close the pipe to capture the output
	w2.Close()
	var buf strings.Builder
	_, err = buf.ReadFrom(r2)
	assert.NoError(t, err)
	
	// Verify the output contains the label
	assert.Contains(t, buf.String(), "Enter:")
	
	// Verify the input is captured correctly (with the newline)
	assert.Equal(t, "test input\n", result)
}

func TestLoadDocument(t *testing.T) {
	// Save original stdin and restore it after the test
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate document input
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdin = r

	// Write test document to the pipe
	testDoc := "This is a test document"
	go func() {
		w.Write([]byte(testDoc))
		w.Close()
	}()

	// Test loading the document
	doc, err := LoadDocument()
	assert.NoError(t, err)
	assert.Equal(t, testDoc, doc)
}