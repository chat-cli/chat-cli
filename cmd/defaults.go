/*
Copyright © 2024 Micah Walter
*/
package cmd

import "strings"

// DefaultModelID is the built-in default Bedrock model when neither --model-id
// nor --custom-arn (or persisted config) is set. Sonnet 5 is invoked via its
// US inference profile ID because on-demand foundation-model IDs are not
// supported for this model.
const DefaultModelID = "us.anthropic.claude-sonnet-5"

// isInferenceProfileID reports whether id is a Bedrock inference profile
// identifier (or cross-region prefix) rather than a foundation model ID.
// GetFoundationModel cannot look these up, so validation is skipped and the
// id is passed through to Converse directly.
func isInferenceProfileID(id string) bool {
	switch {
	case strings.HasPrefix(id, "arn:aws:bedrock:"):
		return strings.Contains(id, ":inference-profile/")
	case strings.HasPrefix(id, "us."), strings.HasPrefix(id, "global."), strings.HasPrefix(id, "eu."):
		return true
	default:
		return false
	}
}
