/*
Copyright © 2024 Micah Walter
*/
package cmd

import "testing"

func TestIsInferenceProfileID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"us.anthropic.claude-sonnet-5", true},
		{"global.anthropic.claude-sonnet-5", true},
		{"arn:aws:bedrock:us-east-1:123:inference-profile/us.anthropic.claude-sonnet-5", true},
		{"anthropic.claude-sonnet-4-20250514-v1:0", false},
		{"anthropic.claude-3-5-sonnet-20240620-v1:0", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := isInferenceProfileID(tt.id); got != tt.want {
				t.Fatalf("isInferenceProfileID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}
