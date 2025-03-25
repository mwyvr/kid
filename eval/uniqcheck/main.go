// Package main provides a test to determine if ID generation delivers unique
// IDs for Go applications utilizing concurrency.
//
// As getTS() is goroutine safe and generates a unique timestamp+sequence pair
// for each subsequent run, even without two bytes of random data appended, this
// test should never produce a duplicate.
//
// Usage:
//
//	$ go run main.go -count 10000000 -goroutines 5
//	uniqcheck - run with -h to see available options.
//	Generating 10,000,000 IDs per 5 goroutines:
//	Total keys: 50,000,000. Keys in last time tick: 3,836. Number of dupes: 0
//
// Single-threaded testing option, direct the output of the cmd/kid evaluation
// tool to sort | uniq:
//
//	$ go run ../../cmd/kid/main.go -c 10000000 | sort | uniq -d
//	(no output, meaning no duplicates)
package main

import (
	"flag"
	"sync"
	"time"

	"github.com/mwyvr/kid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	dupes int64
	// since the underlying structure of ID is an array, not a slice, kid.ID can be a key
	exists = check{lastTick: 0, keys: make(map[kid.ID]bool)}
	// for thousands separator
	fmt = message.NewPrinter(language.English)
)

type check struct {
	keys      map[kid.ID]bool
	lastTick  int64
	totalKeys int
	mu        sync.RWMutex
}

func main() {
	var (
		wg          sync.WaitGroup
		numRoutines = 4
		count       = 1000000
	)

	flag.IntVar(&numRoutines, "goroutines", numRoutines, "Number of goroutines")
	flag.IntVar(&count, "count", count, "Generate count IDs per goroutine")
	flag.Parse()
	fmt.Printf("uniqcheck - run with -h to see available options.\n\n")
	fmt.Printf("Generating %d IDs per %d goroutines:\n", count, numRoutines)

	for range numRoutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			generate(count)
		}()
	}
	wg.Wait()
	fmt.Printf("Total keys: %d. Keys in last time tick: %d. Number of dupes: %d\n", exists.totalKeys, len(exists.keys), dupes)
	if dupes > 0 {
		fmt.Println("!!! Dupes detected !!!")
	}
}

func generate(count int) {
	var id kid.ID
	for range count {
		id = kid.New()
		tmpTimestamp := time.Now().UnixMilli()
		exists.mu.Lock()
		if exists.lastTick != tmpTimestamp {
			exists.lastTick = tmpTimestamp
			// reset each new millisecond
			exists.keys = make(map[kid.ID]bool)
		}
		if !exists.keys[id] {
			exists.keys[id] = true
			exists.totalKeys++
		} else {
			dupes++
			exists.totalKeys++
			fmt.Printf("Generated: %d, found duplicate: %v\n", exists.totalKeys, id)
		}
		exists.mu.Unlock()
	}
}
