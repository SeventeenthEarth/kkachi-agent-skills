"""Hermes plugin registration for Kkachi Agent Skills.

This source checkout doubles as a local-development Hermes plugin when linked
under ``$HERMES_HOME/plugins/kkachi-agent-skills``. The plugin registers the
source-controlled KAS skills as read-only plugin-qualified skills without
copying them into profile-local skill trees.
"""

from __future__ import annotations

from pathlib import Path
from typing import Any

try:  # Hermes ships PyYAML; keep a clearer failure if the environment drifts.
    import yaml
except Exception as exc:  # pragma: no cover - import-time environment guard
    yaml = None  # type: ignore[assignment]
    _yaml_import_error = exc
else:
    _yaml_import_error = None


_ROOT = Path(__file__).resolve().parent
_MANIFEST = _ROOT / "skill-pack.yaml"


def _load_manifest() -> dict[str, Any]:
    if yaml is None:  # pragma: no cover - defensive runtime guard
        raise RuntimeError(f"PyYAML is required to load {_MANIFEST}: {_yaml_import_error}")
    with _MANIFEST.open(encoding="utf-8") as handle:
        data = yaml.safe_load(handle) or {}
    if not isinstance(data, dict):
        raise RuntimeError(f"Invalid KAS skill-pack manifest at {_MANIFEST}")
    return data


def _frontmatter_description(skill_md: Path) -> str:
    try:
        text = skill_md.read_text(encoding="utf-8")
    except OSError:
        return ""
    if not text.startswith("---"):
        return ""
    parts = text.split("---", 2)
    if len(parts) < 3 or yaml is None:
        return ""
    try:
        meta = yaml.safe_load(parts[1]) or {}
    except Exception:
        return ""
    if isinstance(meta, dict) and isinstance(meta.get("description"), str):
        return meta["description"]
    return ""


def _register_skill_with_aliases(ctx: Any, skill_id: str) -> None:
    skill_md = _ROOT / "skills" / skill_id / "SKILL.md"
    description = _frontmatter_description(skill_md)
    ctx.register_skill(skill_id, skill_md, description)

    # SOT-facing ergonomic alias: ``kkachi-agent-skills:plan`` resolves to the
    # same canonical base body as ``kkachi-agent-skills:kkachi-plan``. The
    # canonical source id remains the source-controlled ``kkachi-*`` skill id.
    if skill_id.startswith("kkachi-"):
        alias = skill_id.removeprefix("kkachi-")
        if alias:
            ctx.register_skill(alias, skill_md, description)


def register(ctx: Any) -> None:
    """Register KAS source skills as plugin-qualified read-only skills."""
    manifest = _load_manifest()
    for skill_id in manifest.get("skills") or []:
        if not isinstance(skill_id, str) or not skill_id:
            raise RuntimeError(f"Invalid skill id in {_MANIFEST}: {skill_id!r}")
        _register_skill_with_aliases(ctx, skill_id)

    for guide_id in manifest.get("guides") or []:
        if not isinstance(guide_id, str) or not guide_id:
            raise RuntimeError(f"Invalid guide id in {_MANIFEST}: {guide_id!r}")
        _register_skill_with_aliases(ctx, guide_id)
