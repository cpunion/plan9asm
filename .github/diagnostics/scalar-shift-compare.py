"""Fork-only measurement; never a native execution or full-CI certificate."""

import collections
import hashlib
import json
import math
import os
from pathlib import Path
import re
import shutil
import shlex
import statistics
import subprocess


root = Path.cwd()
out = root / "_out/scalar-shift-cloud"
out.mkdir(parents=True, exist_ok=False)
baseline = Path(os.environ["RUNNER_TEMP"]) / "scalar-shift-baseline"
baseline.mkdir(exist_ok=False)
env = dict(os.environ)
(out / "bootstrap.json").write_text(json.dumps({
    "product": env["PRODUCT_REF"], "baseline": env["BASELINE_REF"],
    "expected_Go": env["EXPECTED_GO"],
    "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
}, indent=2) + "\n")


def command(args, cwd=root):
    result = subprocess.run(
        args, cwd=cwd, env=env, text=True, capture_output=True, timeout=120,
    )
    with (out / "setup-commands.jsonl").open("a") as stream:
        stream.write(json.dumps({
            "args": args, "exit": result.returncode,
            "stdout": result.stdout, "stderr": result.stderr,
        }) + "\n")
    result.check_returncode()
    return result.stdout.strip()


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def sources(directory):
    names = subprocess.check_output(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"],
        cwd=directory,
    ).split(b"\0")
    return {os.fsdecode(n): sha(directory / os.fsdecode(n)) for n in names if n}


def run(args, cwd, stem):
    with (out / (stem + ".stdout")).open("wb") as stdout:
        with (out / (stem + ".stderr")).open("wb") as stderr:
            result = subprocess.run(
                args, cwd=cwd, env=env, stdout=stdout, stderr=stderr, timeout=300,
            )
    assert result.returncode == 0, (stem, result.returncode)


assert command(["go", "env", "GOVERSION"]) == "go" + env["EXPECTED_GO"]
assert command(["go", "env", "GOOS"]) == "linux"
assert command(["go", "env", "GOARCH"]) == "amd64"
assert command([env["LLVM_CONFIG"], "--version"]).split(".")[0] == "22"

product = env["PRODUCT_REF"]
before = env["BASELINE_REF"]
command(["git", "fetch", "--depth=1", "origin", product, before])
extra = {
    ".github/workflows/scalar-shift-translate-diagnostic.yml",
    ".github/diagnostics/scalar-shift-compare.py",
}
actual_delta = set(command(["git", "diff", "--name-only", product, "HEAD"]).splitlines())
assert actual_delta == extra, actual_delta
command(["git", "init", str(baseline)])
command(["git", "-C", str(baseline), "fetch", "--depth=1", str(root), before])
command(["git", "-C", str(baseline), "checkout", "--detach", "FETCH_HEAD"])
for name in (
    "amd64_scalar_shift_pipeline_bench_test.go",
    "amd64_scalar_shift_form_reuse_test.go",
):
    shutil.copyfile(root / name, baseline / name)

frozen = {"A": sources(baseline), "B": sources(root)}
all_go = set(n for n in frozen["A"] | frozen["B"] if n.endswith(".go"))
go_delta = {n for n in all_go if frozen["A"].get(n) != frozen["B"].get(n)}
assert go_delta == {"amd64_lower_shiftrotate.go"}, go_delta
tools = {}
for name in ("go", "clang-22", "llvm-config-22", "llc-22"):
    path = Path(shutil.which(name)).resolve()
    tools[str(path)] = sha(path)
goroot = command(["go", "env", "GOROOT"])
for name in ("compile", "asm", "link", "cgo"):
    path = Path(command(["go", "tool", "-n", name]))
    tools[str(path)] = sha(path)
for library in shlex.split(command([env["LLVM_CONFIG"], "--link-shared", "--libfiles"])):
    path = Path(library).resolve()
    tools[str(path)] = sha(path)
metadata = {
    "product": product, "baseline": before,
    "diagnostic_head": command(["git", "rev-parse", "HEAD"]),
    "sources": frozen, "tools": tools, "GOROOT": goroot,
    "Go": command(["go", "version"]),
    "LLVM": command([env["LLVM_CONFIG"], "--version"]),
    "load": list(os.getloadavg()), "cpu": command(["lscpu"]),
    "environment": {n: env.get(n) for n in (
        "GOFLAGS", "GOENV", "GOWORK", "GOTOOLCHAIN", "GOMAXPROCS",
        "CGO_ENABLED", "CC", "CXX", "CGO_CPPFLAGS", "CGO_LDFLAGS",
    )},
}
(out / "source-tools.json").write_text(json.dumps(metadata, indent=2) + "\n")
binaries = {tag: out / (tag + ".test") for tag in ("A", "B")}
for tag, directory in (("A", baseline), ("B", root)):
    run(["go", "test", "-p=1", "-tags=byollvm", "-buildvcs=false", "-c",
         "-o", str(binaries[tag]), "."], directory, "build-" + tag)
binary_sha = {tag: sha(path) for tag, path in binaries.items()}

selector = "^(BenchmarkX86ScalarShiftTranslation|BenchmarkX86BaselineGPTranslation)$"
expected = {"BenchmarkX86BaselineGPTranslation"}
for row in (
    "SHLQ $7,AX", "SARL CX,AX", "RCLB $10,AL", "RCRW CX,AX",
    "SHLW $16,BX,AX", "SHRL CX,BX,AX", "RCRQ $0,AX", "RORXL $7,BX,AX",
):
    expected.add("BenchmarkX86ScalarShiftTranslation/" + row.replace(" ", "_"))
pattern = re.compile(
    r"^(Benchmark\S+?)(?:-\d+)?\s+(\d+)\s+([\d.]+)\s+ns/op\s+"
    r"(?:[\d.]+\s+\S+/s\s+)?(\d+)\s+B/op\s+(\d+)\s+allocs/op$", re.M,
)
samples = collections.defaultdict(lambda: collections.defaultdict(list))
for block in range(3):
    for order, tag in enumerate("ABBA"):
        stem = f"{block}-{order}-{tag}"
        run([str(binaries[tag]), "-test.run=^$", "-test.bench=" + selector,
             "-test.benchtime=500ms", "-test.count=1", "-test.cpu=1"], root, stem)
        output = (out / (stem + ".stdout")).read_text()
        assert len(re.findall(r"(?m)^PASS\s*$", output)) == 1, stem
        assert not re.search(r"(?m)^(?:---\s+)?(?:FAIL|SKIP)\b", output), stem
        records = pattern.findall(output)
        assert len(records) == 9 and {r[0] for r in records} == expected
        for name, iterations, ns, allocated, allocations in records:
            assert int(iterations) > 0 and float(ns) > 0, (stem, name)
            samples[name][tag].append({
                "ns": float(ns), "bytes": int(allocated), "allocs": int(allocations),
                "iterations": int(iterations), "stem": stem,
            })
        assert sources(root) == frozen["B"] and sources(baseline) == frozen["A"]
        assert all(sha(Path(p)) == digest for p, digest in tools.items())
        assert all(sha(path) == binary_sha[t] for t, path in binaries.items())
        print("complete sample", stem, "all nine cases", flush=True)

rows = {}
for name, values in samples.items():
    assert len(values["A"]) == len(values["B"]) == 6
    medians = {tag: {k: statistics.median(v[k] for v in values[tag])
                     for k in ("ns", "bytes", "allocs")} for tag in ("A", "B")}
    rows[name] = {
        "medians": medians, "time_ratio": medians["B"]["ns"] / medians["A"]["ns"],
        "allocation_delta": medians["B"]["allocs"] - medians["A"]["allocs"],
        "samples": dict(values),
    }
group = [r["time_ratio"] for n, r in rows.items()
         if n.startswith("BenchmarkX86ScalarShiftTranslation/") and "/RORXL" not in n]
receipt = {
    "kind": "CONTROLLED_JOB_PAIRED_TRANSLATION_MEASUREMENT_NOT_NATIVE_EXECUTION",
    "source_and_tools_unchanged": True, "cases": rows, "binary_sha256": binary_sha,
    "ABBA_blocks": 3, "samples_per_case_per_revision": 6,
    "Group2_geomean_time_ratio": math.exp(sum(map(math.log, group)) / len(group)),
    "interpretation": "Timing requires review with controls; this command never certifies native instructions or complete CI.",
}
(out / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
print(json.dumps({k: v for k, v in receipt.items() if k != "cases"}), flush=True)
for path in binaries.values():
    path.unlink()
