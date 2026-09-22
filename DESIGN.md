# Design notes

See [CHANGELOG.md](CHANGELOG.md) for what changed release to release.

## Design objectives

- Short and URL-safe: shorter than [rs/xid](https://github.com/rs/xid) or uuid,
  trading away cross-machine coordination and cryptographic unguessability to
  get there.
- The next produced ID isn't a pure function of the last one, unlike
  counter-based schemes.
- K-sortable: encoded and binary forms sort identically, in generation order.
- No dependencies outside the standard library.
- Lock-free, allocation-free generation that scales with cores.
- Uniqueness is structural, not probabilistic: per-process, from a single
  atomic counter.
- Small enough to read start to finish in one sitting.

## Byte layout

An ID is 10 bytes: a 48-bit millisecond timestamp, then a 32-bit field
holding a 12-bit sequence and 20 bits of randomness, sequence in the high
bits so ordering is unaffected.

```
byte:    0    1    2    3    4    5    6    7    8    9
      +----+----+----+----+----+----+----+----+----+----+
      |        unix_ts_ms (48 bits)      |  seq + rnd   |
      +----+----+----+----+----+----+----+----+----+----+
                                     \_________________/
                                              |
        bytes 6-9, a 32-bit big-endian field--+

 bit: 31                  20 19                         0
      +---------------------+---------------------------+
      |   sequence (12)     |      randomness (20)      |
      +---------------------+---------------------------+
        higher bits: sorts first within a millisecond, so
        randomness never disturbs k-sortability
```

Max date at 48 bits: 10889-08-02.

`New()`/`getTS()` derive milli and the sub-millisecond fraction via
`UnixMilli()` + `Nanosecond()`, not `UnixNano()`: `UnixNano()`'s int64
range is documented as undefined past ~year 2262, far short of the format's
own 10889 ceiling.

## Why 12+20, not 16+16

kid v2 decodes, compares, sorts, and round-trips v1 IDs correctly; only
Sequence() and Random() on a pre-v2 ID won't recover the original v1 split,
since the bit boundaries moved.

v1 ported google/uuid's getV7Time() bit widths directly: a 12-bit sequence, with
the remaining 4 bits of the 2-byte sequence field always zero. That constraint
isn't kid's — it's UUIDv7's version nibble, which kid's own layout has no need
of. v2 reclaims those 4 bits as randomness, raising cross-process collision
resistance from 16 random bits (1 in 65,536) to 20 (1 in 1,048,576), at no cost
to k-sortability: the sequence stays in the higher, more significant bits of the
shared field.

## Uniqueness

Within a process, uniqueness is guaranteed.

In-process uniqueness is delivered by the timestamp and sequence generator; the
trailing random sequence is not there to enforce uniqueness.

Across processes or machines, there's no coordination, intentionally, as kid
IDs remain short by skipping machine ID and PID bytes or additional entropy some
other ID schemes use.

If two processes generate an ID within the same millisecond, the odds of a full
collision are only one in 4 billion:

      seq        randomness
    (1/3,907) × (1/1,048,576) = 1 in 4,096,786,432 (per millisecond)

In contrast, even under the same conservative standard, stdlib UUID v7's
worst-case probability of collision is at least one in 4,611,686,018,427,387,904.

Moral of the story: Use a coordinated or longer ID (xid, uuid) where
cross-machine uniqueness is required.

**Capacity**: a process can generate up to 4,096 IDs per millisecond (~4.1
million per second). Past that point, which easy in a benchmark but unlikely in
a real application, the embedded timestamp runs ahead of the real clock to keep
every ID unique and sortable. Ordering is unaffected; the timestamp becomes an
approximation rather than an exact "created at" instant at that rate.

**Timestamp ceiling**: the 48-bit timestamp field itself allows dates
only up to ~year 10889. `NewWithTime` range-checks against this and
returns `ErrTimestampOutOfRange`; `New` does not — its signature
returns only an `ID`, in keeping with the Go standard library's
[`uuid.NewV7`](https://pkg.go.dev/uuid#NewV7),

### Verifying uniqueness

- **Concurrency**: `go test -race -run TestNewUniqueParallel .`
- **Decoding**: `go test -fuzz '^FuzzParse$' -fuzztime 60s .` (and the
  other `Fuzz*` targets)
- **Brute force**: [`eval/uniqcheck`](eval/uniqcheck/main.go) generates
  millions of IDs across many goroutines and checks afterward for
  duplicates and ordering violations:

      go run ./eval/uniqcheck -count 2000000 -goroutines 20
      # Total IDs: 40,000,000  ts+seq dupes: 0  full-ID dupes: 0  ordering violations: 0

  Single-threaded, without building anything:

      go install github.com/mwyvr/kid/v2/cmd/kid@latest
      kid -c 2000000 | sort | uniq -d
      # no output means no duplicates

## Package comparisons

kid was born out of a desire for a short, url-friendly, k-sortable unique
ID where global uniqueness or inter-process coordination is not required.

| Package                                                                     | BLen | ELen | K-Sort | Encoded ID and Next 3                                                                                                                                                | Unique                                    | Components                                                                            |
| --------------------------------------------------------------------------- | ---- | ---- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------- |
| [mwyvr/kid](https://github.com/mwyvr/kid)                                   | 10   | 16   | true   | `06hd9gqjl6v47tg3`<br>`06hd9gqjl6v5j5pb`<br>`06hd9gqjl6v6vmgx`<br>`06hd9gqjl6v7d1c0`                                                                                 | unique (ts(ms) + sequence) + math/rand/v2 | 6 byte ts(millisecond) : 12 bit sequence : 20 bit random (shared 4 bytes)             |
| [rs/xid](https://github.com/rs/xid)                                         | 12   | 20   | true   | `daolfisi5pnnin79nr80`<br>`daolfisi5pnnin79nr8g`<br>`daolfisi5pnnin79nr90`<br>`daolfisi5pnnin79nr9g`                                                                 | ts(sec) + machineID + pid + counter       | 4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter               |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)                       | 20   | 27   | true   | `3Je2WMHVruCf2BTAhqm0gsQhpmc`<br>`3Je2WLrSODLH7iM0addtkE0ttSP`<br>`3Je2WNkckxFBKzh67YKOFczX4s4`<br>`3Je2WNvM1eYqHTMPmrGUw5hM9O4`                                     | ts + crypto/rand                          | 4 byte ts(sec) : 16 byte random                                                       |
| [uuid](https://pkg.go.dev/uuid) (Go stdlib) V4                              | 16   | 36   | false  | `d9d5cec4-3bb8-4899-a6d4-2c52293d4536`<br>`57b58cb7-94b2-42fa-a043-657aa6ebb7b9`<br>`3b27e0a3-2e88-4fd4-a377-6905b9661b2d`<br>`980ffff0-aba2-4b4a-a6bc-7a4171c34269` | crypto/rand                               | v4: 122 bits random; 6 bits embedding version & variant                               |
| [uuid](https://pkg.go.dev/uuid) (Go stdlib) V7                              | 16   | 36   | true   | `01a0c4be-f199-7c52-be61-01bb4cfc3580`<br>`01a0c4be-f199-7c53-90a0-8376670a0d52`<br>`01a0c4be-f199-7c54-b41b-e0787dcaf1df`<br>`01a0c4be-f199-7c55-9278-dee046499aa0` | ts(ms) + crypto/rand                      | v7: 16 bytes : 48 bits time, 12 bits sequence, 6 bits version/variant, 62 bits random |
| [chilts/sid](https://github.com/chilts/sid)                                 | 16   | 23   | true   | `1ZNOXfI2AF8-2uB5~gvD6d6`<br>`1ZNOXfI2AUl-3AK4CePBg00`<br>`1ZNOXfI2AUl-3AK4CePBg01`<br>`1ZNOXfI2AjO-2GB5u_VcsZq`                                                     | ts + math/rand                            | 8 byte ts(nanosecond) 8 byte random                                                   |
| [matoous/go-nanoid/v2](https://github.com/matoous/go-nanoid/)               | 21   | 21   | false  | `lfvbe617GrfsMs6mwRrKj`<br>`bX0OhNpS4ad3fWzINtfan`<br>`mc_dZmKcdkIy-0-H1mFoN`<br>`OiqThXtoPC-eRAMcxxvYT`                                                             | ts + crypto/rand                          | 21 byte rand (adjustable)                                                             |
| [sony/sonyflake](https://github.com/sony/sonyflake) (kid's base32 encoding) | 8    | 13   | true   | `13ex5ltw04gem`<br>`13ex5ltw08gem`<br>`13ex5ltw0dgem`<br>`13ex5ltw0hgem`                                                                                             | ts + sequence + machine id                | 39 bit ts(10ms) : 8 bit seq : 16 bit mach id                                          |
| [oklog/ulid](https://github.com/oklog/ulid)                                 | 16   | 26   | true   | `01M32BXWCSWFVRZM3ZVZ3VT897`<br>`01M32BXWCSEFGA8FTQJD4A9KZV`<br>`01M32BXWCSGWR0Y87ACG8XHESK`<br>`01M32BXWCSC2027WG775GY9FQJ`                                         | ts + user-definable rand src              | 6 byte ts(ms) : 10 byte monotonic counter random init per ts(ms)                      |
| [devjefster/GoShortUniqueID](https://github.com/devjefster/GoShortUniqueID) | 14   | 22   | false  | `260921091403b9j5mh0002`<br>`260921091403yg4JFb0003`<br>`2609210914031ho7kQ0004`<br>`260921091403mbbROv0005`                                                         | ts + math/rand + counter                  | 6 byte ts(second) : 6 base62 random : 2 byte counter (mod 10000)                      |

xid and sonyflake carry no randomness beyond the timestamp — the next ID from
a live process is exactly computable, not merely difficult to predict. ulid's
monotonic mode has the same property within a single millisecond (random only
once per ms, then incremented). The remaining packages, kid included, include
genuine random bits, so the next ID cannot be computed — only guessed from a
space of possibilities.

## Benchmarks

See [eval/bench/BENCHMARKS.md](eval/bench/BENCHMARKS.md), or run the
suite yourself: [eval/bench](eval/bench/bench_test.go).
