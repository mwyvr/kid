module github.com/mwyvr/kid/eval/bench

// The 1.27 floor (higher than the 1.24 used by the root module and
// eval/uniqcheck) is deliberate: this module benchmarks against the
// stdlib "uuid" package, which first ships in Go 1.27.
go 1.27

require (
	github.com/devjefster/GoShortUniqueID v1.1.1
	github.com/mwyvr/kid/v2 v2.0.0
	github.com/oklog/ulid v1.3.1
	github.com/rs/xid v1.6.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sony/sonyflake v1.3.0
)

replace github.com/mwyvr/kid/v2 => ../../
