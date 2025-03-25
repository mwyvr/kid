package bench

import (
	"crypto/rand"
	"log"
	"testing"
	"time"

	guuid "github.com/google/uuid"
	"github.com/kjk/betterguid"
	"github.com/mwyvr/kid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

// kid ids incorporate a timestamp in milliseconds + sequence + a 2-byte random value supplied by crypto/rand
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

// uuid ids incorporate crypto/rand generated numbers
var resultUUID guuid.UUID

func BenchmarkGoogleUuid(b *testing.B) {
	var r guuid.UUID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// https://pkg.go.dev/github.com/google/UUID#NewRandom
			// uuid v4, equiv to NewRandom()
			r = guuid.New()
		}
		resultUUID = r
	})
}

// uuid V7 ids are k-sortable
var resultUUIDV7 guuid.UUID

func BenchmarkGoogleUuidV7(b *testing.B) {
	var r guuid.UUID
	var err error
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, err = guuid.NewV7()
			if err != nil {
				log.Fatal(err)
			}
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

// https://github.com/kjk/betterguid
// like rs/xid, uses a monotonically incrementing counter rather than
// true randomness
var resultBGUID string

func BenchmarkBetterguid(b *testing.B) {
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = betterguid.New()
		}
		resultBGUID = r
	})
}
