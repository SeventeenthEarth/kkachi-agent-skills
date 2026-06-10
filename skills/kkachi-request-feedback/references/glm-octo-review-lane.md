# GLM Octo Review Lane

This reference expands the official Octo procedure in `../SKILL.md`.

1. Use official GLM `/octo:review` after the normal KAS/KAH path has reached docs or verification, first Blue/Red/Orange/Gray color review, and any required improvement pass, unless 주군 explicitly asks for earlier feedback.
2. Use an explicit KAB config/evidence artifact for the run, including `backend_type: glm`, adapter identity, GLM model, resolved `glm_command` from the real user home/PATH, approved args/caveats, and the KAB session controller path.
3. Before start/send, run KAB and GLM HOME/auth preflight from the real user home, for example `HOME=<real-user-home> zsh -lc 'command -v kkachi-agent-bridge && command -v glm && glm --version'`. This proves availability only and must not perform the review outside KAB.
4. Start and send through KAB only. Evidence must include the KAB session id, backend type `glm`, bridge/tmux/session controller evidence, `selected-cli.json`, `capability-check.md`, `bridge-session-snapshot.json`, and `bridge-events.md` or equivalent KAH artifacts.
5. The actual prompt sent to the KAB GLM backend must begin with `/octo:review` as the first command text. Immediately after that command, state the requirements-plus-implemented-code-only review scope and the command prohibitions.
6. If the prompt appears pasted but not submitted, send Enter to the bridge tmux/session and re-read status/TUI until `prompt_confirmed: true` or a bounded failure is recorded.
7. During permission handling, approve only read-only inspection commands within scope, reject prohibited commands, and record every rejection plus any `response_fidelity_warning` in bridge evidence.
8. Use a bounded watcher after activation is confirmed; Octo may take up to 30 minutes. On completion, copy feedback into the KAH run directory, append the KAH feedback event, create or dispatch Blue/Red/Orange/Gray triage or re-review cards as required, then remove any watcher.
9. Parse verdicts only from the actual verdict heading or field, not examples or requested-output text.
10. If GLM/Octo feedback requires repository changes, route the fix through `handle-feedback` and the selected implementer lane. After mutation, rerun affected verification and the required post-change color review.
11. Clean generated local sidecars such as `.claude/`, `.claude-octopus/`, or GLM/Octo scratch state before final status unless they are intentionally in scope and ignored or recorded.
