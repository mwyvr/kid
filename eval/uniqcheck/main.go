// Command uniqcheck verifies, under concurrent load, the two guarantees made
// by kid.New:
//
//  1. Uniqueness: no two calls ever return the same ID. Specifically, the
//     timestamp+sequence (the leading 8 bytes) must never repeat — that is
//     the guarantee getTS makes, independent of the two trailing random
//     bytes. Full-ID duplicates are reported separately.
//  2. Ordering: within any single goroutine, each call to New returns an ID
//     that sorts after the one before it.
//
// Generation runs lock-free: each goroutine records IDs into its own
// preallocated slice, so New experiences genuine concurrent contention
// rather than being serialized behind a checker mutex. Verification is
// post-hoc: all IDs are merged, sorted, and scanned for adjacent duplicate
// timestamp+sequence prefixes, which detects a repeat no matter when, or on
// which goroutine, the two colliding IDs were produced.
//
// Memory: IDs are 10 bytes each; the defaults (4 goroutines x 1,000,000)
// use roughly 40MB. Size -count and -goroutines to available memory.
//
// Usage:
//
//	$ go run . -count 2000000 -goroutines 20
//	uniqcheck: generating 2,000,000 IDs on each of 20 goroutines...
//	Total IDs: 40,000,000  ts+seq dupes: 0  full-ID dupes: 0  ordering violations: 0
//
// Single-threaded alternative using the cmd/kid tool and OS utilities:
//
//	$ go run ../../cmd/kid/main.go -c 10000000 | sort | uniq -d
//	(no output, meaning no duplicates)
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/mwyvr/kid"
)

func main() {
	var (
		numRoutines = 4
		count       = 1000000
	)
	flag.IntVar(&numRoutines, "goroutines", numRoutines, "Number of goroutines")
	flag.IntVar(&count, "count", count, "Generate count IDs per goroutine")
	flag.Parse()

	fmt.Printf("uniqcheck: generating %s IDs on each of %s goroutines...\n",
		commas(count), commas(numRoutines))

	var (
		wg         sync.WaitGroup
		results    = make([][]kid.ID, numRoutines)
		violations = make([]int, numRoutines)
	)
	for g := range numRoutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ids := make([]kid.ID, count)
			var prev kid.ID // nil ID; New always sorts after it
			for i := range count {
				ids[i] = kid.New()
				// per-goroutine ordering: calls within one goroutine are
				// serialized, so each ID must sort after its predecessor
				if ids[i].Compare(prev) <= 0 {
					violations[g]++
				}
				prev = ids[i]
			}
			results[g] = ids
		}()
	}
	wg.Wait()

	// merge and sort, then scan adjacent entries for duplicate ts+seq
	all := make([]kid.ID, 0, numRoutines*count)
	for _, r := range results {
		all = append(all, r...)
	}
	kid.Sort(all)

	tsSeqDupes, fullDupes, ordering := 0, 0, 0
	for _, v := range violations {
		ordering += v
	}
	for i := 1; i < len(all); i++ {
		if bytes.Equal(all[i-1][:8], all[i][:8]) {
			tsSeqDupes++
			if all[i-1] == all[i] {
				fullDupes++
			}
			fmt.Printf("duplicate ts+seq: %v / %v\n", all[i-1], all[i])
		}
	}

	fmt.Printf("Total IDs: %s  ts+seq dupes: %s  full-ID dupes: %s  ordering violations: %s\n",
		commas(len(all)), commas(tsSeqDupes), commas(fullDupes), commas(ordering))
	if tsSeqDupes > 0 || fullDupes > 0 || ordering > 0 {
		fmt.Println("!!! FAILURES DETECTED !!!")
		os.Exit(1)
	}
}

// commas renders n with thousands separators, e.g. 1234567 -> "1,234,567".
func commas(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var b []byte
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, s[i])
	}
	return string(b)
}
