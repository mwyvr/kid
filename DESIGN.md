# Design notes

_See [CHANGELOG.md](CHANGELOG.md) for what changed release to release._

## Design objectives

- Short, URL-safe and K-sortable: encoded and binary forms sort identically, in generation order.
- Encoding uses a single case and is resistant to accidental rudeness
- No dependencies outside the standard library.
- Performant: Lock-free, allocation-free generation that scales with cores.
- Uniqueness is structural, not probabilistic: per-process, from a single
  atomic counter.
- Small enough to read start to finish in one sitting.

Non-objectives:

- global uniqueness, kid does not aspire to reach cross-machine coordination and
  cryptographic unguessability to get there.

Example kid ID:

    encoded: 06hbdg48wytbs1ln
    binary:  ID{  0x1, 0xa0, 0xa6, 0x3c, 0x88, 0xe7, 0xb4, 0xac, 0x86, 0x75 }

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

## Encoding

A custom alphabet lacking all vowels except for "e" is used for Base32 encoding.

## Uniqueness

Within a process, uniqueness is guaranteed, not probabilistic: the timestamp and
sequence come from one shared atomic counter, so two calls can no more return
the same value than two goroutines incrementing a counter can.

Across processes or machines, there's no coordination, intentionally, as kid
IDs remain short by skipping machine ID and PID bytes or more entropy some other
ID schemes use. Use a coordinated or longer ID (xid, uuid) where cross-machine
uniqueness is required.

**Capacity**: a process can generate up to 4,096 IDs per millisecond
(~4.1 million per second). Past that — easy in a benchmark, unlikely in a
real application — the embedded timestamp runs ahead of the real clock to
keep every ID unique and sortable. Ordering is unaffected; the timestamp
becomes an approximation rather than an exact "created at" instant at
that rate.

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

| Package                                                                     | BLen | ELen | K-Sort | Encoded ID and Next                                                                                                                                                  | Unique                                    | Components                                                                            |
| --------------------------------------------------------------------------- | ---- | ---- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------- |
| [mwyvr/kid](https://github.com/mwyvr/kid)                                   | 10   | 16   | true   | `06h9pzbgh055wlxm`<br>`06h9pzbgh055zq02`<br>`06h9pzbgh05614t3`<br>`06h9pzbgh05621rh`                                                                                 | unique (ts(ms) + sequence) + math/rand/v2 | 6 byte ts(millisecond) : 12 bit sequence : 20 bit random (shared 4 bytes)             |
| [rs/xid](https://github.com/rs/xid)                                         | 12   | 20   | true   | `dajcg0si5pnm303lsb70`<br>`dajcg0si5pnm303lsb7g`<br>`dajcg0si5pnm303lsb80`<br>`dajcg0si5pnm303lsb8g`                                                                 | ts(sec) + machineID + pid + counter       | 4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter               |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)                       | 20   | 27   | true   | `3JHPZTliitkD87kI8qOnSze76ev`<br>`3JHPZSLQp9e6m3T9TRk7mAELRCI`<br>`3JHPZU5mF3LU5s3e2pxI5WOqFd3`<br>`3JHPZQHcjS1DLrQTQb039JWIeFY`                                     | ts + crypto/rand                          | 4 byte ts(sec) : 16 byte random                                                       |
| [uuid](https://pkg.go.dev/uuid) (Go stdlib) V4                              | 16   | 36   | false  | `aac7ab78-a973-4d38-98ea-212241cb070e`<br>`d4036157-accd-466e-8c64-efcf25a55d0d`<br>`e9737579-62a9-4def-8f77-5293e848bb77`<br>`687afd55-c8ba-43c4-b60c-bec03e3315ca` | crypto/rand                               | v4: 122 bits random; 6 bits embedding version & variant                               |
| [uuid](https://pkg.go.dev/uuid) (Go stdlib) V7                              | 16   | 36   | true   | `01a09b7d-4f80-7b34-9b65-d2e15dc3a207`<br>`01a09b7d-4f80-7b35-8a56-6b3a4c7742f1`<br>`01a09b7d-4f80-7b36-a28b-a91b959d2f85`<br>`01a09b7d-4f80-7b37-91d6-21eb314c665d` | ts(ms) + crypto/rand                      | v7: 16 bytes : 48 bits time, 12 bits sequence, 6 bits version/variant, 62 bits random |
| [chilts/sid](https://github.com/chilts/sid)                                 | 16   | 23   | true   | `1ZKw9JMNR68-6p48iWIT_PB`<br>`1ZKw9JMNRLl-1yPbykqwhJM`<br>`1ZKw9JMNRLl-1yPbykqwhJN`<br>`1ZKw9JMNRaO-3BSf4cTC60Q`                                                     | ts + math/rand                            | 8 byte ts(nanosecond) 8 byte random                                                   |
| [matoous/go-nanoid/v2](https://github.com/matoous/go-nanoid/)               | 21   | 21   | false  | `NsVQctw0NmE6GcxOvu6ud`<br>`vEYpx28WjykDI4G5Bg1Lq`<br>`amAaTNAg9Aq5dfYiWkb8n`<br>`jTbWN0yEwV7_BU7BfsY26`                                                             | ts + crypto/rand                          | 21 byte rand (adjustable)                                                             |
| [sony/sonyflake](https://github.com/sony/sonyflake)                         | 8    | 13   | true   | `BDL3FJMMAEAQ4`<br>`BDL3FJMMAIAQ4`<br>`BDL3FJMMAMAQ4`<br>`BDL3FJMMAQAQ4`                                                                                             | ts + sequence + machine id                | 39 bit ts(10ms) : 8 bit seq : 16 bit mach id                                          |
| [oklog/ulid](https://github.com/oklog/ulid)                                 | 16   | 26   | true   | `01M2DQTKW0A29MB8086M3F4Y22`<br>`01M2DQTKW03YHHB6NA20WC35TX`<br>`01M2DQTKW03K678SVFG4ZCFBVY`<br>`01M2DQTKW0BAH9D65JBXWKXKQK`                                         | ts + user-definable rand src              | 6 byte ts(ms) : 10 byte monotonic counter random init per ts(ms)                      |
| [devjefster/GoShortUniqueID](https://github.com/devjefster/GoShortUniqueID) | 14   | 22   | false  | `260913085755GXgyDB0002`<br>`260913085755fh1TPQ0003`<br>`2609130857559cVOat0004`<br>`260913085755TyyNyr0005`                                                         | ts + math/rand + counter                  | 6 byte ts(second) : 6 base62 random : 2 byte counter (mod 10000)                      |

## Benchmarks

See [eval/bench/BENCHMARKS.md](eval/bench/BENCHMARKS.md), or run the
suite yourself: [eval/bench](eval/bench/bench_test.go).
