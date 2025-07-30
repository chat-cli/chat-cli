/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	conf "github.com/chat-cli/chat-cli/config"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// isCapacitySuffix checks if a string represents a capacity/context size suffix
func isCapacitySuffix(s string) bool {
	// Filter out capacity suffixes like "24k", "200k", "1000k", "4k", "8k", "12k", "28k", "48k", "128k", "300k", "512"
	if strings.HasSuffix(s, "k") {
		return true
	}
	// Filter out purely numeric suffixes that are likely capacities (but keep version numbers like 0, 1, 2)
	if isNumeric(s) {
		num, _ := strconv.Atoi(s)
		// Keep small version numbers (0, 1, 2, etc.) but filter larger capacity numbers
		return num > 10
	}
	// Filter out specific capacity suffixes
	capacitySuffixes := []string{"mm"} // multimodal suffix
	for _, suffix := range capacitySuffixes {
		if s == suffix {
			return true
		}
	}
	return false
}

// isNumeric checks if a string contains only numeric characters
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// requiresCrossRegionProfile checks if a model requires cross-region inference profile
func requiresCrossRegionProfile(modelID, modelArn string) bool {
	// First check if the ARN explicitly indicates cross-region capability
	if modelArn != "" && (strings.Contains(modelArn, "inference-profile") || strings.Contains(modelArn, "us.")) {
		return true
	}

	// Check for specific models that require cross-region inference profiles
	// These are models that are only available through inference profiles
	crossRegionModels := []string{
		"anthropic.claude-sonnet-4-20250514-v1:0",
		"anthropic.claude-opus-4-20250514-v1:0",
		"anthropic.claude-3-7-sonnet-20250219-v1:0",
		// Add other models that require cross-region profiles as they're released
	}

	for _, crModel := range crossRegionModels {
		if modelID == crModel {
			return true
		}
	}

	return false
}

// generateInferenceProfileArn generates the correct inference profile ARN for cross-region models
func generateInferenceProfileArn(modelID string) string {
	// Map specific models to their known inference profile ARNs
	// These ARNs are typically in the format: arn:aws:bedrock:us:anthropic::inference-profile/...
	inferenceProfileMap := map[string]string{
		"anthropic.claude-sonnet-4-20250514-v1:0":   "arn:aws:bedrock:us:anthropic::inference-profile/claude-sonnet-4-20250514-v1:0",
		"anthropic.claude-opus-4-20250514-v1:0":     "arn:aws:bedrock:us:anthropic::inference-profile/claude-opus-4-20250514-v1:0",
		"anthropic.claude-3-7-sonnet-20250219-v1:0": "arn:aws:bedrock:us:anthropic::inference-profile/claude-3-7-sonnet-20250219-v1:0",
	}

	if profileArn, exists := inferenceProfileMap[modelID]; exists {
		return profileArn
	}

	// Fallback: if we don't have a specific mapping, try to construct a reasonable ARN
	// This is a best-effort attempt and may need adjustment based on actual AWS patterns
	return fmt.Sprintf("arn:aws:bedrock:us:anthropic::inference-profile/%s", modelID)
}

type item struct {
	modelID     string
	modelName   string
	provider    string
	status      string
	crossRegion bool
	modelArn    string
}

func (i item) Title() string { return i.modelName }
func (i item) Description() string {
	desc := fmt.Sprintf("%s • %s", i.provider, i.modelID)
	if i.crossRegion {
		desc += " • Cross-Region"
	}
	return desc
}
func (i item) FilterValue() string { return i.modelName + " " + i.provider + " " + i.modelID }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			selectedItem, ok := m.list.SelectedItem().(item)
			if ok {
				// For cross-region models, set custom-arn instead of model-id
				if selectedItem.crossRegion {
					// Generate the correct inference profile ARN for cross-region models
					inferenceProfileArn := generateInferenceProfileArn(selectedItem.modelID)
					if err := setCustomArnInConfig(inferenceProfileArn); err != nil {
						fmt.Printf("Error setting model: %v\n", err)
						return m, tea.Quit
					}
					fmt.Printf("✓ Cross-region model set via inference profile: %s\n", selectedItem.modelID)
				} else {
					// Regular models use model-id
					if err := setModelInConfig(selectedItem.modelID); err != nil {
						fmt.Printf("Error setting model: %v\n", err)
						return m, tea.Quit
					}
					fmt.Printf("✓ Model set to: %s\n", selectedItem.modelID)
				}
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func setModelInConfig(modelID string) error {
	// Initialize configuration
	fm, err := conf.NewFileManager("chat-cli")
	if err != nil {
		return err
	}

	if initErr := fm.InitializeViper(); initErr != nil {
		return initErr
	}

	// Clear custom-arn if it exists since we're setting model-id
	if viper.IsSet("custom-arn") {
		viper.Set("custom-arn", "")
	}

	// Set the model-id
	viper.Set("model-id", modelID)

	// Write the configuration to file
	if err := viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

func setCustomArnInConfig(modelArn string) error {
	// Initialize configuration
	fm, err := conf.NewFileManager("chat-cli")
	if err != nil {
		return err
	}

	if initErr := fm.InitializeViper(); initErr != nil {
		return initErr
	}

	// Clear model-id if it exists since we're setting custom-arn
	if viper.IsSet("model-id") {
		viper.Set("model-id", "")
	}

	// Set the custom-arn
	viper.Set("custom-arn", modelArn)

	// Write the configuration to file
	if err := viper.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// modelsInteractiveCmd represents the interactive models command
var modelsInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactively select and configure a model",
	Long:  `Display an interactive list of available models and set the selected one as the default.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInteractiveModelSelector(); err != nil {
			log.Fatalf("Error running interactive model selector: %v", err)
		}
	},
}

func runInteractiveModelSelector() error {
	return runInteractiveModelSelectorWithRegion("us-east-1")
}

func runInteractiveModelSelectorWithRegion(region string) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %w", err)
	}

	// Create Bedrock client
	svc := bedrock.NewFromConfig(cfg)

	// Fetch models
	result, err := svc.ListFoundationModels(context.TODO(), &bedrock.ListFoundationModelsInput{})
	if err != nil {
		return fmt.Errorf("error listing models: %w", err)
	}

	// Convert models to list items, filtering for active base models only
	var items []list.Item
	seenModels := make(map[string]bool) // Track models to avoid duplicates

	for i := range result.ModelSummaries {
		model := &result.ModelSummaries[i]

		// Only include active models
		if model.ModelLifecycle != nil && string(model.ModelLifecycle.Status) == "ACTIVE" {
			modelID := aws.ToString(model.ModelId)

			// Filter out model variants with capacity/context size suffixes
			// Keep base models and legitimate version identifiers
			if strings.Contains(modelID, ":") {
				parts := strings.Split(modelID, ":")
				if len(parts) >= 3 {
					// For patterns like "model:version:suffix", check the last part
					lastPart := parts[len(parts)-1]
					if isCapacitySuffix(lastPart) {
						continue
					}
				} else if len(parts) == 2 {
					// For patterns like "model:suffix", check if suffix is a capacity
					suffix := parts[1]
					if isCapacitySuffix(suffix) {
						continue
					}
				}
			}

			// Avoid duplicate base models
			if seenModels[modelID] {
				continue
			}
			seenModels[modelID] = true

			// Check for cross-region inference capability
			crossRegion := false
			modelArn := aws.ToString(model.ModelArn)

			// Check if this model requires cross-region inference profile
			if requiresCrossRegionProfile(modelID, modelArn) {
				crossRegion = true
			}

			items = append(items, item{
				modelID:     modelID,
				modelName:   aws.ToString(model.ModelName),
				provider:    aws.ToString(model.ProviderName),
				status:      "", // Remove status since we only show active models
				crossRegion: crossRegion,
				modelArn:    modelArn,
			})
		}
	}

	// Sort items by provider first, then by model name within each provider
	sort.Slice(items, func(i, j int) bool {
		itemI := items[i].(item)
		itemJ := items[j].(item)

		// First sort by provider
		if itemI.provider != itemJ.provider {
			return itemI.provider < itemJ.provider
		}

		// Within same provider, sort by model name
		return itemI.modelName < itemJ.modelName
	})

	// Create list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a Model"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Bold(true)

	m := model{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}

	return nil
}

func init() {
	modelsCmd.AddCommand(modelsInteractiveCmd)
}
