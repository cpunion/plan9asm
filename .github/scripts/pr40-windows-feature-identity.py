"""Dated fork verification of actual Windows CLI failures; no production gate."""

from collections import Counter
import hashlib
import json
import os
from pathlib import Path
import subprocess


OUT = Path("_out/windows-feature-identity")
ROOTS = json.loads(Path(".github/scripts/pr40-windows-feature-roots.json").read_text())
PACKAGE = "github.com/xgo-dev/plan9asm/cmd/plan9asmll"


def save(name, value):
    (OUT / name).write_text(json.dumps(value, indent=2) + "\n")


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    summary = {"status": "failed", "expected_roots": ROOTS}
    try:
        head = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
        assert head == os.environ["GITHUB_SHA"]
        assert os.environ["GITHUB_REPOSITORY"] == "cpunion/plan9asm"
        assert not subprocess.check_output(["git", "status", "--porcelain"], text=True)
        summary["head"] = head
        assert len(ROOTS) == len(set(ROOTS)) == 46
        selector = "^(" + "|".join(ROOTS) + ")$"
        listing = subprocess.run(["go", "test", "-C", "cmd/plan9asmll", "-p=1", "-list", selector],
                                 capture_output=True, text=True)
        (OUT / "inventory.txt").write_text(listing.stdout)
        (OUT / "inventory.stderr").write_text(listing.stderr)
        assert listing.returncode == 0
        assert Counter(line for line in listing.stdout.splitlines() if line.startswith("Test")) == Counter(ROOTS)
        command = ["go", "test", "-C", "cmd/plan9asmll", "-p=1", "-run", selector,
                   "-count=1", "-timeout=30m", "-json"]
        save("command.json", command)
        with (OUT / "tests.jsonl").open("wb") as stdout, (OUT / "tests.stderr").open("wb") as stderr:
            result = subprocess.run(command, stdout=stdout, stderr=stderr, timeout=1860)
        events = [json.loads(line) for line in (OUT / "tests.jsonl").read_text().splitlines()]
        assert all(event.get("Package") == PACKAGE for event in events)
        runs = Counter(event["Test"] for event in events if event.get("Test") and event["Action"] == "run")
        ends = [event for event in events if event.get("Test") and event["Action"] in ("pass", "fail", "skip")]
        terminals = Counter(event["Test"] for event in ends)
        actual = {name for name in runs if "/" not in name}
        summary.update(exit_code=result.returncode, roots=len(actual), children=sum("/" in name for name in runs),
                       outcomes=dict(Counter(event["Action"] for event in ends)),
                       bad=[event["Test"] for event in ends if event["Action"] != "pass"])
        assert actual == set(ROOTS)
        assert runs == terminals and all(count == 1 for count in runs.values())
        assert result.returncode == 0 and not summary["bad"]
        assert [event["Action"] for event in events if not event.get("Test") and event["Action"] in ("pass", "fail", "skip")] == ["pass"]
        assert subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip() == head
        assert not subprocess.check_output(["git", "status", "--porcelain"], text=True)
        summary["status"] = "passed"
    except Exception as error:
        summary["error"] = repr(error)
        raise
    finally:
        save("audit.json", summary)
        save("sha256.json", {path.name: hashlib.sha256(path.read_bytes()).hexdigest()
                            for path in OUT.iterdir() if path.is_file() and path.name != "sha256.json"})
        print(json.dumps(summary, indent=2), flush=True)


if __name__ == "__main__":
    main()
