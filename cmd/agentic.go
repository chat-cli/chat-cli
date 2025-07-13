package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chat-cli/chat-cli/agents"
	"github.com/chat-cli/chat-cli/db"
	"github.com/chat-cli/chat-cli/factory"
	"github.com/chat-cli/chat-cli/repository"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"

	conf "github.com/chat-cli/chat-cli/config"
)

// wrapText wraps text to fit within the specified width, preserving newlines
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	lines := strings.Split(text, "\n")
	var result strings.Builder
	
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		
		// Handle empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}
		
		currentLine := words[0]
		// Handle words longer than width
		if len(currentLine) > width {
			result.WriteString(currentLine)
			currentLine = ""
		}
		
		for _, word := range words[1:] {
			if len(word) > width {
				// Word is too long, just add it on its own line
				if currentLine != "" {
					result.WriteString(currentLine + "\n")
				}
				result.WriteString(word)
				currentLine = ""
			} else if len(currentLine)+len(word)+1 <= width {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine + "\n")
				currentLine = word
			}
		}
		
		if currentLine != "" {
			result.WriteString(currentLine)
		}
	}
	
	return result.String()
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type agenticModel struct {
	input      string
	cursor     int
	responses  []string
	thinking   bool
	spinner    spinner.Model
	agent      *agents.FileEditAgent
	chatRepo   *repository.ChatRepository
	sessionId  string
	fm         *conf.FileManager
	width      int
	height     int
}

func (m agenticModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.spinner.Tick)
}

func (m agenticModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if strings.TrimSpace(m.input) == "" {
				return m, nil
			}
			if strings.TrimSpace(m.input) == "quit" {
				return m, tea.Quit
			}
			m.thinking = true
			return m, m.executeTask()
		case "backspace":
			if m.cursor > 0 {
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.input) {
				m.cursor++
			}
		default:
			if len(msg.String()) == 1 {
				m.input = m.input[:m.cursor] + msg.String() + m.input[m.cursor:]
				m.cursor++
			}
		}
	case taskResult:
		m.thinking = false
		response := msg.response
		if msg.success {
			response = fmt.Sprintf("✅ %s", response)
		} else {
			response = fmt.Sprintf("❌ %s", response)
		}
		m.responses = append(m.responses, fmt.Sprintf("Task: %s", msg.task))
		m.responses = append(m.responses, response)
		
		// Save to database
		m.saveToDatabase(msg.task, response)
		
		// Clear input
		m.input = ""
		m.cursor = 0
	case spinner.TickMsg:
		if m.thinking {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}
	
	// Always update spinner if we're thinking
	if m.thinking {
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	
	return m, nil
}

func (m agenticModel) View() string {
	// Use default width if terminal size not detected yet
	width := m.width
	if width <= 0 {
		width = 80
	}
	
	// Calculate content width accounting for borders and padding
	contentWidth := width - 6 // Account for borders and padding
	if contentWidth < 20 {
		contentWidth = 20 // Minimum width
	}
	
	// Create responsive styles
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1).
		Width(width - 4) // Account for border width
	
	responseStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1).
		Margin(1, 0).
		Width(width - 4) // Account for border width
	
	title := titleStyle.Render("🤖 Agentic Chat CLI")
	
	var responseText string
	if len(m.responses) > 0 {
		for _, resp := range m.responses {
			// Wrap text to fit within content area
			wrappedResp := wrapText(resp, contentWidth)
			responseText += responseStyle.Render(wrappedResp) + "\n"
		}
	}
	
	var inputPrompt string
	if m.thinking {
		spinnerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)
		promptText := fmt.Sprintf("%s Processing your request...", spinnerStyle.Render(m.spinner.View()))
		wrappedPrompt := wrapText(promptText, contentWidth)
		inputPrompt = inputStyle.Render(wrappedPrompt)
	} else {
		inputText := m.input
		if m.cursor < len(inputText) {
			inputText = inputText[:m.cursor] + "│" + inputText[m.cursor:]
		} else {
			inputText += "│"
		}
		promptText := fmt.Sprintf("Enter task: %s", inputText)
		wrappedPrompt := wrapText(promptText, contentWidth)
		inputPrompt = inputStyle.Render(wrappedPrompt)
	}
	
	helpText := "Press Esc/Ctrl+C to quit, Enter to execute task, type 'quit' to exit"
	wrappedHelp := wrapText(helpText, width)
	help := helpStyle.Render(wrappedHelp)
	
	return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", title, responseText, inputPrompt, help)
}

type taskResult struct {
	task     string
	response string
	success  bool
}

func (m agenticModel) executeTask() tea.Cmd {
	task := strings.TrimSpace(m.input)
	
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.agent.Execute(context.Background(), task, nil)
			if err != nil {
				return taskResult{
					task:     task,
					response: fmt.Sprintf("Execution failed: %v", err),
					success:  false,
				}
			}
			
			response := result.Message
			if result.Error != "" {
				response = result.Error
			}
			
			// Add tool results to response if available
			if len(result.ToolResults) > 0 {
				response += "\n\n🔧 Tool Results:\n"
				for _, toolResult := range result.ToolResults {
					status := "✅"
					if !toolResult.Success {
						status = "❌"
					}
					response += fmt.Sprintf("  %s %s", status, toolResult.ToolName)
					if toolResult.Error != "" {
						response += fmt.Sprintf(" - %s", toolResult.Error)
					}
					response += "\n"
				}
			}
			
			return taskResult{
				task:     task,
				response: response,
				success:  result.Success,
			}
		},
	)
}

func (m agenticModel) saveToDatabase(task, response string) {
	if m.chatRepo == nil {
		return
	}
	
	// Save user task
	userChat := &repository.Chat{
		ChatId:  m.sessionId,
		Persona: "User",
		Message: task,
	}
	if err := m.chatRepo.Create(userChat); err != nil {
		log.Printf("Failed to save user message: %v", err)
	}
	
	// Save assistant response
	assistantChat := &repository.Chat{
		ChatId:  m.sessionId,
		Persona: "Assistant",
		Message: response,
	}
	if err := m.chatRepo.Create(assistantChat); err != nil {
		log.Printf("Failed to save assistant message: %v", err)
	}
}

// agenticCmd represents the agentic command for quick file operations
var agenticCmd = &cobra.Command{
	Use:   "agentic",
	Short: "Start an interactive agentic session",
	Long: `Start an interactive agentic session using AI agents for file operations.

The TUI provides:
- Interactive task input with visual feedback
- Session history saved to database
- Elegant display of agent responses
- Tool execution results

Press Esc/Ctrl+C to quit, Enter to execute tasks, type 'quit' to exit.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize configuration
		fm, err := conf.NewFileManager("chat-cli")
		if err != nil {
			log.Fatal(err)
		}

		if err := fm.InitializeViper(); err != nil {
			log.Fatal(err)
		}

		// Get SQLite database path
		dbPath := fm.GetDBPath()
		driver := fm.GetDBDriver()

		// Get configuration values
		region, err := cmd.Parent().PersistentFlags().GetString("region")
		if err != nil {
			log.Fatalf("unable to get region flag: %v", err)
		}

		modelIdFlag, err := cmd.PersistentFlags().GetString("model-id")
		if err != nil {
			modelIdFlag = ""
		}

		sessionIdFlag, err := cmd.PersistentFlags().GetString("session-id")
		if err != nil {
			log.Fatalf("unable to get session-id flag: %v", err)
		}

		// Get configuration values with precedence order (flag -> config -> default)
		modelId := fm.GetConfigValue("model-id", modelIdFlag, "anthropic.claude-3-sonnet-20240229-v1:0").(string)

		// Create file edit agent
		fileAgent, err := agents.NewFileEditAgent(region, modelId)
		if err != nil {
			log.Fatalf("Failed to create file agent: %v", err)
		}

		// Initialize database
		config := db.Config{
			Driver: driver,
			Name:   dbPath,
		}

		database, err := factory.CreateDatabase(config)
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		defer database.Close()

		// Run migrations to ensure tables exist
		if err := database.Migrate(); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		// Create repositories
		chatRepo := repository.NewChatRepository(database)

		// Generate session ID if not provided
		var sessionId string
		if sessionIdFlag == "" {
			sessionUUID := uuid.NewV4()
			sessionId = sessionUUID.String()
		} else {
			sessionId = sessionIdFlag
		}

		// Initialize spinner
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))

		// Create model
		model := agenticModel{
			agent:     fileAgent,
			chatRepo:  chatRepo,
			sessionId: sessionId,
			fm:        fm,
			spinner:   s,
			width:     80,  // Default width
			height:    24,  // Default height
		}

		// Load previous session if exists
		if sessionIdFlag != "" {
			if chats, err := chatRepo.GetMessages(sessionId); err != nil {
				log.Printf("Failed to load session: %v", err)
			} else {
				for _, chat := range chats {
					if chat.Persona == "User" {
						model.responses = append(model.responses, fmt.Sprintf("Task: %s", chat.Message))
					} else {
						model.responses = append(model.responses, chat.Message)
					}
				}
			}
		}

		// Start the TUI
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("TUI error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(agenticCmd)
	agenticCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-sonnet-20240229-v1:0", "set the model id")
	agenticCmd.PersistentFlags().String("session-id", "", "pass a valid session-id to load a previous conversation")
}
