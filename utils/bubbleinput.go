package utils

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
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
	ta.Prompt = "> "
	ta.ShowLineNumbers = false

	// Get terminal width
	width, _, err := term.GetSize(int(syscall.Stdin))
	if err != nil || width <= 0 {
		// Default if we can't get terminal width
		width = 80
	}

	// Use terminal width with padding
	padding := 4 // 2 chars on each side
	ta.SetWidth(width - padding)

	// Start with a single line and allow expansion
	ta.SetHeight(1)

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
	// Handle terminal resize events
	case tea.WindowSizeMsg:
		// Update width based on new terminal size with padding
		padding := 4 // 2 chars on each side
		m.textarea.SetWidth(msg.Width - padding)
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

// clearInputBox clears just the current input box and moves cursor up
func clearInputBox() {
	// Get terminal height to determine how many lines to move up
	_, height, err := term.GetSize(int(syscall.Stdin))

	// Default to moving up 3 lines if we can't get terminal info
	linesToMoveUp := 3

	// If we can get terminal info, calculate a reasonable number of lines
	if err == nil && height > 0 {
		// Assume the input box takes at most 3 lines
		linesToMoveUp = 3
	}

	// Move cursor up linesToMoveUp lines and clear those lines
	for i := 0; i < linesToMoveUp; i++ {
		// Move cursor up one line
		fmt.Print("\033[1A")
		// Clear entire line
		fmt.Print("\033[2K")
	}
}

// BubbleInput provides a BubbleTea-powered input field
func BubbleInput() (string, bool) {
	m := NewInputField()
	p := tea.NewProgram(m)

	_, err := p.Run()
	if err != nil {
		return "", false
	}

	input := m.Value()

	// Clear just the input box lines
	clearInputBox()

	// Special commands handling
	input = strings.TrimSpace(input) + "\n"
	if input == "/quit\n" {
		return "quit\n", true
	}

	return input, true
}
