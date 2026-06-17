#!/usr/bin/env python3
"""Local stdlib-only Multi-Agent Review helper.

This MVP renders local prompts and normalizes fixture/mock review evidence. It
does not execute providers.
"""

import argparse
import json
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


def parse_review_path(path, raw_cap):
    raw = load_text(path)
    capped = cap_raw_output(raw, raw_cap)
    try:
        parsed = json.loads(raw)
    except json.JSONDecodeError as exc:
        return structured_failure(
            "DEGRADED",
            "parse_failure",
            path=str(path),
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
            path=str(path),
            raw_output=capped,
        )

    status = parsed.get("status")
    if not valid_status(status):
        return structured_failure(
            "FAILED",
            "invalid_status",
            path=str(path),
            observed_status=status,
            allowed_statuses=list(STATUS_VOCABULARY),
            raw_output=capped,
        )

    result = dict(parsed)
    result.setdefault("no_provider_execution", True)
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
        "subcommands": ["doctor", "render", "validate", "merge-pack"],
        "allowed_statuses": list(STATUS_VOCABULARY),
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


def cmd_merge_pack(args):
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
        description="Local fixture/mock MAR helper; provider execution is out of scope."
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
    merge_pack.set_defaults(func=cmd_merge_pack, guarded=True)

    return parser


def main(argv):
    parser = build_parser()
    args = parser.parse_args(argv)
    payload = with_mutation_guard(args.func, args) if args.guarded else args.func(args)
    emit(payload)
    return status_exit(payload)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
