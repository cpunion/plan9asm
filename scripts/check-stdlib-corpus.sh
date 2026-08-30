#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
cd "$repo_root"

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

targets=(
  "linux 386 i386-unknown-linux-gnu"
  "linux amd64 x86_64-unknown-linux-gnu"
  "linux arm64 aarch64-unknown-linux-gnu"
  "darwin amd64 x86_64-apple-macosx"
  "darwin arm64 arm64-apple-macosx"
)

# Include Windows by default. CI may disable these two extra targets for old
# Go compatibility lanes; the latest toolchain and the Windows host lane still
# scan and compile both COFF corpora.
if [[ "${PLAN9ASM_CORPUS_INCLUDE_WINDOWS:-1}" != "0" ]]; then
  targets+=(
    "windows 386 i686-pc-windows-msvc"
    "windows amd64 x86_64-pc-windows-msvc"
    "windows arm64 aarch64-pc-windows-msvc"
  )
fi

for target in "${targets[@]}"; do
  set -- $target
  goos=$1
  goarch=$2
  triple=$3

  echo "==> scan $goos/$goarch"
  json="$tmp_root/$goos-$goarch.json"
  go run ./cmd/plan9asmscan -goos="$goos" -goarch="$goarch" -repo-root . -format json -out "$json"
  "$python_cmd" - "$json" "$goos/$goarch" <<'PY'
import json
import sys

path, target = sys.argv[1], sys.argv[2]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)

unsupported = data.get("unsupported", [])
parse_errs = data.get("parse_errs") or []
print(f"scan {target}: packages={data['std_pkgs_with_sfile']} files={data['asm_files']} unsupported={len(unsupported)} parse_errs={len(parse_errs)}")
if unsupported:
    top = ", ".join(f"{item['op']}({item['count']})" for item in unsupported[:12])
    raise SystemExit(f"{target}: unsupported ops remain: {top}")
if parse_errs:
    top = ", ".join(f"{item['File']}: {item['Err']}" for item in parse_errs[:8])
    raise SystemExit(f"{target}: parse errors remain: {top}")
PY

  echo "==> transpile+compile $goos/$goarch"
  out_dir="$tmp_root/$goos-$goarch-ll"
  meta="$tmp_root/$goos-$goarch-meta.json"
  go run -C cmd/plan9asm . transpile -goos="$goos" -goarch="$goarch" -dir "$out_dir" -meta "$meta" std >/dev/null

  ll_count=$(find "$out_dir" -name '*.ll' | wc -l | tr -d ' ')
  if [ "$ll_count" -eq 0 ]; then
    echo "$goos/$goarch: no .ll files generated" >&2
    exit 1
  fi
  echo "compiled corpus $goos/$goarch: ll_files=$ll_count"

  while IFS= read -r ll; do
    obj="${ll%.ll}.o"
    "$llc_cmd" -mtriple="$triple" -filetype=obj "$ll" -o "$obj"
  done < <(find "$out_dir" -name '*.ll' | sort)
done
