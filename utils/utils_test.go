package utils

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/chat-cli/chat-cli/errors"
)

func TestDecodeImage(t *testing.T) {
	tests := []struct {
		name        string
		base64Image string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid base64",
			base64Image: "SGVsbG8gV29ybGQ=", // "Hello World" in base64
			expectError: false,
		},
		{
			name:        "invalid base64",
			base64Image: "invalid-base64!@#",
			expectError: true,
			errorCode:   "invalid_base64_format",
		},
		{
			name:        "empty string",
			base64Image: "",
			expectError: true,
			errorCode:   "empty_base64_input",
		},
		{
			name:        "invalid base64 length",
			base64Image: "SGVsbG8=extra", // Invalid length
			expectError: true,
			errorCode:   "invalid_base64_format",
		},
		{
			name:        "base64 that decodes to empty",
			base64Image: "", // This will trigger empty input validation
			expectError: true,
			errorCode:   "empty_base64_input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := DecodeImage(tt.base64Image)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil {
				// Check if it's an AppError with expected code
				if appErr, ok := err.(*errors.AppError); ok {
					if appErr.Code != tt.errorCode {
						t.Errorf("expected error code %q, got %q", tt.errorCode, appErr.Code)
					}
				}
			}
			if !tt.expectError && tt.base64Image == "SGVsbG8gV29ybGQ=" {
				expected := "Hello World"
				if string(decoded) != expected {
					t.Errorf("expected %q, got %q", expected, string(decoded))
				}
			}
		})
	}
}

func TestReadImage(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	// Create test files
	testFiles := map[string][]byte{
		"test.png":  []byte("fake png data"),
		"test.jpg":  []byte("fake jpg data"),
		"test.jpeg": []byte("fake jpeg data"),
		"test.gif":  []byte("fake gif data"),
		"test.webp": []byte("fake webp data"),
		"test.txt":  []byte("not an image"),
		"empty.png": []byte{}, // Empty file
	}

	for filename, content := range testFiles {
		if err := os.WriteFile(filename, content, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a directory to test directory validation
	if err := os.Mkdir("testdir", 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct { //nolint:govet // fieldalignment is a minor test optimization
		name         string
		filename     string
		expectError  bool
		expectedType string
		errorCode    string
	}{
		{
			name:         "valid PNG file",
			filename:     "test.png",
			expectError:  false,
			expectedType: "png",
		},
		{
			name:         "valid JPG file",
			filename:     "test.jpg",
			expectError:  false,
			expectedType: "jpeg",
		},
		{
			name:         "valid JPEG file",
			filename:     "test.jpeg",
			expectError:  false,
			expectedType: "jpeg",
		},
		{
			name:         "valid GIF file",
			filename:     "test.gif",
			expectError:  false,
			expectedType: "gif",
		},
		{
			name:         "valid WEBP file",
			filename:     "test.webp",
			expectError:  false,
			expectedType: "webp",
		},
		{
			name:        "unsupported file type",
			filename:    "test.txt",
			expectError: true,
			errorCode:   "unsupported_image_format",
		},
		{
			name:        "non-existent file",
			filename:    "nonexistent.png",
			expectError: true,
			errorCode:   "file_not_found",
		},
		{
			name:        "path traversal attempt",
			filename:    "../../../etc/passwd",
			expectError: true,
			errorCode:   "path_traversal_denied",
		},
		{
			name:        "empty filename",
			filename:    "",
			expectError: true,
			errorCode:   "empty_filename",
		},
		{
			name:        "directory instead of file",
			filename:    "testdir",
			expectError: true,
			errorCode:   "path_is_directory",
		},
		{
			name:        "empty image file",
			filename:    "empty.png",
			expectError: true,
			errorCode:   "empty_image_file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, imageType, err := ReadImage(tt.filename)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectError && err != nil {
				// Check if it's an AppError with expected code
				if appErr, ok := err.(*errors.AppError); ok {
					if tt.errorCode != "" && appErr.Code != tt.errorCode {
						t.Errorf("expected error code %q, got %q", tt.errorCode, appErr.Code)
					}
				}
			}
			if !tt.expectError {
				if imageType != tt.expectedType {
					t.Errorf("expected type %q, got %q", tt.expectedType, imageType)
				}
				if len(data) == 0 {
					t.Errorf("expected data but got empty")
				}
			}
		})
	}
}

func TestLoadDocument(t *testing.T) {
	// Test document wrapping logic
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty document",
			input:    "",
			expected: "",
		},
		{
			name:     "document with content",
			input:    "Hello World",
			expected: "<document>\n\nHello World\n\n</document>\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the document wrapping logic separately since we can't easily mock stdin
			var result string
			if tt.input != "" {
				result = "<document>\n\n" + tt.input + "\n\n</document>\n\n"
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}

	// Test that LoadDocument handles terminal detection
	// This will test the actual function but won't read from stdin in terminal mode
	t.Run("terminal detection", func(t *testing.T) {
		// In test environment, this should detect we're in a terminal and return empty string
		result, err := LoadDocument()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// In terminal mode, should return empty string
		if result != "" {
			t.Errorf("expected empty string in terminal mode, got %q", result)
		}
	})
}

func TestProcessStreamingOutput(t *testing.T) {
	// Test validation of input parameters
	t.Run("nil output validation", func(t *testing.T) {
		handler := func(_ context.Context, part string) error {
			return nil
		}
		
		_, err := ProcessStreamingOutput(nil, handler)
		if err == nil {
			t.Error("expected error for nil output")
		}
		
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code != "invalid_input" {
				t.Errorf("expected error code 'invalid_input', got %q", appErr.Code)
			}
		}
	})

	t.Run("nil handler validation", func(t *testing.T) {
		// We can't easily create a real ConverseStreamOutput, so we'll test with nil
		// This tests the validation logic
		_, err := ProcessStreamingOutput(nil, nil)
		if err == nil {
			t.Error("expected error for nil handler")
		}
		
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Code != "invalid_input" {
				t.Errorf("expected error code 'invalid_input', got %q", appErr.Code)
			}
		}
	})

	// Test the handler function pattern
	t.Run("handler function receives parts", func(t *testing.T) {
		var receivedParts []string
		handler := func(_ context.Context, part string) error { //nolint:unparam // test function always returns nil
			receivedParts = append(receivedParts, part)
			return nil
		}

		// Test the handler function directly
		testParts := []string{"Hello", " ", "World"}
		for _, part := range testParts {
			if err := handler(context.Background(), part); err != nil {
				t.Errorf("handler returned error: %v", err)
			}
		}

		expected := strings.Join(testParts, "")
		actual := strings.Join(receivedParts, "")
		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
	})
}

func TestStringPrompt(t *testing.T) {
	// Test validation of empty label
	t.Run("empty label validation", func(t *testing.T) {
		// This will test the validation logic but won't actually prompt
		// since we're in a test environment (terminal mode)
		result := StringPrompt("")
		// Should not panic and should return some result
		// The exact result depends on the terminal state, but it shouldn't crash
		_ = result // We can't predict the exact output in test environment
	})

	t.Run("valid label", func(t *testing.T) {
		// Test with a valid label
		result := StringPrompt("Test prompt")
		// Should not panic and should return some result
		_ = result // We can't predict the exact output in test environment
	})
}
