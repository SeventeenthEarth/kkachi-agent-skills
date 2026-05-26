import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def env_with_src_and_harness_guard() -> dict[str, str]:
    env = os.environ.copy()
    env["PYTHONPATH"] = str(ROOT / "src")
    env["KAS_ALLOW_PROFILE_ROOT_OVERRIDE"] = "1"
    return env


def current_direct_skill_count() -> int:
    return sum(1 for path in (ROOT / "skills").glob("*/SKILL.md") if path.is_file())


class RealRepoE2ETests(unittest.TestCase):
    def test_real_repo_list_counts_current_skills_and_does_not_write_profile(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            profile_root = Path(tmp) / "profiles" / "e2e"

            proc = subprocess.run(
                [
                    str(ROOT / "bin" / "kkachi-hermes-skills"),
                    "list",
                    "--repo",
                    str(ROOT),
                    "--profile",
                    "e2e",
                    "--profile-root",
                    str(profile_root),
                    "--json",
                ],
                text=True,
                capture_output=True,
                env=env_with_src_and_harness_guard(),
                check=False,
            )

            self.assertFalse(profile_root.exists())

        self.assertEqual(proc.returncode, 0, proc.stderr)
        payload = json.loads(proc.stdout)
        self.assertTrue(payload["ok"])
        self.assertEqual(len(payload["packs"]), current_direct_skill_count())
        self.assertEqual(payload["target_profile"]["manifest_state"], "manifest_missing")
        self.assertEqual({pack["installed_state"] for pack in payload["packs"]}, {"manifest_missing"})


if __name__ == "__main__":
    unittest.main()
