module github.com/mwyvr/kid/eval/compare

// The 1.27 floor (higher than the 1.24 used by the root module and
// eval/uniqcheck) is deliberate: this module compares against the
// stdlib "uuid" package, which first ships in Go 1.27.
go 1.27

require (
	github.com/chilts/sid v0.0.0-20190607042430-660e94789ec9
	github.com/devjefster/GoShortUniqueID v1.1.1
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/mwyvr/kid v1.4.1
	github.com/oklog/ulid v1.3.1
	github.com/rs/xid v1.6.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sony/sonyflake v1.3.0
)

replace github.com/mwyvr/kid => ../../
