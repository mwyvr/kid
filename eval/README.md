kid/eval contains:

- bench - benchmarking against compared packages
- compare - generate comparison table for pkg README
- uniqcheck - concurrent uniqueness and ordering verification for mass ID generation

Note: You'll need to run `go mod tidy` to pull in external packages for `bench`
and `compare`; Like kid itself, `uniqcheck` requires only the standard library.
