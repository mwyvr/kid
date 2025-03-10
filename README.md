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
  - Timestamp + sequence is guaranteed to be unique.
  - 2 bytes of trailing randomness to prevent simple counter attacks.
  - K-orderable in both binary and base32 encoded representations.
  - URL-friendly custom encoding without the vowels a, i, o, and u.
  - Automatic (un)/marshalling for SQL and JSON.
  - The cmd/kid tool for ID generation and introspection.

## Example usage

```go
func main() {
    id := kid.New()
	  fmt.Printf("%s %s %03v\n", id, id.String(), id[:])
	  // Example output: 06bq7xhnr03mlz6r 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]

	  id, err := kid.FromString("06bq7xhnr03mlz6r")
	  if err != nil {
	  	// do something
	  }
	  fmt.Printf("%s %s %03v\n", id, id.String(), id[:])
	  // Output: 06bq7xhnr03mlz6r 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]
}
```

## Acknowledgments

- While the ID payload differs greatly, the API and much of this package
borrows heavily from [github.com/rs/xid](https://github.com/rs/xid), a
zero-configuration globally-unique ID generator.

- Unique timestamp+sequence pairs are generated by the
[github.com/google/uuid](https://github.com/google/uuid/blob/master/version7.go#L88) getV7Time() algorithm.

## Uniqueness
 
Each call to `kid.New()` is guaranteed to return a unique ID with a
timestamp+sequence greater than any previous call.

To satisfy whether kid.IDs are unique, run [eval/uniqcheck/main.go](eval/uniqcheck/main.go):

  $ go run eval/uniqcheck/main.go -count 2000000 -goroutines 20

  Generating 2,000,000 IDs per 20 goroutines:
  Total keys: 40,000,000. Keys in last time tick: 1,380. Number of dupes: 0

Or, at the command line, produce IDs and use OS utilities to check (single-threaded):

    $ kid -c 2000000 | sort | uniq -d
    // None output

## CLI

Package `kid` also provides a tool for id generation and inspection:

```bash
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

- 2025-03-08 v1.2.0 released. Requires Go 1.24+.
- 2025-03-06 Forked [rid](https://github.com/mwyvr/rid) in favour of kid for
  true k-sortability, requiring a new ID payload, now expected to remain static.
  Improved code coverage and documentation.

## Contributing

Contributions are welcome.

## Package Comparisons

`kid` was born out of a desire for a short, k-sortable unique ID where global
uniqueness or inter-process ID generation coordination is not required.

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

An article presenting various Go-based unique ID solutions can be found at:
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
