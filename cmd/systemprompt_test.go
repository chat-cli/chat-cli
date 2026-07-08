/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

func TestBuildSystemContentBlocks(t *testing.T) {
	tests := []struct {
		name         string
		systemPrompt string
		wantNil      bool
		wantText     string
	}{
		{
			name:         "empty string yields no content blocks",
			systemPrompt: "",
			wantNil:      true,
		},
		{
			name:         "non-empty string yields a single text content block",
			systemPrompt: "You are a terse assistant",
			wantNil:      false,
			wantText:     "You are a terse assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := buildSystemContentBlocks(tt.systemPrompt)

			if tt.wantNil {
				if blocks != nil {
					t.Errorf("expected nil content blocks, got %v", blocks)
				}
				return
			}

			if len(blocks) != 1 {
				t.Fatalf("expected exactly 1 content block, got %d", len(blocks))
			}

			textBlock, ok := blocks[0].(*types.SystemContentBlockMemberText)
			if !ok {
				t.Fatalf("expected block to be *types.SystemContentBlockMemberText, got %T", blocks[0])
			}

			if textBlock.Value != tt.wantText {
				t.Errorf("expected text %q, got %q", tt.wantText, textBlock.Value)
			}
		})
	}
}
