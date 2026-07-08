# Logical Components — Unit 2 (Tool Use / Function Calling)

No infrastructure-style logical components (queues, caches, circuit breakers, load balancers) apply to this unit — chat-cli is a local, single-process CLI (confirmed globally in `inception/plans/execution-plan.md`'s Infrastructure Design: SKIP).

The only "logical component" introduced is the in-process `tools.Registry`, already fully specified in `application-design/components.md` and `functional-design/domain-entities.md`. No additional logical/infrastructure component is needed to satisfy this unit's NFRs — the security and reliability patterns in `nfr-design-patterns.md` are implemented entirely as code-level patterns (shared validation function, fail-closed dispatch, bounded loop), not as separate infrastructure.
