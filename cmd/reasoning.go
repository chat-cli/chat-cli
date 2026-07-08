/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const defaultThinkingEffort = "medium"

var validThinkingEfforts = map[string]struct{}{
	"low":    {},
	"medium": {},
	"high":   {},
}

// usesAdaptiveThinking reports whether modelID expects the adaptive thinking
// request shape (thinking.type=adaptive + output_config.effort) rather than
// the legacy enabled + budget_tokens shape.
func usesAdaptiveThinking(modelID string) bool {
	id := strings.ToLower(modelID)

	adaptiveMarkers := []string{
		"claude-sonnet-5",
		"claude-sonnet-4-6",
		"claude-opus-4-6",
		"claude-opus-4-7",
		"claude-opus-4-8",
		"claude-fable-5",
	}

	for _, marker := range adaptiveMarkers {
		if strings.Contains(id, marker) {
			return true
		}
	}

	return false
}

func normalizeThinkingEffort(effort string) (string, error) {
	if effort == "" {
		return defaultThinkingEffort, nil
	}

	effort = strings.ToLower(strings.TrimSpace(effort))
	if _, ok := validThinkingEfforts[effort]; !ok {
		return "", fmt.Errorf("invalid thinking effort %q: must be low, medium, or high", effort)
	}

	return effort, nil
}

// buildReasoningConfig builds the AdditionalModelRequestFields payload that
// enables extended thinking / reasoning mode. Returns nil when disabled, so
// request shape is unchanged (NFR1) unless --thinking is set.
func buildReasoningConfig(modelID string, enabled bool, budgetTokens int32, effort string) document.Interface {
	if !enabled {
		return nil
	}

	if usesAdaptiveThinking(modelID) {
		return document.NewLazyDocument(map[string]interface{}{
			"thinking": map[string]interface{}{
				"type": "adaptive",
			},
			"output_config": map[string]interface{}{
				"effort": effort,
			},
		})
	}

	return document.NewLazyDocument(map[string]interface{}{
		"thinking": map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": budgetTokens,
		},
	})
}

// printReasoningBlock prints a finalized reasoning content block's text,
// visually distinct from the final answer (Rule 2,
// functional-design/business-logic-model.md, unit-5-extended-thinking).
// Redacted content (encrypted bytes) is intentionally not printed - it
// isn't human-readable text (Rule 5).
func printReasoningBlock(block *types.ContentBlockMemberReasoningContent) {
	reasoningText, ok := block.Value.(*types.ReasoningContentBlockMemberReasoningText)
	if !ok {
		return
	}

	fmt.Printf("\033[90m[thinking] %s\033[0m\n\n", aws.ToString(reasoningText.Value.Text))
}
