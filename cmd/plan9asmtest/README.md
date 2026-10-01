# Exhaustive root-package test partitions

`plan9asmtest` runs one exact Go package in fresh, serial child processes. It
discovers the complete current `Test`, runnable `Example` and `Fuzz` inventory
with actual `go test -list=.`; no checked-in test list is used. Go's separately
listed benchmarks are retained in the report but not run, matching ordinary
`go test` without `-bench`. Fuzz targets run their normal seed corpus, not a
fuzzing campaign. Every emitted subtest and seed must also terminate.

From the repository root, select the required Go/LLVM tools described in
[validation](../../docs/development/validation.md), then use a new output path:

```sh
mkdir -p _out
GOFLAGS= go run ./cmd/plan9asmtest \
  -out _out/root-partitions-current -shards 32 -timeout 45m
```

This covers **only the root package**, not root `./...`, the nested CLI
modules, strict stdlib/object gates or cross-runtime jobs. Those still require
their existing complete commands. To run every other root-module package
without repeating the root:

```sh
go list ./... > _out/non-root-packages.txt || exit 1
suite_status=0
while IFS= read -r pkg; do
  if [ "$pkg" != github.com/xgo-dev/plan9asm ]; then
    go test "$pkg" -count=1 -timeout=45m || suite_status=1
  fi
done < _out/non-root-packages.txt
test "$suite_status" -eq 0
```

Use the same race/coverage/build flags for those other packages when required.
The tool does not change workflows, and its coverage profile is root-package
coverage only, not a replacement for the original module-wide profile.

Build/instrumentation flags follow `--`, for example:

```sh
GOFLAGS= go run ./cmd/plan9asmtest \
  -out _out/root-partitions-race-cover -shards 32 -timeout 60m -- \
  -race -covermode=atomic -coverprofile=coverage-root.out -p=1
```

The tool can be built/run from a separate tool checkout while targeting a
clean frozen source checkout through `-dir /path/to/frozen-source`. Neither
checkout is edited. An absolute `-out` separates tool evidence from source.

The actual list and every execution use identical build flags. Each nonempty
partition has an independently generated coverage profile. Only after all
children and JSON validations succeed are the profiles merged into a **new**
destination, checking identical block/statement inventories and modes. `set`
counters combine by OR; `count`/`atomic` combine by checked uint64 addition.
Failed runs retain their separate profiles, not a misleading merged result.

Supported flags are `-race`, `-cover`, `-covermode`, `-coverpkg`, `-tags`, `-p`,
`-gcflags`, `-asmflags`, `-ldflags`, `-mod`, `-modfile`, `-overlay`, `-trimpath`,
`-asan`, `-msan` and `-work`. Filters, `-count`, `-fuzz`, benchmarks, custom
executors and test arguments are rejected. `GOFLAGS` must be empty: pass its
supported build flags explicitly instead of allowing hidden test filters.
The runner forces `-count=1`; successful cached runs cannot replace execution.

## Proof and failure behavior

The sorted inventory hashes newline-separated names with SHA-256. The first
eight bytes of each name's SHA-256, interpreted big-endian modulo the shard
count, determine its partition. Sorting/sharding never deletes a name. Empty
partitions are explicitly recorded; an entirely empty inventory fails.
The exact escaped alternation selects every planned root in its partition.
A pattern over 16,000 bytes fails and asks for more shards; it is never cut to
fit command-line limits.

`report.json` retains the complete plan, commands, per-root/subtest/seed
outcomes, elapsed times, incomplete counts and errors. Raw actual inventory,
Go version, JSON and stderr logs remain alongside it. Each root must have one
actual run and one terminal pass/fail/skip; unknown, missing or repeated roots,
unterminated seeds, malformed JSON, failed package/child exits and timeouts
fail. Other partitions are still attempted after a failure. Existing Go skips
are counted separately from assertions passed; they do not prove execution of
the skipped semantics. Required cross/native counterpart jobs remain required.

The wall deadline also covers compilation; Go receives the same test timeout.
Combined stdout/stderr capture is bounded at 64 MiB per command. On Unix every
child owns a process group, terminated on cancellation **and ordinary exit**,
with a bounded output-pipe wait. Interrupt, TERM and HUP cancel owned groups.
This releases LLVM contexts between shards on Linux and Darwin.
The initial tool intentionally fails on non-Unix systems until equivalent
descendant-cleanup proof exists; it does not pretend Windows execution passed.

Freeze source, environment and tools throughout a real suite run. Reports
describe the invoked Go tests, not assembly-ledger or external-module coverage.
The runner never deletes caller caches or modifies baseline/scan inputs.
