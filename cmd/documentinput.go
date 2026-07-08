/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

var disallowedDocumentNameChars = regexp.MustCompile(`[^A-Za-z0-9 ()\[\]-]`)

// sanitizeDocumentName derives a neutral document name for Bedrock's
// DocumentBlock.Name field from a filename. The raw filename can't be used
// directly: DocumentBlock.Name only allows alphanumeric, whitespace,
// hyphens, parentheses, and square brackets (a "." isn't valid), and the SDK
// itself flags unsanitized names as a prompt-injection vector (Functional
// Design Decision 1, unit-4-document-input).
func sanitizeDocumentName(filename string) string {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	cleaned := disallowedDocumentNameChars.ReplaceAllString(base, " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	if cleaned == "" {
		return "attached-document"
	}
	return cleaned
}

// buildDocumentContentBlock builds a Bedrock document content block from
// document bytes, a format string (as returned by utils.ReadDocument), and a
// sanitized name.
func buildDocumentContentBlock(data []byte, format, name string) *types.ContentBlockMemberDocument {
	return &types.ContentBlockMemberDocument{
		Value: types.DocumentBlock{
			Name:   aws.String(name),
			Format: types.DocumentFormat(format),
			Source: &types.DocumentSourceMemberBytes{Value: data},
		},
	}
}
