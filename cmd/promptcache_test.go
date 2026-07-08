/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

func TestWithSystemCachePoint(t *testing.T) {
	t.Run("nil input yields nil output", func(t *testing.T) {
		if got := withSystemCachePoint(nil); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("non-empty input gets a cache point appended", func(t *testing.T) {
		in := buildSystemContentBlocks("be terse")
		out := withSystemCachePoint(in)

		if len(out) != 2 {
			t.Fatalf("expected 2 blocks (text + cache point), got %d", len(out))
		}
		if _, ok := out[0].(*types.SystemContentBlockMemberText); !ok {
			t.Errorf("expected first block to be text, got %T", out[0])
		}
		if _, ok := out[1].(*types.SystemContentBlockMemberCachePoint); !ok {
			t.Errorf("expected second block to be a cache point, got %T", out[1])
		}
	})
}

func TestHasSystemCachePoint(t *testing.T) {
	t.Run("nil has no cache point", func(t *testing.T) {
		if hasSystemCachePoint(nil) {
			t.Error("expected false for nil input")
		}
	})

	t.Run("text-only has no cache point", func(t *testing.T) {
		if hasSystemCachePoint(buildSystemContentBlocks("be terse")) {
			t.Error("expected false when no cache point is present")
		}
	})

	t.Run("with cache point returns true", func(t *testing.T) {
		if !hasSystemCachePoint(withSystemCachePoint(buildSystemContentBlocks("be terse"))) {
			t.Error("expected true when a cache point is present")
		}
	})
}

func TestStripSystemCachePoints(t *testing.T) {
	t.Run("strips the cache point, keeps the text", func(t *testing.T) {
		withCache := withSystemCachePoint(buildSystemContentBlocks("be terse"))
		stripped := stripSystemCachePoints(withCache)

		if len(stripped) != 1 {
			t.Fatalf("expected 1 block after stripping, got %d", len(stripped))
		}
		if _, ok := stripped[0].(*types.SystemContentBlockMemberText); !ok {
			t.Errorf("expected remaining block to be text, got %T", stripped[0])
		}
	})

	t.Run("no-op when there's nothing to strip", func(t *testing.T) {
		in := buildSystemContentBlocks("be terse")
		stripped := stripSystemCachePoints(in)

		if len(stripped) != len(in) {
			t.Errorf("expected unchanged length %d, got %d", len(in), len(stripped))
		}
	})
}

func TestHasContentCachePoint(t *testing.T) {
	t.Run("no document has no cache point", func(t *testing.T) {
		if hasContentCachePoint(buildQuestionContent("", "hi")) {
			t.Error("expected false when no document/cache point is present")
		}
	})

	t.Run("document present has a cache point", func(t *testing.T) {
		if !hasContentCachePoint(buildQuestionContent("doc text", "hi")) {
			t.Error("expected true when a document (and its cache point) is present")
		}
	})
}

func TestStripContentCachePoints(t *testing.T) {
	blocks := buildQuestionContent("some document text", "what does this say?")
	stripped := stripContentCachePoints(blocks)

	if len(stripped) != 2 {
		t.Fatalf("expected 2 text blocks after stripping the cache point, got %d", len(stripped))
	}
	for _, b := range stripped {
		if _, ok := b.(*types.ContentBlockMemberText); !ok {
			t.Errorf("expected remaining blocks to be text, got %T", b)
		}
	}
}

func TestBuildQuestionContent(t *testing.T) {
	t.Run("no document yields a single text block", func(t *testing.T) {
		blocks := buildQuestionContent("", "how are you?")
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		textBlock, ok := blocks[0].(*types.ContentBlockMemberText)
		if !ok {
			t.Fatalf("expected text block, got %T", blocks[0])
		}
		if textBlock.Value != "how are you?" {
			t.Errorf("expected question text, got %q", textBlock.Value)
		}
	})

	t.Run("document present yields document, cache point, question in order", func(t *testing.T) {
		blocks := buildQuestionContent("<document>stuff</document>", "summarize this")
		if len(blocks) != 3 {
			t.Fatalf("expected 3 blocks, got %d", len(blocks))
		}

		docBlock, ok := blocks[0].(*types.ContentBlockMemberText)
		if !ok || docBlock.Value != "<document>stuff</document>" {
			t.Errorf("expected first block to be the document text, got %+v", blocks[0])
		}
		if _, ok := blocks[1].(*types.ContentBlockMemberCachePoint); !ok {
			t.Errorf("expected second block to be a cache point, got %T", blocks[1])
		}
		questionBlock, ok := blocks[2].(*types.ContentBlockMemberText)
		if !ok || questionBlock.Value != "summarize this" {
			t.Errorf("expected third block to be the question text, got %+v", blocks[2])
		}
	})
}
