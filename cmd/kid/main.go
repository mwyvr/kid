// A utility to generate or inspect kid.IDs.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/mwyvr/kid"
)

func main() {
	count := 1
	showVersion := false
	flag.IntVar(&count, "c", count, "Generate N-count IDs")
	flag.BoolVar(&showVersion, "version", showVersion, "Print version and exit")
	flag.Usage = func() {
		fs := flag.CommandLine
		fcount := fs.Lookup("c")

		fmt.Printf("Usage: kid\n\n")
		fmt.Printf("Options:\n")
		fmt.Printf("  kid 06bpk9h5kd17xd7z\t\tDecode the supplied Base32 ID\n")
		fmt.Printf("  kid -%s N\t\t\t%s default: %s\n", fcount.Name, fcount.Usage, fcount.DefValue)
		fmt.Printf("  echo 06bpk9h5kd17xd7z | kid\t\tDecode IDs from stdin\n")
		fmt.Printf("  kid -version\t\t\tPrint version and exit\n\n")
		fmt.Printf("With no parameters and no piped input, kid generates %s random ID encoded as Base32.\n", fcount.DefValue)
		fmt.Printf("Generate and inspect 4 random IDs using Linux/Unix command substitution:\n")
		fmt.Printf("  kid `kid -c 4`\n")
	}
	flag.Parse()
	args := flag.Args()

	if showVersion {
		fmt.Printf("kid %s (%s %s/%s)\n", version(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return
	}

	if count > 1 && len(args) > 0 {
		fmt.Fprintf(flag.CommandLine.Output(),
			"kid: Error, cannot generate ID(s) and inspect at the same time.\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(args) > 0 {
		os.Exit(decodeIDs(args))
	}

	if isTerminal(os.Stdin) {
		// no piped input: generate one or -c N ids
		for c := 1; c <= count; c++ {
			fmt.Fprintf(os.Stdout, "%s\n", kid.New())
		}
		return
	}

	// piped input: decode every whitespace-separated kid value
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kid: reading stdin: %v\n", err)
		os.Exit(1)
	}
	os.Exit(decodeIDs(strings.Fields(string(data))))
}

// decodeIDs decodes and prints each kid value. It returns a non-zero exit
// code if any value fails to decode, so pipelines can detect bad input.
func decodeIDs(vals []string) int {
	failed := false
	for _, arg := range vals {
		id, err := kid.FromString(arg)
		if err != nil {
			fmt.Printf("[%s] %s\n", arg, err)
			failed = true
			continue
		}

		fmt.Printf("%s ts:%d seq:%4d rnd:%5d %s ID{%s }\n", arg,
			id.Timestamp(), id.Sequence(), id.Random(), id.Time(), asHex(id.Bytes()))
	}
	if failed {
		return 1
	}
	return 0
}

// isTerminal reports whether f is attached to a terminal, as opposed to
// receiving piped input.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func asHex(b []byte) string {
	return hex.EncodeToString(b)
}

// version reports the module version recorded by the Go toolchain: the tagged
// version (e.g. v1.3.0) when installed via `go install .../cmd/kid@<tag>`, a
// pseudo-version for untagged commits, or "(devel)" for local builds.
func version() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
		return bi.Main.Version
	}
	return "(unknown)"
}
