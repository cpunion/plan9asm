#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
progress=0
if [[ "${1:-}" == --progress ]]; then
  progress=1
  shift
fi
reports_dir=${1:-"$repo_root/_out/discovered-library-corpus"}
shard_count=${2:-64}
if (( $# > 2 )) || ! [[ "$shard_count" =~ ^[1-9][0-9]*$ ]]; then
  echo "usage: $0 [--progress] [reports-directory] [shard-count]" >&2
  exit 2
fi

if (( progress )); then
  if [[ "$reports_dir" != /* ]]; then
    reports_dir="$PWD/$reports_dir"
  fi
  reports_dir=$(go run -C "$repo_root" ./cmd/plan9asmcorpus \
    -repo-root "$repo_root" \
    -discovery-ledger "$repo_root/testdata/discovery/ledger" \
    -compare-assembly-ledger "$repo_root/testdata/discovery/assembly-ledger" \
    -discovery-progress-directory "$reports_dir")
  mkdir -p "$reports_dir"
  reports_dir=$(cd "$reports_dir" && pwd -P)
fi
corpus_args=(
  -manifest "$repo_root/testdata/corpus/reported-libraries.json"
  -repo-root "$repo_root"
  -discovery-ledger "$repo_root/testdata/discovery/ledger"
  -discovery-shard-count "$shard_count"
  -verify-discovery-reports "$reports_dir"
  -compare-assembly-ledger "$repo_root/testdata/discovery/assembly-ledger"
)
if (( progress )); then
  if [[ -e "$reports_dir/scan-status.json" || -L "$reports_dir/scan-status.json" ||
        -e "$reports_dir/assembly-progress.json" || -L "$reports_dir/assembly-progress.json" ]]; then
    echo "progress output already exists; preserve it and use a fresh reports directory" >&2
    exit 1
  fi
  pending_dir=$(mktemp -d "$reports_dir/.status-XXXXXX")
  exec 8<"$pending_dir"
  exec 9>"$pending_dir/scan-status.json"
  status_identity() {
    stat -c '%d:%i' -- "$1" 2>/dev/null || stat -f '%d:%i' -- "$1"
  }
  directory_identity=$(status_identity "$pending_dir")
  file_identity=$(status_identity "$pending_dir/scan-status.json")
  cleanup_status() {
    local status=$?
    if [[ "$(status_identity "$pending_dir")" == "$directory_identity" &&
          "$(status_identity "$pending_dir/scan-status.json")" == "$file_identity" ]]; then
      rm -- "$pending_dir/scan-status.json" || status=1
      rmdir -- "$pending_dir" || status=1
    else
      echo "owned status temporary directory or file changed; preserve replacement" >&2
      status=1
    fi
    exec 9>&-
    exec 8<&-
    exit "$status"
  }
  trap cleanup_status EXIT
  go run -C "$repo_root" ./cmd/plan9asmdiscover \
    -status -out-dir "$repo_root/testdata/discovery/ledger" \
    >&9
  if ! [[ "$(status_identity "$pending_dir")" == "$directory_identity" &&
          "$(status_identity "$pending_dir/scan-status.json")" == "$file_identity" ]]; then
    echo "owned status temporary directory or file changed" >&2
    exit 1
  fi
  # The directory form appends the fixed source basename exactly once. Passing
  # the final leaf would let ln treat a late directory as another destination.
  ln "$pending_dir/scan-status.json" "$reports_dir/"

  # The CLI atomically publishes only fully audited progress. It still exits
  # nonzero for incomplete/failed reports or a committed-ledger mismatch.
  corpus_args+=(-discovery-progress-output "$reports_dir/assembly-progress.json")
  status=0
  go run -C "$repo_root" ./cmd/plan9asmcorpus "${corpus_args[@]}" || status=$?
  if (( status == 0 )) && ! [[ -f "$reports_dir/assembly-progress.json" ]]; then
    echo "strict reader returned success without its requested progress output" >&2
    exit 1
  fi
  exit "$status"
fi

go run -C "$repo_root" ./cmd/plan9asmdiscover \
  -status \
  -out-dir "$repo_root/testdata/discovery/ledger"
go run -C "$repo_root" ./cmd/plan9asmcorpus "${corpus_args[@]}"
