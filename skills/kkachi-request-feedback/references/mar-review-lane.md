# MAR Review Lane

This reference expands the MAR procedure in `../SKILL.md`.

1. Run MAR after the first Blue + Red/Orange/Gray color review and feedback handling unless 주군 explicitly waives or replaces the lane before start and the decision is recorded in KAH/run evidence artifacts.
2. Use KAS to prepare request-bundle, prompt/input refs, provider registry refs, role matrix, and approval/retry/waiver metadata only.
3. Verify effective KAH `mar` capability/readback evidence, then trigger KAH `mar start`; KAS must not run provider CLIs or count local provider attempts as coverage.
4. Preserve KAH-owned input bundle, role request, bounded raw output, parse result, findings, compact merge pack/status, and Blue disposition refs under the run evidence directory.
5. Treat degraded providers, failed providers, parse failures, mutation detection, and unresolved required roles as fail-closed coverage gaps.
6. If MAR feedback requires repository changes, route the fix through `handle-feedback` and the selected implementer lane. After mutation, rerun affected verification and the required post-change color review.
7. Clean generated local sidecars or provider scratch state before final status unless they are intentionally in scope and ignored or recorded.
