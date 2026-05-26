from __future__ import annotations

import json
import stat
from dataclasses import dataclass
from hashlib import sha256
from pathlib import Path


class DiscoveryError(RuntimeError):
    """Raised when KAS source packs cannot be discovered safely."""


@dataclass(frozen=True)
class SourcePack:
    pack_id: str
    category: str
    name: str
    source_path: str
    description: str | None
    checksum: str

    def as_dict(self) -> dict[str, str]:
        data = {
            "pack_id": self.pack_id,
            "category": self.category,
            "name": self.name,
            "source_path": self.source_path,
        }
        if self.description:
            data["description"] = self.description
        return data


def find_source_repo(start: Path | None = None) -> Path:
    current = (start or Path.cwd()).resolve()
    if not current.exists():
        raise DiscoveryError(f"source repo path does not exist: {current}")
    candidates = [current, *current.parents] if current.is_dir() else [current.parent, *current.parent.parents]
    for candidate in candidates:
        if (candidate / "skills").is_dir():
            return candidate
    raise DiscoveryError(f"source repo not found from {current}")


def discover_source_packs(repo: Path | str | None = None) -> list[SourcePack]:
    source_repo = find_source_repo(Path(repo)) if repo is not None else find_source_repo()
    skills_dir = source_repo / "skills"
    if not skills_dir.is_dir():
        raise DiscoveryError(f"source repo has no skills directory: {source_repo}")

    packs: list[SourcePack] = []
    for child in sorted(path for path in skills_dir.iterdir() if path.is_dir()):
        direct_skill = child / "SKILL.md"
        if direct_skill.is_file():
            packs.append(_read_pack(source_repo, child, "core", child.name))
            continue
        for grandchild in sorted(path for path in child.iterdir() if path.is_dir()):
            if (grandchild / "SKILL.md").is_file():
                packs.append(_read_pack(source_repo, grandchild, child.name, f"{child.name}/{grandchild.name}"))

    if not packs:
        raise DiscoveryError(f"no readable KAS skill packs found under {skills_dir}")
    return packs


def _read_pack(source_repo: Path, pack_dir: Path, category: str, pack_id: str) -> SourcePack:
    skill_md = pack_dir / "SKILL.md"
    try:
        metadata = parse_frontmatter(skill_md.read_text(encoding="utf-8"))
    except OSError as exc:
        raise DiscoveryError(f"cannot read pack metadata: {skill_md}") from exc

    return SourcePack(
        pack_id=pack_id,
        category=category,
        name=metadata.get("name") or pack_dir.name,
        description=metadata.get("description"),
        source_path=pack_dir.relative_to(source_repo).as_posix(),
        checksum=compute_pack_checksum(pack_dir),
    )


def parse_frontmatter(text: str) -> dict[str, str]:
    lines = text.splitlines()
    if not lines or lines[0].strip() != "---":
        return {}

    metadata: dict[str, str] = {}
    for line in lines[1:]:
        if line.strip() == "---":
            break
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        key = key.strip()
        if key not in {"name", "description"}:
            continue
        metadata[key] = _clean_scalar(value.strip())
    return metadata


def compute_pack_checksum(pack_dir: Path) -> str:
    entries = []
    for file_path in sorted(path for path in pack_dir.rglob("*") if path.is_file()):
        relative_path = file_path.relative_to(pack_dir).as_posix()
        if _excluded(relative_path):
            continue
        data = file_path.read_bytes()
        entries.append(
            {
                "path": relative_path,
                "bytes": len(data),
                "mode": f"{stat.S_IMODE(file_path.stat().st_mode):04o}",
                "sha256": sha256(data).hexdigest(),
            }
        )
    payload = json.dumps(entries, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return sha256(payload).hexdigest()


def _clean_scalar(value: str) -> str:
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {"'", '"'}:
        return value[1:-1]
    return value


def _excluded(relative_path: str) -> bool:
    parts = relative_path.split("/")
    if any(part in {".git", ".kkachi", "__pycache__"} for part in parts):
        return True
    return any(part == ".DS_Store" or part.endswith((".swp", ".swo")) for part in parts)
