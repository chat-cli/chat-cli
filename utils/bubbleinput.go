package utils

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputField model for handling text input using BubbleTea
type InputField struct {
	textarea  textarea.Model
	submitted bool
	err       error
	input     string
}

// Initialize a new InputField
func NewInputField() *InputField {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.Prompt = "│ "
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	ta.SetHeight(3)

	// Disable newlines when Enter is pressed - we'll use Enter to submit
	ta.KeyMap.InsertNewline.SetEnabled(false)

	// Customize the styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Foreground(lipgloss.Color("15"))

	// Style the placeholder by customizing the TextArea styles
	ta.Placeholder = "Type your message..."

	return &InputField{
		textarea:  ta,
		submitted: false,
		input:     "",
	}
}

// Init initializes the input field
func (m *InputField) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles UI events
func (m *InputField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Value() == "" {
				// If input is empty, treat Esc as quit
				m.input = "quit\n"
				m.submitted = true
				return m, tea.Quit
			}
			// Otherwise clear the input
			m.textarea.Reset()
			return m, nil
		case tea.KeyEnter:
			// Submit current input
			m.input = m.textarea.Value() + "\n"
			m.submitted = true
			return m, tea.Quit
		case tea.KeyCtrlC:
			// Exit program
			m.input = "quit\n"
			m.submitted = true
			return m, tea.Quit
		}
	}

	// Handle other textarea updates
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the input field
func (m *InputField) View() string {
	return fmt.Sprintf("%s", m.textarea.View())
}

// Value returns the current input text
func (m *InputField) Value() string {
	return m.input
}

// BubbleInput provides a BubbleTea-powered input field
func BubbleInput() string {
	m := NewInputField()
	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		return ""
	}

	input := m.Value()

	// Special commands handling
	input = strings.TrimSpace(input) + "\n"
	if input == "/quit\n" {
		return "quit\n"
	}

	return input
}
