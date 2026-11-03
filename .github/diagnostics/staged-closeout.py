"""Fork-only gate commands and artifact-directory layout; original verifiers stay unchanged."""

import argparse
import glob
import json
import os
from pathlib import Path
import re
import shutil
import subprocess


inputs = Path(__file__).resolve().parent
manifest = json.loads((inputs / "staged-closeout.json").read_text())


def product_head():
    head = manifest["product_head"]
    if not re.fullmatch(r"[0-9a-f]{40}", head):
        raise SystemExit("a final fixed product SHA is required; this draft cannot execute")
    assert os.environ["GITHUB_REPOSITORY"] == "cpunion/plan9asm"
    assert manifest["main_root_parts"] == 16 and manifest["cli_root_parts"] == 4
    assert manifest["cli_modules"] == ["plan9asm", "plan9asmll"]
    return head


def guard():
    assert subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip() == product_head()
    assert not subprocess.check_output(["git", "status", "--porcelain"], text=True).strip()


def run_gate(name):
    guard()
    expected_go = manifest["minimum_go"] if name == "minimum-fivearch" else manifest["current_go"]
    assert subprocess.check_output(["go", "version"], text=True).strip() == "go version go" + expected_go + " linux/amd64"
    if name == "policy":
        workflow = Path(".github/workflows/go-ci.yml").read_text()
        selector = re.search(r"run: go test \. -run '([^']+)' -count=1", workflow).group(1)
        commands = [
            ["node", "--test"] + sorted(glob.glob(".github/scripts/priority-*.test.cjs")),
            ["go", "test", ".", "-run", selector, "-count=1", "-json"],
            ["go", "test", ".", "-tags=plan9asm_native_avx512", "-run", "^TestCIAVX512NativeQualificationInventory$", "-count=1"],
            ["bash", "scripts/check-go-asm-coverage.sh"],
            ["bash", "scripts/check-arm64-plan9-corpus.sh"],
        ]
        assert not subprocess.check_output(["gofmt", "-l", "."], text=True).strip()
    elif name == "module-compile":
        commands = [["go", "build", "./..."],
                    ["go", "-C", "cmd/plan9asm", "build", "./..."],
                    ["go", "-C", "cmd/plan9asmll", "build", "./..."]]
    else:
        assert name == "minimum-fivearch"
        commands = [["bash", "scripts/check-go-asm-coverage.sh"],
                    ["bash", "scripts/check-arm64-plan9-corpus.sh"]]
    out = Path("_out/staged-closeout-gates")
    out.mkdir(parents=True, exist_ok=False)
    result = {"product_head": product_head(), "workflow_head": (inputs / "workflow-head.txt").read_text().strip(),
              "gate": name, "partial_scope": manifest["scope"], "full325": False,
              "native65": manifest["native65_execution"], "commands": []}
    try:
        for command in commands:
            print("EXEC", json.dumps(command), flush=True)
            execution = subprocess.run(command)
            result["commands"].append({"argv": command, "exit": execution.returncode})
            assert execution.returncode == 0, command
            guard()
    finally:
        (out / "execution.json").write_text(json.dumps(result, indent=2) + "\n")


def assemble(module):
    guard()
    if module == "main":
        source = Path("_out/staged-downloads-main")
        destination = Path("_out/module-coverage-owners")
        expected = {f"staged-main-coverage-root-{part}": f"owner-0000/part-{part:04d}"
                    for part in range(manifest["main_root_parts"])}
        expected.update({"staged-main-coverage-profile": "owner-0001", "staged-main-coverage-other": "owner-0002"})
    else:
        assert module in manifest["cli_modules"]
        source = Path("_out/staged-downloads-cli-" + module)
        destination = Path(f"_out/linux-cli-{module}-coverage-owners")
        expected = {f"staged-cli-{module}-coverage-root-{part}": f"owner-0000/part-{part:04d}"
                    for part in range(manifest["cli_root_parts"])}
        expected.update({f"staged-cli-{module}-coverage-owner-1": "owner-0001",
                         f"staged-cli-{module}-coverage-owner-2": "owner-0002"})
    assert {path.name for path in source.iterdir()} == set(expected), "missing or foreign producer artifact"
    assert not destination.exists()
    for name, relative in expected.items():
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.move(str(source / name), str(target))


parser = argparse.ArgumentParser()
parser.add_argument("action", choices=["prepare", "guard", "gate", "assemble"])
parser.add_argument("value", nargs="?")
args = parser.parse_args()
if args.action == "prepare":
    assert args.value in (manifest["current_go"], manifest["minimum_go"])
    print("product_head=" + product_head())
elif args.action == "guard":
    guard()
elif args.action == "gate":
    run_gate(args.value)
else:
    assemble(args.value)
