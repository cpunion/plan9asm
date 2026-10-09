"""First-party counterexamples for the decoder script's actual probe loop.

This frozen RED adapter executes the original loop AST, rather than copying its
classification rule. Controlled tools are infrastructure counterexamples only;
they are never certificates of Go source rejection. Two independent positive
controls invoke the explicitly selected stock Go assembler.
"""

import ast
import contextlib
import hashlib
import json
import os
import pathlib
import signal
import subprocess
import sys
import tempfile
import time
import types
import unittest


REPOSITORY = pathlib.Path(__file__).resolve().parent.parent
SCRIPT = REPOSITORY / "scripts" / "check-arm64-plan9-corpus.sh"


def original_probe_loop():
    data = SCRIPT.read_text(encoding="utf-8")
    blocks = data.split("<<'PY'\n")[1:]
    matches = []
    for block in blocks:
        program = ast.parse(block.split("\nPY\n", 1)[0])
        for node in program.body:
            if isinstance(node, ast.For) and ast.unparse(node.iter) == "enumerate(selected)":
                matches.append(node)
    if len(matches) != 1:
        raise AssertionError("one complete original per-case Go probe loop required")
    module = ast.Module(body=matches, type_ignores=[])
    return compile(ast.fix_missing_locations(module), str(SCRIPT), "exec")


class ARM64Plan9OracleTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.loop = original_probe_loop()
        cls.go = os.environ.get("PLAN9ASM_ORACLE_GO")
        cls.goroot = os.environ.get("PLAN9ASM_ORACLE_GOROOT")
        if not cls.go or not cls.goroot:
            raise AssertionError("explicit stock Go1.27.1 driver and SDK required")
        result = subprocess.run(
            [cls.go, "version"], capture_output=True, timeout=5, check=True
        )
        if not result.stdout.startswith(b"go version go1.27.1 "):
            raise AssertionError("the selected oracle must be actual Go1.27.1")
        cls.records = []

    @classmethod
    def tearDownClass(cls):
        directory = os.environ.get("PLAN9ASM_ORACLE_EVIDENCE_DIR")
        if directory:
            output = pathlib.Path(directory) / "first-party-commands.json"
            with output.open("x", encoding="utf-8") as stream:
                json.dump(cls.records, stream, indent=2, ensure_ascii=False)
                stream.write("\n")

    def probe(self, program=None, instruction="MOVD R0,R1"):
        with tempfile.TemporaryDirectory(prefix="arm64-decoder-oracle-") as temporary:
            root = pathlib.Path(temporary)
            source, obj = root / "probe.s", root / "probe.o"
            accepted, rejected, requested = [], [], []

            def run(argv, **kwargs):
                requested.append(kwargs)
                command = [self.go, *argv[1:]]
                if program is not None:
                    command = [sys.executable, "-c", program, str(source), str(obj)]
                environment = kwargs["env"].copy()
                environment["GOTOOLCHAIN"] = "local"
                environment["GOENV"] = "off"
                start = time.monotonic()
                child = subprocess.Popen(
                    command, env=environment, stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE, start_new_session=True,
                )
                pgid = os.getpgid(child.pid)
                try:
                    stdout, stderr = child.communicate(timeout=5)
                except subprocess.TimeoutExpired:
                    os.killpg(pgid, signal.SIGKILL)
                    child.communicate(timeout=2)
                    raise
                finally:
                    with contextlib.suppress(ProcessLookupError):
                        os.killpg(pgid, signal.SIGKILL)
                self.records.append({
                    "test": self.id(), "kind": "actual-stock-Go" if program is None else "controlled-infrastructure-tool",
                    "argv": command, "pid": child.pid, "pgid": pgid,
                    "exit": child.returncode, "duration_seconds": time.monotonic() - start,
                    "source_sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                    "source": source.read_text(encoding="utf-8"),
                    "stdout": stdout.decode("utf-8"), "stderr": stderr.decode("utf-8"),
                })
                return subprocess.CompletedProcess(command, child.returncode, stdout, stderr)

            namespace = {
                "selected": ["00000000 | " + instruction], "tmp_root": root,
                "source": source, "obj": obj, "goroot": pathlib.Path(self.goroot),
                "env": os.environ.copy(), "accepted": accepted, "rejected": rejected,
                "subprocess": types.SimpleNamespace(run=run, DEVNULL=subprocess.DEVNULL),
            }
            namespace["env"].update({"GOOS": "linux", "GOARCH": "arm64"})
            exec(self.loop, namespace)
            object_info = None
            if obj.exists():
                info = obj.lstat()
                object_info = {"regular": obj.is_file(), "size": info.st_size}
            return accepted, rejected, requested, object_info

    def assert_infrastructure_failure(self, program):
        accepted, rejected, _, _ = self.probe(program)
        self.assertEqual((accepted, rejected), ([], []),
                         "infrastructure was classified as accepted/source-rejected")

    def test_actual_go_acceptance(self):
        accepted, rejected, _, obj = self.probe()
        self.assertEqual((len(accepted), len(rejected)), (1, 0))
        self.assertTrue(obj and obj["regular"] and obj["size"] > 0)

    def test_actual_go_source_rejection(self):
        accepted, rejected, _, _ = self.probe(instruction="MOVD (R0),RSP")
        self.assertEqual((len(accepted), len(rejected)), (0, 1))
        actual = self.records[-1]
        self.assertEqual(actual["exit"], 1)
        self.assertIn("probe.s:3", actual["stderr"])
        self.assertNotIn("panic:", actual["stderr"])

    def test_unknown_exit_one(self):
        self.assert_infrastructure_failure("raise SystemExit(1)")

    def test_unknown_exit_two(self):
        self.assert_infrastructure_failure("raise SystemExit(2)")

    def test_signal(self):
        self.assert_infrastructure_failure("import os,signal; os.kill(os.getpid(),signal.SIGTERM)")

    def test_footer_only(self):
        self.assert_infrastructure_failure("import sys; sys.stderr.write('asm: assembly failed\\n'); raise SystemExit(1)")

    def test_foreign_source(self):
        self.assert_infrastructure_failure("import sys; sys.stderr.write('foreign.s:3: bad operand\\n'); raise SystemExit(1)")

    def test_panic_after_fixture_diagnostic(self):
        self.assert_infrastructure_failure("import sys; sys.stderr.write(sys.argv[1]+':3: bad operand\\npanic: backend failure\\n'); raise SystemExit(1)")

    def test_wrong_instruction_listing(self):
        self.assert_infrastructure_failure("import sys; sys.stderr.write('asm: bad constant for MOVK\\n00000 ('+sys.argv[1]+':3) MOVK $65537,R0\\n'); raise SystemExit(1)")

    def test_success_missing_object(self):
        self.assert_infrastructure_failure("raise SystemExit(0)")

    def test_success_empty_object(self):
        self.assert_infrastructure_failure("import pathlib,sys; pathlib.Path(sys.argv[2]).touch()")

    def test_success_nonregular_object(self):
        self.assert_infrastructure_failure("import pathlib,sys; pathlib.Path(sys.argv[2]).mkdir()")

    def test_per_probe_deadline(self):
        _, _, calls, _ = self.probe()
        self.assertTrue(0 < calls[0].get("timeout", 0) <= 60,
                        "original Go probe has no bounded deadline")

    def test_result_output_is_retained(self):
        _, _, calls, _ = self.probe()
        self.assertNotEqual(calls[0].get("stdout"), subprocess.DEVNULL)
        self.assertNotEqual(calls[0].get("stderr"), subprocess.DEVNULL)

    def test_owned_group(self):
        _, _, calls, _ = self.probe()
        self.assertTrue(calls[0].get("start_new_session", False),
                        "original probe does not own its fresh process group")


if __name__ == "__main__":
    unittest.main(verbosity=2)
