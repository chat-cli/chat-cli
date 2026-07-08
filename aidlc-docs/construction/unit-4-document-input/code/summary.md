# Unit 4 — Native Document Input — Code Generation Summary

**Issue**: [#84](https://github.com/chat-cli/chat-cli/issues/84) | **Story**: 4.1 | **Plan**: `aidlc-docs/construction/plans/unit-4-document-input-code-generation-plan.md`

## Files Created
- `cmd/documentinput.go` — `sanitizeDocumentName` (strips path/extension, replaces disallowed characters, falls back to a neutral name), `buildDocumentContentBlock` — both 100% covered
- `cmd/documentinput_test.go`

## Files Modified
- `utils/utils.go` — new `ReadDocument`, mirroring `ReadImage`'s shape exactly (reuses `ValidateLocalPath`, validates against the 9 supported `DocumentFormat` extensions)
- `utils/utils_test.go` — `TestReadDocument` (all 9 formats, unsupported type, nonexistent file, path traversal)
- `cmd/prompt.go` — new `-d, --document` flag; when set, reads and attaches the document independently of `--image`; on read failure, fails clearly with no retry (a document is requested content, not an optional cache)
- `README.md`, `docs/usage.md` — documented `--document`, supported formats, and independence from `--image`

## Verification
- `make test`: all green, no regressions
- `make lint`: clean
- `go test -tags=integration -v .`: all 7 pass; `--document`/`-d` confirmed in `prompt --help` output
- Coverage: `utils` 44.7% → 49.3%; total statement coverage 64.7% → 66.2%
- **Not verified in this environment**: an actual document-input request against real Bedrock (no AWS credentials available) — the format validation, name sanitization, and content-block construction are all unit-tested directly; the untested seam is whether a real model actually accepts the resulting `DocumentBlock` shape.

## Story 4.1 Acceptance Criteria Status
- [x] `--document`/`-d` attaches a supported file as `ContentBlockMemberDocument`
- [x] Unsupported extensions and out-of-bounds paths rejected with a clear error before calling Bedrock (reuses `ValidateLocalPath`)
- [x] `--image` behavior completely unchanged; both flags usable together
- [x] Document name sanitized (never the raw filename) — addresses both the SDK's character restriction and its documented prompt-injection concern
