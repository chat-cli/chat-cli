# NFR Design Patterns — Unit 2 (Tool Use / Function Calling)

Presented alongside `../nfr-requirements/` (see that file for why these two stages are combined for this unit).

## Security Pattern: Shared Path Validation (addresses SEC-1)
`utils.ValidateLocalPath` (introduced in this unit, extracted from `utils.ReadImage`'s existing inline logic) is the single choke point every file-reading capability in this unit goes through. `ReadFileTool.Execute` calls it before any `os.ReadFile`; no other path resolution logic exists in the `tools` package. This is a "validate at the boundary, once" pattern — there is exactly one place a path can be rejected, making it easy to audit.

## Security Pattern: Fail-Closed Tool Dispatch (addresses SEC-2, SEC-3)
`Registry.Dispatch` is structured so every exit path produces a `types.ToolResultBlock` — there is no code path where an unknown tool or a bad input silently does nothing or panics. Go's `recover()` is **not** used here deliberately: a panicking `Tool.Execute` implementation is a bug in that tool, not an expected runtime condition, and should surface loudly during development/testing rather than being silently swallowed. (This only applies to the one built-in tool shipped in this unit; if/when #86 adds more tools including ones that shell out, that unit should revisit whether `recover()` becomes warranted.)

## Reliability Pattern: Bounded Retry-Like Loop (addresses REL-1)
The tool round-trip loop (`functional-design/business-logic-model.md`'s algorithm) is a bounded loop with an explicit counter (max 10), not a `for {}` with an external break condition only — this is the same category of pattern as a bounded retry loop, chosen so the cap is impossible to accidentally remove during a future edit (the loop literally cannot execute an 11th time without a code change to the constant).

## Explicitly Not Applied (and why)
- **Circuit breaker**: N/A — there's no repeated external dependency being protected against cascading failure; a single Bedrock call failing is already handled by existing `log.Fatal` behavior, unchanged by this unit.
- **Caching**: N/A for tool results — the whole point of a tool call is to get fresh data (e.g. current file contents); caching would be incorrect here (contrast with Unit 3's prompt caching, which caches static prompt content, not tool outputs).
- **Rate limiting**: N/A — no multi-tenant or shared-resource concern for a local, single-user CLI.
