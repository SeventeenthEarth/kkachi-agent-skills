from __future__ import annotations

import json
import subprocess
from pathlib import Path
from typing import Any

from .discovery import SourcePack, discover_source_packs, find_source_repo


NEXT_ACTION = "Run install --dry-run before any profile writes."


def build_list_result(
    repo: Path | str | None = None,
    *,
    category: str | None = None,
    profile: str | None = None,
    profile_root: Path | str | None = None,
) -> dict[str, Any]:
    effective_repo = find_source_repo(Path(repo)) if repo is not None else find_source_repo()
    packs = discover_source_packs(effective_repo)

    diagnostics: list[dict[str, str]] = []
    if category:
        filtered = [pack for pack in packs if pack.category == category]
        if not filtered:
            diagnostics.append(
                {
                    "level": "info",
                    "code": "unknown_category",
                    "message": f"카테고리 '{category}'에 해당하는 KAS pack이 없습니다.",
                }
            )
        packs = filtered

    target_profile, install_entries = _load_profile(profile, profile_root)
    pack_payloads = [_pack_payload(pack, install_entries, target_profile) for pack in packs]

    result: dict[str, Any] = {
        "ok": True,
        "command": "list",
        "source_repo": _source_repo_info(effective_repo),
        "packs": pack_payloads,
        "diagnostics": diagnostics,
        "next_action": NEXT_ACTION,
    }
    if target_profile is not None:
        result["target_profile"] = target_profile
    return result


def render_human_list(result: dict[str, Any]) -> str:
    packs = result["packs"]
    profile = result.get("target_profile")
    lines = [f"상태: 조회 완료 — KAS pack {len(packs)}개 발견."]
    if profile:
        counts = _installed_counts(packs)
        lines.append(
            "설치 상태: "
            f"current {counts.get('installed_current', 0)}, "
            f"missing {counts.get('not_installed', 0) + counts.get('manifest_missing', 0)}, "
            f"drifted {counts.get('installed_drifted', 0)}, "
            f"unknown {counts.get('installed_unknown', 0)}, "
            f"conflict {counts.get('conflict', 0)}, "
            f"error {counts.get('error', 0)}."
        )
    source = result["source_repo"]
    lines.append(f"소스: {source['path']} @ {source.get('git_commit') or 'unknown'}")
    for diagnostic in result["diagnostics"]:
        lines.append(f"진단: {diagnostic['message']}")
    lines.append("다음: 설치 전 `install --dry-run`으로 변경 경로를 확인하세요.")
    return "\n".join(lines)


def _pack_payload(
    pack: SourcePack,
    install_entries: dict[str, dict[str, Any]] | None,
    target_profile: dict[str, Any] | None,
) -> dict[str, Any]:
    payload = pack.as_dict()
    if target_profile is None:
        return payload

    manifest_state = target_profile["manifest_state"]
    if manifest_state == "manifest_missing":
        payload["installed_state"] = "manifest_missing"
        return payload
    if manifest_state == "manifest_unreadable":
        payload["installed_state"] = "error"
        return payload

    install = (install_entries or {}).get(pack.pack_id)
    if install is None:
        payload["installed_state"] = "not_installed"
        return payload

    target_path = install.get("target_path")
    if isinstance(target_path, str):
        payload["installed_path"] = target_path
        if _invalid_relative_path(target_path):
            payload["installed_state"] = "conflict"
            return payload

    installed_checksum = install.get("pack_checksum")
    if not isinstance(installed_checksum, str) or not installed_checksum:
        payload["installed_state"] = "installed_unknown"
    elif installed_checksum == pack.checksum:
        payload["installed_state"] = "installed_current"
    else:
        payload["installed_state"] = "installed_drifted"
    return payload


def _load_profile(
    profile: str | None,
    profile_root: Path | str | None,
) -> tuple[dict[str, Any] | None, dict[str, dict[str, Any]] | None]:
    if profile is None:
        return None, None

    root = Path(profile_root).expanduser().resolve() if profile_root is not None else Path.home() / ".hermes" / "profiles" / profile
    manifest_path = root / ".kas" / "skill-pack-manifest.json"
    target_profile = {
        "name": profile,
        "root": str(root),
        "manifest_path": str(manifest_path),
        "manifest_state": "manifest_missing",
    }
    if not manifest_path.exists():
        return target_profile, None

    try:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        target_profile["manifest_state"] = "manifest_unreadable"
        return target_profile, None

    target_profile["manifest_state"] = "manifest_present"
    installs = manifest.get("installs")
    if not isinstance(installs, list):
        return target_profile, {}

    entries: dict[str, dict[str, Any]] = {}
    for install in installs:
        if isinstance(install, dict) and isinstance(install.get("pack_id"), str):
            entries[install["pack_id"]] = install
    return target_profile, entries


def _source_repo_info(repo: Path | None) -> dict[str, Any]:
    path = str(repo) if repo is not None else str(Path.cwd().resolve())
    return {
        "path": path,
        "git_commit": _git_value(["git", "rev-parse", "HEAD"], repo),
        "dirty": _git_dirty(repo),
    }


def _git_value(cmd: list[str], cwd: Path | None) -> str | None:
    try:
        proc = subprocess.run(cmd, cwd=cwd, text=True, capture_output=True, check=False)
    except OSError:
        return None
    if proc.returncode != 0:
        return None
    return proc.stdout.strip() or None


def _git_dirty(cwd: Path | None) -> bool | None:
    try:
        proc = subprocess.run(["git", "status", "--porcelain"], cwd=cwd, text=True, capture_output=True, check=False)
    except OSError:
        return None
    if proc.returncode != 0:
        return None
    return bool(proc.stdout.strip())


def _installed_counts(packs: list[dict[str, Any]]) -> dict[str, int]:
    counts: dict[str, int] = {}
    for pack in packs:
        state = pack.get("installed_state")
        if isinstance(state, str):
            counts[state] = counts.get(state, 0) + 1
    return counts


def _invalid_relative_path(value: str) -> bool:
    parts = value.split("/")
    path = Path(value)
    return not value or path.is_absolute() or ".." in parts or "" in parts
