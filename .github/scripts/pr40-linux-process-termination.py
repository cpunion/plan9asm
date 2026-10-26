"""One dated fork probe; no instruction, ledger, or production CI changes."""

from collections import Counter
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import time


BASE = "8c85729c415c3eef26161c3d02ec7faa74343caa"
OUT = Path("_out/linux-process-termination")
EXPECTED = {
    "internal/ownedprocess": {
        "TestLinuxOwnedGroupWaitsForTerminationWithUnreapedZombie": [],
        "TestLinuxProcMembershipRequiresCompleteVisibleView": [
            "none", "hidden", "wrong_filesystem", "missing_mount",
            "missing_process", "wrong_pid", "unknown_state", "absent",
        ],
        "TestLinuxProcStatIdentityAndStates": [
            "state_modifier", "negative_pid", "negative_group", "missing_command",
            "starttime_overflow", "truncated", "unknown_state", "Darwin-only_state",
        ],
        "TestOwnedGroupAbsentFastPathDoesNotInspect": [
            "already_absent", "parent_still_needs_Wait",
        ],
        "TestOwnedGroupOtherSignalOutcomesRemainExact": [],
        "TestOwnedGroupPermissionRequiresExitedMembership": [
            "reaped_absent", "reaped_zombie_descendants", "zombie_leader_before_wait",
            "leader_termination_unknown", "leader_missing_with_zombie_descendant",
            "live_leader", "live_descendant", "stopped_descendant", "unknown_state",
            "wrong_group", "unknown_pid", "duplicate_member",
            "inaccessible_membership", "snapshot_unavailable",
        ],
        "TestOwnedGroupRejectsBroadSignalTargets": [],
        "TestOwnedGroupSuccessfulSignalRequiresTerminalMembership": [
            "terminal_members", "signal_delivery_is_asynchronous",
            "new_descendant_requires_another_complete_observation",
            "reused_pid_requires_another_complete_observation",
            "live_member_exceeds_the_bounded_join", "live_then_inaccessible",
            "unknown_kernel_state", "foreign_group",
        ],
    },
    "cmd/plan9asmtest": {
        "TestModuleOtherWorkersActualWholePackagesAndStableUnion": [
            "failure-false", "failure-true",
        ],
        "TestModuleOtherWorkersActualCancellationStopsDispatchAndJoins": [],
        "TestOwnedCommandKillsDescendantsAfterExitAndTimeout": ["success", "timeout"],
    },
}


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def save(name, value):
    (OUT / name).write_text(json.dumps(value, indent=2, sort_keys=True) + "\n")


def run(stem, argv, timeout=180):
    started = time.monotonic()
    with (OUT / (stem + ".stdout")).open("wb") as output:
        with (OUT / (stem + ".stderr")).open("wb") as error:
            completed = subprocess.run(argv, stdout=output, stderr=error, timeout=timeout)
    save(stem + ".command.json", {
        "argv": argv, "exit_code": completed.returncode,
        "seconds": time.monotonic() - started,
    })
    return completed.returncode, (OUT / (stem + ".stdout")).read_text()


def query(stem, argv):
    status, output = run(stem, argv)
    require(status == 0, f"required observation failed: {stem}")
    return output.strip()


def source():
    head = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
    status = subprocess.check_output(["git", "status", "--porcelain"], text=True)
    require(not status, "source changed during validation: " + status)
    return head


def audit(events, package, expected):
    names = set(expected)
    names.update(root + "/" + child for root, children in expected.items() for child in children)
    runs, passes = Counter(), Counter()
    package_actions = Counter()
    for event in events:
        require(event.get("Package") == package, "foreign package event")
        action, name = event["Action"], event.get("Test")
        require(action not in ("fail", "skip"), "failed or skipped event: " + str(event))
        if name:
            require(name in names, "unrequested test or child: " + name)
            if action == "run":
                runs[name] += 1
            elif action == "pass":
                require(runs[name] == 1 and passes[name] == 0, "terminal without one run: " + name)
                passes[name] += 1
        elif action in ("start", "pass"):
            package_actions[action] += 1
    required = Counter({name: 1 for name in names})
    require(runs == required and passes == required, "missing or duplicate execution/terminal")
    require(package_actions == {"start": 1, "pass": 1}, "incomplete package lifecycle")
    return {"roots": sorted(expected), "children": sorted(names - set(expected)), "status": "passed"}


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    result = {"status": "failed", "packages": {}}
    try:
        head = source()
        require(head == os.environ["GITHUB_SHA"], "checkout does not match workflow SHA")
        require(os.environ["GITHUB_REPOSITORY"] == "cpunion/plan9asm", "not the allowed fork")
        require(os.environ["GITHUB_REF"] == "refs/heads/codex/pr40-linux-process-termination-20261026", "wrong branch")
        require(query("base", ["git", "rev-parse", "HEAD^"]) == BASE, "unexpected production base")
        changed = query("delta", ["git", "diff", "--name-only", "HEAD^", "HEAD"]).splitlines()
        require(set(changed) == {
            ".github/workflows/pr40-linux-process-termination.yml",
            ".github/scripts/pr40-linux-process-termination.py",
        }, "probe changed production or unrelated files")
        save("source.json", {"head": head, "production_base": BASE, "changed_files": changed,
                             "run_id": os.environ["GITHUB_RUN_ID"], "repository": os.environ["GITHUB_REPOSITORY"]})
        require(query("go-version", ["go", "version"]) == "go version go1.27.2 linux/amd64", "wrong Go")
        require(query("node-version", ["node", "--version"]) == "v22.23.4", "wrong Node")
        version = query("llvm-version", ["llvm-config-22", "--version"])
        require(version.startswith("22."), "wrong LLVM")
        bindir = Path(query("llvm-bindir", ["llvm-config-22", "--bindir"]))
        require(bindir.is_absolute(), "relative LLVM bindir")
        for tool in ("llc", "clang", "llvm-as", "opt", "llvm-dis", "llvm-objdump", "llvm-nm", "ld.lld"):
            observed = query(tool + "-version", [str(bindir / tool), "--version"])
            require(re.search(r"(?:version |LLD )22\.", observed), "wrong backend: " + tool)
        query("go-env", ["go", "env", "-json"])
        save("expected.json", EXPECTED)
        errors = []
        for relative, expected in EXPECTED.items():
            stem = relative.replace("/", "-")
            pattern = "^Test" if relative == "internal/ownedprocess" else "^(" + "|".join(expected) + ")$"
            try:
                listing = query(stem + "-inventory", ["go", "test", "-p=1", "./" + relative, "-list", pattern])
                actual = [line for line in listing.splitlines() if line.startswith("Test")]
                require(Counter(actual) == Counter({name: 1 for name in expected}), "fresh inventory differs")
                status, raw = run(stem, ["go", "test", "-p=1", "-json", "./" + relative,
                                         "-run", pattern, "-count=1", "-timeout=8m"], timeout=540)
                require(status == 0, "go test failed: " + relative)
                events = [json.loads(line) for line in raw.splitlines()]
                result["packages"][relative] = audit(events, "github.com/xgo-dev/plan9asm/" + relative, expected)
            except Exception as error:
                errors.append(relative + ": " + str(error))
        require(source() == head, "source head changed")
        require(not errors, "\n".join(errors))
        result["status"] = "passed"
    except Exception as error:
        result["error"] = str(error)
        raise
    finally:
        save("audit.json", result)
        save("sha256.json", {path.name: hashlib.sha256(path.read_bytes()).hexdigest()
                             for path in sorted(OUT.iterdir()) if path.is_file() and path.name != "sha256.json"})
        print(json.dumps(result, indent=2), flush=True)


if __name__ == "__main__":
    main()
