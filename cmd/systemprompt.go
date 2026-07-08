/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// buildSystemContentBlocks builds the Bedrock SystemContentBlocks for a
// Converse/ConverseStream request from a resolved system prompt string.
// An empty systemPrompt yields no content blocks, so request shape is
// unchanged from before system prompt support existed.
func buildSystemContentBlocks(systemPrompt string) []types.SystemContentBlock {
	if systemPrompt == "" {
		return nil
	}

	return []types.SystemContentBlock{
		&types.SystemContentBlockMemberText{
			Value: systemPrompt,
		},
	}
}
