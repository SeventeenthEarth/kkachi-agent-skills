from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any

from .discovery import DiscoveryError
from .listing import build_list_result, render_human_list


def main(argv: list[str] | None = None) -> int:
    parser = _build_parser()
    args = parser.parse_args(argv)
    if args.command != "list":
        parser.error("only the list command is implemented")

    if args.profile_root and os.environ.get("KAS_ALLOW_PROFILE_ROOT_OVERRIDE") != "1":
        return _emit_error(
            "profile_root_override_rejected",
            "--profile-root is only allowed under an explicit test/harness guard.",
            json_output=args.json,
        )

    try:
        result = build_list_result(
            Path(args.repo) if args.repo else None,
            category=args.category,
            profile=args.profile,
            profile_root=Path(args.profile_root) if args.profile_root else None,
        )
    except DiscoveryError as exc:
        return _emit_error("discovery_failed", str(exc), json_output=args.json)

    if args.json:
        print(json.dumps(result, ensure_ascii=False, sort_keys=True))
    else:
        print(render_human_list(result))
    return 0


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="kkachi-hermes-skills")
    subparsers = parser.add_subparsers(dest="command", required=True)

    list_parser = subparsers.add_parser("list", help="list available KAS skill packs")
    list_parser.add_argument("--repo", help="source KAS repo path")
    list_parser.add_argument("--profile", help="Hermes target profile name")
    list_parser.add_argument("--profile-root", help="test/harness-only explicit profile root")
    list_parser.add_argument("--category", help="filter packs by category")
    list_parser.add_argument("--json", action="store_true", help="emit machine-readable JSON")
    list_parser.add_argument("--no-color", action="store_true", help="accepted for stable CLI shape; output is uncolored")
    return parser


def _emit_error(code: str, message: str, *, json_output: bool) -> int:
    payload: dict[str, Any] = {
        "ok": False,
        "command": "list",
        "source_repo": None,
        "packs": [],
        "diagnostics": [{"level": "error", "code": code, "message": message}],
        "next_action": "Fix the reported issue and rerun list.",
    }
    if json_output:
        print(json.dumps(payload, ensure_ascii=False, sort_keys=True), file=sys.stderr)
    else:
        print(f"오류: {message}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
