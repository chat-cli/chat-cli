# Application Design Plan

**Scope** (per `execution-plan.md`): Application Design is executing specifically because Tool Use (#82) introduces genuinely new components. The other 4 features (system prompt, prompt caching, document input, extended thinking) are small, direct additions to the existing `cmd/chat.go`/`cmd/prompt.go` request-building code and don't need separate component/interface design — their detailed logic is deferred to per-unit Functional Design in Construction, per this stage's own scope note ("detailed business logic design happens later").

## Design Decisions Stated as Assumptions (in place of a Q&A round)

Consistent with the calibration used throughout this session (solo-maintained CLI, not a multi-stakeholder project):

1. **Component boundaries**: One new package, `tools/` (or `internal/tools/` — final placement decided in Code Generation), holding a `Tool` interface and a `Registry` that implements the tool-dispatch loop described in Story 2.1. Kept out of `cmd` so it's independently unit-testable without Cobra/Bedrock wiring.
2. **Path-safety reuse**: `utils.ReadImage`'s inline path-traversal validation is extracted into a standalone `utils.ValidateLocalPath(filename string) (string, error)` helper, reused by both the new `read_file` tool (#82/FR2.5) and the new document-attachment flow (#84/FR4.2) — avoiding a second copy of that logic (echoing the existing "deduplicate validation logic" technical debt item, #92, so this design doesn't add to that pile).
3. **Cache-point handling**: A small `utils.CachePoint`-style helper (exact shape decided in Code Generation) is responsible for appending a cache checkpoint and retrying once without it on rejection (FR3.3), reused by both `chat` and `prompt`.
4. **No new persistence component**: Tool-call turns reuse `repository.ChatRepository.Create` as-is (Story 2.1/FR2.4) — no new repository or schema component.
5. **Service layer**: Given this is a CLI (not a multi-service backend), "services" in this design map to the request-orchestration functions already living in `cmd/chat.go`'s `Run` and `cmd/prompt.go`'s `Run` — no new service layer is introduced; `services.md` documents how these two existing orchestration points now call into the new components.

## Mandatory Artifacts
- [ ] `components.md` — component definitions and responsibilities
- [ ] `component-methods.md` — method signatures (business rules deferred to Functional Design)
- [ ] `services.md` — orchestration patterns (existing `chat`/`prompt` Run functions, extended)
- [ ] `component-dependency.md` — dependency matrix and data flow
- [ ] `application-design.md` — consolidated summary of the above
