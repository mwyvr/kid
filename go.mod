module github.com/mwyvr/kid/v2

// kid has no dependencies outside the Go standard library.

// The 1.24 floor is deliberate: kid_test.go's own benchmarks use
// testing.B.Loop. (eval/bench is a separate module with its own floor
// and intentionally doesn't use B.Loop — see its go.mod.)
go 1.24
