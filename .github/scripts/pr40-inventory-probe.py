"""Fork-only diagnostics; not a replacement for the complete contribution gates."""

import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys


def main():
    role, ordinal = sys.argv[1], int(sys.argv[2])
    root = Path.cwd()
    directory = root / "cmd/plan9asmll" if role == "cli" else root
    package = "." if role == "cli" else "./gotoolprofile"
    artifact = root / "_out" / f"windows-{role}-{ordinal}"
    artifact.mkdir(parents=True, exist_ok=False)
    listed = subprocess.run(
        ["go", "test", "-list=.", package],
        cwd=directory, text=True, capture_output=True, timeout=300,
    )
    (artifact / "inventory.stdout").write_text(listed.stdout, encoding="utf-8")
    (artifact / "inventory.stderr").write_text(listed.stderr, encoding="utf-8")
    if listed.returncode:
        raise RuntimeError(listed.stdout + listed.stderr)
    names = sorted(
        line for line in listed.stdout.splitlines()
        if re.fullmatch(r"(?:Test|Example|Fuzz)[A-Za-z0-9_]*", line)
    )
    if not names or len(names) != len(set(names)) or ordinal not in range(4):
        raise RuntimeError("missing or duplicate actual Go inventory / invalid owner")
    assignments = [
        [name for name in names if hashlib.sha256(name.encode()).digest()[0] % 4 == index]
        for index in range(4)
    ]
    if sorted(name for shard in assignments for name in shard) != names:
        raise RuntimeError("partition changed the actual inventory")
    planned = assignments[ordinal]
    if not planned:
        raise RuntimeError("unexpected empty diagnostic owner")
    (artifact / "plan.json").write_text(
        json.dumps({"inventory": names, "assignments": assignments, "owner": ordinal}, indent=2),
        encoding="utf-8",
    )
    print(f"Actual inventory {len(names)}; owner {ordinal}/4 runs {len(planned)} roots", flush=True)
    command = [
        "go", "test", "-count=1", "-timeout=30m", "-json",
        "-run=^(?:" + "|".join(re.escape(name) for name in planned) + ")$", package,
    ]
    outcomes = {}
    invalid = []
    # The GitHub-hosted job retains its existing process cleanup boundary.
    # This probe does not enable the Unix-only owned-process suite runner.
    with (artifact / "events.jsonl").open("w", encoding="utf-8") as output:
        process = subprocess.Popen(
            command, cwd=directory, stdout=subprocess.PIPE, text=True,
        )
        for line in process.stdout:
            output.write(line)
            output.flush()
            event = json.loads(line)
            name, action = event.get("Test", ""), event.get("Action")
            if action in ("fail", "skip") or action == "pass" and "/" not in name:
                print(json.dumps(event), flush=True)
            if name and "/" not in name and action in ("pass", "fail", "skip"):
                if name in outcomes or name not in planned:
                    invalid.append(name)
                outcomes[name] = action
        status = process.wait()
    summary = {"status": status, "outcomes": outcomes, "invalid": invalid,
               "missing": sorted(set(planned) - set(outcomes))}
    (artifact / "summary.json").write_text(json.dumps(summary, indent=2), encoding="utf-8")
    if status or invalid or summary["missing"] or "fail" in outcomes.values():
        raise RuntimeError(json.dumps(summary))
    print(f"Owner {ordinal}: {sum(value == 'pass' for value in outcomes.values())} pass; "
          f"{sum(value == 'skip' for value in outcomes.values())} skipped roots (see raw reasons)", flush=True)


if __name__ == "__main__":
    main()
