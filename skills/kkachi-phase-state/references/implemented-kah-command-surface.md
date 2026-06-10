# Implemented KAH Command Surface

This reference expands the command catalog in `../SKILL.md`.

```bash
kkachi-agent-helper project init ... [--force] [--json]
kkachi-agent-helper project status [--json]
kkachi-agent-helper project doctor [--json]

kkachi-agent-helper run create --title <title> --work-path <A_development_execution|B_discovery_shaping> --work-mode <standard|light> --urgency <normal|urgent|critical> --sot-policy <existing_sot_basis|minimal_sot_before_code|full_sot_before_code> --execution-mode <production_write|adapter_qa|readiness_hardening|research|verification|docs_only> --commander <profile> [--backend-evidence <auto|required|not_applicable>] [--task-id <id>] [--redteam <profile>] [--json]
kkachi-agent-helper run list [--json]
kkachi-agent-helper run activate <run_id-or-prefix> [--json]
kkachi-agent-helper run close <run_id-or-prefix> [--json]
kkachi-agent-helper run abort <run_id-or-prefix> [--json]
kkachi-agent-helper run show <run_id-or-prefix> [--json]

kkachi-agent-helper artifact init <run_id-or-prefix> [--json]
kkachi-agent-helper artifact list <run_id-or-prefix> [--json]
kkachi-agent-helper artifact validate <run_id-or-prefix> [--gate intake] [--json]
kkachi-agent-helper artifact write <run_id-or-prefix> <artifact_path> --from <repo-relative-file> [--json]
kkachi-agent-helper artifact append <run_id-or-prefix> <artifact_path> --from <repo-relative-file> [--json]
kkachi-agent-helper artifact set-status <run_id-or-prefix> <artifact_path> --status <pending|complete|not_applicable> [--reason <text>] [--json]

kkachi-agent-helper gate check <run_id-or-prefix> <intake|sot|roadmap|plan|backend|implementation|review|verification|docs|final> [--json]
kkachi-agent-helper gate final <run_id-or-prefix> [--json]

kkachi-agent-helper event append <event_type> --run <run_id-or-prefix> --payload '<json-object>' [--json]
kkachi-agent-helper schema validate <file> --schema <config|status|event|run-metadata|selected-cli|bridge-session-snapshot> [--json]
kkachi-agent-helper schema export [--schema <name>|--all] [--dry-run] [--json]
kkachi-agent-helper schema migrate --from <version> --to <version> [--dry-run] [--json]
kkachi-agent-helper lock recover <active-run|project-write|all> --reason <text> [--run <run_id-or-prefix>] [--json]
kkachi-agent-helper diagnostics export [--run <run_id-or-prefix>] [--output <repo-relative-path>] [--json]

kkachi-agent-helper phase-plan init <run_id-or-prefix> [--json]
kkachi-agent-helper phase-plan show <run_id-or-prefix> [--json]
kkachi-agent-helper phase-plan set <run_id-or-prefix> <phase-id> --status <pending|in_progress|complete|skipped|not_applicable|blocked> [--evidence <path>] [--reason <text>] [--approval-required true|false] [--json]
kkachi-agent-helper phase-plan validate <run_id-or-prefix> [--final] [--json]

kkachi-agent-helper graph init --from-template <template-id-or-path> [--output .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph validate [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph explain [--file .kkachi-workflow.yaml] [--json]
kkachi-agent-helper graph diff --from <repo-relative-graph> --to <repo-relative-graph> [--semantic] [--json]
kkachi-agent-helper graph propose --candidate-file <repo-relative-candidate-graph> --reason <text> [--json]
kkachi-agent-helper graph apply --proposal <proposal-id> --approval <evidence-ref> [--json]
kkachi-agent-helper graph export --format mermaid|plantuml [--output <path>] [--json]

kkachi-agent-helper approval request <run_id-or-prefix> --phase <phase-id> --reason <reason> [--evidence <ref>] [--json]
kkachi-agent-helper approval record <run_id-or-prefix> --phase <phase-id> --decision <approved|rejected> --by <approver> --evidence <ref> [--reason <reason>] [--json]
kkachi-agent-helper approval show <run_id-or-prefix> [--phase <phase-id>] [--json]
```
