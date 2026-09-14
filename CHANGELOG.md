# Change Log

## v2.0.0

Module path is now `github.com/mwyvr/kid/v2`; update imports accordingly.

### Breaking

- **ID layout**: the trailing 4 bytes (previously a clean 2-byte sequence
  + 2-byte random split) are now a single big-endian 32-bit field: a
  12-bit sequence in the high bits, followed by 20 bits of randomness in
  the low bits, with the sequence kept more significant so k-sortability
  is unaffected. v1's 2-byte sequence field only ever used its low 12
  bits — a constraint inherited from porting google/uuid's getV7Time(),
  where those 4 spare bits are forced by UUIDv7's version nibble, a
  constraint that never applied to kid's own layout. v2 reclaims them as
  randomness, raising the cross-process collision-avoidance figure from
  16 random bits (1 in 65,536) to 20 (1 in 1,048,576).

  IDs generated under v1 remain completely valid under v2 — they still
  decode, compare, sort, and round-trip correctly, since those operations
  treat the 10 bytes as opaque. Only `Sequence()`/`Random()` introspection
  is affected: calling either on a *pre-v2* ID under v2 code will not
  recover the original v1 sequence/random split, since the bit boundaries
  moved. Identity, ordering, and storage are unaffected either way.
- `Sequence()` now returns `uint16` (was `int32`); `Random()` now returns
  `uint32` (was `int32`, and needed widening regardless since 20 bits no
  longer fits `uint16`).
- `FromString` renamed to `Parse`, aligning with `time.Parse`/`uuid.Parse`
  convention; no alias retained.

### Added

- `MarshalBinary`/`UnmarshalBinary` (`encoding.BinaryMarshaler`/
  `BinaryUnmarshaler`), so `gob` and anything else that dispatches on
  that interface pair can use `ID` directly.
- `AppendText`/`AppendBinary` (`encoding.TextAppender`/`BinaryAppender`,
  new in Go 1.24): append the encoded/binary form to an existing buffer
  without an intermediate allocation.
- Two more fuzz seeds for `FuzzParse` (long input, non-ASCII).

## v1.4.1

- Fix inaccurate doc comments incorrectly identifying the overflow year
  as 2262, the overflow year for a 64-bit _nanosecond_ timestamp, not this
  48-bit millisecond one.
- `IsZero` is now the canonical method (matching the `ZeroID` name and the
  stdlib `time.Time`/`reflect.Value` convention for non-nilable value
  types); `IsNil` is documented as an alias, kept for readers coming from
  other ID libraries' "nil sentinel" terminology.

## v1.4.0

- Require Go 1.24+ (committed benchmarks use `testing.B.Loop`).
- Export `ZeroID` (previously the unexported `nilID`); document that it is
  both the nil sentinel and a valid, decodable ID.
- Add `ValueBinary()` to write the 10-byte binary form through
  `driver.Valuer`; document the `Value`/`Scan` asymmetry.
- `Scan` now resets the ID to `ZeroID` when given an unsupported type.
- Add `NewWithTime(t)` for generating an ID with a fixed timestamp (tests,
  backfills, replays); out-of-range timestamps (pre-epoch or beyond the
  6-byte field, ~year 10889) return `ErrTimestampOutOfRange`.
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
- `eval/bench`, `eval/compare`: drop discontinued `betterguid` and
  `google/uuid` in favor of the stdlib `uuid` package; add `sony/sonyflake`
  and `devjefster/GoShortUniqueID`; sonyflake row now encodes the uint64
  properly instead of its string form; eval modules require Go 1.27.
- CI: build and vet the `eval/*` modules; upload coverage to codecov.

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
