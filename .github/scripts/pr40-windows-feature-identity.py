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


def run_group(roots, directory, stem, package, minutes, summary):
    selector = "^(" + "|".join(roots) + ")$"
    prefix = ["go", "test", "-C", directory, "-p=1"]
    listing = subprocess.run(prefix + ["-list", selector], capture_output=True, text=True)
    (OUT / (stem + "-inventory.txt")).write_text(listing.stdout)
    (OUT / (stem + "-inventory.stderr")).write_text(listing.stderr)
    assert listing.returncode == 0
    assert Counter(line for line in listing.stdout.splitlines() if line.startswith("Test")) == Counter(roots)
    command = prefix + ["-run", selector, "-count=1", f"-timeout={minutes}m", "-json"]
    save(stem + "-command.json", command)
    with (OUT / (stem + "-tests.jsonl")).open("wb") as stdout, (OUT / (stem + "-tests.stderr")).open("wb") as stderr:
        result = subprocess.run(command, stdout=stdout, stderr=stderr, timeout=minutes * 60 + 60)
    events = [json.loads(line) for line in (OUT / (stem + "-tests.jsonl")).read_text().splitlines()]
    assert all(event.get("Package") == package for event in events)
    runs = Counter(event["Test"] for event in events if event.get("Test") and event["Action"] == "run")
    ends = [event for event in events if event.get("Test") and event["Action"] in ("pass", "fail", "skip")]
    terminals = Counter(event["Test"] for event in ends)
    actual = {name for name in runs if "/" not in name}
    group = dict(exit_code=result.returncode, roots=len(actual), children=sum("/" in name for name in runs),
                 outcomes=dict(Counter(event["Action"] for event in ends)),
                 bad=[event["Test"] for event in ends if event["Action"] != "pass"])
    summary[stem] = group
    assert actual == set(roots)
    assert runs == terminals and all(count == 1 for count in runs.values())
    assert result.returncode == 0 and not group["bad"]
    assert [event["Action"] for event in events if not event.get("Test") and event["Action"] in ("pass", "fail", "skip")] == ["pass"]


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
        run_group(ROOTS, "cmd/plan9asmll", "cli", PACKAGE, 30, summary)
        run_group(["TestGoAMD64FrameSourceRetainsCommandSDKIdentity"], ".", "frame",
                  "github.com/xgo-dev/plan9asm", 5, summary)
        assert summary["frame"]["children"] == 2
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
