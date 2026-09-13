package bench

import (
	"crypto/rand"
	"testing"
	"time"

	idgen "github.com/devjefster/GoShortUniqueID/idgen"
	"github.com/mwyvr/kid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"
	"uuid"
)

// kid ids incorporate a timestamp in milliseconds + sequence + a 2-byte random value from math/rand/v2
var resultKID kid.ID

func BenchmarkKid(b *testing.B) {
	var r kid.ID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = kid.New()
		}
		resultKID = r
	})
}

// https://github.com/rs/xid xid ids incorporate time + machine ID + pid +
// random-initialized (once only) monotonically increasing counter
var resultXID xid.ID

func BenchmarkXid(b *testing.B) {
	var r xid.ID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = xid.New()
		}
		resultXID = r
	})
}

// https://github.com/segmentio/ksuid
// ksuid ids incorporate crypto/rand generated numbers
var resultKSUID ksuid.KSUID

func BenchmarkKsuid(b *testing.B) {
	var r ksuid.KSUID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = ksuid.New()
		}
		resultKSUID = r
	})
}

// https://pkg.go.dev/uuid stdlib (Go 1.27+) v4 ids incorporate crypto/rand
// generated numbers
var resultUUIDV4 uuid.UUID

func BenchmarkUuidV4(b *testing.B) {
	var r uuid.UUID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = uuid.NewV4()
		}
		resultUUIDV4 = r
	})
}

// stdlib uuid V7 ids are k-sortable
var resultUUIDV7 uuid.UUID

func BenchmarkUuidV7(b *testing.B) {
	var r uuid.UUID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = uuid.NewV7()
		}
		resultUUIDV7 = r
	})
}

// as configured here, for a good comparison, ulid ids incorporate crypto/rand
// generated numbers
var resultULID ulid.ULID

func BenchmarkUlid(b *testing.B) {
	var r ulid.ULID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rand.Reader)
		}
		resultULID = r
	})
}

// https://github.com/sony/sonyflake
// sonyflake ids incorporate a 10-msec-resolution timestamp + sequence number
// + machine id; generation is mutex-guarded
var resultSonyflake uint64

func BenchmarkSonyflake(b *testing.B) {
	fl, err := sonyflake.New(sonyflake.Settings{})
	if err != nil {
		b.Fatal(err)
	}
	var r uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var err error
			r, err = fl.NextID()
			if err != nil {
				b.Error(err)
			}
		}
		resultSonyflake = r
	})
}

// https://github.com/devjefster/GoShortUniqueID
// goshortuniqueid ids incorporate a second-resolution timestamp + a
// math/rand random string + a per-generator counter; generation is
// mutex-guarded
var resultGSUID string

func BenchmarkGoShortUniqueID(b *testing.B) {
	gen := idgen.New(0, "", "")
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = gen.Generate()
		}
		resultGSUID = r
	})
}
