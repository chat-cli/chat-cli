# NFR Requirements + Design — Unit 4 (Native Document Input)

Combined into one document (see `../nfr-design/` for the equivalent Unit 2 precedent this follows) — this unit's security requirement is fully satisfied by reusing an already-designed, already-tested pattern, so a fresh design pass isn't needed.

## Security Requirements
- **SEC-1 (identical to Unit 2's SEC-1)**: `--document` must not be able to read files outside the current working directory. **Satisfied by reuse**: `utils.ValidateLocalPath` (introduced in Unit 2, 100%-path-tested via `TestValidateLocalPath` and `TestReadImage`) is called for `--document` exactly as it already is for `--image` and the `read_file` tool — no new validation logic is written for this unit.
- **SEC-2 (new for this unit)**: The document's `Name` field (sent to the model) must not echo the raw filename verbatim, both because the SDK rejects most filename characters outright (`.` isn't in the allowed set) and because the SDK's own documentation flags unsanitized names as a prompt-injection vector. Satisfied by `sanitizeDocumentName` (Functional Design Decision 1).

## Scalability / Performance / Availability
**N/A**, same rationale as Units 2/3 — single-user local CLI, no multi-tenant or uptime concept.

## Reliability
No new reliability pattern needed — Design Decision 2 (no retry-without-document) means errors surface immediately and clearly via the existing `log.Fatalf` pattern, consistent with how `--image` failures are handled today.

## Design Pattern Reused, Not Redesigned
`utils.ValidateLocalPath`'s "validate at the boundary, once" pattern (documented in Unit 2's `nfr-design/nfr-design-patterns.md`) extends naturally to a third caller (`--document`, alongside `--image`'s `ReadImage` and the `read_file` tool) without modification — this is exactly the kind of reuse that pattern was designed for.

## Tech Stack
No new dependencies — `types.DocumentBlock`/`DocumentSourceMemberBytes`/`ContentBlockMemberDocument`/`DocumentFormat` are all already available in the upgraded SDK (Unit 3's prerequisite).
