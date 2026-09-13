# Change Log

## Unreleased (main)

- Export `ZeroID` (previously the unexported `nilID`); document that it is
  both the nil sentinel and a valid, decodable ID.
- Add `ValueBinary()` to write the 10-byte binary form through
  `driver.Valuer`; document the `Value`/`Scan` asymmetry.
- `Scan` now resets the ID to `ZeroID` when given an unsupported type.
- Package doc now states: decoding is case-sensitive (uppercase rejected),
  uniqueness is per process, and the 6-byte timestamp overflows around 2262.
- `Timestamp()`: single big-endian 64-bit load + shift instead of six
  byte-shifts and ORs.
- Require Go 1.24+ (committed benchmarks use `testing.B.Loop`).
- `cmd/kid`: exit 1 when any supplied ID fails to decode; decode IDs from
  piped stdin (`kid -c N | kid` round-trips); terminal stdin still generates.
- `TestNewUnique`: drop the wall-clock delta assertion that flaked on
  forward clock steps; drift behaviors are covered by the stubbed-clock
  tests.
- Add `BenchmarkUnmarshalJSON`; benchmarks confirm the decode paths are
  allocation-free.
- CI: `actions/checkout@v5`/`setup-go@v5`, `go vet`, `go test -race -count=1`
  on [1.24.x, stable], and a golangci-lint job; `.golangci.yml` migrated to
  the v2 schema (govet `shadow`, documented gosec exclusions for the
  deliberate `math/rand/v2` choice).
- `eval/*` submodules now carry tracked `go.mod`/`go.sum`; `.gitignore`
  covers build artifacts.
- Trim doc comments in `kid.go` and `cmd/kid`; README updated to match.

Thanks to @sergeevabc for the heads up on the minimum supported Go version.

## v1.3.1

- `TestNewUnique` rewritten: 10,000 → 100,000 IDs, O(n) duplicate detection,
  strict ordering.
- New tests: `TestNewUniqueParallel` (concurrent ts+seq uniqueness under
  -race), `TestGetTSClockRegression` (wall clock steps backwards),
  `TestGetTSSequenceBorrow` (sequence overflow carries into the timestamp),
  `TestGetTSBurstMonotonic` (saturated burst on a frozen clock), and
  `TestEncodingPreservesOrder` (encoded form preserves byte order).
- README: expanded Uniqueness section with the verification recipes.

## v1.3.0 (2025-03-24)

- `New()` is lock-free and allocation-free: atomic CAS + wait-free
  increment replaces the mutex; trailing bytes from `math/rand/v2` replace
  per-call `crypto/rand`.
- `Compare`/`Sort` consider all 10 bytes, consistent with `==`.
- `UnmarshalJSON` rejects non-string JSON values (bug fix).
- `Scan` accepts the 10-byte binary form.
- Fuzz targets: `FuzzFromString`, `FuzzUnmarshalJSON`, `FuzzFromBytes`.
- `eval/uniqcheck` rewritten: lock-free generation, post-hoc verification.

## v1.2.1 (2025-03-24)

- Add golangci-lint configuration; lint-inspired improvements.
- CI workflow: `go test` on Go 1.23.x and stable, all major OSes.
- Drop minimum required Go version to 1.23.
- Split license files; third-party notices consolidated into NOTICES.
- Updated benchmarks and package comparisons.

## v1.2.0 (2025-03-06)

- Forked from [rid](https://github.com/mwyvr/rid) in favour of kid for true
  k-sortability, with a new ID payload (6-byte ms timestamp + 2-byte
  sequence + 2-byte random) now expected to remain static.
- Improved code coverage and documentation.
