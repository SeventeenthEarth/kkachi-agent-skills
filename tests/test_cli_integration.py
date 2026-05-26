import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def env_with_src() -> dict[str, str]:
    env = os.environ.copy()
    env["PYTHONPATH"] = str(ROOT / "src")
    return env


def write_skill(path: Path, *, name: str = "Sample") -> None:
    path.mkdir(parents=True)
    (path / "SKILL.md").write_text(f"---\nname: {name}\n---\n# {name}\n", encoding="utf-8")


class CliIntegrationTests(unittest.TestCase):
    def test_python_module_list_json_shape(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            write_skill(repo / "skills" / "alpha", name="Alpha")

            proc = subprocess.run(
                [
                    sys.executable,
                    "-m",
                    "kkachi_hermes_skills.cli",
                    "list",
                    "--repo",
                    str(repo),
                    "--json",
                ],
                text=True,
                capture_output=True,
                env=env_with_src(),
                check=False,
            )

        self.assertEqual(proc.returncode, 0, proc.stderr)
        payload = json.loads(proc.stdout)
        self.assertTrue(payload["ok"])
        self.assertEqual(payload["command"], "list")
        self.assertEqual(payload["packs"][0]["pack_id"], "alpha")
        self.assertEqual(payload["diagnostics"], [])
        self.assertIn("next_action", payload)

    def test_console_wrapper_list_json(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            write_skill(repo / "skills" / "alpha", name="Alpha")

            proc = subprocess.run(
                [
                    str(ROOT / "bin" / "kkachi-hermes-skills"),
                    "list",
                    "--repo",
                    str(repo),
                    "--json",
                ],
                text=True,
                capture_output=True,
                env=env_with_src(),
                check=False,
            )

        self.assertEqual(proc.returncode, 0, proc.stderr)
        payload = json.loads(proc.stdout)
        self.assertTrue(payload["ok"])
        self.assertEqual(payload["packs"][0]["name"], "Alpha")

    def test_profile_root_override_requires_harness_guard(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp) / "repo"
            profile_root = Path(tmp) / "profile"
            write_skill(repo / "skills" / "alpha", name="Alpha")

            proc = subprocess.run(
                [
                    str(ROOT / "bin" / "kkachi-hermes-skills"),
                    "list",
                    "--repo",
                    str(repo),
                    "--profile",
                    "demo",
                    "--profile-root",
                    str(profile_root),
                    "--json",
                ],
                text=True,
                capture_output=True,
                env=env_with_src(),
                check=False,
            )

        self.assertEqual(proc.returncode, 2)
        payload = json.loads(proc.stderr)
        self.assertFalse(payload["ok"])
        self.assertEqual(payload["diagnostics"][0]["code"], "profile_root_override_rejected")


if __name__ == "__main__":
    unittest.main()
