package bench

import (
	"crypto/rand"
	"testing"
	"time"
	"uuid"

	"github.com/devjefster/GoShortUniqueID/idgen"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"

	"github.com/mwyvr/kid/v2"
)

// Unless noted otherwise, package id creation is zero allocation.

// kid ids incorporate a timestamp in milliseconds + sequence + a 20-bit random value from math/rand/v2
func BenchmarkKid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var r kid.ID
		for pb.Next() {
			r = kid.New()
		}
		_ = r
	})
}

// https://github.com/rs/xid xid ids incorporate time + machine ID + pid +
// random-initialized (once only) monotonically increasing counter
func BenchmarkXid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var r xid.ID
		for pb.Next() {
			r = xid.New()
		}
		_ = r
	})
}

// https://github.com/segmentio/ksuid
// ksuid ids incorporate crypto/rand generated numbers
func BenchmarkKsuid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var r ksuid.KSUID
		for pb.Next() {
			r = ksuid.New()
		}
		_ = r
	})
}

// https://pkg.go.dev/uuid stdlib (Go 1.27+) v4 ids incorporate crypto/rand
// generated numbers
func BenchmarkUuidV4(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var r uuid.UUID
		for pb.Next() {
			r = uuid.NewV4()
		}
		_ = r
	})
}

// stdlib uuid V7 ids are k-sortable
func BenchmarkUuidV7(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var r uuid.UUID
		for pb.Next() {
			r = uuid.NewV7()
		}
		_ = r
	})
}

// as configured here, ulid ids incorporate crypto/rand
// generated numbers. MustNew() is not zero allocation.
func BenchmarkUlid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var r ulid.ULID
		for pb.Next() {
			r = ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rand.Reader)
		}
		_ = r
	})
}

// https://github.com/sony/sonyflake
// sonyflake ids incorporate a 10-msec-resolution timestamp + sequence number
// + machine id; generation is mutex-guarded
func BenchmarkSonyflake(b *testing.B) {
	fl, err := sonyflake.New(sonyflake.Settings{})
	if err != nil {
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		var r uint64
		for pb.Next() {
			var err error
			r, err = fl.NextID()
			if err != nil {
				b.Error(err)
			}
		}
		_ = r
	})
}

// https://github.com/devjefster/GoShortUniqueID
// goshortuniqueid ids incorporate a second-resolution timestamp + a
// math/rand random string + a per-generator counter; generation is
// mutex-guarded. Generate() is not zero allocation.
func BenchmarkGoShortUniqueID(b *testing.B) {
	gen := idgen.New(0, "", "")
	b.RunParallel(func(pb *testing.PB) {
		var r string
		for pb.Next() {
			r = gen.Generate()
		}
		_ = r
	})
}
