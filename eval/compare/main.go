// Package main produces for comparison purposes a markdown formatted table
// illustrating key differences between a number of unique ID packages.
package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"time"
	"uuid"

	"github.com/chilts/sid"
	idgen "github.com/devjefster/GoShortUniqueID/idgen"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"

	"github.com/mwyvr/kid"
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
			"unique (ts(ms) + sequence) + math/rand/v2",
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
			"ts(sec) + machineID + pid + counter",
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
			"ts + crypto/rand",
			"4 byte ts(sec) : 16 byte random",
		},
		{
			"[uuid](https://pkg.go.dev/uuid) (Go stdlib) V4",
			len(uuid.NewV4()),
			len(uuid.NewV4().String()),
			false,
			uuid.NewV4().String(),
			uuid.NewV4().String(),
			uuid.NewV4().String(),
			uuid.NewV4().String(),
			"crypto/rand",
			"v4: 122 bits random; 6 bits embedding version & variant",
		},
		{
			"[uuid](https://pkg.go.dev/uuid) (Go stdlib) V7",
			len(uuid.NewV7()),
			len(uuid.NewV7().String()),
			true,
			uuid.NewV7().String(),
			uuid.NewV7().String(),
			uuid.NewV7().String(),
			uuid.NewV7().String(),
			"ts(ms) + crypto/rand",
			"v7: 16 bytes : 48 bits time, 12 bits sequence, 6 bits version/variant, 62 bits random",
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
			"ts + math/rand",
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
			"ts + crypto/rand",
			"21 byte rand (adjustable)",
		},
		{
			"[sony/sonyflake](https://github.com/sony/sonyflake)",
			8,
			len(newSonyFlake()),
			true,
			newSonyFlake(),
			newSonyFlake(),
			newSonyFlake(),
			newSonyFlake(),
			"ts + sequence + machine id",
			"39 bit ts(10ms) : 8 bit seq : 16 bit mach id",
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
			"ts + user-definable rand src",
			"6 byte ts(ms) : 10 byte monotonic counter random init per ts(ms)",
		},
		{
			"[devjefster/GoShortUniqueID](https://github.com/devjefster/GoShortUniqueID)",
			6 + 6 + 2, // only available as a string
			len(newGSUID()),
			false,
			newGSUID(),
			newGSUID(),
			newGSUID(),
			newGSUID(),
			"ts + math/rand + counter",
			"6 byte ts(second) : 6 base62 random : 2 byte counter (mod 10000)",
		},
	}

	fmt.Printf("| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Unique | Components |\n")
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

var (
	sonygen       = newSonygen()
	base32Encoder = base32.StdEncoding.WithPadding(base32.NoPadding)
)

func newSonygen() *sonyflake.Sonyflake {
	fl, err := sonyflake.New(sonyflake.Settings{})
	if err != nil {
		panic(err)
	}
	return fl
}

// SonyFlake has no built-in string encoding; encode the 8-byte big-endian
// form with base32, no padding. The base32 alphabet is ASCII-ascending and
// the time bits lead, so the encoding preserves ID order.
func newSonyFlake() string {
	id, err := sonygen.NextID()
	if err != nil {
		panic(err)
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	return base32Encoder.EncodeToString(b[:])
}

func newGSUID() string {
	return gsuidGen.Generate()
}

var gsuidGen = idgen.New(0, "", "")

func newNanoID() string {
	id, err := gonanoid.New()
	if err != nil {
		panic(err)
	}
	return id
}
