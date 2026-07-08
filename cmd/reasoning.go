/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// buildReasoningConfig builds the AdditionalModelRequestFields payload that
// enables extended thinking / reasoning mode. Returns nil when disabled, so
// request shape is unchanged (NFR1) unless --thinking is set.
//
// WARNING: AdditionalModelRequestFields is an untyped, provider-specific
// field - unlike every other request shape in this codebase, this one
// could not be verified against the SDK's type definitions (see
// functional-design/business-logic-model.md, unit-5-extended-thinking).
// This shape is a best-effort assumption, not a confirmed contract.
func buildReasoningConfig(enabled bool, budgetTokens int32) document.Interface {
	if !enabled {
		return nil
	}

	return document.NewLazyDocument(map[string]interface{}{
		"reasoning_config": map[string]interface{}{
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
