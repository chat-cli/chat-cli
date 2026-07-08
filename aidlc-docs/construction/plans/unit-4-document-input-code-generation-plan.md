# Unit 4 — Native Document Input — Code Generation Plan

**Unit**: Unit 4, issue [#84](https://github.com/chat-cli/chat-cli/issues/84) | **Story**: 4.1
**FR/NFR coverage**: FR4.1-FR4.4, SEC-1 (reused), SEC-2 (new)
**Dependencies**: Unit 2's `utils.ValidateLocalPath` (reused as-is); SDK upgrade (prerequisite) complete

## Layering (mirrors existing conventions)
- `utils.ReadDocument` — file I/O + format validation, mirrors `utils.ReadImage`'s shape exactly.
- `cmd/documentinput.go` — SDK-type-specific construction (`sanitizeDocumentName`, the `ContentBlockMemberDocument` builder), mirrors how `buildSystemContentBlocks`/`promptcache.go` keep Bedrock-type-specific logic in `cmd`, separate from `utils`'s file-IO layer.

## TDD Order

### Step 1-2 — `utils.ReadDocument`
- [ ] Test (`utils/utils_test.go`, new cases, same temp-dir pattern as `TestReadImage`): valid `.pdf`/`.csv`/`.doc`/`.docx`/`.xls`/`.xlsx`/`.html`/`.txt`/`.md` files → correct format string, no error; unsupported extension → error; nonexistent file → error; path traversal attempt → error (reusing `ValidateLocalPath`'s existing protection, same as `ReadImage`).
- [ ] Implement `ReadDocument(filename string) (data []byte, format string, err error)` in `utils/utils.go`, calling `ValidateLocalPath` then mapping the extension to one of the 9 supported formats (case-insensitive, matching `ReadImage`'s existing `strings.ToLower` convention).

### Step 3-4 — `sanitizeDocumentName`
- [ ] Test (`cmd/documentinput_test.go`, new): `"report.pdf"` → `"report"`; a name with disallowed characters (e.g. `"my_file!.txt"`) → only allowed characters remain, collapsed/trimmed; an all-disallowed-characters input → falls back to `"attached-document"`.
- [ ] Implement `sanitizeDocumentName(filename string) string` in `cmd/documentinput.go` (new file): `filepath.Base`, strip extension, replace disallowed characters with a space, collapse repeated spaces, trim, fall back to `"attached-document"` if empty.

### Step 5-6 — Document content-block builder
- [ ] Test: given data/format/name, returns a `*types.ContentBlockMemberDocument` with the correct `Format` (mapped from the `utils.ReadDocument` format string to the matching `types.DocumentFormat` constant), `Name`, and `Source` (`DocumentSourceMemberBytes{Value: data}`).
- [ ] Implement `buildDocumentContentBlock(data []byte, format, name string) *types.ContentBlockMemberDocument` in `cmd/documentinput.go`.

### Step 7 — Wire into `cmd/prompt.go`
- [ ] Add `-d, --document ""` flag (mirrors `-i, --image` registration style).
- [ ] When set: call `utils.ReadDocument`, `sanitizeDocumentName`, `buildDocumentContentBlock`; append the resulting content block to `userMsg.Content` (after the document/cache-point/question blocks from Unit 3, alongside the existing `--image` block if both are set — Rule 5, independent of each other).
- [ ] On `ReadDocument` error: `log.Fatalf` with a clear message (Rule 4, no retry — Design Decision 2).
- [ ] On a Bedrock API error with `--document` set: existing `log.Fatalf("error from Bedrock, %v", err)` already covers this — no special-casing needed since Design Decision 2 says no retry.
- [ ] `chat.go` and `image.go` are **not** touched — document input is `prompt`-only per Functional Design's Rule 6.

### Step 8 — Full test suite, lint, coverage, integration
- [ ] `make test`, `make lint`, `make test-coverage`, `make cli && go test -tags=integration -v .`.

### Step 9 — Documentation
- [ ] `README.md`/`docs/usage.md`: document `--document`/`-d`, the supported formats, and that it can be combined with `--image`.

### Step 10 — Unit Documentation Summary
- [ ] `aidlc-docs/construction/unit-4-document-input/code/summary.md`.

## Story Traceability
- Story 4.1 (all 4 acceptance criteria) → Steps 1-7 implement and test them; Step 8 verifies; Step 9 satisfies NFR6.
