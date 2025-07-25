package utils

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test the creation and initialization of the InputField model
func TestNewInputField(t *testing.T) {
	inputField := NewInputField()

	if inputField == nil {
		t.Fatal("NewInputField returned nil")
	}

	if inputField.textarea.Placeholder != "Type your message..." {
		t.Errorf("Expected placeholder 'Type your message...', got '%s'", inputField.textarea.Placeholder)
	}

	if inputField.textarea.Prompt != "> " {
		t.Errorf("Expected prompt '> ', got '%s'", inputField.textarea.Prompt)
	}

	if inputField.textarea.ShowLineNumbers {
		t.Error("Line numbers should be disabled")
	}

	// Verify that newlines are disabled for Enter key
	if inputField.textarea.KeyMap.InsertNewline.Enabled() {
		t.Error("InsertNewline should be disabled")
	}
}

// Test basic model methods without requiring user interaction
func TestInputFieldBasics(t *testing.T) {
	inputField := NewInputField()

	// Test Init
	cmd := inputField.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}

	// Test View
	view := inputField.View()
	if view == "" {
		t.Error("View should not return empty string")
	}
}

// TestBubbleInput itself is hard to test without user interaction
// Similar to StringPrompt, we'll need integration tests for full coverage
func TestBubbleInput(t *testing.T) {
	t.Skip("BubbleInput requires user interaction - should be tested in integration tests")
}

// Test handling of escape key
func TestInputFieldEscape(t *testing.T) {
	inputField := NewInputField()

	// Test Escape when input is empty
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	model, cmd := inputField.Update(escMsg)

	updatedField, ok := model.(*InputField)
	if !ok {
		t.Fatal("Update should return *InputField")
	}

	if !updatedField.submitted {
		t.Error("Escape on empty input should set submitted to true")
	}

	if updatedField.input != "quit\n" {
		t.Errorf("Escape on empty input should set input to 'quit\\n', got '%s'", updatedField.input)
	}

	if cmd == nil {
		t.Error("Escape should return a quit command")
	}
}
