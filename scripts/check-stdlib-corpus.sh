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

# Build the package tool once, then run it from the repository root. Invoking
# it through `go run -C cmd/plan9asm` would make old Go lanes auto-select the
# nested module's newer toolchain, and the corpus would silently come from the
# wrong GOROOT.
if [[ -n "${PLAN9ASM_CMD:-}" ]]; then
  plan9asm_cmd=$PLAN9ASM_CMD
  if [[ ! -x "$plan9asm_cmd" ]]; then
    echo "PLAN9ASM_CMD is not executable: $plan9asm_cmd" >&2
    exit 1
  fi
else
  plan9asm_cmd="$tmp_root/plan9asm"
  if [[ "${RUNNER_OS:-}" == "Windows" ]]; then
    plan9asm_cmd+=".exe"
  fi
  go build -C cmd/plan9asm -o "$plan9asm_cmd" .
fi

all_targets=(
  "linux 386 i386-unknown-linux-gnu"
  "linux amd64 x86_64-unknown-linux-gnu"
  "linux arm armv7-unknown-linux-gnueabihf"
  "linux arm64 aarch64-unknown-linux-gnu"
  "darwin amd64 x86_64-apple-macosx"
  "darwin arm64 arm64-apple-macosx"
)

# Include Windows by default. Linux CI cross-compiles the COFF corpora, while
# the latest Windows host lane remains an auxiliary native-host integration.
if [[ "${PLAN9ASM_CORPUS_INCLUDE_WINDOWS:-1}" != "0" ]]; then
  all_targets+=(
    "windows 386 i686-pc-windows-msvc"
    "windows amd64 x86_64-pc-windows-msvc"
    "windows arm64 aarch64-pc-windows-msvc"
  )
fi

targets=()
if [[ -z "${PLAN9ASM_CORPUS_TARGETS:-}" ]]; then
  targets=("${all_targets[@]}")
else
  IFS=',' read -r -a requested_targets <<< "$PLAN9ASM_CORPUS_TARGETS"
  for requested in "${requested_targets[@]}"; do
    requested=${requested//[[:space:]]/}
    matched=0
    for target in "${all_targets[@]}"; do
      read -r goos goarch _ <<< "$target"
      if [[ "$requested" == "$goos/$goarch" ]]; then
        targets+=("$target")
        matched=1
        break
      fi
    done
    if [[ "$matched" -eq 0 ]]; then
      echo "unsupported PLAN9ASM_CORPUS_TARGETS entry: $requested" >&2
      exit 1
    fi
  done
fi

for target in "${targets[@]}"; do
  read -r goos goarch triple <<< "$target"

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
unsupported_forms = data.get("unsupported_by_form", [])
parse_errs = data.get("parse_errs") or []
print(f"scan {target}: packages={data['std_pkgs_with_sfile']} files={data['asm_files']} unsupported={len(unsupported)} unsupported_forms={len(unsupported_forms)} parse_errs={len(parse_errs)}")
if unsupported:
    top = ", ".join(f"{item['op']}({item['count']})" for item in unsupported[:12])
    raise SystemExit(f"{target}: unsupported ops remain: {top}")
if unsupported_forms:
    top = ", ".join(
        f"{item['form']} ({item['examples'][0] if item.get('examples') else 'no example'})"
        for item in unsupported_forms[:12]
    )
    raise SystemExit(f"{target}: unsupported operand forms remain: {top}")
if parse_errs:
    top = ", ".join(f"{item['File']}: {item['Err']}" for item in parse_errs[:8])
    raise SystemExit(f"{target}: parse errors remain: {top}")
PY

  echo "==> transpile+compile $goos/$goarch"
  out_dir="$tmp_root/$goos-$goarch-ll"
  meta="$tmp_root/$goos-$goarch-meta.json"
  GOTOOLCHAIN=local "$plan9asm_cmd" transpile -goos="$goos" -goarch="$goarch" -dir "$out_dir" -meta "$meta" std >/dev/null

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
