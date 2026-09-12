#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
corpus_root="$repo_root/testdata/corpus"

if [[ -n "${LLVM_CONFIG:-}" ]]; then
  llvm_bin_dir=$("$LLVM_CONFIG" --bindir)
  llc_cmd="$llvm_bin_dir/llc"
else
  llc_cmd=$(command -v llc || command -v llc-23 || command -v llc-22 || command -v llc-21 || command -v llc-20 || command -v llc-19 || true)
fi
if [[ ! -x "$llc_cmd" ]]; then
  echo "llc not found through LLVM_CONFIG or PATH" >&2
  exit 1
fi

if command -v python3 >/dev/null 2>&1; then
  python_cmd=python3
elif command -v python >/dev/null 2>&1; then
  python_cmd=python
else
  echo "Python 3 not found in PATH" >&2
  exit 1
fi

tmp_root=$(mktemp -d)
trap 'rm -rf "$tmp_root"' EXIT
tool="$tmp_root/plan9asmll"
go build -C "$repo_root/cmd/plan9asmll" -o "$tool" .
go -C "$corpus_root" mod download \
  github.com/tmthrgd/go-hex@v0.0.0-20190904060850-447a3041c3bc \
  golang.org/x/sys@v0.36.0

report="$tmp_root/report.json"
(
  cd "$corpus_root"
  "$tool" \
    -targets=linux/amd64,darwin/amd64,windows/amd64 \
    -patterns=github.com/tmthrgd/go-hex \
    -out="$tmp_root/out" \
    -compile \
    -llc="$llc_cmd" \
    -report="$report" \
    -repo-root="$repo_root"
)

"$python_cmd" - "$report" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as source:
    report = json.load(source)

expected = {
    ("linux", "amd64"): 2,
    ("darwin", "amd64"): 2,
    ("windows", "amd64"): 2,
}
actual = {
    (item["goos"], item["goarch"]): item["total_asm"]
    for item in report["targets"]
}
if actual != expected:
    raise SystemExit(f"go-hex assembly inventory changed: expected {expected}, got {actual}")
if report["failed"] or report["success"] != 6:
    raise SystemExit(
        f"go-hex corpus failed: success={report['success']} failed={report['failed']}"
    )
print("go-hex: translated and LLVM-compiled both amd64 assembly files for Linux, macOS, and Windows")
PY
