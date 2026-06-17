#!/usr/bin/env python3
"""Local stdlib-only Multi-Agent Review helper.

This helper renders local prompts, normalizes fixture/mock review evidence, and
prepares fail-closed provider-attempt artifacts for validated MAR lanes.
"""

import argparse
import datetime
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path


STATUS_VOCABULARY = (
    "PASS",
    "PASS_WITH_FINDINGS",
    "REQUEST_CHANGES",
    "BLOCKED",
    "DEGRADED",
    "FAILED",
)

DEFAULT_RAW_CAP = 4096
DEFAULT_TIMEOUT = 10
DEFAULT_PROVIDER_REGISTRY = "registries/mar-provider-lanes.json"
DEFAULT_TOOLCHAIN = ".kkachi/toolchain.yaml"
PROVIDER_LANES_SCHEMA_VERSION = "mar.provider_lanes.v1"
PROVIDER_TOOLS_SCHEMA_VERSION = "mar.provider_tools.v1"
PROVIDER_ATTEMPT_SCHEMA_VERSION = "mar.provider_attempt.v1"
PROVIDER_FAILURE_REASONS = (
    "auth_failed",
    "token_exhausted",
    "quota_exhausted",
    "rate_limited",
    "cli_missing",
    "model_unavailable",
    "timeout",
    "nonzero_exit",
    "parse_failure",
    "mutation_detected",
    "adapter_proof_required",
    "unknown_provider_failure",
)
COVERED_STATUSES = ("PASS", "PASS_WITH_FINDINGS")
NON_CLEAN_STATUSES = ("BLOCKED", "DEGRADED", "FAILED")


def run_git_status():
    completed = subprocess.run(
        ["git", "status", "--porcelain=v1", "--untracked-files=all"],
        shell=False,
        timeout=DEFAULT_TIMEOUT,
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return {
        "available": completed.returncode == 0,
        "returncode": completed.returncode,
        "stdout": completed.stdout,
        "stderr": completed.stderr,
    }


def cap_raw_output(raw, cap):
    data = raw.encode("utf-8", "replace")
    truncated = len(data) > cap
    if truncated:
        data = data[:cap]
    return {
        "text": data.decode("utf-8", "replace"),
        "bytes_original": len(raw.encode("utf-8", "replace")),
        "bytes_returned": len(data),
        "cap_bytes": cap,
        "truncated": truncated,
    }


def emit(payload):
    print(json.dumps(payload, indent=2, sort_keys=True))


def status_exit(payload):
    if payload.get("mutation_guard", {}).get("detected"):
        return 2
    return 0


def valid_status(value):
    return value in STATUS_VOCABULARY


def valid_provider_failure_reason(value):
    return value in PROVIDER_FAILURE_REASONS


def structured_failure(status, reason, **extra):
    payload = {
        "status": status,
        "reason": reason,
        "no_provider_execution": True,
    }
    payload.update(extra)
    return payload


def load_text(path):
    return Path(path).read_text(encoding="utf-8")


def utc_now():
    return datetime.datetime.now(datetime.timezone.utc).replace(microsecond=0).isoformat()


def repo_root():
    return Path(__file__).resolve().parents[1]


def resolve_repo_path(path):
    candidate = Path(path)
    if candidate.is_absolute():
        return candidate
    return repo_root() / candidate


def parse_review_path(path, raw_cap):
    raw = load_text(path)
    return parse_review_raw(raw, raw_cap, path=str(path))


def parse_review_raw(raw, raw_cap, path=None):
    capped = cap_raw_output(raw, raw_cap)
    try:
        parsed = json.loads(raw)
    except json.JSONDecodeError as exc:
        return structured_failure(
            "DEGRADED",
            "parse_failure",
            path=str(path) if path is not None else None,
            parse_failure={
                "message": str(exc),
                "line": exc.lineno,
                "column": exc.colno,
            },
            raw_output=capped,
        )

    if not isinstance(parsed, dict):
        return structured_failure(
            "FAILED",
            "review_json_must_be_object",
            path=str(path) if path is not None else None,
            raw_output=capped,
        )

    status = parsed.get("status")
    if not valid_status(status):
        return structured_failure(
            "FAILED",
            "invalid_status",
            path=str(path) if path is not None else None,
            observed_status=status,
            allowed_statuses=list(STATUS_VOCABULARY),
            raw_output=capped,
        )

    result = dict(parsed)
    result.setdefault("no_provider_execution", True)
    if path is not None:
        result["path"] = str(path)
    result["raw_output"] = capped
    return result


def aggregate_status(reviews, required_reviewers):
    if not reviews:
        return "BLOCKED", "no_reviews"

    statuses = [review.get("status") for review in reviews]
    if required_reviewers and len(reviews) < required_reviewers:
        return "BLOCKED", "insufficient_coverage"
    if statuses and all(status == "FAILED" for status in statuses):
        return "FAILED", "all_reviewers_failed"
    if "BLOCKED" in statuses:
        return "BLOCKED", "blocked_review"
    if "REQUEST_CHANGES" in statuses:
        return "REQUEST_CHANGES", "actionable_findings"
    if "FAILED" in statuses or "DEGRADED" in statuses:
        return "DEGRADED", "degraded_coverage"
    if "PASS_WITH_FINDINGS" in statuses:
        return "PASS_WITH_FINDINGS", "non_blocking_findings"
    return "PASS", "all_reviews_passed"


def load_provider_registry(path):
    registry_path = resolve_repo_path(path)
    try:
        data = json.loads(registry_path.read_text(encoding="utf-8"))
    except OSError as exc:
        raise ValueError(f"provider registry read failure: {exc}") from exc
    except json.JSONDecodeError as exc:
        raise ValueError(f"provider registry parse failure: {exc}") from exc

    if not isinstance(data, dict):
        raise ValueError("provider registry must be a JSON object")
    if data.get("schema_version") != PROVIDER_LANES_SCHEMA_VERSION:
        raise ValueError("provider registry schema_version must be mar.provider_lanes.v1")
    if not isinstance(data.get("default_reviewers"), list):
        raise ValueError("provider registry default_reviewers must be a list")
    if not isinstance(data.get("reviewers"), dict):
        raise ValueError("provider registry reviewers must be an object")
    return data


def parse_toolchain_scalar(value):
    value = value.strip()
    if value == "":
        return ""
    if value in ("null", "Null", "NULL", "~"):
        return None
    if value in ("true", "True", "TRUE"):
        return True
    if value in ("false", "False", "FALSE"):
        return False
    if (value.startswith("\"") and value.endswith("\"")) or (
        value.startswith("'") and value.endswith("'")
    ):
        return value[1:-1]
    return value


def parse_minimal_toolchain_yaml(raw):
    """Parse the small YAML subset used by .kkachi/toolchain.yaml.

    This script intentionally stays stdlib-only. It only needs to understand the
    existing toolchain file plus the MAR provider proof section; unsupported YAML
    constructs are ignored rather than treated as executable shell snippets.
    """
    data = {}
    mar_tools = None
    providers = None
    shell_probe = None
    current_provider = None
    current_list_key = None
    in_mar_tools = False
    in_providers = False
    in_shell_probe = False

    for raw_line in raw.splitlines():
        line_without_comment = raw_line.split("#", 1)[0].rstrip()
        if not line_without_comment.strip():
            continue
        indent = len(line_without_comment) - len(line_without_comment.lstrip(" "))
        line = line_without_comment.strip()

        if indent == 0:
            in_mar_tools = False
            in_providers = False
            in_shell_probe = False
            current_provider = None
            current_list_key = None
            if line in ("mar_provider_tools:", "provider_tools:"):
                mar_tools = {}
                data["mar_provider_tools"] = mar_tools
                in_mar_tools = True
            else:
                if ":" in line:
                    key, value = line.split(":", 1)
                    data[key.strip()] = parse_toolchain_scalar(value)
            continue

        if not in_mar_tools or mar_tools is None:
            continue

        if indent == 2 and ":" in line:
            key, value = line.split(":", 1)
            key = key.strip()
            value = value.strip()
            in_providers = False
            in_shell_probe = False
            current_provider = None
            current_list_key = None
            if key == "providers" and value == "":
                providers = {}
                mar_tools["providers"] = providers
                in_providers = True
            elif key == "shell_probe" and value == "":
                shell_probe = {}
                mar_tools["shell_probe"] = shell_probe
                in_shell_probe = True
            else:
                mar_tools[key] = parse_toolchain_scalar(value)
            continue

        if in_shell_probe and shell_probe is not None and indent == 4 and ":" in line:
            key, value = line.split(":", 1)
            shell_probe[key.strip()] = parse_toolchain_scalar(value)
            continue

        if in_providers and providers is not None:
            if indent == 4 and line.endswith(":"):
                current_provider = line[:-1].strip()
                providers[current_provider] = {}
                current_list_key = None
                continue
            if current_provider and indent == 6 and ":" in line:
                key, value = line.split(":", 1)
                key = key.strip()
                value = value.strip()
                if value == "":
                    providers[current_provider][key] = [] if key == "resolved_argv" else {}
                    current_list_key = key
                else:
                    providers[current_provider][key] = parse_toolchain_scalar(value)
                    current_list_key = None
                continue
            if (
                current_provider
                and current_list_key
                and indent == 8
                and line.startswith("- ")
                and isinstance(providers[current_provider].get(current_list_key), list)
            ):
                providers[current_provider][current_list_key].append(
                    parse_toolchain_scalar(line[2:])
                )
                continue

    return data


def load_toolchain_config(path):
    if not path:
        return {}
    toolchain_path = resolve_repo_path(path)
    if not toolchain_path.exists():
        return {}
    try:
        raw = toolchain_path.read_text(encoding="utf-8")
    except OSError as exc:
        raise ValueError(f"toolchain read failure: {exc}") from exc

    stripped = raw.lstrip()
    if stripped.startswith("{"):
        try:
            data = json.loads(raw)
        except json.JSONDecodeError as exc:
            raise ValueError(f"toolchain JSON parse failure: {exc}") from exc
    else:
        data = parse_minimal_toolchain_yaml(raw)

    if not isinstance(data, dict):
        raise ValueError("toolchain config must be an object")
    return data


def overlay_provider_toolchain(registry, toolchain_path):
    registry_copy = json.loads(json.dumps(registry))
    toolchain = load_toolchain_config(toolchain_path)
    tools = toolchain.get("mar_provider_tools") or toolchain.get("provider_tools")
    if not tools:
        return registry_copy
    if not isinstance(tools, dict):
        raise ValueError("toolchain mar_provider_tools must be an object")
    if tools.get("schema_version") != PROVIDER_TOOLS_SCHEMA_VERSION:
        raise ValueError("toolchain mar_provider_tools schema_version must be mar.provider_tools.v1")
    providers = tools.get("providers") or {}
    if not isinstance(providers, dict):
        raise ValueError("toolchain mar_provider_tools.providers must be an object")

    reviewers = registry_copy.get("reviewers", {})
    applied = []
    for reviewer_id, provider_tools in providers.items():
        if reviewer_id not in reviewers:
            continue
        if not isinstance(provider_tools, dict):
            raise ValueError(f"toolchain provider {reviewer_id} must be an object")
        config = reviewers[reviewer_id]
        if "resolved_argv" in provider_tools:
            resolved_argv = provider_tools["resolved_argv"]
            if not isinstance(resolved_argv, list) or not resolved_argv or not all(
                isinstance(item, str) and item for item in resolved_argv
            ):
                raise ValueError(f"toolchain provider {reviewer_id} resolved_argv must be a non-empty string list")
            config["resolved_argv"] = list(resolved_argv)
        for key in (
            "command_lane",
            "selected_model",
            "validated",
            "version",
            "reason",
            "validation_evidence",
            "adapter_proof_evidence",
        ):
            if key in provider_tools:
                target_key = "adapter_proof_evidence" if key == "validation_evidence" else key
                config[target_key] = provider_tools[key]
        config["toolchain_overlay"] = {
            "path": str(resolve_repo_path(toolchain_path)),
            "schema_version": tools.get("schema_version"),
        }
        applied.append(reviewer_id)
    if applied:
        registry_copy["toolchain_overlay"] = {
            "path": str(resolve_repo_path(toolchain_path)),
            "applied_reviewers": applied,
        }
    return registry_copy


def load_provider_registry_with_toolchain(registry_path, toolchain_path):
    registry = load_provider_registry(registry_path)
    return overlay_provider_toolchain(registry, toolchain_path)


def provider_base_command(config):
    resolved_argv = config.get("resolved_argv")
    if resolved_argv:
        return [str(item) for item in resolved_argv]
    executable = config.get("executable") or config.get("command_lane")
    return [str(executable)] if executable else []


def command_head_available(command):
    if not command:
        return False
    head = str(command[0])
    head_path = Path(head)
    if head_path.is_absolute() or "/" in head:
        return head_path.exists()
    return shutil.which(head) is not None


def registry_payload_for_readback(registry):
    payload = dict(registry)
    payload["status"] = "PASS"
    payload["provider_failure_reasons"] = list(PROVIDER_FAILURE_REASONS)
    payload["no_provider_execution"] = True
    return payload


def parse_reviewers_arg(value):
    if not value:
        return None
    reviewers = [item.strip() for item in value.split(",") if item.strip()]
    return reviewers or None


def selected_reviewers_or_blocked(args, registry):
    default_reviewers = list(registry.get("default_reviewers", []))
    requested = parse_reviewers_arg(getattr(args, "reviewers", None))
    if not requested:
        return default_reviewers, None
    if requested != default_reviewers and not getattr(args, "pre_scoped_evidence", None):
        return requested, structured_failure(
            "BLOCKED",
            "pre_scoped_evidence_required",
            required_default_reviewers=default_reviewers,
            requested_reviewers=requested,
            no_provider_execution=True,
        )
    return requested, None


def reviewer_config(registry, reviewer_id):
    reviewers = registry.get("reviewers", {})
    if reviewer_id not in reviewers:
        raise ValueError(f"unknown reviewer_id: {reviewer_id}")
    config = dict(reviewers[reviewer_id])
    config["reviewer_id"] = reviewer_id
    return config


def redacted_command(command):
    redacted = []
    skip_next = False
    secret_flags = {"--token", "--api-key", "--apikey", "--password", "--secret"}
    for item in command:
        if skip_next:
            redacted.append("<redacted>")
            skip_next = False
            continue
        if item in secret_flags:
            redacted.append(item)
            skip_next = True
            continue
        lowered = item.lower()
        if "token=" in lowered or "api_key=" in lowered or "apikey=" in lowered:
            redacted.append("<redacted>")
        else:
            redacted.append(item)
    return redacted


def build_provider_command(config, prompt_path, prompt_text, timeout_seconds):
    args = config.get("command_args") or []
    model = config.get("selected_model")
    replacements = {
        "{model}": "" if model is None else str(model),
        "{prompt_path}": "" if prompt_path is None else str(prompt_path),
        "{prompt_text}": "" if prompt_text is None else str(prompt_text),
        "{timeout_seconds}": str(timeout_seconds),
    }
    command = provider_base_command(config)
    for item in args:
        value = str(item)
        for old, new in replacements.items():
            value = value.replace(old, new)
        command.append(value)
    return command


def classify_provider_failure(output):
    lowered = (output or "").lower()
    if "auth" in lowered or "login" in lowered or "unauthorized" in lowered:
        return "auth_failed"
    if "token" in lowered and ("exhaust" in lowered or "limit" in lowered):
        return "token_exhausted"
    if "quota" in lowered or "billing" in lowered:
        return "quota_exhausted"
    if "rate limit" in lowered or "rate_limited" in lowered or "too many requests" in lowered:
        return "rate_limited"
    if "model" in lowered and (
        "unavailable" in lowered
        or "not found" in lowered
        or "unsupported" in lowered
        or "unknown" in lowered
    ):
        return "model_unavailable"
    return "nonzero_exit"


def base_provider_attempt(args, config, attempt_id, timeout_seconds):
    started_at = utc_now()
    selected_model = config.get("selected_model")
    command = build_provider_command(
        config,
        getattr(args, "prompt", None),
        getattr(args, "prompt_text", None),
        timeout_seconds,
    )
    return {
        "schema_version": PROVIDER_ATTEMPT_SCHEMA_VERSION,
        "run_id": getattr(args, "run_id", None),
        "task_id": getattr(args, "task_id", None),
        "attempt_id": attempt_id,
        "reviewer_id": config["reviewer_id"],
        "command_lane": config.get("command_lane"),
        "selected_model": selected_model,
        "started_at": started_at,
        "ended_at": started_at,
        "timeout_seconds": timeout_seconds,
        "exit_code": None,
        "terminal_status": "BLOCKED",
        "provider_failure_reason": None,
        "parser_status": "not_run",
        "mutation_check": {
            "checked": False,
            "detected": False,
        },
        "redacted_command": redacted_command(command),
        "preflight_evidence_path": getattr(args, "preflight_evidence_path", None),
        "raw_output_path": None,
        "parsed_finding_path": None,
        "capped_output_note": {
            "cap_bytes": getattr(args, "raw_cap", DEFAULT_RAW_CAP),
            "truncated": False,
        },
        "retry_of_attempt_id": getattr(args, "retry_of_attempt_id", None),
        "alternate_for_reviewer_id": getattr(args, "alternate_for_reviewer_id", None),
        "approval_evidence": getattr(args, "approval_evidence", None),
        "waiver_evidence": getattr(args, "waiver_evidence", None),
        "no_provider_execution": True,
    }


def write_attempt_outputs(args, attempt, raw_text=None, parsed=None):
    output_dir = getattr(args, "output_dir", None)
    if not output_dir:
        return attempt

    root = Path(output_dir)
    raw_dir = root / "raw"
    parsed_dir = root / "parsed"
    attempts_dir = root / "attempts"
    raw_dir.mkdir(parents=True, exist_ok=True)
    parsed_dir.mkdir(parents=True, exist_ok=True)
    attempts_dir.mkdir(parents=True, exist_ok=True)

    attempt_id = attempt["attempt_id"]
    if raw_text is not None:
        raw_path = raw_dir / f"{attempt_id}.txt"
        raw_path.write_text(raw_text, encoding="utf-8")
        attempt["raw_output_path"] = str(raw_path)
    if parsed is not None:
        parsed_path = parsed_dir / f"{attempt_id}.json"
        parsed_path.write_text(json.dumps(parsed, indent=2, sort_keys=True), encoding="utf-8")
        attempt["parsed_finding_path"] = str(parsed_path)

    attempt_path = attempts_dir / f"{attempt_id}.json"
    attempt_path.write_text(json.dumps(attempt, indent=2, sort_keys=True), encoding="utf-8")
    attempt["provider_attempt_path"] = str(attempt_path)
    return attempt


def provider_preflight_attempt(args, config, attempt_id=None):
    timeout_seconds = getattr(args, "timeout", None) or config.get("timeout_seconds") or DEFAULT_TIMEOUT
    attempt = base_provider_attempt(
        args,
        config,
        attempt_id or f"{config['reviewer_id']}-preflight",
        timeout_seconds,
    )

    before = run_git_status()
    after = before
    attempt["mutation_check"] = {
        "checked": before["available"],
        "detected": False,
    }

    if not config.get("selected_model"):
        attempt["ended_at"] = utc_now()
        attempt["terminal_status"] = "BLOCKED"
        attempt["provider_failure_reason"] = "model_unavailable"
        return attempt

    command = provider_base_command(config)
    if not command_head_available(command):
        attempt["ended_at"] = utc_now()
        attempt["terminal_status"] = "DEGRADED"
        attempt["provider_failure_reason"] = "cli_missing"
        return attempt

    if config.get("validation_required_before_success_coverage") and not config.get("validated"):
        attempt["ended_at"] = utc_now()
        attempt["terminal_status"] = "BLOCKED"
        attempt["provider_failure_reason"] = "adapter_proof_required"
        attempt["preflight_evidence_path"] = config.get("adapter_proof_evidence")
        return attempt

    if getattr(args, "preflight_only", False):
        attempt["ended_at"] = utc_now()
        attempt["terminal_status"] = "PASS"
        attempt["parser_status"] = "not_applicable"
        attempt["provider_failure_reason"] = None
        return attempt

    command = attempt["redacted_command"]
    raw_text = ""
    parsed = None
    try:
        completed = subprocess.run(
            build_provider_command(
                config,
                getattr(args, "prompt", None),
                getattr(args, "prompt_text", None),
                timeout_seconds,
            ),
            shell=False,
            timeout=timeout_seconds,
            check=False,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=os.environ.copy(),
        )
        raw_text = (completed.stdout or "") + (completed.stderr or "")
        attempt["exit_code"] = completed.returncode
        attempt["no_provider_execution"] = False
    except subprocess.TimeoutExpired as exc:
        raw_text = ((exc.stdout or "") if isinstance(exc.stdout, str) else "") + (
            (exc.stderr or "") if isinstance(exc.stderr, str) else ""
        )
        attempt["ended_at"] = utc_now()
        attempt["terminal_status"] = "BLOCKED"
        attempt["provider_failure_reason"] = "timeout"
        capped = cap_raw_output(raw_text, getattr(args, "raw_cap", DEFAULT_RAW_CAP))
        attempt["capped_output_note"] = {
            "cap_bytes": capped["cap_bytes"],
            "truncated": capped["truncated"],
        }
        return write_attempt_outputs(args, attempt, raw_text=raw_text)
    finally:
        after = run_git_status()

    attempt["mutation_check"] = {
        "checked": before["available"] and after["available"],
        "detected": before["available"] and after["available"] and before["stdout"] != after["stdout"],
    }
    capped = cap_raw_output(raw_text, getattr(args, "raw_cap", DEFAULT_RAW_CAP))
    attempt["capped_output_note"] = {
        "cap_bytes": capped["cap_bytes"],
        "truncated": capped["truncated"],
    }

    if attempt["mutation_check"]["detected"]:
        attempt["ended_at"] = utc_now()
        attempt["terminal_status"] = "BLOCKED"
        attempt["provider_failure_reason"] = "mutation_detected"
        return write_attempt_outputs(args, attempt, raw_text=raw_text)

    if attempt["exit_code"] != 0:
        attempt["ended_at"] = utc_now()
        attempt["terminal_status"] = "DEGRADED"
        attempt["provider_failure_reason"] = classify_provider_failure(raw_text)
        return write_attempt_outputs(args, attempt, raw_text=raw_text)

    parsed = parse_review_raw(raw_text, getattr(args, "raw_cap", DEFAULT_RAW_CAP))
    if parsed.get("reason") == "parse_failure" or parsed.get("status") in ("FAILED", "DEGRADED") and parsed.get("reason") == "invalid_status":
        attempt["terminal_status"] = "DEGRADED"
        attempt["provider_failure_reason"] = "parse_failure"
        attempt["parser_status"] = "parse_failure"
    else:
        attempt["terminal_status"] = parsed.get("status", "DEGRADED")
        attempt["provider_failure_reason"] = None
        attempt["parser_status"] = "parsed"

    if attempt["terminal_status"] in NON_CLEAN_STATUSES and not attempt["provider_failure_reason"]:
        attempt["provider_failure_reason"] = "unknown_provider_failure"
    attempt["ended_at"] = utc_now()
    # Keep a local variable use so future command-redaction changes are covered by
    # the attempt object without emitting raw command arguments.
    attempt["redacted_command"] = command
    return write_attempt_outputs(args, attempt, raw_text=raw_text, parsed=parsed)


def with_mutation_guard(fn, args):
    before = run_git_status()
    payload = fn(args)
    after = run_git_status()

    guard = {
        "checked": before["available"] and after["available"],
        "detected": False,
    }
    if before["available"] and after["available"]:
        guard["before"] = before["stdout"]
        guard["after"] = after["stdout"]
        guard["detected"] = before["stdout"] != after["stdout"]
    else:
        guard["degraded_reason"] = "git_status_unavailable"
        guard["before_returncode"] = before["returncode"]
        guard["after_returncode"] = after["returncode"]

    if guard["detected"]:
        return structured_failure(
            "BLOCKED",
            "mutation_detected",
            mutation_guard=guard,
            no_provider_execution=True,
        )

    payload["mutation_guard"] = guard
    payload.setdefault("no_provider_execution", True)
    return payload


def cmd_doctor(args):
    payload = {
        "status": "PASS",
        "capability_status": "PASS",
        "tool": "mar.py",
        "stdlib_only": True,
        "no_provider_execution": True,
        "subcommands": [
            "doctor",
            "render",
            "validate",
            "merge-pack",
            "provider-lanes",
            "provider-preflight",
            "provider-attempt",
        ],
        "allowed_statuses": list(STATUS_VOCABULARY),
        "provider_failure_reasons": list(PROVIDER_FAILURE_REASONS),
        "raw_output_cap_bytes": DEFAULT_RAW_CAP,
    }
    if args.fixture:
        try:
            payload["fixture_evidence"] = parse_review_path(args.fixture, args.raw_cap)
        except OSError as exc:
            payload["fixture_evidence"] = structured_failure(
                "FAILED",
                "fixture_read_failure",
                path=str(args.fixture),
                detail=str(exc),
            )
    return payload


def parse_set_values(values):
    context = {}
    for item in values:
        if "=" not in item:
            raise ValueError("--set values must use name=value")
        key, value = item.split("=", 1)
        context[key] = value
    return context


def render_template(template, context):
    rendered = template
    for key, value in context.items():
        rendered = rendered.replace("{{ ." + key + " }}", str(value))
        rendered = rendered.replace("{{." + key + "}}", str(value))
    return rendered


def cmd_render(args):
    try:
        context = parse_set_values(args.set)
        if args.context_json:
            loaded = json.loads(load_text(args.context_json))
            if not isinstance(loaded, dict):
                return structured_failure("FAILED", "context_json_must_be_object")
            context.update(loaded)
    except (OSError, ValueError, json.JSONDecodeError) as exc:
        return structured_failure("FAILED", "render_input_failure", detail=str(exc))

    try:
        template = load_text(args.template)
    except OSError as exc:
        return structured_failure("FAILED", "template_read_failure", detail=str(exc))

    rendered = render_template(template, context)
    return {
        "status": "PASS",
        "template": str(args.template),
        "rendered_prompt": rendered,
        "no_provider_execution": True,
    }


def cmd_validate(args):
    try:
        return parse_review_path(args.input, args.raw_cap)
    except OSError as exc:
        return structured_failure(
            "FAILED",
            "input_read_failure",
            path=str(args.input),
            detail=str(exc),
        )


def cmd_provider_lanes(args):
    try:
        registry = load_provider_registry_with_toolchain(args.registry, args.toolchain)
    except ValueError as exc:
        return structured_failure("FAILED", "provider_registry_failure", detail=str(exc))
    return registry_payload_for_readback(registry)


def cmd_provider_preflight(args):
    try:
        registry = load_provider_registry_with_toolchain(args.registry, args.toolchain)
        reviewers, blocked = selected_reviewers_or_blocked(args, registry)
        if blocked:
            return blocked
        attempts = []
        for reviewer_id in reviewers:
            config = reviewer_config(registry, reviewer_id)
            preflight_args = argparse.Namespace(**vars(args))
            preflight_args.preflight_only = True
            attempts.append(provider_preflight_attempt(preflight_args, config))
    except ValueError as exc:
        return structured_failure("FAILED", "provider_registry_failure", detail=str(exc))

    statuses = [attempt.get("terminal_status") for attempt in attempts]
    reasons = [
        attempt.get("provider_failure_reason")
        for attempt in attempts
        if attempt.get("provider_failure_reason")
    ]
    if statuses and all(status in COVERED_STATUSES for status in statuses):
        status, reason = "PASS", "provider_preflight_passed"
    elif statuses and all(status in NON_CLEAN_STATUSES for status in statuses):
        status, reason = "FAILED", "all_provider_preflight_failed"
    else:
        status, reason = "DEGRADED", "provider_preflight_degraded"
    return {
        "status": status,
        "reason": reason,
        "schema_version": "mar.provider_preflight.v1",
        "default_reviewers": registry.get("default_reviewers", []),
        "requested_reviewers": reviewers,
        "provider_failure_reasons": reasons,
        "attempts": attempts,
        "no_provider_execution": True,
    }


def cmd_provider_attempt(args):
    try:
        registry = load_provider_registry_with_toolchain(args.registry, args.toolchain)
        config = reviewer_config(registry, args.reviewer)
    except ValueError as exc:
        return structured_failure("FAILED", "provider_registry_failure", detail=str(exc))

    attempt_id = args.attempt_id or f"{args.reviewer}-001"
    return provider_preflight_attempt(args, config, attempt_id=attempt_id)


def parse_waivers(values):
    waivers = {}
    for value in values:
        if "=" not in value:
            raise ValueError("--waiver must use reviewer_id=evidence_ref")
        reviewer_id, evidence = value.split("=", 1)
        reviewer_id = reviewer_id.strip()
        evidence = evidence.strip()
        if not reviewer_id or not evidence:
            raise ValueError("--waiver must include reviewer_id and evidence_ref")
        waivers[reviewer_id] = evidence
    return waivers


def load_json_object(path, raw_cap):
    raw = load_text(path)
    capped = cap_raw_output(raw, raw_cap)
    try:
        parsed = json.loads(raw)
    except json.JSONDecodeError as exc:
        return structured_failure(
            "DEGRADED",
            "parse_failure",
            path=str(path),
            parse_failure={"message": str(exc), "line": exc.lineno, "column": exc.colno},
            raw_output=capped,
        )
    if not isinstance(parsed, dict):
        return structured_failure(
            "FAILED",
            "review_json_must_be_object",
            path=str(path),
            raw_output=capped,
        )
    parsed.setdefault("path", str(path))
    return parsed


def is_provider_attempt(item):
    return item.get("schema_version") == PROVIDER_ATTEMPT_SCHEMA_VERSION


def aggregate_provider_coverage(attempts, required_reviewers, waivers, pre_scoped_evidence=None):
    coverage_by_reviewer = {}
    unresolved = []
    resolved = []
    all_terminal_statuses = [attempt.get("terminal_status") for attempt in attempts]

    for reviewer_id in required_reviewers:
        direct_attempts = [
            attempt
            for attempt in attempts
            if attempt.get("reviewer_id") == reviewer_id
            and not attempt.get("alternate_for_reviewer_id")
        ]
        direct_success = [
            attempt
            for attempt in direct_attempts
            if attempt.get("terminal_status") in COVERED_STATUSES
        ]
        retry_success = [
            attempt
            for attempt in direct_success
            if attempt.get("retry_of_attempt_id")
        ]
        alternate_success = [
            attempt
            for attempt in attempts
            if attempt.get("alternate_for_reviewer_id") == reviewer_id
            and attempt.get("terminal_status") in COVERED_STATUSES
            and attempt.get("approval_evidence")
        ]
        alternate_without_approval = [
            attempt
            for attempt in attempts
            if attempt.get("alternate_for_reviewer_id") == reviewer_id
            and attempt.get("terminal_status") in COVERED_STATUSES
            and not attempt.get("approval_evidence")
        ]

        if retry_success:
            resolution = {
                "reviewer_id": reviewer_id,
                "state": "resolved",
                "resolution": "same_provider_retry_success",
                "attempt_id": retry_success[-1].get("attempt_id"),
                "retry_of_attempt_id": retry_success[-1].get("retry_of_attempt_id"),
            }
            coverage_by_reviewer[reviewer_id] = resolution
            resolved.append(resolution)
        elif direct_success:
            resolution = {
                "reviewer_id": reviewer_id,
                "state": "covered",
                "resolution": "default_provider_success",
                "attempt_id": direct_success[-1].get("attempt_id"),
            }
            coverage_by_reviewer[reviewer_id] = resolution
            resolved.append(resolution)
        elif alternate_success:
            resolution = {
                "reviewer_id": reviewer_id,
                "state": "resolved",
                "resolution": "approved_alternate_success",
                "attempt_id": alternate_success[-1].get("attempt_id"),
                "approval_evidence": alternate_success[-1].get("approval_evidence"),
            }
            coverage_by_reviewer[reviewer_id] = resolution
            resolved.append(resolution)
        elif reviewer_id in waivers:
            resolution = {
                "reviewer_id": reviewer_id,
                "state": "resolved",
                "resolution": "explicit_waiver",
                "waiver_evidence": waivers[reviewer_id],
            }
            coverage_by_reviewer[reviewer_id] = resolution
            resolved.append(resolution)
        else:
            failure_reasons = [
                attempt.get("provider_failure_reason")
                for attempt in direct_attempts
                if attempt.get("provider_failure_reason")
            ]
            reason = "unresolved_default_coverage"
            if alternate_without_approval:
                reason = "approval_evidence_required"
            elif not direct_attempts:
                reason = "missing_provider_attempt"
            unresolved_item = {
                "reviewer_id": reviewer_id,
                "state": "unresolved",
                "reason": reason,
                "provider_failure_reasons": failure_reasons,
            }
            coverage_by_reviewer[reviewer_id] = unresolved_item
            unresolved.append(unresolved_item)

    actionable_statuses = [
        status for status in all_terminal_statuses if status == "REQUEST_CHANGES"
    ]
    if actionable_statuses:
        status, reason = "REQUEST_CHANGES", "actionable_findings"
    elif unresolved:
        attempted_required = [
            attempt
            for attempt in attempts
            if attempt.get("reviewer_id") in required_reviewers
            and not attempt.get("alternate_for_reviewer_id")
        ]
        if attempted_required and len(unresolved) == len(required_reviewers):
            status, reason = "FAILED", "all_default_reviewers_unresolved"
        else:
            status, reason = "DEGRADED", "unresolved_default_reviewer_coverage"
    else:
        status, reason = "PASS", "all_required_provider_coverage_resolved"

    return {
        "status": status,
        "reason": reason,
        "coverage": {
            "required_reviewers": required_reviewers,
            "observed_reviewers": sorted(
                {
                    attempt.get("reviewer_id")
                    for attempt in attempts
                    if attempt.get("reviewer_id")
                }
            ),
            "minimum_met": not unresolved,
            "pre_scoped_evidence": pre_scoped_evidence,
            "resolved": resolved,
            "by_reviewer": coverage_by_reviewer,
            "unresolved_default_reviewers": [
                item["reviewer_id"] for item in unresolved
            ],
        },
        "provider_attempts": attempts,
        "no_provider_execution": True,
    }


def cmd_merge_pack(args):
    if args.provider_coverage:
        try:
            registry = load_provider_registry(args.registry)
            reviewers, blocked = selected_reviewers_or_blocked(args, registry)
            if blocked:
                return blocked
            waivers = parse_waivers(args.waiver)
        except ValueError as exc:
            return structured_failure("FAILED", "provider_registry_failure", detail=str(exc))

        attempts = []
        for path in args.inputs:
            try:
                attempts.append(load_json_object(path, args.raw_cap))
            except OSError as exc:
                attempts.append(
                    structured_failure(
                        "FAILED",
                        "input_read_failure",
                        path=str(path),
                        detail=str(exc),
                    )
                )
        provider_attempts = [item for item in attempts if is_provider_attempt(item)]
        return aggregate_provider_coverage(
            provider_attempts,
            reviewers,
            waivers,
            pre_scoped_evidence=args.pre_scoped_evidence,
        )

    reviews = []
    for path in args.inputs:
        try:
            reviews.append(parse_review_path(path, args.raw_cap))
        except OSError as exc:
            reviews.append(
                structured_failure(
                    "FAILED",
                    "input_read_failure",
                    path=str(path),
                    detail=str(exc),
                )
            )

    status, reason = aggregate_status(reviews, args.required_reviewers)
    return {
        "status": status,
        "reason": reason,
        "coverage": {
            "required_reviewers": args.required_reviewers,
            "observed_reviewers": len(reviews),
            "minimum_met": not args.required_reviewers
            or len(reviews) >= args.required_reviewers,
        },
        "reviews": reviews,
        "no_provider_execution": True,
    }


def build_parser():
    parser = argparse.ArgumentParser(
        description="Local fixture/mock MAR helper with fail-closed provider attempt scaffolding."
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    doctor = subparsers.add_parser("doctor", help="report local MAR MVP capability")
    doctor.add_argument("--fixture")
    doctor.add_argument("--raw-cap", type=int, default=DEFAULT_RAW_CAP)
    doctor.set_defaults(func=cmd_doctor, guarded=True)

    render = subparsers.add_parser("render", help="render a local prompt template")
    render.add_argument("--template", required=True)
    render.add_argument("--context-json")
    render.add_argument("--set", action="append", default=[])
    render.set_defaults(func=cmd_render, guarded=True)

    validate = subparsers.add_parser("validate", help="validate local review JSON or raw output")
    validate.add_argument("--input", required=True)
    validate.add_argument("--raw-cap", type=int, default=DEFAULT_RAW_CAP)
    validate.set_defaults(func=cmd_validate, guarded=True)

    merge_pack = subparsers.add_parser(
        "merge-pack", help="merge local review JSON/raw evidence into a compact pack"
    )
    merge_pack.add_argument("inputs", nargs="+")
    merge_pack.add_argument("--required-reviewers", type=int, default=0)
    merge_pack.add_argument("--raw-cap", type=int, default=DEFAULT_RAW_CAP)
    merge_pack.add_argument("--provider-coverage", action="store_true")
    merge_pack.add_argument("--registry", default=DEFAULT_PROVIDER_REGISTRY)
    merge_pack.add_argument("--reviewers")
    merge_pack.add_argument("--pre-scoped-evidence")
    merge_pack.add_argument("--waiver", action="append", default=[])
    merge_pack.set_defaults(func=cmd_merge_pack, guarded=True)

    provider_lanes = subparsers.add_parser(
        "provider-lanes", help="read back MAR provider lane registry"
    )
    provider_lanes.add_argument("--registry", default=DEFAULT_PROVIDER_REGISTRY)
    provider_lanes.add_argument("--toolchain", default=DEFAULT_TOOLCHAIN)
    provider_lanes.set_defaults(func=cmd_provider_lanes, guarded=True)

    provider_preflight = subparsers.add_parser(
        "provider-preflight", help="preflight MAR provider lanes without review prompts"
    )
    provider_preflight.add_argument("--registry", default=DEFAULT_PROVIDER_REGISTRY)
    provider_preflight.add_argument("--toolchain", default=DEFAULT_TOOLCHAIN)
    provider_preflight.add_argument("--reviewers")
    provider_preflight.add_argument("--pre-scoped-evidence")
    provider_preflight.add_argument("--run-id")
    provider_preflight.add_argument("--task-id")
    provider_preflight.add_argument("--prompt")
    provider_preflight.add_argument("--prompt-text")
    provider_preflight.add_argument("--timeout", type=int)
    provider_preflight.add_argument("--raw-cap", type=int, default=DEFAULT_RAW_CAP)
    provider_preflight.add_argument("--output-dir")
    provider_preflight.add_argument("--preflight-evidence-path")
    provider_preflight.set_defaults(func=cmd_provider_preflight, guarded=True)

    provider_attempt = subparsers.add_parser(
        "provider-attempt", help="run one MAR provider attempt with fail-closed evidence"
    )
    provider_attempt.add_argument("--registry", default=DEFAULT_PROVIDER_REGISTRY)
    provider_attempt.add_argument("--toolchain", default=DEFAULT_TOOLCHAIN)
    provider_attempt.add_argument("--reviewer", required=True)
    provider_attempt.add_argument("--run-id")
    provider_attempt.add_argument("--task-id")
    provider_attempt.add_argument("--attempt-id")
    provider_attempt.add_argument("--prompt")
    provider_attempt.add_argument("--prompt-text")
    provider_attempt.add_argument("--timeout", type=int)
    provider_attempt.add_argument("--raw-cap", type=int, default=DEFAULT_RAW_CAP)
    provider_attempt.add_argument("--output-dir")
    provider_attempt.add_argument("--preflight-evidence-path")
    provider_attempt.add_argument("--retry-of-attempt-id")
    provider_attempt.add_argument("--alternate-for-reviewer-id")
    provider_attempt.add_argument("--approval-evidence")
    provider_attempt.add_argument("--waiver-evidence")
    provider_attempt.set_defaults(func=cmd_provider_attempt, guarded=True)

    return parser


def main(argv):
    parser = build_parser()
    args = parser.parse_args(argv)
    payload = with_mutation_guard(args.func, args) if args.guarded else args.func(args)
    emit(payload)
    return status_exit(payload)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
