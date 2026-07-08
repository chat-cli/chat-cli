/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// withSystemCachePoint appends a cache point after the given system content
// blocks, so the (typically large, unchanging) system prompt can be reused
// across requests instead of being reprocessed every time. Returns nil
// unchanged - if there's no system prompt, there's nothing to cache and the
// request shape is identical to before this unit (NFR1).
func withSystemCachePoint(blocks []types.SystemContentBlock) []types.SystemContentBlock {
	if len(blocks) == 0 {
		return blocks
	}

	return append(blocks, &types.SystemContentBlockMemberCachePoint{
		Value: types.CachePointBlock{Type: types.CachePointTypeDefault},
	})
}

// hasSystemCachePoint reports whether blocks contains a cache point, so
// callers can skip a pointless retry when there was nothing to strip.
func hasSystemCachePoint(blocks []types.SystemContentBlock) bool {
	for _, b := range blocks {
		if _, ok := b.(*types.SystemContentBlockMemberCachePoint); ok {
			return true
		}
	}
	return false
}

// hasContentCachePoint reports whether blocks contains a cache point, the
// message-content counterpart to hasSystemCachePoint.
func hasContentCachePoint(blocks []types.ContentBlock) bool {
	for _, b := range blocks {
		if _, ok := b.(*types.ContentBlockMemberCachePoint); ok {
			return true
		}
	}
	return false
}

// stripSystemCachePoints returns a copy of blocks with any cache-point
// blocks removed, used to retry a request once without caching if the
// model/request combination doesn't support it (Rule 3,
// functional-design/business-logic-model.md).
func stripSystemCachePoints(blocks []types.SystemContentBlock) []types.SystemContentBlock {
	out := make([]types.SystemContentBlock, 0, len(blocks))
	for _, b := range blocks {
		if _, ok := b.(*types.SystemContentBlockMemberCachePoint); ok {
			continue
		}
		out = append(out, b)
	}
	return out
}

// stripContentCachePoints returns a copy of blocks with any cache-point
// blocks removed, the message-content counterpart to
// stripSystemCachePoints.
func stripContentCachePoints(blocks []types.ContentBlock) []types.ContentBlock {
	out := make([]types.ContentBlock, 0, len(blocks))
	for _, b := range blocks {
		if _, ok := b.(*types.ContentBlockMemberCachePoint); ok {
			continue
		}
		out = append(out, b)
	}
	return out
}

// buildQuestionContent builds the content blocks for a prompt/document pair.
// When document is empty, this is a single text block - identical to today's
// behavior (NFR1). When a document is present, it's split into its own text
// block, followed by a cache point, followed by the question - so the
// (typically large, unchanging) document can be cached separately from the
// (always different) question, per Functional Design Decision 2.
func buildQuestionContent(document, question string) []types.ContentBlock {
	if document == "" {
		return []types.ContentBlock{&types.ContentBlockMemberText{Value: question}}
	}

	return []types.ContentBlock{
		&types.ContentBlockMemberText{Value: document},
		&types.ContentBlockMemberCachePoint{Value: types.CachePointBlock{Type: types.CachePointTypeDefault}},
		&types.ContentBlockMemberText{Value: question},
	}
}
