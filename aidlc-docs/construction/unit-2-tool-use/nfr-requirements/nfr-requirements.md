# NFR Requirements — Unit 2 (Tool Use / Function Calling)

**Note**: NFR Requirements and NFR Design are presented together for this unit (see `../nfr-design/` for the design side) — the NFR profile is narrow enough that splitting them into two separate approval gates would add ceremony without adding clarity. chat-cli is a single-process, single-user CLI with no network-facing surface, no multi-tenancy, and no deployment/scaling model, so most NFR categories genuinely don't apply here. Security is the one category with real, non-trivial requirements for this specific unit (arbitrary tool execution + file access).

## Scalability Requirements
**N/A.** Single-user, single-process CLI invocation — there is no concurrent load to scale for.

## Performance Requirements
**N/A beyond what already exists.** Tool dispatch is a synchronous, in-process function call; the dominant latency is the Bedrock round-trip itself (unchanged from today). No new performance target is introduced by this unit.

## Availability Requirements
**N/A.** No uptime/SLA concept applies to a local CLI tool.

## Security Requirements (the applicable category)
- **SEC-1**: The built-in `read_file` tool must not be able to read files outside the current working directory (path traversal protection), reusing the validation already proven in `utils.ReadImage` and being extracted as `utils.ValidateLocalPath` in this unit (per Application Design).
- **SEC-2**: A tool the model requests but that isn't registered must be rejected safely (Rule 2, `functional-design/business-rules.md`) — the model cannot cause the CLI to execute arbitrary, unregistered code paths.
- **SEC-3**: Tool input (the JSON the model supplies) must be treated as untrusted input — parsed defensively (Rule 4), never passed to a shell or `exec`-like call in this unit (no tool in this unit shells out; that's explicitly out of scope, tracked separately in issue #86's "run a shell command" tool, which will need its own NFR pass when it's built).
- **SEC-4**: No tool result content is ever interpreted as instructions to skip the existing safety checks elsewhere in the codebase (e.g. a malicious file's contents returned by `read_file` are just text data sent back to the model — they don't get executed or evaluated by chat-cli itself).

## Reliability Requirements
- **REL-1**: The tool round-trip cap (Rule 5, `functional-design/business-rules.md` — max 10 per turn) is the reliability safeguard for this unit; carried over here as the formal NFR justification for a rule introduced during Functional Design.

## Maintainability / Usability / Tech Stack Selection
**N/A / already covered.** No new dependency, no new build/deploy tooling. See `tech-stack-decisions.md`.
