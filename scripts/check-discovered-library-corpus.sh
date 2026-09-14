#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
shard_index=${1:-}
shard_count=${2:-}
report_path=${3:-}
if [[ -z "$shard_index" || -z "$shard_count" || $# -gt 3 ]]; then
  echo "usage: $0 <zero-based-shard-index> <shard-count> [report.json]" >&2
  exit 2
fi
if ! [[ "$shard_index" =~ ^[0-9]+$ && "$shard_count" =~ ^[1-9][0-9]*$ ]] || (( shard_index >= shard_count )); then
  echo "invalid shard $shard_index/$shard_count" >&2
  exit 2
fi

go_version=$(go env GOVERSION)
if [[ ! "$go_version" =~ ^go1[.]27([.]|$) ]]; then
  echo "discovered third-party corpus requires Go 1.27, got $go_version" >&2
  exit 1
fi

# Resolved relative to the checked-out repository.
# shellcheck disable=SC1091
source "$repo_root/scripts/llvm22.sh"
if ! llc_cmd=$(find_llvm22_llc); then
  exit 1
fi

tmp_root=$(mktemp -d)
trap 'rm -rf "$tmp_root"' EXIT
translator="$tmp_root/plan9asmll"
runner="$tmp_root/plan9asmcorpus"
go build -C "$repo_root/cmd/plan9asmll" -o "$translator" .
go build -C "$repo_root" -o "$runner" ./cmd/plan9asmcorpus

if [[ -z "$report_path" ]]; then
  report_path="$repo_root/_out/discovered-library-corpus/shard-$shard_index.json"
fi

runner_args=(
  -manifest="$repo_root/testdata/corpus/reported-libraries.json"
  -discovery-ledger="$repo_root/testdata/discovery/ledger"
  -repo-root="$repo_root"
  -translator="$translator"
  -llc="$llc_cmd"
  -discovery-shard-index="$shard_index"
  -discovery-shard-count="$shard_count"
  -candidate-timeout="${PLAN9ASM_DISCOVERY_CANDIDATE_TIMEOUT:-5m}"
  -discovery-report="$report_path"
)
if [[ -n "${PLAN9ASM_DISCOVERY_TARGETS:-}" ]]; then
  runner_args+=("-discovery-targets=$PLAN9ASM_DISCOVERY_TARGETS")
fi
"$runner" "${runner_args[@]}"
