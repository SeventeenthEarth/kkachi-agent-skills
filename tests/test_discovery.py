import json
import os
import stat
import sys
import tempfile
import unittest
from hashlib import sha256
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "src"))


def write_skill(path: Path, *, title: str | None = None, description: str | None = None) -> None:
    path.mkdir(parents=True)
    lines = ["---"]
    if title is not None:
        lines.append(f"name: {title}")
    if description is not None:
        lines.append(f"description: {description}")
    lines.extend(["---", "# Skill", "", "Body"])
    (path / "SKILL.md").write_text("\n".join(lines), encoding="utf-8")


def expected_pack_checksum(pack_dir: Path) -> str:
    entries = []
    for file_path in sorted(path for path in pack_dir.rglob("*") if path.is_file()):
        relative_path = file_path.relative_to(pack_dir).as_posix()
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


class DiscoveryTests(unittest.TestCase):
    def test_discovers_direct_skill_layout_as_core_pack(self) -> None:
        from kkachi_hermes_skills.discovery import discover_source_packs

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            write_skill(repo / "skills" / "alpha", title="Alpha Pack", description="Alpha description")

            packs = [pack.as_dict() for pack in discover_source_packs(repo)]

        self.assertEqual(len(packs), 1)
        self.assertEqual(packs[0]["pack_id"], "alpha")
        self.assertEqual(packs[0]["category"], "core")
        self.assertEqual(packs[0]["name"], "Alpha Pack")
        self.assertEqual(packs[0]["description"], "Alpha description")
        self.assertEqual(packs[0]["source_path"], "skills/alpha")

    def test_discovers_category_skill_layout(self) -> None:
        from kkachi_hermes_skills.discovery import discover_source_packs

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            write_skill(repo / "skills" / "software-development" / "roadmap", description="Roadmap work")

            packs = [pack.as_dict() for pack in discover_source_packs(repo)]

        self.assertEqual(len(packs), 1)
        self.assertEqual(packs[0]["pack_id"], "software-development/roadmap")
        self.assertEqual(packs[0]["category"], "software-development")
        self.assertEqual(packs[0]["name"], "roadmap")
        self.assertEqual(packs[0]["description"], "Roadmap work")
        self.assertEqual(packs[0]["source_path"], "skills/software-development/roadmap")

    def test_fails_closed_when_no_readable_pack_metadata_exists(self) -> None:
        from kkachi_hermes_skills.discovery import DiscoveryError, discover_source_packs

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            (repo / "skills" / "empty").mkdir(parents=True)

            with self.assertRaises(DiscoveryError):
                discover_source_packs(repo)


class ListResultTests(unittest.TestCase):
    def test_category_filter_and_unknown_category_diagnostic(self) -> None:
        from kkachi_hermes_skills.listing import build_list_result

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            write_skill(repo / "skills" / "alpha")
            write_skill(repo / "skills" / "docs" / "beta")

            docs_result = build_list_result(repo, category="docs")
            unknown_result = build_list_result(repo, category="missing")

        self.assertTrue(docs_result["ok"])
        self.assertEqual([pack["pack_id"] for pack in docs_result["packs"]], ["docs/beta"])
        self.assertTrue(unknown_result["ok"])
        self.assertEqual(unknown_result["packs"], [])
        self.assertEqual(unknown_result["diagnostics"][0]["code"], "unknown_category")

    def test_profile_manifest_missing_and_present_states(self) -> None:
        from kkachi_hermes_skills.listing import build_list_result

        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            repo = base / "repo"
            write_skill(repo / "skills" / "alpha")
            write_skill(repo / "skills" / "beta")

            missing = build_list_result(repo, profile="demo", profile_root=base / "profiles" / "demo")

            profile_root = base / "profile-present"
            manifest_path = profile_root / ".kas" / "skill-pack-manifest.json"
            manifest_path.parent.mkdir(parents=True)
            manifest = {
                "version": "0.1",
                "kind": "kas_profile_skill_manifest",
                "installs": [
                    {
                        "pack_id": "alpha",
                        "target_path": "skills/alpha",
                        "pack_checksum": expected_pack_checksum(repo / "skills" / "alpha"),
                    },
                    {
                        "pack_id": "beta",
                        "target_path": "skills/beta",
                        "pack_checksum": "not-current",
                    },
                ],
            }
            manifest_path.write_text(json.dumps(manifest), encoding="utf-8")

            present = build_list_result(repo, profile="demo", profile_root=profile_root)

        self.assertEqual(missing["target_profile"]["manifest_state"], "manifest_missing")
        self.assertEqual({pack["installed_state"] for pack in missing["packs"]}, {"manifest_missing"})
        states = {pack["pack_id"]: pack["installed_state"] for pack in present["packs"]}
        self.assertEqual(present["target_profile"]["manifest_state"], "manifest_present")
        self.assertEqual(states["alpha"], "installed_current")
        self.assertEqual(states["beta"], "installed_drifted")

    def test_profile_manifest_rejects_non_normalized_target_paths(self) -> None:
        from kkachi_hermes_skills.listing import build_list_result

        with tempfile.TemporaryDirectory() as tmp:
            base = Path(tmp)
            repo = base / "repo"
            write_skill(repo / "skills" / "alpha")
            write_skill(repo / "skills" / "beta")
            profile_root = base / "profile"
            manifest_path = profile_root / ".kas" / "skill-pack-manifest.json"
            manifest_path.parent.mkdir(parents=True)
            manifest = {
                "version": "0.1",
                "kind": "kas_profile_skill_manifest",
                "installs": [
                    {"pack_id": "alpha", "target_path": "", "pack_checksum": "checksum"},
                    {"pack_id": "beta", "target_path": "skills//beta", "pack_checksum": "checksum"},
                ],
            }
            manifest_path.write_text(json.dumps(manifest), encoding="utf-8")

            result = build_list_result(repo, profile="demo", profile_root=profile_root)

        states = {pack["pack_id"]: pack["installed_state"] for pack in result["packs"]}
        self.assertEqual(states["alpha"], "conflict")
        self.assertEqual(states["beta"], "conflict")


if __name__ == "__main__":
    unittest.main()
