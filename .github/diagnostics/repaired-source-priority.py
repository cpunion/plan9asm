"""Fork-only selected-owner execution; never a complete Go CI acceptance."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--go-version", choices=["1.27.3", "1.20.14"], required=True)
    parser.add_argument("--owner", type=int, choices=range(4), required=True)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    inputs = Path(__file__).resolve().parent
    manifest = json.loads((inputs / "repaired-source-priority.json").read_text())
    head = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
    assert head == manifest["product_head"], "diagnostic code is not the product checkout"
    assert not subprocess.check_output(["git", "status", "--porcelain"], text=True).strip()
    version = subprocess.check_output(["go", "version"], text=True).strip()
    assert version == "go version go" + args.go_version + " linux/amd64", version
    args.out.mkdir(parents=True, exist_ok=False)
    result = {
        "kind": "CURRENT_PRODUCT_SELECTED_OWNER_ONLY",
        "head": head,
        "workflow_head": (inputs / "workflow-head.txt").read_text().strip(),
        "go_version": version,
        "owner": args.owner,
        "owners": 4,
        "manifest_sha256": hashlib.sha256((inputs / "repaired-source-priority.json").read_bytes()).hexdigest(),
        "full_325_CI_acceptance": False,
        "native65_execution": "PENDING_NOT_EXECUTED",
        "packages": [],
    }
    failed = False
    for scope in manifest["packages"]:
        roots = scope["roots"]
        assert roots == sorted(set(roots)) and roots
        assert all(re.fullmatch(r"Test[^/\s]+", root) for root in roots)
        directory = args.out / scope["id"]
        command = [
            "go", "run", "./cmd/plan9asmtest",
            "-dir", scope["directory"], "-package", ".", "-out", str(directory),
            "-ci-selector", "^(" + "|".join(roots) + ")$",
            "-ci-isolate-root", scope["isolated_root"],
            "-ci-owner", str(args.owner), "-shards", "4", "-timeout", "45m",
        ]
        with (args.out / (scope["id"] + ".stdout")).open("wb") as stdout, \
                (args.out / (scope["id"] + ".stderr")).open("wb") as stderr:
            execution = subprocess.run(command, stdout=stdout, stderr=stderr)
        record = {"package": scope["package"], "command": command, "exit": execution.returncode}
        try:
            report = json.loads((directory / "report.json").read_text())
            plan = report["plan"]
            assert plan["scope"]["git_head"] == head, "foreign product source"
            assert plan["package"] == scope["package"], "foreign package"
            assert plan["selected_inventory"] == roots, "missing or foreign selected owner"
            assert plan["owners"] == 4 and report["owner"] == args.owner
            assert execution.returncode == 0 and report["status"] == "passed"
            record.update(status="passed", roots_executed=len(report["owned_root_receipts"]))
        except (AssertionError, KeyError, OSError, ValueError) as error:
            failed = True
            record.update(status="failed", error=str(error))
        result["packages"].append(record)
        (args.out / "partial-owner.json").write_text(json.dumps(result, indent=2) + "\n")
        print(scope["package"], record["status"], flush=True)
    assert not subprocess.check_output(["git", "status", "--porcelain"], text=True).strip()
    raise SystemExit(1 if failed else 0)


if __name__ == "__main__":
    main()
