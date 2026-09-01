#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
cd "$repo_root"

if [[ -n "${GOROOT:-}" ]]; then
  go_root=$GOROOT
else
  go_root=$(go env GOROOT)
fi

if [[ "${RUNNER_OS:-}" == "Windows" ]] && command -v python >/dev/null 2>&1; then
  python_cmd=python
elif command -v python3 >/dev/null 2>&1; then
  python_cmd=python3
elif command -v python >/dev/null 2>&1; then
  python_cmd=python
else
  echo "Python 3 not found in PATH" >&2
  exit 1
fi

tmp_root=$(mktemp -d)
trap 'rm -rf "$tmp_root"' EXIT

for goarch in 386 amd64 arm arm64; do
  echo "==> official Go assembler coverage $goarch"
  go run ./cmd/plan9asmscan \
    -corpus=go-asm \
    -goroot="$go_root" \
    -goos=linux \
    -goarch="$goarch" \
    -repo-root=. \
    -format=json \
    -out="$tmp_root/$goarch.json"
done

"$python_cmd" - testdata/coverage/go-asm-baseline.json "$tmp_root" <<'PY'
import json
import pathlib
import re
import sys

baseline_path = pathlib.Path(sys.argv[1])
report_dir = pathlib.Path(sys.argv[2])
baseline = json.loads(baseline_path.read_text(encoding="utf-8"))

fields = (
    "official_opcodes",
    "unique_ops",
    "unique_forms",
    "supported_forms",
    "context_forms",
    "unsupported_forms",
    "parse_err_count",
    "runtime_verified_forms",
    "coverage_fingerprint",
)

for report_path in sorted(report_dir.glob("*.json")):
    report = json.loads(report_path.read_text(encoding="utf-8"))
    match = re.match(r"^(go\d+\.\d+)", report["go_version"])
    if not match:
        raise SystemExit(f"cannot normalize Go version {report['go_version']!r}")
    version = match.group(1)
    arch = report["goarch"]
    expected = baseline.get("versions", {}).get(version, {}).get(arch)
    if expected is None:
        raise SystemExit(
            f"missing coverage baseline for {version}/{arch}; "
            "inspect the full JSON report and update the baseline intentionally"
        )
    changed = [
        field for field in fields
        if report.get(field) != expected.get(field)
    ]
    print(
        f"{version}/{arch}: official={report['official_opcodes']} "
        f"ops={report['unique_ops']} forms={report['unique_forms']} "
        f"supported={report['supported_forms']} "
        f"context={report['context_forms']} "
        f"unsupported={report['unsupported_forms']} "
        f"runtime_verified={report['runtime_verified_forms']} "
        f"parse_errors={report['parse_err_count']}"
    )
    if changed:
        details = ", ".join(
            f"{field}: expected {expected.get(field)!r}, got {report.get(field)!r}"
            for field in changed
        )
        raise SystemExit(
            f"{version}/{arch}: instruction coverage changed ({details}); "
            "review the form-level report before updating the baseline"
        )
PY
