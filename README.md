![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mwyvr/kid)[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/mwyvr/kid?tab=doc)[![Test](https://github.com/mwyvr/kid/actions/workflows/test.yaml/badge.svg)](https://github.com/mwyvr/kid/actions/workflows/test.yaml)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)![Coverage](https://img.shields.io/badge/coverage-92.6%25-brightgreen)
# kid

Package kid (K-sortable ID) provides a goroutine-safe generator
of short (10 byte binary, 16 bytes when base32 encoded), url-safe,
[k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence) unique IDs.

The 10-byte binary representation of an ID is composed of:

- 6-byte value representing Unix time in milliseconds
- 2-byte sequence, and,
- 2-byte random value.

IDs encode (base32) as 16-byte url-friendly strings that look like:

    06bqj05bhh2lcbdb

## kid.ID features

- Size: 10 bytes as binary, 16 bytes if stored/transported as an encoded string.
- Timestamp + sequence is guaranteed to be unique and monotonically increasing
  for each call to New(), even if the wall clock steps backwards.
- 2 bytes of trailing randomness to avoid counter-based attacks, drawn
  from math/rand/v2's per-goroutine ChaCha8 generator, seeded by the Go
  runtime from OS entropy.
- K-orderable in both binary and base32 encoded representations; the encoding
  alphabet is in ascending ASCII order, so encoded strings sort identically to
  the underlying bytes.
- Lock-free, allocation-free ID generation that scales with cores; no mutex
  in the New() path.
- URL-friendly custom encoding without the vowels a, i, o, and u.
- Automatic (un)/marshalling for SQL and JSON.
- cmd/kid tool for ID generation and introspection.

Requires Go 1.23+; no newer version is needed for performance — benchmarks
published here were produced with Go 1.26. kid has no dependencies outside
the standard library.

**Security note**: an ID carries only 16 bits of randomness alongside values
derived from the clock; IDs are predictable by design. Do not use kid IDs
where unguessability matters, such as session tokens, API keys, or password
reset codes.

## Example usage

```go
func main() {
	id := kid.New()
	fmt.Printf("%s %03v\n", id, id[:])
	// Example output: 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]

	id, err := kid.FromString("06bq7xhnr03mlz6r")
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

- The timestamp+sequence encoding is derived from the
  [github.com/google/uuid](https://github.com/google/uuid/blob/master/version7.go#L88)
  getV7Time() algorithm; kid replaces its mutex protection with a lock-free
  atomic claim.

Third-party copyright notices and license texts are reproduced in
[NOTICES](NOTICES).

## Uniqueness

Each call to `kid.New()` is guaranteed to return a unique ID with a
timestamp+sequence greater than any previous call, even across goroutines
and even if the system clock steps backwards.

To satisfy whether kid.IDs are unique, run
[eval/uniqcheck/main.go](eval/uniqcheck/main.go), which generates IDs
concurrently without locking, then verifies post-hoc that no
timestamp+sequence pair ever repeats and that each goroutine observed
strictly increasing IDs:

    $ go run eval/uniqcheck/main.go -count 2000000 -goroutines 20
    # example output:
    uniqcheck: generating 2,000,000 IDs on each of 20 goroutines...
    Total IDs: 40,000,000  ts+seq dupes: 0  full-ID dupes: 0  ordering violations: 0

Or, at the command line, produce IDs and use OS utilities to check (single-threaded):

    $ go install github.com/mwyvr/kid/cmd/kid@latest
    $ kid -c 2000000 | sort | uniq -d
    // None output

### Uniqueness scope

The guarantee is per process: the timestamp+sequence pair is claimed from a
single atomic value, so two calls within one process can no more collide
than two atomic increments can return the same number — uniqueness is
structural, not probabilistic, and does not depend on the random bytes.
Across processes or machines there is no coordination (kid deliberately
omits xid's machine ID and PID bytes in exchange for shortness): two
processes that derive the same timestamp+sequence in the same ~256ns window
are separated only by the two random bytes, a 1-in-65,536 chance per such
coincidence. If you need cross-machine uniqueness at high sustained rates,
use a coordinated or longer ID (xid, uuid).

### Capacity and timestamp drift

An ID's uniqueness is carried entirely by the timestamp+sequence pair: 4,096
sequence slots per millisecond, or a sustained capacity of ~4.1 million IDs
per second, per process. (The random bytes multiply the namespace, not the
capacity.) A clock reading derives the sequence from the fractional
nanoseconds at 256ns granularity, yielding values 0-3906; the remaining
slots up to 4095 are headroom consumed under load before the sequence
borrows into the next millisecond.

Generation bursts exceeding ~4.1M IDs/s — trivially reached by benchmarks,
rarely by applications — push the internal clock ahead of real time: each
second of full-rate generation consumes several seconds of timestamp space,
and the embedded timestamps lead the wall clock until generation slows and
real time catches up. Nothing about uniqueness or ordering is affected; IDs
remain strictly k-sortable in generation order at any rate. The practical
guidance: kid is a good fit for systems that treat the embedded time as
approximate metadata, and a poor fit for systems that require ID timestamps
to be exact wall-clock instants under extreme generation rates.

### For the doubtful

The race detector verifies concurrent generation analytically (ordering and
ts+seq uniqueness across goroutines):

    go test -race -run TestNewUniqueParallel -count=20 .

Fuzzing hammers the decode paths (each target runs separately; failing
inputs, should one ever appear, land in testdata/fuzz/ and become permanent
regression tests):

    go test -fuzz '^FuzzFromString$'    -fuzztime 60s .
    go test -fuzz '^FuzzUnmarshalJSON$' -fuzztime 60s .
    go test -fuzz '^FuzzFromBytes$'     -fuzztime 60s .

And uniqcheck brute-forces the uniqueness and ordering guarantees under
real contention — one large run, then a burst loop whose oversubscribed
goroutines explore scheduler interleavings a single run never hits:

    go run ./eval/uniqcheck -count 4000000 -goroutines 32
    for i in $(seq 1 20); do go run ./eval/uniqcheck -count 500000 -goroutines 64 || break; done

uniqcheck holds every generated ID in memory (~10 bytes each) for post-hoc
verification; size count x goroutines to available RAM.

## CLI

Package `kid` also provides a tool for id generation and inspection:

```bash
$ kid -version
kid v1.3.0 (go1.26.3 linux/amd64)

$ kid
06bpwm8x107evvh9

$ kid -c 2
06bpwm3hkm371gz4
06bpwm3hkm3d5ezr

# produce 4 and inspect
kid $(kid -c 4)
06bpwlvhb86bypp7 ts:1741312454738 seq:3247 rnd:23239 2025-03-07 01:54:14.738 +0000 UTC ID{  0x1, 0x95, 0x6e, 0x4f, 0x70, 0x52,  0xc, 0xaf, 0x5a, 0xc7 }
06bpwlvhb86gcdw6 ts:1741312454738 seq:3317 rnd:45958 2025-03-07 01:54:14.738 +0000 UTC ID{  0x1, 0x95, 0x6e, 0x4f, 0x70, 0x52,  0xc, 0xf5, 0xb3, 0x86 }
06bpwlvhb86gkmks ts:1741312454738 seq:3320 rnd:53817 2025-03-07 01:54:14.738 +0000 UTC ID{  0x1, 0x95, 0x6e, 0x4f, 0x70, 0x52,  0xc, 0xf8, 0xd2, 0x39 }
06bpwlvhb86gmb73 ts:1741312454738 seq:3322 rnd:10467 2025-03-07 01:54:14.738 +0000 UTC ID{  0x1, 0x95, 0x6e, 0x4f, 0x70, 0x52,  0xc, 0xfa, 0x28, 0xe3 }
```

## Change Log

v1.3.0: lock-free New(), full-width Compare, hardened decode paths
- New() is lock-free and allocation-free: atomic CAS + wait-free
  increment replaces the mutex; trailing bytes from math/rand/v2
  (ChaCha8) replace per-call crypto/rand
- Compare/Sort consider all 10 bytes, consistent with ==
- UnmarshalJSON rejects non-string JSON values (bug fix)
- Scan accepts the 10-byte binary form
- New tests: clock regression, sequence borrow, parallel CAS stress,
  three fuzz targets; eval/uniqcheck rewritten; CI runs -race on
  1.23.x and stable across linux/macos/windows"

main:
- Drop minimum supported Go version to 1.23, thanks to heads up from @sergeevabc.

v1.2.0 released:
- 2025-03-06 Forked [rid](https://github.com/mwyvr/rid) in favour of kid for
  true k-sortability, requiring a new ID payload, now expected to remain static.
  Improved code coverage and documentation.

## Contributing

Contributions are welcome.

## Package Comparisons

`kid` was born out of a desire for a short, k-sortable unique ID where global
uniqueness or inter-process ID generation coordination is not required.

A comparison of various Go ID generators:

| Package                                                       | BLen | ELen | K-Sort | Encoded ID and Next                                                                                                                                                  | Unique                                   | Components                                                                            |
| ------------------------------------------------------------- | ---- | ---- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- | ------------------------------------------------------------------------------------- |
| [mwyvr/kid](https://github.com/mwyvr/kid)                     | 10   | 16   | true   | `06bwz2qyzm14d070`<br>`06bwz2qyzm14fnte`<br>`06bwz2qyzm14hxmf`<br>`06bwz2qyzm14kdl1`                                                                                 | unique (ts(ms) + sequence) + chacha8 rand | 6 byte ts(millisecond) : 2 byte sequence : 2 byte random                              |
| [rs/xid](https://github.com/rs/xid)                           | 12   | 20   | true   | `cvhjc0tq9fa75iaa3d00`<br>`cvhjc0tq9fa75iaa3d0g`<br>`cvhjc0tq9fa75iaa3d10`<br>`cvhjc0tq9fa75iaa3d1g`                                                                 | ts(sec) + machineID + pid + counter      | 4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter               |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)         | 20   | 27   | true   | `2upRtyliBRn6UnfS2RsdkEIhqbg`<br>`2upRu1TTpojt5KQDykjTjreGXGE`<br>`2upRu0IZ0RbjFMSS1lb0Io3aQ8A`<br>`2upRu2AjlZoy3rnU6MJdqSuDs1H`                                     | ts + crypto/rand                         | 4 byte ts(sec) : 16 byte random                                                       |
| [google/uuid](https://github.com/google/uuid) V4              | 16   | 36   | false  | `f03fed10-c632-4d06-95b5-6783796e6aaa`<br>`1c6c044b-66db-44e8-ac45-0e4e358dcc1f`<br>`9b26d7cd-d85d-4696-beec-0483d6446e7f`<br>`f2cf8c26-4e3c-4612-865c-2270883456fd` | crypt/rand                               | v4: 122 bits random; 6 bits embedding version & variant                               |
| [google/uuid](https://github.com/google/uuid) V7              | 16   | 36   | true   | `0195cf8a-fefd-725a-8655-0910a814aa54`<br>`0195cf8a-fefd-725b-b7ee-b178436e1aa0`<br>`0195cf8a-fefd-725e-b663-bc0ae6ee45f9`<br>`0195cf8a-fefd-725f-8936-57d2937b3ecc` | ts(ms) + crypt/rand                      | v7: 16 bytes : 48 bits time, 12 bits sequence, 6 bits version/variant, 62 bits random |
| [chilts/sid](https://github.com/chilts/sid)                   | 16   | 23   | true   | `1WlBXfeKc31-1L10__76tcP`<br>`1WlBXfeKcCQ-4_us_ZM7ZtY`<br>`1WlBXfeKcMr-0HIZhTn1J_0`<br>`1WlBXfeKcUN-4AYyBPKLA5P`                                                     | ts + math/rand                           | 8 byte ts(nanosecond) 8 byte random                                                   |
| [matoous/go-nanoid/v2](https://github.com/matoous/go-nanoid/) | 21   | 21   | false  | `iaGKMTIcslXAPNkbIp4ho`<br>`hM2E2H2y56NtaljmWCfNs`<br>`SpTDFWVxLW_Rrk_S93A87`<br>`kPxkTozTenuSfwAO0s9rb`                                                             | ts + crypto/rand                         | 21 byte rand (adjustable)                                                             |
| [sony/sonyflake](https://github.com/sony/sonyflake)           | 16   | 29   | true   | `GU2TSMZXGYYTCNZVGYYDANBZHEZTA`<br>`GU2TSMZXGYYTCNZVGYYDCMJVGQ3DM`<br>`GU2TSMZXGYYTCNZVGYYDCOBRGAYDE`<br>`GU2TSMZXGYYTCNZVGYYDENBWGUZTQ`                             | ts + counter                             | 39 bit ts(10msec) 8 bit seq, 16 bit mach id                                           |
| [oklog/ulid](https://github.com/oklog/ulid)                   | 16   | 26   | true   | `01JQ7RNZQXCZ605958V4E9XMF0`<br>`01JQ7RNZQXK35FDQBEQM86VMEF`<br>`01JQ7RNZQXJV8JBXKFE1VVPJJT`<br>`01JQ7RNZQX1GJ4EQSG1KC3RY8N`                                         | ts + user-definable rand src             | 6 byte ts(ms) : 10 byte monotonic counter random init per ts(ms)                      |
| [kjk/betterguid](https://github.com/kjk/betterguid)           | 17   | 20   | true   | `-OMEXjvxqWZMbCF4xNP6`<br>`-OMEXjvxqWZMbCF4xNP7`<br>`-OMEXjvxqWZMbCF4xNP8`<br>`-OMEXjvxqWZMbCF4xNP9`                                                                 | ts + rand-init counter                   | 8 byte ts(ms) : 9 byte counter random init per ts(ms)                                 |

## Package Benchmarks

kid's `New()` is lock-free and allocation-free, and generation scales with
cores; against rs/xid — the fastest comparable generator — it is effectively
a draw on amd64 and within ~20% on Apple silicon, while additionally
guaranteeing strictly increasing timestamp+sequence ordering.

Benchmarked with Go 1.26 (`go test -cpu 1,2,4,8,16,32 -test.benchmem -bench .`
in [eval/bench](eval/bench/bench_test.go)). On Linux, set the scaling
governor to `performance`; on macOS laptops, use High Power mode:

    echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkKid                    36166317                28.90 ns/op            0 B/op          0 allocs/op
BenchmarkKid-2                  35779724                30.58 ns/op            0 B/op          0 allocs/op
BenchmarkKid-4                  38237456                32.65 ns/op            0 B/op          0 allocs/op
BenchmarkKid-8                  32437368                33.14 ns/op            0 B/op          0 allocs/op
BenchmarkKid-16                 33878541                35.40 ns/op            0 B/op          0 allocs/op
BenchmarkKid-32                 51138286                22.47 ns/op            0 B/op          0 allocs/op
BenchmarkXid                    44153774                27.32 ns/op            0 B/op          0 allocs/op
BenchmarkXid-2                  43354214                28.38 ns/op            0 B/op          0 allocs/op
BenchmarkXid-4                  36619646                28.60 ns/op            0 B/op          0 allocs/op
BenchmarkXid-8                  38258444                29.05 ns/op            0 B/op          0 allocs/op
BenchmarkXid-16                 48193314                32.18 ns/op            0 B/op          0 allocs/op
BenchmarkXid-32                 56045674                20.66 ns/op            0 B/op          0 allocs/op
```

```
goos: darwin
goarch: arm64
cpu: Apple M4 Max
BenchmarkKid                    38023658                31.19 ns/op            0 B/op          0 allocs/op
BenchmarkKid-2                  37417369                31.95 ns/op            0 B/op          0 allocs/op
BenchmarkKid-4                  39099517                30.96 ns/op            0 B/op          0 allocs/op
BenchmarkKid-8                  19487215                60.74 ns/op            0 B/op          0 allocs/op
BenchmarkKid-16                 15209077                79.36 ns/op            0 B/op          0 allocs/op
BenchmarkXid                    38485384                31.06 ns/op            0 B/op          0 allocs/op
BenchmarkXid-2                  36064644                30.28 ns/op            0 B/op          0 allocs/op
BenchmarkXid-4                  44846259                28.37 ns/op            0 B/op          0 allocs/op
BenchmarkXid-8                  26422333                46.76 ns/op            0 B/op          0 allocs/op
BenchmarkXid-16                 18502420                65.51 ns/op            0 B/op          0 allocs/op
```

For a broader comparison including ksuid, google/uuid (V4 and V7), ulid, and
betterguid, run the suite in [eval/bench](eval/bench/bench_test.go) on your
own hardware.
