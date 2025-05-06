package utils

import (
	"testing"
)

// Testing utils functions
func TestStringPrompt(t *testing.T) {
	// This function requires user input via os.Stdin
	// We'll only test that it exists and can be called
	// A more comprehensive test would require mocking os.Stdin
	t.Run("StringPrompt function exists", func(t *testing.T) {
		// Just verify the function exists
		var _ = StringPrompt
	})
}

func TestDecodeImage(t *testing.T) {
	t.Run("DecodeImage with invalid input", func(t *testing.T) {
		_, err := DecodeImage("not-base64")
		if err == nil {
			t.Error("Expected error when decoding invalid base64, got nil")
		}
	})
}

func TestLoadDocument(t *testing.T) {
	// This function reads from a file
	// A proper test would require a test file
	t.Run("LoadDocument function exists", func(t *testing.T) {
		var _ = LoadDocument
	})
}