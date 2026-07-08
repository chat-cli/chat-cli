# Functional Design — Unit 4 (Native Document Input)

Combined into one document, consistent with Units 2/3's calibration for scope.

## SDK Facts Verified by Direct Inspection (v1.55.0)
- `types.DocumentBlock{Name *string, Source DocumentSource, Citations *CitationsConfig, Context *string, Format DocumentFormat}`; `types.DocumentSourceMemberBytes{Value []byte}` implements `DocumentSource`; `types.ContentBlockMemberDocument{Value DocumentBlock}` implements `ContentBlock`.
- `types.DocumentFormat` values: `pdf, csv, doc, docx, xls, xlsx, html, txt, md` — exactly what `requirements.md`'s FR4.1 already specified.
- **`DocumentBlock.Name` has a hard character restriction**: "alphanumeric, whitespace (no more than one in a row), hyphens, parentheses, square brackets" only — a raw filename like `report.pdf` is **not valid as-is** (the `.` isn't allowed). The SDK's own doc comment also flags this field as "vulnerable to prompt injections... recommend a neutral name." Both facts mean the filename can't be passed through untouched.
- `bedrock` control-plane's `ModelModality` enum has exactly `TEXT, IMAGE, EMBEDDING` — **no `DOCUMENT` value exists**, confirming the same pattern already found for tool use (Unit 2) and caching (Unit 3): Bedrock exposes no pre-flight "supports documents" capability check. Consistent with those units, this design attempts the request and surfaces a real API error rather than fabricating a client-side check.

## Design Decision 1: Derive a Sanitized, Neutral Document Name
Since the raw filename can't be used directly (character restrictions + injection risk), `--document report.pdf` produces a document name like `"report"` (extension stripped, any character outside the allowed set replaced with a space, collapsed/trimmed) rather than the literal filename. If sanitization leaves nothing usable, fall back to a fixed neutral name (`"attached-document"`).

## Design Decision 2: No Retry-Without-Document on Error (unlike Unit 3's caching)
Unit 3's cache points are an optional performance optimization — dropping them on failure and retrying is safe because the content is unchanged either way. A document is the user's actual requested content — silently dropping it and retrying without it would answer a question the user didn't ask. **Decision**: if the model rejects the document (format unsupported, size too large, etc.), surface the Bedrock error clearly (FR4.3) rather than retrying without it.

## Design Decision 3: No Explicit File-Size Pre-Check
`utils.ReadImage` today has no size validation despite `README.md`/`CLAUDE.md` documenting a "<5MB" limit for images — the API itself enforces size limits and returns a clear error. For consistency with that existing precedent (and to avoid inventing a size threshold that could drift from Bedrock's actual, service-side limit), `--document` follows the same pattern: no client-side size check, rely on Bedrock's error response.

## Business Rules
- **Rule 1 (FR4.1)**: `--document`/`-d <path>` on `prompt` reads the file, validates its extension is in the supported `DocumentFormat` allow-list, and attaches it as `ContentBlockMemberDocument`.
- **Rule 2 (FR4.2, reuses Unit 2's pattern)**: Path resolution goes through the existing `utils.ValidateLocalPath` (already extracted in Unit 2) — no new path-safety logic is written for this unit, it's reused as-is.
- **Rule 3 (Design Decision 1)**: The document's `Name` field is derived via a new `sanitizeDocumentName` function, never the raw filename.
- **Rule 4 (FR4.3, Design Decision 2)**: A rejected document (unsupported format detected client-side, or a Bedrock API error) produces a clear, specific error message and no retry.
- **Rule 5 (FR4.4)**: `--document` and `--image` are independent and can be combined in the same `prompt` invocation; `--image` behavior is completely unchanged.
- **Rule 6 (scope note)**: `--document` is `prompt`-only in this pass, consistent with `application-design.md`'s original scoping — `chat` has no document-input capability at all yet (issue #46 covers that separately).
