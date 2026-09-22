![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mwyvr/kid) [![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/mwyvr/kid/v2?tab=doc) [![Test](https://github.com/mwyvr/kid/actions/workflows/test.yaml/badge.svg)](https://github.com/mwyvr/kid/actions/workflows/test.yaml) [![codecov](https://codecov.io/gh/mwyvr/kid/branch/main/graph/badge.svg)](https://codecov.io/gh/mwyvr/kid) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# kid

Package kid (K-sortable ID) provides a goroutine-safe generator of short
(10 byte binary, 16 bytes when base32 encoded), url-safe,
[k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence) unique IDs.

Encoded IDs look like:

    06hbcv4y3d9df8yt

## Install

    go get github.com/mwyvr/kid/v2

## Usage

```go
import "github.com/mwyvr/kid/v2"

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

Notes on v2:

- Importing `github.com/mwyvr/kid` (no /v2) resolves to the old v1.x line.
- kid v2 will decode kid v1 IDs into the same binary representation and
  timestamp values, but sequence and random ID segments are not comparable between
  v1 and v2.

## Features

- 10 bytes binary, 16 bytes encoded. No dependencies outside the standard
  library. Requires Go 1.24+.
- Timestamp + sequence unique and strictly increasing per process, even if
  the wall clock steps backwards.
- Lock-free, allocation-free generation; no mutex in the New() path.
- K-orderable in both binary and encoded form.
- URL-friendly encoding without the vowels a, i, o, and u.
- (Un)marshalling for SQL, JSON, text, and binary.
- `cmd/kid` for generation and introspection.

Uniqueness is per process; two processes landing on the same
timestamp+sequence are separated only by 20 bits of randomness (1 in
1,048,576). See [DESIGN.md](DESIGN.md) for the byte layout, capacity
limits, and how to verify the claims yourself.

**Security note**: an ID carries only 20 bits of randomness alongside
values derived from the clock; IDs are predictable by design. Do not use
kid IDs where unguessability matters — session tokens, API keys, password
reset codes.

## CLI

```bash
$ kid
06hb42zvde2y9csj

$ kid -c 2
06hb42zvdkce6e6y
06hb42zvdktcjhwm
```

`kid <id>` decodes and inspects; `echo <id> | kid` decodes from stdin.

    go install github.com/mwyvr/kid/v2/cmd/kid@latest

## More

- [CHANGELOG.md](CHANGELOG.md) — release history.
- [DESIGN.md](DESIGN.md) — byte layout, v1/v2 rationale, uniqueness
  guarantees, package comparisons.
- [eval/bench/BENCHMARKS.md](eval/bench/BENCHMARKS.md) — benchmark
  numbers; run the suite yourself in
  [eval/bench](eval/bench/bench_test.go).

## Acknowledgments

The API borrows from [github.com/rs/xid](https://github.com/rs/xid). The
lock-free ts+seq mechanism derives from
[google/uuid](https://github.com/google/uuid/blob/master/version7.go#L88)'s
getV7Time(). Third-party license texts are in [NOTICES](NOTICES).
