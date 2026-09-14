# Go module assembly discovery ledger

`ledger/` is the repository-owned checkpoint for scans of the official Go
module index. It records every completed exact `module@version`, including
modules without Plan 9 assembly. Assembly matches retain every `.s` path and
architecture hints for all current Go ports, including architectures plan9asm
does not support yet; failures remain eligible for retry.

Results are deliberately stored as uncompressed, line-oriented JSON instead
of a binary gzip blob:

```text
ledger/
  manifest.json
  records/00.jsonl ... records/ff.jsonl
```

`sha256(module)[0]` selects one of 256 shards, so every version of a module
stays in one stable file. Records are sorted by module, Go semantic version,
and result kind. Updating the ledger merges newly resolved exact versions
with the existing records; a later success clears a retryable failure for the
same exact version. Repository tests validate the layout, shard ownership,
ordering, and counts, and reject committed `.gz` discovery results.

Continue from this checkpoint without downloading completed versions again:

```sh
go run ./cmd/plan9asmdiscover \
  -since "$(jq -r .next_since testdata/discovery/ledger/manifest.json)" \
  -limit 20000 \
  -workers 24 \
  -out-dir testdata/discovery/ledger
```

When `-out-dir` already exists, it is automatically used as the completed
exact-version checkpoint and updated in place. `-seen-report` remains
available for additional legacy JSON, gzip-compressed JSON, or sharded import
sources and is repeatable. Every unique module discovered in the index is
resolved to an exact `@latest` version before the seen check; only that exact
version's ZIP inspection can be reused. If `@latest` resolves to a new
version, it is added while older versions remain in the ledger and corpus.

After adding a new supported target, its external assembly corpus can be
replayed directly from the saved exact versions without reading the module
index or revisiting no-assembly modules:

```sh
for shard in $(seq 0 31); do
  PLAN9ASM_DISCOVERY_TARGETS=linux/riscv64 \
    scripts/check-discovered-library-corpus.sh "$shard" 32 \
    "_out/discovered-library-corpus/riscv64-shard-$shard.json"
done
```

The target filter uses the recorded file paths. It conservatively retains
unsuffixed and custom-suffixed files, then fetches only those matched exact
versions and applies the current Go source/build-constraint checks before
translation and LLVM 22 compilation.

The external-library discovery and compilation corpus uses Go 1.27 only. The
Go 1.20–1.27 matrix belongs exclusively to the separate official Go
toolchain/standard-library assembly corpus.
