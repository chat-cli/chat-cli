# NFR Requirements and Design (Combined): AGENTS.md Convention (#88)

Combined into one document, same as Units 2 and 4 in Initiative 1, given the narrow NFR profile of a single-user local CLI - Security and Reliability are the only applicable categories. Scalability/Performance/Availability are N/A for the same reasons documented throughout Initiative 1 (see `aidlc-docs/construction/build-and-test/performance-test-instructions.md`): no concurrent load, no throughput target, single local process.

## Security

### Requirement
File discovery must never read content outside the two directories the user is already implicitly trusting: their current working directory, and the root of the git repository they're standing inside. It must not be usable to read arbitrary files elsewhere on the filesystem, and it must not silently scan an unbounded number of ancestor directories.

### Design
- **Bounded to exactly two directories** (business-logic-model.md's Phase A/B): cwd and, if found, the git-root boundary. No other directory's contents are ever enumerated or read.
- **Phase A only ever checks for `.git`'s existence** (a stat call, not a content read) while walking upward - this walk is bounded by the filesystem's actual depth (or, defensively, capped - see Design Note below) and never reads file *content* outside the two Phase B directories.
- **Fixed, known filenames only** - unlike `--document`/`read_file` (Initiative 1, Units 2/4) which take a user- or model-supplied path and need `utils.ValidateLocalPath`'s path-traversal defenses, this feature only ever checks for exact, hardcoded (or explicitly user-configured via `context-files`) filenames within the two known directories. There is no user-supplied *path* to validate - the attack surface `ValidateLocalPath` defends against (`../../etc/passwd`-style traversal) doesn't apply here, since candidates are filenames/relative-subpaths joined against a directory chat-cli itself determined, never a raw external input path.
- **Design Note - walk-up cap**: Phase A's upward walk is theoretically unbounded if `chat` is run from a very deeply nested cwd with no `.git` anywhere above it (e.g. accidentally run from `/`). Cap the walk at a generous fixed depth (e.g. 64 levels) as a defensive belt-and-suspenders measure against pathological input, even though in practice this only costs cheap `stat` calls, never file reads.
- **Local, self-authored content**: unlike tool-call results or model output, the file content here is something the user (or their team) wrote directly into their own repository - there's no untrusted-input/prompt-injection concern analogous to Unit 4's document-name sanitization, since this isn't material being described *to* Bedrock as coming from an external document, it's used exactly like an explicit `--system` value already is today.

### Compliance
✅ Compliant - bounded directory scope, no traversal surface, no untrusted-content handling gap.

## Reliability

### Requirement
A missing, unreadable, empty, or oddly-permissioned candidate file must never prevent `chat` from starting. Discovery is a nice-to-have convenience, not a hard dependency.

### Design
- BR10: a read error (permission denied, race-condition deletion, etc.) is treated as "no match," search continues to the next candidate - never a fatal error, never even a warning (this is expected/silent, since e.g. checking for `CLAUDE.md` in a repo that doesn't have one is the overwhelmingly common case).
- BR8: an empty-after-trim file is likewise treated as "no match," not as "found but empty" - avoids a confusing FR6 notice for a file that contributes nothing.
- The one place this feature *can* affect startup is the flag-parsing/config-read calls it adds (`--no-context-file`, `context-files`) - these follow the exact same `log.Fatalf` pattern already used for every other flag in `chat.go` today (a genuinely malformed flag value is a usage error, consistent with existing behavior, not a new failure mode).

### Compliance
✅ Compliant - filesystem-level failures degrade to "no match," never fatal; flag/config errors use the codebase's existing, already-accepted error-handling convention.

## Non-Applicable Categories
- **Scalability**: N/A - single local process, no concurrent users.
- **Performance**: N/A beyond "runs once at `chat` startup, bounded to ≤2 directory checks + a capped upward stat-walk" - no load/throughput concept, consistent with `performance-test-instructions.md`'s established rationale for this whole project.
- **Availability**: N/A - no uptime/SLA concept for a local CLI.
- **Maintainability**: Addressed structurally (one new file following the established per-feature-file convention) rather than as a distinct NFR requiring its own analysis.
