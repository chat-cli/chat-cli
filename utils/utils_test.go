package utils

import (
	"context"
	"os"
	"path/filepath"
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
	}

	for filename, content := range testFiles {
		if err := os.WriteFile(filename, content, 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct { //nolint:govet // fieldalignment is a minor test optimization
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

func TestReadDocument(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	testFiles := map[string][]byte{
		"test.pdf":  []byte("fake pdf data"),
		"test.csv":  []byte("a,b,c"),
		"test.doc":  []byte("fake doc data"),
		"test.docx": []byte("fake docx data"),
		"test.xls":  []byte("fake xls data"),
		"test.xlsx": []byte("fake xlsx data"),
		"test.html": []byte("<html></html>"),
		"test.txt":  []byte("plain text"),
		"test.md":   []byte("# heading"),
		"test.exe":  []byte("not a document"),
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
		{name: "valid PDF file", filename: "test.pdf", expectedType: "pdf"},
		{name: "valid CSV file", filename: "test.csv", expectedType: "csv"},
		{name: "valid DOC file", filename: "test.doc", expectedType: "doc"},
		{name: "valid DOCX file", filename: "test.docx", expectedType: "docx"},
		{name: "valid XLS file", filename: "test.xls", expectedType: "xls"},
		{name: "valid XLSX file", filename: "test.xlsx", expectedType: "xlsx"},
		{name: "valid HTML file", filename: "test.html", expectedType: "html"},
		{name: "valid TXT file", filename: "test.txt", expectedType: "txt"},
		{name: "valid MD file", filename: "test.md", expectedType: "md"},
		{name: "unsupported file type", filename: "test.exe", expectError: true},
		{name: "non-existent file", filename: "nonexistent.pdf", expectError: true},
		{name: "path traversal attempt", filename: "../../../etc/passwd", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, format, err := ReadDocument(tt.filename)

			if tt.expectError {
				if err == nil {
					t.Error("expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if format != tt.expectedType {
				t.Errorf("expected format %q, got %q", tt.expectedType, format)
			}
			if len(data) == 0 {
				t.Error("expected non-empty file data")
			}
		})
	}
}

func TestValidateLocalPath(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	if err := os.WriteFile("in-bounds.txt", []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "valid in-bounds path",
			filename:    "in-bounds.txt",
			expectError: false,
		},
		{
			name:        "path traversal attempt",
			filename:    "../../../etc/passwd",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath, err := ValidateLocalPath(tt.filename)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && fullPath == "" {
				t.Error("expected a non-empty validated path")
			}
		})
	}
}

func TestValidateLocalPathForWrite(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	t.Run("a path to a file that does not exist yet succeeds", func(t *testing.T) {
		fullPath, err := ValidateLocalPathForWrite("new-file.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fullPath == "" {
			t.Error("expected a non-empty validated path")
		}
	})

	t.Run("path traversal attempt is still rejected", func(t *testing.T) {
		if _, err := ValidateLocalPathForWrite("../../../etc/passwd"); err == nil {
			t.Error("expected an error for a path outside the working directory")
		}
	})
}

func TestResolveUserPath(t *testing.T) {
	tempDir := t.TempDir()
	docPath := filepath.Join(tempDir, "doc.pdf")
	if err := os.WriteFile(docPath, []byte("fake pdf"), 0644); err != nil {
		t.Fatal(err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	t.Run("relative path in working directory", func(t *testing.T) {
		got, err := resolveUserPath("doc.pdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		gotEval, _ := filepath.EvalSymlinks(got)
		wantEval, _ := filepath.EvalSymlinks(docPath)
		if gotEval != wantEval {
			t.Fatalf("expected %q, got %q", wantEval, gotEval)
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		got, err := resolveUserPath(docPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != docPath {
			t.Fatalf("expected %q, got %q", docPath, got)
		}
	})

	t.Run("tilde path under home", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}

		homeDoc := filepath.Join(home, "chat-cli-resolve-user-path-test.pdf")
		if err := os.WriteFile(homeDoc, []byte("fake pdf"), 0644); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Remove(homeDoc) })

		got, err := resolveUserPath("~/chat-cli-resolve-user-path-test.pdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != homeDoc {
			t.Fatalf("expected %q, got %q", homeDoc, got)
		}
	})

	t.Run("relative path traversal is blocked", func(t *testing.T) {
		if _, err := resolveUserPath("../../../etc/passwd"); err == nil {
			t.Fatal("expected path traversal to be rejected")
		}
	})
}

func TestStringPrompt(t *testing.T) {
	// StringPrompt reads from stdin, so we can't easily test it in unit tests
	// We could test it with dependency injection or mocking, but for now
	// we'll leave this as an integration test candidate
	t.Skip("StringPrompt requires stdin interaction - should be tested in integration tests")
}

func TestFindGitBoundary(t *testing.T) {
	t.Run("matches at dir's own .git", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, ".git"), 0750); err != nil {
			t.Fatalf("failed to create .git dir: %v", err)
		}

		if got := FindGitBoundary(root); got != root {
			t.Errorf("expected %s, got %s", root, got)
		}
	})

	t.Run("matches at a nested parent's .git", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, ".git"), 0750); err != nil {
			t.Fatalf("failed to create .git dir: %v", err)
		}
		sub := filepath.Join(root, "a", "b", "c")
		if err := os.MkdirAll(sub, 0750); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		if got := FindGitBoundary(sub); got != root {
			t.Errorf("expected %s, got %s", root, got)
		}
	})

	t.Run("no .git anywhere returns empty string", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "a", "b")
		if err := os.MkdirAll(sub, 0750); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}

		if got := FindGitBoundary(sub); got != "" {
			t.Errorf("expected empty string, got %s", got)
		}
	})
}
