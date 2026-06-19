import json
import os
import shutil
import subprocess
import sys
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
TEMPLATE = ROOT / "templates" / "runners" / "direct-codex-sdk-appserver-runner.py.tmpl"


class DirectCodexRunnerGuardTest(unittest.TestCase):
    def setUp(self) -> None:
        self.work = Path(os.environ.get("TEST_TMPDIR", "/tmp")) / f"kas-runner-guard-{os.getpid()}-{self._testMethodName}"
        if self.work.exists():
            shutil.rmtree(self.work)
        self.work.mkdir(parents=True)
        self.project = self.work / "project"
        self.project.mkdir()
        self.prompt = self.project / "prompt.md"
        self.prompt.write_text("Implement a small guarded change.\n", encoding="utf-8")
        self.runner = self.work / "runner.py"
        shutil.copyfile(TEMPLATE, self.runner)
        self.codex_bin = self.work / "codex"
        self.codex_bin.write_text("#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then echo 'codex 0.test'; exit 0; fi\necho unexpected codex $@ >&2; exit 2\n", encoding="utf-8")
        self.codex_bin.chmod(0o755)
        self.sdk = self.work / "sdk"
        (self.sdk / "openai_codex").mkdir(parents=True)
        marker = self.work / "codex-run-called"
        (self.sdk / "openai_codex" / "__init__.py").write_text(textwrap.dedent(f"""
            class _EnumValue:
                def __init__(self, value): self.value = value
            class ApprovalMode:
                deny_all = _EnumValue('deny_all')
                auto_review = _EnumValue('auto_review')
            class Sandbox:
                read_only = _EnumValue('read_only')
                workspace_write = _EnumValue('workspace_write')
            class CodexConfig:
                def __init__(self, **kwargs): self.kwargs = kwargs
            class _Usage:
                def model_dump(self, mode='json'): return {{}}
            class _Result:
                final_response = 'fake codex output\\n'
                status = _EnumValue('completed')
                id = 'turn-1'
                error = None
                started_at = 1
                completed_at = 2
                duration_ms = 1
                usage = _Usage()
            class _Thread:
                id = 'thread-1'
                def run(self, *args, **kwargs):
                    open({str(marker)!r}, 'w', encoding='utf-8').write('called')
                    return _Result()
            class Codex:
                def __init__(self, cfg): self.cfg = cfg
                def __enter__(self): return self
                def __exit__(self, *args): return False
                def thread_start(self, **kwargs): return _Thread()
                def thread_resume(self, *args, **kwargs): return _Thread()
        """), encoding="utf-8")

    def tearDown(self) -> None:
        shutil.rmtree(self.work, ignore_errors=True)

    def write_packet(self, expected_revision=7) -> Path:
        packet = self.project / "packet.json"
        packet.write_text(json.dumps({
            "strict_order": True,
            "workflow_id": "development_full",
            "run_id": "run-test",
            "instance_id": "run-test",
            "instance_revision": expected_revision,
            "node_id": "implement",
            "status": "ready_for_declared_lane",
            "fallback_policy": "none_fail_closed",
            "completion_authority": "kah_only",
            "direct_kah_state_write": False,
            "expected_start_revision": expected_revision,
        }), encoding="utf-8")
        return packet

    def base_args(self, kah_bin: Path, packet: Path, metadata: Path, output: Path) -> list[str]:
        return [
            sys.executable, str(self.runner),
            "--project-dir", str(self.project),
            "--phase", "implement",
            "--run-id", "run-test",
            "--prompt-path", str(self.prompt),
            "--output-path", str(output),
            "--metadata-path", str(metadata),
            "--codex-bin", str(self.codex_bin),
            "--codex-sdk-src", str(self.sdk),
            "--sandbox", "workspace_write",
            "--no-kab-codex-rationale", "unit test",
            "--dispatch-packet", str(packet),
            "--kah-bin", str(kah_bin),
            "--completion-evidence", str(output.relative_to(self.project)),
        ]

    def test_kah_start_failure_happens_before_codex_thread_run(self) -> None:
        packet = self.write_packet(expected_revision=7)
        kah_log = self.work / "kah.log"
        kah = self.work / "kah-fail-start"
        kah.write_text(f"#!/bin/sh\necho \"$@\" >> {kah_log}\nexit 3\n", encoding="utf-8")
        kah.chmod(0o755)
        metadata = self.work / "metadata.json"
        output = self.project / "cli-output.md"

        proc = subprocess.run(self.base_args(kah, packet, metadata, output), text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        self.assertNotEqual(proc.returncode, 0, proc.stderr)
        self.assertFalse((self.work / "codex-run-called").exists(), "Codex thread.run must not run before KAH start succeeds")
        meta = json.loads(metadata.read_text(encoding="utf-8"))
        self.assertEqual(meta["status"], "kah_start_failed")
        self.assertIn("workflow node start --run run-test --node implement --expect-revision 7 --json", kah_log.read_text(encoding="utf-8"))

    def test_successful_guard_starts_then_runs_then_completes_with_post_start_revision(self) -> None:
        packet = self.write_packet(expected_revision=7)
        kah_log = self.work / "kah.log"
        kah = self.work / "kah-success"
        kah.write_text(textwrap.dedent(f"""#!/bin/sh
            echo "$@" >> {kah_log}
            if [ "$1 $2 $3" = "workflow node start" ]; then
              printf '{{"ok":true,"instance":{{"revision":8}}}}'
              exit 0
            fi
            if [ "$1 $2 $3" = "workflow node complete" ]; then
              printf '{{"ok":true,"instance":{{"revision":9}}}}'
              exit 0
            fi
            exit 4
        """), encoding="utf-8")
        kah.chmod(0o755)
        metadata = self.work / "metadata.json"
        output = self.project / "cli-output.md"

        proc = subprocess.run(self.base_args(kah, packet, metadata, output), text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        self.assertEqual(proc.returncode, 0, proc.stderr)
        self.assertTrue((self.work / "codex-run-called").exists(), "Codex should run after KAH start succeeds")
        calls = kah_log.read_text(encoding="utf-8").splitlines()
        self.assertEqual(calls[0], "workflow node start --run run-test --node implement --expect-revision 7 --json")
        self.assertEqual(calls[1], f"workflow node complete --run run-test --node implement --expect-revision 8 --evidence {output.relative_to(self.project)} --json")
        meta = json.loads(metadata.read_text(encoding="utf-8"))
        self.assertEqual(meta["kah_start_revision"], 8)
        self.assertEqual(meta["kah_complete_revision"], 9)
        self.assertTrue(meta["kah_completion_claimed"])

    def test_kah_complete_failure_preserves_no_completion_claim(self) -> None:
        packet = self.write_packet(expected_revision=7)
        kah_log = self.work / "kah.log"
        kah = self.work / "kah-fail-complete"
        kah.write_text(textwrap.dedent(f"""#!/bin/sh
            echo "$@" >> {kah_log}
            if [ "$1 $2 $3" = "workflow node start" ]; then
              printf '{{"ok":true,"instance":{{"revision":8}}}}'
              exit 0
            fi
            if [ "$1 $2 $3" = "workflow node complete" ]; then
              printf '{{"ok":false,"reason":"stale"}}'
              exit 5
            fi
            exit 4
        """), encoding="utf-8")
        kah.chmod(0o755)
        metadata = self.work / "metadata.json"
        output = self.project / "cli-output.md"

        proc = subprocess.run(self.base_args(kah, packet, metadata, output), text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        self.assertNotEqual(proc.returncode, 0, proc.stderr)
        self.assertTrue((self.work / "codex-run-called").exists(), "Codex may run after KAH start succeeds")
        meta = json.loads(metadata.read_text(encoding="utf-8"))
        self.assertEqual(meta["status"], "kah_complete_failed")
        self.assertFalse(meta["kah_completion_claimed"])
        self.assertEqual(meta["kah_start_revision"], 8)
        self.assertEqual(meta["kah_complete_exit_code"], 5)

    def test_missing_completion_evidence_refuses_kah_complete_claim(self) -> None:
        packet = self.write_packet(expected_revision=7)
        kah_log = self.work / "kah.log"
        kah = self.work / "kah-start-only"
        kah.write_text(textwrap.dedent(f"""#!/bin/sh
            echo "$@" >> {kah_log}
            if [ "$1 $2 $3" = "workflow node start" ]; then
              printf '{{"ok":true,"instance":{{"revision":8}}}}'
              exit 0
            fi
            exit 4
        """), encoding="utf-8")
        kah.chmod(0o755)
        metadata = self.work / "metadata.json"
        output = self.project / "cli-output.md"
        args = self.base_args(kah, packet, metadata, output)
        args = args[:-2]  # remove --completion-evidence and its value

        proc = subprocess.run(args, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        self.assertNotEqual(proc.returncode, 0, proc.stderr)
        meta = json.loads(metadata.read_text(encoding="utf-8"))
        self.assertEqual(meta["status"], "kah_complete_failed")
        self.assertFalse(meta["kah_completion_claimed"])
        self.assertIn("at least one --completion-evidence", meta["failure_excerpt"])
        self.assertEqual(kah_log.read_text(encoding="utf-8").splitlines(), ["workflow node start --run run-test --node implement --expect-revision 7 --json"])


if __name__ == "__main__":
    unittest.main()
