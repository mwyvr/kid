![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mwyvr/kid)[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/mwyvr/kid?tab=doc)[![Test](https://github.com/mwyvr/kid/actions/workflows/test.yaml/badge.svg)](https://github.com/mwyvr/kid/actions/workflows/test.yaml)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)![Coverage](https://img.shields.io/badge/coverage-92.6%25-brightgreen)

# kid

Package kid provides a performant, goroutine-safe generator of short
[k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence) unique IDs
suitable for use where inter-process ID generation coordination is not
required.

Using a non-standard character set (fewer vowels), IDs Base-32 encode as a
16-character URL-friendly, case-insensitive representation like
`dfp7qt0v2pwt0v2x`.

An ID is a:

  - 4-byte timestamp value representing seconds since the Unix epoch, plus a
  - 6-byte random value; see the [Random Source](#random-source) discussion.

Built-in (de)serialization simplifies interacting with SQL databases and JSON.
`cmd/kid` provides the `kid` utility to generate or inspect IDs. Thanks to
`internal/fastrand` introduced in Go 1.19 and made the default `math/rand` source in Go
1.20, ID generation starts fast and scales well as cores are added. De-serialization
has also been optimized. See [Package Benchmarks](#package-benchmarks).

Why `kid` instead of [alternatives](#package-comparisons)?

  - At 10 bytes binary, 16 bytes base32-encoded, kid.IDs are case-insensitive
    and short, yet with 48 bits of uniqueness *per second*, are unique
    enough for many use cases.
  - IDs have a random component rather than a potentially guessable
    monotonic counter found in some libraries.

_**Acknowledgment**: This package borrows heavily from rs/xid
(https://github.com/rs/xid), a zero-configuration globally-unique
high-performance ID generator that leverages ideas from MongoDB
(https://docs.mongodb.com/manual/reference/method/ObjectId/)._

## Example:

```go
id := kid.New()
fmt.Printf("%s\n", id.String()) // example: 06bpwnfe2h3edlj7

id2, err := kid.FromString("06bpwnfe2h3edlj7")
if err != nil {
	fmt.Println(err)
}

fmt.Printf("equal: %v\n", id == id2)
// Output: equal: true

fmt.Printf("%s %d %d %v\n", id2.Time().UTC(), id2.Sequence(), id2.Random(), id2[:])
// Output: 2025-03-07 02:01:11.7 +0000 UTC 1750 20039 [1 149 110 85 205 20 6 214 78 71]
```

## Uniqueness
 
To satisfy whether kid.IDs are unique, run [eval/uniqcheck/main.go](eval/uniqcheck/main.go):

  $ go run eval/uniqcheck/main.go -count 2000000 -goroutines 20

  Generating 2,000,000 IDs per 20 goroutines:
  Total keys: 40,000,000. Keys in last time tick: 1,380. Number of dupes: 0

Or, at the command line, produce IDs and use OS utilities to check (single-threaded):

    $ kid -c 2000000 | sort | uniq -d
    // None output

## CLI

Package `kid` also provides a tool for id generation and inspection:

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

## Change Log

- 2025-03-08 v1.2.0 released. Requires Go 1.24+.
- 2025-03-06 Forked [rid](https://github.com/mwyvr/rid) in favour of kid for
  true k-sortability, requiring a new ID payload, now expected to remain static.
  Improved code coverage and documentation.

## Contributing

Contributions are welcome.

## Package Comparisons

`kid` was born out of a desire for a short, not-guessable, k-sortable unique ID.
`kid` is not designed to be globally unique.

A comparison of a variety of ID generators:

| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Method | Components |
|-----------------------------------------------------------|----|----|-------|---------------------|--------|------------|
| [mwyvr/kid](https://github.com/mwyvr/kid)                 | 10 | 16 |  true | `06bqhl9qdw2ltsc7`<br>`06bqhl9qdw2lx33b`<br>`06bqhl9qdw2lzwvx`<br>`06bqhl9qdw2m0307`  | crypto/rand | 6 byte ts(millisecond) : 2 byte sequence : 2 byte random |
| [rs/xid](https://github.com/rs/xid)                       | 12 | 20 |  true | `cv6e14dq9fa1afbuoev0`<br>`cv6e14dq9fa1afbuoevg`<br>`cv6e14dq9fa1afbuof00`<br>`cv6e14dq9fa1afbuof0g`  | globally | 4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)     | 20 | 27 |  true | `2u3bD06RFqKmjRocArG0nYFNbi1`<br>`2u3bD4lwRryOVcUn0RVf3N9Dfp3`<br>`2u3bCygWvDhLgo6VSydaDZHXi6q`<br>`2u3bD3jOEtbm2ftvHLt9fECjRu1`  | math/rand | 4 byte ts(sec) : 16 byte random |
| [google/uuid](https://github.com/google/uuid) V4          | 16 | 36 | false | `782f8fea-ec19-41c0-a459-cf4f795b3071`<br>`759e72fd-ef8f-4e01-86ad-261263b02e3a`<br>`f7e787ed-8c56-4819-a494-7107c59c033b`<br>`89b0956e-50cb-488e-9641-ae2fc5526c8b`  | crypt/rand UUID | v4: 16 bytes random with version & variant embedded |
| [google/uuid](https://github.com/google/uuid) V7          | 16 | 36 |  true | `0195784d-3767-7551-8e7c-3b584bcc78d0`<br>`0195784d-3767-7552-9429-7c1f9a06de03`<br>`0195784d-3767-7555-865e-b951a7df6853`<br>`0195784d-3767-7556-9fa3-d40367d02981`  | crypt/rand UUID | v7: 16 bytes : 8 bytes time+sequence, version/variant, random |
| [chilts/sid](https://github.com/chilts/sid)               | 16 | 23 |  true | `1WfzjWveCSO-4x2g1m260I~`<br>`1WfzjWveC_x-2LqZ~YcsV7L`<br>`1WfzjWveChs-4K0ebZwbnSv`<br>`1WfzjWveCpf-5G57fDXwfEW`  | math/rand | 8 byte ts(nanosecond) 8 byte random |
| [matoous/go-nanoid/v2](https://github.com/matoous/go-nanoid/) | 21 | 21 | false | `IzV_WtDKMeJXByVzCStWI`<br>`Mm_QfTYzATNAj0tjwJ3gD`<br>`7cx9ReEADqWWUEK27y11v`<br>`0TF6y6BsGapagy15dXxV4`  | crypto/rand | 21 byte rand (adjustable) |
| [sony/sonyflake](https://github.com/sony/sonyflake)       | 16 | 29 |  true | `GU2TMOJSGA2DSMRVGIZTKOBVHAYDE`<br>`GU2TMOJSGA2DSMRVGIZTMNJRGMZTQ`<br>`GU2TMOJSGA2DSMRVGIZTOMJWHA3TI`<br>`GU2TMOJSGA2DSMRVGIZTOOBSGQYTA`  | ts+counter | 39 bit ts(10msec) 8 bit seq, 16 bit mach id |
| [oklog/ulid](https://github.com/oklog/ulid)               | 16 | 26 |  true | `01JNW4TDV7ZFQKNKYYBRSQ3BYB`<br>`01JNW4TDV7CMDMJ0FB5N1XCN3G`<br>`01JNW4TDV7422RPDYADKD0DASE`<br>`01JNW4TDV78NHZHTKKY2AARERT`  | crypt/rand | 6 byte ts(ms) : 10 byte counter random init per ts(ms) |
| [kjk/betterguid](https://github.com/kjk/betterguid)       | 17 | 20 |  true | `-OKsIISbpc0n250oHsGf`<br>`-OKsIISbpc0n250oHsGg`<br>`-OKsIISbpc0n250oHsGh`<br>`-OKsIISbpc0n250oHsGi`  | counter | 8 byte ts(ms) : 9 byte counter random init per ts(ms) |

| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Unique? | Components |
|-----------------------------------------------------------|----|----|-------|---------------------|---------|------------|
| [mwyvr/kid](https://github.com/mwyvr/kid)                 | 10 | 16 |  true | `06bqhh5rnr5h3sj2`<br>`06bqhh5rnr5h42b2`<br>`06bqhh5rnr5h69sk`<br>`06bqhh5rnr5h9z8l`  | crypto/rand | 6 byte ts(millisecond) : 2 byte sequence : 2 byte random |
| [rs/xid](https://github.com/rs/xid)                       | 12 | 20 |  true | `cv6dqnlq9fa04csdltq0`<br>`cv6dqnlq9fa04csdltqg`<br>`cv6dqnlq9fa04csdltr0`<br>`cv6dqnlq9fa04csdltrg`  | counter | 4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)     | 20 | 27 |  true | `2u3ZY4o2ptB7yVyUktXPAVOWdoL`<br>`2u3ZYADxXLQX6QOb9X2d0eC88tQ`<br>`2u3ZY8RqqrV4y8FJDhirejW3xbZ`<br>`2u3ZYAX8WpSutXVZKdsNwTsVTjv`  | math/rand | 4 byte ts(sec) : 16 byte random |
| [google/uuid](https://github.com/google/uuid) V4          | 16 | 36 | false | `e71335bd-d8e3-4778-b0fc-8cef8f92e05d`<br>`367b7633-8deb-4388-bedf-e504a794552d`<br>`4728c53d-fdb1-4392-b101-4d34e3329dcf`<br>`5123ba29-468c-4d85-b78e-cfb04774c9a4`  | crypt/rand | v4: 16 bytes random with version & variant embedded |
| [google/uuid](https://github.com/google/uuid) V7          | 16 | 36 |  true | `01957840-b8ae-7b15-8b86-30ad78faaba7`<br>`01957840-b8ae-7b16-87dc-a816c6e348d2`<br>`01957840-b8ae-7b17-b548-783d043bf258`<br>`01957840-b8ae-7b18-af6d-5260449a70e3`  | crypt/rand | v7: 16 bytes : 8 bytes time+sequence, version/variant, random |
| [chilts/sid](https://github.com/chilts/sid)               | 16 | 23 |  true | `1WfzYbI22St-5~ZoYnbV6vy`<br>`1WfzYbI22b7-6AYuiO6j4zC`<br>`1WfzYbI22iy-5Rhe7fz0TLG`<br>`1WfzYbI22qw-4TThCW26eki`  | math/rand | 8 byte ts(nanosecond) 8 byte randmo |
| [matoous/go-nanoid/v2](https://github.com/matoous/go-nanoid/) | 21 | 21 |  true | `XqPRWxlP7oOcvaTxaB4b8`<br>`88Zx-c44rjLrrnw95Ma1l`<br>`jsr1UYgJ0smO5EEO8OvAw`<br>`K_BSHtABuQTK_L7kM-7Vx`  | crypto/rand | 21 byte rand (adjustable) |
| [sony/sonyflake](https://github.com/sony/sonyflake)       | 16 | 29 |  true | `GU2TMOJRHEYTCOBWHA3TMOJZGIYTA`<br>`GU2TMOJRHEYTCOBWHA3TONRUG42DM`<br>`GU2TMOJRHEYTCOBWHA3TQMZQGI4DE`<br>`GU2TMOJRHEYTCOBWHA3TQOJVHAYTQ`  | counter | 39 bit ts(10msec) 8 bit seq, 16 bit mach id |
| [oklog/ulid](https://github.com/oklog/ulid)               | 16 | 26 |  true | `01JNW41E5EHJ2X4D1JEQ0B5G2Q`<br>`01JNW41E5ET3Q0JC13CCRG55GC`<br>`01JNW41E5EKTCFG3C8H4MVEPVQ`<br>`01JNW41E5EBNJ0XXHWNK725X6Q`  | crypt/rand | 6 byte ts(ms) : 10 byte counter random init per ts(ms) |
| [kjk/betterguid](https://github.com/kjk/betterguid)       | 17 | 20 |  true | `-OKsFAXizLIVMi8MGojv`<br>`-OKsFAXizLIVMi8MGojw`<br>`-OKsFAXizLIVMi8MGojx`<br>`-OKsFAXizLIVMi8MGojy`  | counter | 8 byte ts(ms) : 9 byte counter random init per ts(ms) |

Another comparison of various Go-based unique ID solutions:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

## Package Benchmarks

A benchmark suite comparing some of the above-noted packages can be found in
[eval/bench/bench_test.go](eval/bench/bench_test.go). All runs were done with
scaling_governor set to `performance`:

    echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

```bash
$ go test -cpu 1,2,4,8,16,32 -test.benchmem -bench .
goos: linux
goarch: amd64
pkg: github.com/mwyvr/kid/eval/bench
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkKid                	23253198	       44.42 ns/op	      0 B/op	      0 allocs/op
BenchmarkKid-2              	24385503	       49.65 ns/op	      0 B/op	      0 allocs/op
BenchmarkKid-4              	14705680	       75.48 ns/op	      0 B/op	      0 allocs/op
BenchmarkKid-8              	12582271	       98.48 ns/op	      0 B/op	      0 allocs/op
BenchmarkKid-16             	10654134	      114.5 ns/op	      0 B/op	      0 allocs/op
BenchmarkKid-32             	8707262	      140.4 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid                	40021941	       28.57 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-2              	38214714	       31.09 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-4              	37732369	       31.65 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-8              	37982810	       32.00 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-16             	37114318	       32.72 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-32             	52958653	       22.30 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid              	15703082	       75.19 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-2            	13473422	       81.83 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-4            	11726649	      100.7 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-8            	10216989	      117.4 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-16           	8344321	      148.0 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-32           	6745603	      181.3 ns/op	      0 B/op	      0 allocs/op
BenchmarkGoogleUuid         	23282745	       48.18 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-2       	32059802	       37.19 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-4       	36299127	       32.08 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-8       	38249354	       31.25 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-16      	34422613	       34.56 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-32      	42726945	       28.28 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7       	13948104	       84.37 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-2     	13603436	       84.85 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-4     	11828551	       95.73 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-8     	11514358	      102.4 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-16    	9943927	      121.7 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-32    	8140674	      150.6 ns/op	     16 B/op	      1 allocs/op
BenchmarkUlid               	 201520	     5700 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-2             	 384513	     3085 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-4             	 719776	     1734 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-8             	1000000	     1068 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-16            	 951507	     1206 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-32            	 907669	     1273 ns/op	   5440 B/op	      3 allocs/op
BenchmarkBetterguid         	25408560	       46.06 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-2       	20633276	       52.74 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-4       	18814110	       63.86 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-8       	15597657	       78.82 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-16      	10815397	      105.5 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-32      	9364880	      134.0 ns/op	     24 B/op	      1 allocs/op
PASS
ok  	github.com/mwyvr/kid/eval/bench	53.447s
```
