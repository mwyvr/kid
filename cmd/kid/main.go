// A utility to generate or inspect kid.IDs.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mwyvr/kid"
)

func main() {
	count := 1
	flag.IntVar(&count, "c", count, "Generate N-count IDs")
	flag.Usage = func() {
		fs := flag.CommandLine
		fcount := fs.Lookup("c")

		fmt.Printf("Usage: kid\n\n")
		fmt.Printf("Options:\n")
		fmt.Printf("  kid 06bpk9h5kd17xd7z\t\tDecode the supplied Base32 ID\n")
		fmt.Printf("  kid -%s N\t\t\t%s default: %s\n\n", fcount.Name, fcount.Usage, fcount.DefValue)
		fmt.Printf("With no parameters, kid generates %s random ID encoded as Base32.\n", fcount.DefValue)
		fmt.Printf("Generate and inspect 4 random IDs using Linux/Unix command substitution:\n")
		fmt.Printf("  kid `kid -c 4`\n")
	}
	flag.Parse()
	args := flag.Args()

	if count > 1 && len(args) > 0 {
		fmt.Fprintf(flag.CommandLine.Output(),
			"kid: Error, cannot generate ID(s) and inspect at the same time.\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(args) > 0 {
		// attempt to decode each as an kid
		for _, arg := range args {
			id, err := kid.FromString(arg)
			if err != nil {
				fmt.Printf("[%s] %s\n", arg, err)
				continue
			}

			fmt.Printf("%s ts:%d seq:%4d rnd:%5d %s ID{%s }\n", arg,
				id.Timestamp(), id.Sequence(), id.Random(), id.Time(), asHex(id.Bytes()))
		}
	} else {
		// generate one or -c N ids
		for c := 1; c <= count; c++ {
			fmt.Fprintf(os.Stdout, "%s\n", kid.New())
		}
	}
}

func asHex(b []byte) string {
	s := []string{}
	for _, v := range b {
		s = append(s, fmt.Sprintf(" %#4x", v))
	}

	return strings.Join(s, ",")
}
