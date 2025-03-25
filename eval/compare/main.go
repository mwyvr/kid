// Package main produces for comparison purposes a markdown formatted table
// illustrating key differences between a number of unique ID packages.
package main

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/chilts/sid"
	"github.com/google/uuid"
	"github.com/kjk/betterguid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mwyvr/kid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"
)

type pkg struct {
	name       string
	blen       int
	elen       int
	ksortable  bool
	sample     string
	next       string
	next2      string
	next3      string
	uniq       string
	components string
}

func main() {
	packages := []pkg{
		{
			"[mwyvr/kid](https://github.com/mwyvr/kid)",
			len(kid.New().Bytes()),
			len(kid.New().String()),
			true,
			kid.New().String(),
			kid.New().String(),
			kid.New().String(),
			kid.New().String(),
			"crypto/rand",
			"6 byte ts(millisecond) : 2 byte sequence : 2 byte random",
		},
		{
			"[rs/xid](https://github.com/rs/xid)",
			len(xid.New().Bytes()),
			len(xid.New().String()),
			true,
			xid.New().String(),
			xid.New().String(),
			xid.New().String(),
			xid.New().String(),
			"globally",
			"4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter",
		},
		{
			"[segmentio/ksuid](https://github.com/segmentio/ksuid)",
			len(ksuid.New().Bytes()),
			len(ksuid.New().String()),
			true,
			ksuid.New().String(),
			ksuid.New().String(),
			ksuid.New().String(),
			ksuid.New().String(),
			"math/rand",
			"4 byte ts(sec) : 16 byte random",
		},
		{
			"[google/uuid](https://github.com/google/uuid) V4",
			len(uuid.New()),
			len(uuid.New().String()),
			false,
			uuid.New().String(),
			uuid.New().String(),
			uuid.New().String(),
			uuid.New().String(),
			"crypt/rand UUID",
			"v4: 16 bytes random with version & variant embedded",
		},
		{
			"[google/uuid](https://github.com/google/uuid) V7",
			len(newUUIDV7()),
			len(newUUIDV7().String()),
			true,
			newUUIDV7().String(),
			newUUIDV7().String(),
			newUUIDV7().String(),
			newUUIDV7().String(),
			"crypt/rand UUID",
			"v7: 16 bytes : 8 bytes time+sequence, version/variant, random",
		},
		{
			"[chilts/sid](https://github.com/chilts/sid)",
			16,
			len(sid.IdBase64()),
			true,
			sid.IdBase64(),
			sid.IdBase64(),
			sid.IdBase64(),
			sid.IdBase64(),
			"math/rand",
			"8 byte ts(nanosecond) 8 byte random",
		},
		{
			"[matoous/go-nanoid/v2](https://github.com/matoous/go-nanoid/)",
			21,
			len(newNanoID()),
			false,
			newNanoID(),
			newNanoID(),
			newNanoID(),
			newNanoID(),
			"crypto/rand",
			"21 byte rand (adjustable)",
		},
		{
			"[sony/sonyflake](https://github.com/sony/sonyflake)",
			16,
			len(newSonyFlake()),
			true,
			newSonyFlake(),
			newSonyFlake(),
			newSonyFlake(),
			newSonyFlake(),
			"ts+counter",
			"39 bit ts(10msec) 8 bit seq, 16 bit mach id",
		},
		{
			"[oklog/ulid](https://github.com/oklog/ulid)",
			len(newUlid()),
			len(newUlid().String()),
			true,
			newUlid().String(),
			newUlid().String(),
			newUlid().String(),
			newUlid().String(),
			"user-definable, crypt/rand",
			"6 byte ts(ms) : 10 byte counter random init per ts(ms)",
		},
		{
			"[kjk/betterguid](https://github.com/kjk/betterguid)",
			8 + 9, // only available as a string
			len(betterguid.New()),
			true,
			betterguid.New(),
			betterguid.New(),
			betterguid.New(),
			betterguid.New(),
			"counter",
			"8 byte ts(ms) : 9 byte counter random init per ts(ms)",
		},
	}

	fmt.Printf("| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Method | Components |\n")
	fmt.Printf("|-----------------------------------------------------------|----|----|-------|---------------------|--------|------------|\n")

	for _, v := range packages {
		fmt.Printf("| %-57s | %d | %d | %5v | `%s`<br>`%s`<br>`%s`<br>`%s`  | %s | %s |\n",
			v.name, v.blen, v.elen, v.ksortable, v.sample, v.next, v.next2, v.next3, v.uniq, v.components)
	}
}

// ulid is configured here to be similar (crypto/rand component) to kid
func newUlid() ulid.ULID {
	return ulid.MustNew(ulid.Timestamp(time.Now().UTC()), rand.Reader)
}

func newUUIDV7() uuid.UUID {
	r, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return r
}

// SonyFlake doesn't provide encoding
var (
	sonygen       = sonyflake.NewSonyflake(sonyflake.Settings{})
	base32Encoder = base32.StdEncoding.WithPadding(base32.NoPadding)
)

func newSonyFlake() string {
	if sonygen == nil {
		panic("could not initialize SonyFlake")
	}
	id, err := sonygen.NextID()
	if err != nil {
		panic(err)
	}
	return base32Encoder.EncodeToString([]byte(fmt.Sprintf("%v", id)))
}

func newNanoID() string {
	id, err := gonanoid.New()
	if err != nil {
		panic(err)
	}
	return id
}
