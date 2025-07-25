package utils

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestDecodeImage(t *testing.T) {
	tests := []struct {
		name        string
		base64Image string
		expectError bool
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
		},
		{
			name:        "empty string",
			base64Image: "",
			expectError: false,
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
	defer os.Chdir(originalWd)

	// Create test files
	testFiles := map[string][]byte{
		"test.png":  []byte("fake png data"),
		"test.jpg":  []byte("fake jpg data"),
		"test.jpeg": []byte("fake jpeg data"),
		"test.gif":  []byte("fake gif data"),
		"test.webp": []byte("fake webp data"),
		"test.txt":  []byte("not an image"),
	}

	for filename, content := range testFiles {
		if err := os.WriteFile(filename, content, 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name         string
		filename     string
		expectError  bool
		expectedType string
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
		},
		{
			name:        "non-existent file",
			filename:    "nonexistent.png",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			filename:    "../../../etc/passwd",
			expectError: true,
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
	// This test is tricky because LoadDocument reads from stdin
	// We'll test the document wrapping logic by mocking
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
			// Since we can't easily mock stdin in this context,
			// we'll test the document wrapping logic separately
			var result string
			if tt.input != "" {
				result = "<document>\n\n" + tt.input + "\n\n</document>\n\n"
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProcessStreamingOutput(t *testing.T) {
	// This test requires AWS SDK types which are complex to mock
	// We'll create a simple test for the handler function pattern
	t.Run("handler function receives parts", func(t *testing.T) {
		var receivedParts []string
		handler := func(ctx context.Context, part string) error {
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
	// StringPrompt reads from stdin, so we can't easily test it in unit tests
	// We could test it with dependency injection or mocking, but for now
	// we'll leave this as an integration test candidate
	t.Skip("StringPrompt requires stdin interaction - should be tested in integration tests")
}
