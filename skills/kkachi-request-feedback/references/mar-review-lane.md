# MAR Review Lane

This reference expands the MAR procedure in `../SKILL.md`.

1. Run MAR after the first Blue + Red/Orange/Gray color review and feedback handling unless 주군 explicitly waives or replaces the lane before start and the decision is recorded in KAH/run evidence artifacts.
2. Verify provider toolchain/preflight evidence before counting any provider attempt.
3. Cover required roles `logic`, `security`, `arch`, `cve`, and `test_adequacy` through declared primary/secondary provider lanes only.
4. Preserve the input bundle, role request, bounded raw output, parse result, findings, compact merge pack, and Blue disposition under the run evidence directory.
5. Treat degraded providers, failed providers, parse failures, mutation detection, and unresolved required roles as fail-closed coverage gaps.
6. If MAR feedback requires repository changes, route the fix through `handle-feedback` and the selected implementer lane. After mutation, rerun affected verification and the required post-change color review.
7. Clean generated local sidecars or provider scratch state before final status unless they are intentionally in scope and ignored or recorded.
