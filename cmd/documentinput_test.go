/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

func TestSanitizeDocumentName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{name: "simple filename", filename: "report.pdf", want: "report"},
		{name: "path is reduced to base name", filename: "/some/dir/report.pdf", want: "report"},
		{name: "disallowed characters become spaces", filename: "my_file!.txt", want: "my file"},
		{name: "all disallowed characters falls back to a neutral name", filename: "!!!.pdf", want: "attached-document"},
		{name: "hyphens and parentheses are allowed", filename: "report (final)-v2.pdf", want: "report (final)-v2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeDocumentName(tt.filename)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestBuildDocumentContentBlock(t *testing.T) {
	block := buildDocumentContentBlock([]byte("hello"), "pdf", "report")

	if block.Value.Format != types.DocumentFormatPdf {
		t.Errorf("expected format %v, got %v", types.DocumentFormatPdf, block.Value.Format)
	}
	if block.Value.Name == nil || *block.Value.Name != "report" {
		t.Errorf("expected name 'report', got %v", block.Value.Name)
	}
	source, ok := block.Value.Source.(*types.DocumentSourceMemberBytes)
	if !ok {
		t.Fatalf("expected DocumentSourceMemberBytes, got %T", block.Value.Source)
	}
	if string(source.Value) != "hello" {
		t.Errorf("expected source bytes 'hello', got %q", string(source.Value))
	}
}
