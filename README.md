![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mwyvr/kid) [![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/mwyvr/kid/v2?tab=doc) [![Test](https://github.com/mwyvr/kid/actions/workflows/test.yaml/badge.svg)](https://github.com/mwyvr/kid/actions/workflows/test.yaml) [![codecov](https://codecov.io/gh/mwyvr/kid/branch/main/graph/badge.svg)](https://codecov.io/gh/mwyvr/kid) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# kid

Package kid (K-sortable ID) provides a goroutine-safe generator
of short (10 byte binary, 16 bytes when base32 encoded), url-safe,
[k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence) unique IDs.

Using a custom base32 encoding, IDs encode as url-friendly strings that
look like:

    06hb4fpsz04dpx4p

The 10-byte binary representation of an ID is composed of:

- 6-byte value representing Unix time in milliseconds (max date 10889-08-02)
- 12-bit sequence, and,
- 20 bits of randomness, with k-sortability preserved

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

## kid.ID features

kid has no dependencies outside the standard library, and requires Go 1.24+
(benchmarks use `testing.B.Loop`).

- Size: 10 bytes as binary, 16 bytes if stored/transported as an encoded string.
- Timestamp + sequence is guaranteed to be unique and monotonically increasing
  for each call to New(), even if the wall clock steps backwards.
- 20 bits of trailing randomness to avoid counter-based attacks, drawn
  from math/rand/v2 (seeded by the Go runtime from OS entropy).
- K-orderable in both binary and base32 encoded representations; the encoding
  alphabet is in ascending ASCII order, so encoded strings sort identically to
  the underlying bytes.
- URL-friendly custom encoding without the vowels a, i, o, and u.
- Lock-free, allocation-free ID generation that scales with cores; there is no
  mutex in the New() path.
- Automatic (un)/marshalling for SQL, JSON, text, and binary
  (TextAppender/BinaryAppender, encoding.BinaryMarshaler/BinaryUnmarshaler).
- cmd/kid tool for ID generation and introspection.

**Security note**: an ID carries only 20 bits of randomness alongside values
derived from the clock; IDs are predictable by design. Do not use kid IDs
where unguessability matters, such as session tokens, API keys, or password
reset codes.

## Example usage

```go
func main() {
	id := kid.New()
	fmt.Printf("%s %03v\n", id, id[:])
	// Example output: 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]

	id, err := kid.Parse("06bq7xhnr03mlz6r")
	if err != nil {
		// handle the error
	}
	fmt.Printf("%s %03v\n", id, id[:])
	// Output: 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]
}
```

## Acknowledgments

- While the ID payload differs greatly, the API and much of this package
  borrows heavily from [github.com/rs/xid](https://github.com/rs/xid), a
  zero-configuration globally-unique ID generator.

- The lock-free ts+seq claim in getTS is derived from the
  [github.com/google/uuid](https://github.com/google/uuid/blob/master/version7.go#L88)
  getV7Time() algorithm, with its mutex replaced by a lock-free atomic claim.
  The sequence/randomness packing within the ID's trailing 4 bytes is kid's
  own as of v2: v1 ported getV7Time()'s bit widths directly, which left 4
  bits of every generated ID always zero — a constraint inherited from
  UUIDv7's version nibble, which kid's own layout has no need of. v2 reclaims
  those bits as randomness instead.

Third-party copyright notices and license texts are reproduced in
[NOTICES](NOTICES).

## Uniqueness

Every call to `kid.New()` returns a unique ID, even from many goroutines at
once, and even if the system clock jumps backwards.

Within a process, this is guaranteed, not just likely: the timestamp and
sequence come from one shared atomic counter, so two calls can no more
return the same value than two goroutines incrementing a counter can. It
doesn't depend on the random bytes at all.

Across different processes or machines, there's no coordination — kid
trades that away to stay short (it skips the machine ID and PID bytes some
other ID schemes use). Two processes would need to land on the exact same
millisecond and sequence value to even risk colliding, and even then
they're still separated by 20 bits of randomness — about 1 chance in a
million. If you need guaranteed uniqueness across machines, reach for
something coordinated or longer, like xid or uuid.

**Capacity:** a process can generate up to 4,096 IDs per millisecond (~4.1
million per second). While it is easy to push past that in a benchmark, it is
unlikely to happen in applications kid is a good fit for. in a real application
is unlikely to push past that. Nevertheless, in such conditions the embedded
timestamp starts running ahead of the real clock to keep every ID unique
and sortable. IDs stay correctly ordered no matter what, but the timestamp
becomes an approximation rather than an exact "created at" instant once you're
generating that fast.

### Checking the claims yourself

- **Concurrency:** `go test -race -run TestNewUniqueParallel .` runs ID
  generation under Go's race detector.
- **Decoding:** `go test -fuzz '^FuzzParse$' -fuzztime 60s .` (and the other
  `Fuzz*` targets) fuzz the decode paths.
- **Brute force:** [`eval/uniqcheck`](eval/uniqcheck/main.go) generates
  millions of IDs across many goroutines to confirm uniqueness and ordering.

      go run ./eval/uniqcheck -count 2000000 -goroutines 20
      # Total IDs: 40,000,000  ts+seq dupes: 0  full-ID dupes: 0  ordering violations: 0

  Or, without building anything, a quick single-threaded check with
  standard command-line tools:

      go install github.com/mwyvr/kid/v2/cmd/kid@latest
      kid -c 2000000 | sort | uniq -d
      # no output means no duplicates

## CLI

Package `kid` also provides a tool for id generation and inspection:

```bash
$ kid
06hb42zvde2y9csj

$ kid -c 2
06hb42zvdkce6e6y
06hb42zvdktcjhwm

# produce 4 and inspect
$ kid -c 4 | kid
06hb4gyt7yv261k6 ts:1789428488767 seq:2914 rnd: 198214 2026-09-14 23:28:08.767 +0000 UTC ID{  0x1, 0xa0, 0xa2, 0x3f, 0xda, 0x3f, 0xb6, 0x23,  0x6, 0x46 }
06hb4gyt7yw1w9b7 ts:1789428488767 seq:2945 rnd: 927047 2026-09-14 23:28:08.767 +0000 UTC ID{  0x1, 0xa0, 0xa2, 0x3f, 0xda, 0x3f, 0xb8, 0x1e, 0x25, 0x47 }
06hb4gyt7yw5qprh ts:1789428488767 seq:2949 rnd: 776976 2026-09-14 23:28:08.767 +0000 UTC ID{  0x1, 0xa0, 0xa2, 0x3f, 0xda, 0x3f, 0xb8, 0x5b, 0xdb, 0x10 }
06hb4gyt7yw6ynvm ts:1789428488767 seq:2950 rnd:1005428 2026-09-14 23:28:08.767 +0000 UTC ID{  0x1, 0xa0, 0xa2, 0x3f, 0xda, 0x3f, 0xb8, 0x6f, 0x57, 0x74 }
```

## Change Log

See [CHANGELOG.md](CHANGELOG.md).

## Package Comparisons

`kid` was born out of a desire for a short, url-friendly, k-sortable unique
ID where global uniqueness or inter-process ID generation coordination is not
required.

A comparison of various Go ID generators:

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

## Package Benchmarks

The benchmarks presented are intended to demonstrate nothing more than that the
design of `kid.New()` stands up to heavy use, maintaining low and predictable
cost under concurrent use on both amd64 and arm64 platforms.

See [BENCHMARKS.md](eval/bench/BENCHMARKS.md) for a broader
comparison on Linux amd64 and macOS arm64 including the stdlib uuid
(V4 and V7), ksuid, ulid, sonyflake, and others, or run the suite in
[eval/bench](eval/bench/bench_test.go) on your own hardware.

    ❯ go1.27.1 test -cpu 1,2,4,8,16,32 -test.benchmem -bench .
    goos: linux
    goarch: amd64
    pkg: github.com/mwyvr/kid/v2/eval/bench
    cpu: Intel(R) Core(TM) i9-14900K
    BenchmarkKid                    39246956                30.24 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-2                  40488458                28.21 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-4                  36920178                32.40 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-8                  36412285                32.19 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-16                 34016721                35.04 ns/op            0 B/op          0 allocs/op
    BenchmarkKid-32                 52991698                21.57 ns/op            0 B/op          0 allocs/op
