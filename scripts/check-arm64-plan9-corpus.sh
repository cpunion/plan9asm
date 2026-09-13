#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
corpus_root="$repo_root/testdata/corpus"
family_file="$repo_root/testdata/coverage/arm64-go-assembler-families.txt"
baseline="$repo_root/testdata/coverage/arm64-xarch-plan9-baseline.json"
cd "$repo_root"

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
module_json="$tmp_root/xarch.json"
go -C "$corpus_root" mod download -json golang.org/x/arch@v0.31.0 >"$module_json"
xarch_root=$("$python_cmd" - "$module_json" <<'PY'
import json
import sys
print(json.load(open(sys.argv[1], encoding="utf-8"))["Dir"])
PY
)
source_cases="$xarch_root/arm64/arm64asm/testdata/plan9cases.txt"

full_report="$tmp_root/full.json"
go run "$repo_root/cmd/plan9asmscan" \
  -corpus=arm64-plan9 \
  -input="$source_cases" \
  -goos=linux \
  -goarch=arm64 \
  -repo-root="$repo_root" \
  -format=json \
  -out="$full_report"

"$python_cmd" - "$baseline" "$full_report" <<'PY'
import json
import sys

expected = json.load(open(sys.argv[1], encoding="utf-8"))
actual = json.load(open(sys.argv[2], encoding="utf-8"))
fields = ("asm_files", "unique_ops", "unique_forms", "supported_forms", "context_forms", "unsupported_forms", "parse_err_count", "coverage_fingerprint")
changed = [field for field in fields if actual.get(field) != expected.get(field)]
if changed:
    details = ", ".join(f"{field}: expected {expected.get(field)!r}, got {actual.get(field)!r}" for field in changed)
    raise SystemExit(f"x/arch ARM64 Plan 9 corpus changed ({details}); inspect the form report before updating the baseline")
print(f"x/arch ARM64 Plan 9 decoder corpus: ops={actual['unique_ops']} forms={actual['unique_forms']} supported={actual['supported_forms']} context={actual['context_forms']} unsupported={actual['unsupported_forms']} parse_errors={actual['parse_err_count']}")
PY

accepted_cases="$tmp_root/accepted.txt"
"$python_cmd" - "$source_cases" "$family_file" "$accepted_cases" "$(go env GOROOT)" "$tmp_root" <<'PY'
import os
import pathlib
import re
import subprocess
import sys

source_path = pathlib.Path(sys.argv[1])
family_path = pathlib.Path(sys.argv[2])
accepted_path = pathlib.Path(sys.argv[3])
goroot = pathlib.Path(sys.argv[4])
tmp_root = pathlib.Path(sys.argv[5])
families = {
    line.strip()
    for line in family_path.read_text(encoding="utf-8").splitlines()
    if line.strip() and not line.lstrip().startswith("#")
}
selected = []
for line in source_path.read_text(encoding="utf-8").splitlines():
    if "|" not in line:
        continue
    _, asm = line.split("|", 1)
    asm = asm.strip()
    match = re.match(r"^([A-Z][A-Z0-9.]*)\s", asm)
    if match and match.group(1).split(".", 1)[0] in families:
        selected.append(line)

env = os.environ.copy()
env.update({"GOOS": "linux", "GOARCH": "arm64"})
source = tmp_root / "probe.s"
obj = tmp_root / "probe.o"
accepted = []
rejected = []
for number, line in enumerate(selected):
    _, asm = line.split("|", 1)
    source.write_text(f'#include "textflag.h"\nTEXT ·probe{number}(SB),NOSPLIT,$0-0\n\t{asm.strip()}\n\tRET\n', encoding="utf-8")
    result = subprocess.run(
        ["go", "tool", "asm", "-I", str(goroot / "pkg" / "include"), "-o", str(obj), str(source)],
        env=env,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    (accepted if result.returncode == 0 else rejected).append(line)
accepted_path.write_text("\n".join(accepted) + "\n", encoding="utf-8")
print(f"x/arch required families: selected={len(selected)} native_go_accepted={len(accepted)} decoder_only_rejected={len(rejected)}")
if not accepted or len(selected) < 300:
    raise SystemExit("x/arch ARM64 Plan 9 family selection is unexpectedly small")
PY

accepted_report="$tmp_root/accepted.json"
go run "$repo_root/cmd/plan9asmscan" \
  -corpus=arm64-plan9 \
  -input="$accepted_cases" \
  -goos=linux \
  -goarch=arm64 \
  -repo-root="$repo_root" \
  -format=json \
  -out="$accepted_report"

"$python_cmd" - "$accepted_report" <<'PY'
import json
import sys

report = json.load(open(sys.argv[1], encoding="utf-8"))
if report["parse_err_count"] or report["unsupported_forms"]:
    examples = [item["form"] for item in report.get("unsupported_by_form", [])[:12]]
    raise SystemExit(f"native-Go-accepted x/arch forms are not lowerable: parse_errors={report['parse_err_count']} unsupported={report['unsupported_forms']} examples={examples}")
print(f"native-Go-accepted x/arch family corpus: ops={report['unique_ops']} forms={report['unique_forms']} supported={report['supported_forms']} context={report['context_forms']}")
PY
