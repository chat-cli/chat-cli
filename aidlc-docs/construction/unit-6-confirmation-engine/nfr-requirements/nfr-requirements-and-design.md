# NFR Requirements and Design (Combined): Unit 6, Confirmation and Sticky Approval Engine (#86)

Combined presentation, same pattern as Initiative 1 Units 2/4 and Initiative 2 - Security is the dominant category here (this unit exists specifically to make destructive actions safe), with Reliability and Usability as secondary but real concerns given a broken gate is worse than no gate at all.

## Security

### Requirement
The gate must be the actual, sole control for destructive actions - there must be no code path where `write_file`/`run_shell` execute without a decision having been recorded, and the persisted approval store must not be a soft target (unreadable by other users, not trivially spoofable).

### Design
- **`Registry.Dispatch` is the single choke point** - `write_file`/`run_shell`'s `Execute` methods are never called directly by anything else in the codebase (same principle as Unit 2's `ValidateLocalPath` choke point). A future tool that forgets to implement `RequiresConfirmation()` correctly is the only way to bypass this - mitigated by the interface requiring an explicit `bool` return (can't be silently omitted, Go interfaces are structurally checked at compile time).
- **Fail-closed everywhere**: unparseable `ConfirmationSummary` input → denial (BR4); unrecognized/empty prompt input → denial (BR12); malformed persisted-store file → treated as *no* approvals, not as approve-everything (BR15). Every ambiguous state resolves to "ask again" or "deny," never to "allow."
- **`tool-approvals.yaml` created with `0600`** (owner-only) - matches the existing config file's implicit trust model (chat-cli already stores AWS-adjacent config in the same directory tree with no additional access-control layer; this doesn't introduce a new class of secret, just a new class of "things this user has said yes to").
- **Per-repository scoping is itself a security boundary**, not just a UX nicety (FR7.1) - an "always" approval for `git *` in a trusted personal project must not silently apply when the same user runs `chat-cli` inside a cloned, less-trusted repository. Verified explicitly in Story 8.2's acceptance criteria and will get a dedicated Build and Test scenario.
- **No new attack surface from the persisted file's *content*** - it stores only `toolName:patternKey` strings the user themselves approved; it's not executable, not interpreted as anything beyond an approval-lookup key, and a corrupted file degrades to "no approvals" (BR15) rather than being parsed permissively.

### Compliance
✅ Compliant - single choke point, fail-closed on every ambiguous state, repo-scoped persistence, restrictive file permissions.

## Reliability

### Requirement
A confirmation-engine bug must never crash `chat` outright, and must never silently skip the gate (a crash is preferable to a silent bypass, but "ask again" is preferable to either).

### Design
- BR15: a corrupted/missing `tool-approvals.yaml` degrades to empty approvals (re-prompt for everything) rather than a fatal error at `chat` startup - consistent with #88's established graceful-degradation precedent for file-read problems.
- BR13: no retry-loop on invalid prompt input - a single bad keystroke just means "denied, try again next time the model asks" rather than the CLI hanging on a re-prompt loop or crashing on unexpected input.
- The gate's `Check` call happens synchronously inside the existing tty-loop (no new goroutines, no new concurrency-related failure modes) - consistent with Application Design's stated communication pattern.

### Compliance
✅ Compliant - no fatal-error path introduced by this unit; every failure mode degrades to a safe, expected state (re-prompt or deny).

## Usability

### Requirement
The confirmation prompt must give enough information for an informed decision without becoming an obstacle so heavy that users reflexively approve without reading (NFR5 from requirements.md).

### Design
- The prompt always shows the tool-specific summary (exact command string for `run_shell`; path + content for `write_file`) before asking for a decision - never a generic "allow this tool call?" with no specifics.
- For `write_file`, content over a readable threshold (4KB) is shown truncated with an explicit "N bytes total, shown truncated" note, rather than either flooding the terminal with a huge write or hiding the size entirely - balances informed-decision-making against practical terminal usability (exact threshold documented as a Code Generation detail, not re-litigated here).
- The three approval tiers are presented with their consequence stated plainly (e.g. "session" vs "always" - not just single letters with no explanation) at the prompt itself, so the choice's scope is clear in the moment, not just in documentation.

### Compliance
✅ Compliant - informative by default, with a stated (not silent) truncation policy for oversized content.

## Non-Applicable Categories
- **Scalability/Performance/Availability**: N/A - same rationale as every prior unit in this project (single local process, no concurrent-load concept). The gate's own overhead is one YAML file read at startup and (rarely) one write - negligible.
