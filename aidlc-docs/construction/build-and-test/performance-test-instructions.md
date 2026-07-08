# Performance Test Instructions

## Status: Not Applicable

chat-cli is a single-user, local, single-process terminal CLI — there is no concurrent load, no throughput target, no multi-tenant capacity to plan for, and no uptime/SLA concept. This was established as a cross-cutting NFR finding as early as `requirements.md` and reaffirmed in every unit's NFR assessment throughout this initiative (Units 1, 3, and 5 skipped NFR Requirements/Design entirely for this reason; Units 2 and 4 explicitly marked Scalability/Performance/Availability as N/A in their combined NFR documents).

The one genuinely performance-related feature in this initiative — prompt caching (Unit 3, #83) — is a *cost/latency optimization for individual requests*, not a load-bearing concern; its correctness (cache points inserted/stripped correctly) is covered by unit tests in `cmd/promptcache_test.go`, not load testing.

No performance test suite was created, consistent with this rationale.
