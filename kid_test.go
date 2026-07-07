package kid

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type test struct {
	id      ID
	encoded string
	ts      int64
	seq     int32
	random  int32
	time    string
	iskid   bool
}

var tests = []test{
	// sorted (ascending) should be IDs 2, 3, 0, 5, 4, 1 and then the rest.
	{
		// 03f6nlxczw000000 ts:946684799999 seq:   0 rnd:    0 1999-12-31 23:59:59.999 +0000 UTC ID{  0x0, 0xdc, 0x6a, 0xcf, 0xab, 0xff,  0x0,  0x0,  0x0,  0x0 }
		ID{0x0, 0xdc, 0x6a, 0xcf, 0xab, 0xff, 0x0, 0x0, 0x0, 0x0},
		"03f6nlxczw000000",
		946684799999,
		0,
		0,
		"1999-12-31 23:59:59.999 +0000 UTC",
		true,
	},
	{
		// zzzzzzzzzzzzzzzz ts:281474976710655 seq:65535 rnd:65535 10889-08-02 05:31:50.655 +0000 UTC ID{ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff }
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		"zzzzzzzzzzzzzzzz",
		281474976710655,
		65535,
		65535,
		"10889-08-02 05:31:50.655 +0000 UTC",
		true,
	},
	{
		// 0000000000000000 ts:0 seq:   0 rnd:    0 1970-01-01 00:00:00 +0000 UTC ID{  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0 }
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		"0000000000000000",
		0,
		0,
		0,
		"1970-01-01 00:00:00 +0000 UTC",
		true,
	},
	{
		// 02j4he6ek8000t4f ts:696996122002 seq:   0 rnd:26766 1992-02-02 02:02:02.002 +0000 UTC ID{  0x0, 0xa2, 0x48, 0x34, 0xcd, 0x92,  0x0,  0x0, 0x68, 0x8e }
		ID{0x0, 0xa2, 0x48, 0x34, 0xcd, 0x92, 0x0, 0x0, 0x68, 0x8e},
		"02j4he6ek8000t4f",
		696996122002,
		0,
		26766,
		"1992-02-02 02:02:02.002 +0000 UTC",
		true,
	},
	{
		// 06bpkb8pz0000000 ts:1741226055416 seq:   0 rnd:    0 2025-03-05 17:54:15.416 -0800 PST ID{  0x1, 0x95, 0x69, 0x29, 0x16, 0xf8,  0x0,  0x0,  0x0,  0x0 }
		ID{0x1, 0x95, 0x69, 0x29, 0x16, 0xf8, 0x0, 0x0, 0x0, 0x0},
		"06bpkb8pz0000000",
		1741226055416,
		0,
		0,
		"2025-03-06 01:54:15.416 +0000 UTC",
		true,
	},
	{
		// 05z169vrs40006zf ts:1640998861001 seq:   0 rnd: 7150 2022-01-01 01:01:01.001 +0000 UTC ID{  0x1, 0x7e, 0x13, 0x27, 0x78, 0xc9,  0x0,  0x0, 0x1b, 0xee }
		ID{0x1, 0x7e, 0x13, 0x27, 0x78, 0xc9, 0x0, 0x0, 0x1b, 0xee},
		"05z169vrs40006zf",
		1640998861001,
		0,
		7150,
		"2022-01-01 01:01:01.001 +0000 UTC",
		true,
	},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0x9d, 0x3a, 0xb3}, "06bqer9xnm79tfnl", 1741456227757, 3741, 15027, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xaa, 0x84, 0x0}, "06bqer9xnm7bn100", 1741456227757, 3754, 33792, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xb7, 0xd9, 0x40}, "06bqer9xnm7cgpb0", 1741456227757, 3767, 55616, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xc4, 0xdb, 0xb2}, "06bqer9xnm7d9pxk", 1741456227757, 3780, 56242, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xd1, 0xd5, 0x4e}, "06bqer9xnm7e3nbf", 1741456227757, 3793, 54606, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xe4, 0x19, 0xbb}, "06bqer9xnm7f86ev", 1741456227757, 3812, 6587, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xf2, 0xad, 0x75}, "06bqer9xnm7g5ccn", 1741456227757, 3826, 44405, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xff, 0xc0, 0xb}, "06bqer9xnm7gzh0c", 1741456227757, 3839, 49163, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xf, 0xd, 0xca, 0x3b}, "06bqer9xnm7hvkjv", 1741456227757, 3853, 51771, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xf, 0x21, 0x70, 0x79}, "06bqer9xnm7k2w3s", 1741456227757, 3873, 28793, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xf, 0x3b, 0xac, 0xdb}, "06bqer9xnm7lqc6v", 1741456227757, 3899, 44251, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x12, 0x41, 0x49}, "06bqer9xnr014hb9", 1741456227758, 18, 16713, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x20, 0x75, 0x9b}, "06bqer9xnr020xdv", 1741456227758, 32, 30107, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x2d, 0x8d, 0x95}, "06bqer9xnr02v3dn", 1741456227758, 45, 36245, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x3b, 0xd3, 0xf7}, "06bqer9xnr03qmzq", 1741456227758, 59, 54263, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x48, 0xa4, 0xef}, "06bqer9xnr04j97g", 1741456227758, 72, 42223, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x55, 0x4f, 0x4f}, "06bqer9xnr05bltg", 1741456227758, 85, 20303, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x63, 0xc6, 0x81}, "06bqer9xnr067jm1", 1741456227758, 99, 50817, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x70, 0xd9, 0x2c}, "06bqer9xnr071p9d", 1741456227758, 112, 55596, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x7d, 0x5d, 0xac}, "06bqer9xnr07tqed", 1741456227758, 125, 23980, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x8b, 0x13, 0xb}, "06bqer9xnr08p4rc", 1741456227758, 139, 4875, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x98, 0x7a, 0xe5}, "06bqer9xnr09hyq5", 1741456227758, 152, 31461, "2025-03-08 17:50:27.758 +0000 UTC", true},
	// invalid encoded values
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x98, 0x7a, 0xe5}, "06BQER9XNR09HYQ5", 1741456227758, 152, 31461, "2025-03-08 17:50:27.758 +0000 UTC", false}, // must be lowercase
	{ID{}, "o6bqer9xnr09hyq5", 0, 0, 0, "", false}, // "o" is not a valid character in encoding
	{ID{}, "06bqer9", 0, 0, 0, "", false},          // invalid length
}

func TestNew(t *testing.T) {
	var id ID
	if !id.IsNil() {
		t.Errorf("id is NOT nil")
	}
	id = New()
	if id.IsNil() {
		t.Errorf("id is nil")
	}
}

func TestNewUnique(t *testing.T) {
	// Generate N ids, see if all unique
	// Parallel generation test is in ./cmd/eval/uniqcheck/main.go
	count := 100000
	ids := make([]ID, count)
	seen := make(map[ID]struct{}, count)
	for i := range count {
		ids[i] = New()
		if _, dup := seen[ids[i]]; dup {
			t.Fatalf("generated ID is not unique (%d) %v", i, ids[i])
		}
		seen[ids[i]] = struct{}{}
	}
	for i := 1; i < count; i++ {
		// Each ID must sort strictly after its predecessor
		if ids[i].Compare(ids[i-1]) <= 0 {
			t.Errorf("ID %d does not sort after its predecessor", i)
		}
		// Check that timestamp was incremented and is within 1000 milliseconds of the previous one
		milli := ids[i].Time().Sub(ids[i-1].Time()).Milliseconds()
		if milli < 0 || milli > 1000 {
			t.Error("wrong timestamp in generated ID")
		}
	}
}

func TestID_IsNil(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want bool
	}{
		{name: "ID not nil", id: New(), want: false},
		{name: "Nil ID", id: ID{}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.id.IsNil(), tt.want; got != want {
				t.Errorf("IsNil() = %v, want %v", got, want)
			}
		})
	}
}

func TestID_IsZero(t *testing.T) {
	id := ID{}
	if !id.IsZero() {
		t.Errorf("ID.IsZero() = %v, want %v", id.IsZero(), true)
	}
}

func TestInvalid(t *testing.T) {
	for i, v := range tests {
		if v.iskid {
			continue
		}
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			id, err := FromString(v.encoded)
			if err == nil {
				t.Errorf("invalid encoded %v, FromString() should be err", v.encoded)
			}
			if id != nilID {
				t.Errorf("invalid encoded %v returned %v, FromString() should return nilID", v.encoded, v.id[:])
			}
		})
	}
}

func TestIDComponents(t *testing.T) {
	for i, v := range tests {
		if v.iskid {
			t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
				if got, want := v.id.Time().String(), v.time; got != want {
					t.Errorf("Time() = %v, want %v", got, want)
				}
				if got, want := v.id.Timestamp(), v.ts; got != want {
					t.Errorf("Timestamp() = %v, want %v", got, want)
				}
				if got, want := v.id.Random(), v.random; got != want {
					t.Errorf("Random() = %v, want %v", got, want)
				}
			})
		}
	}
}

// ensure sequencing produces unique ts+seq combos
func TestSequence(t *testing.T) {
	var (
		lastTS  int64
		lastSeq int32
	)
	// Generate 1,000,000 new IDs
	check := []ID{}
	for range 1000000 {
		check = append(check, New())
	}
	for _, id := range check {
		if lastTS != id.Timestamp() {
			lastTS = id.Timestamp()
			lastSeq = id.Sequence()
			continue
		}
		if id.Timestamp() == lastTS && id.Sequence() <= lastSeq {
			t.Errorf("sequence not unique for next ID ts: %d seq: %d last: %d", id.Timestamp(), id.Sequence(), lastTS)
		} else {
			lastSeq = id.Sequence()
		}
	}
}

func TestIDTime(t *testing.T) {
	nilTime := "1970-01-01 00:00:00 +0000 UTC"
	if nilID.Time().String() != nilTime {
		t.Errorf("got: %s, want:%s", nilID.Time(), nilTime)
	}
}

func TestIDString(t *testing.T) {
	for _, v := range tests {
		if v.iskid {
			if got, want := v.encoded, v.id.String(); got != want {
				t.Errorf("String() = %v, want %v", got, want)
			}
		}
	}
}

func TestIDEncode(t *testing.T) {
	id := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	text := make([]byte, encodedLen)
	if got, want := string(id.Encode(text)), "dfp7emzzzzy30ey2"; got != want {
		t.Errorf("Encode() = %v, want %v", got, want)
	}
}

func TestFromString(t *testing.T) {
	// 06bprdfln4x281hd ts:1741276959657 seq:14884 rnd: 1548 2025-03-06 16:02:39.657 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x31, 0xd3, 0xa9, 0x3a, 0x24,  0x6,  0xc }
	got, err := FromString("06bprdfln4x281hd")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x31, 0xd3, 0xa9, 0x3a, 0x24, 0x6, 0xc}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// nil ID
	got, err = FromString("0000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// max ID
	got, err = FromString("zzzzzzzzzzzzzzzz")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
}

func TestFromStringInvalid(t *testing.T) {
	_, err := FromString("012345")
	if err != ErrInvalidID {
		t.Errorf("FromString(invalid length) err=%v, want %v", err, ErrInvalidID)
	}
	id, err := FromString("062ez870acdtzd2y3qajilou") // i, l, o, u never in our IDs
	if err != ErrInvalidID {
		t.Errorf("FromString(062ez870acdtzd2y3qajilou - invalid chars) err=%v, want %v", err, ErrInvalidID)
	}
	if id != nilID {
		t.Errorf("FromString() =%v, there want %v", id, nilID)
	}
}

func TestID_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		id      ID
		wantErr bool
	}{
		{ // 0000000000000000 ts:0 seq:   0 rnd:    0 1970-01-01 00:00:00 +0000 UTC ID{  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0 }
			"valid", "0000000000000000", ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, false,
		},
		{ // 0000000000000000 ts:0 seq:   0 rnd:    0 1970-01-01 00:00:00 +0000 UTC ID{  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0 }
			"valid", "zzzzzzzzzzzzzzzz", ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, false,
		},
		{ // zzzzzzzzzzzzzzzz ts:281474976710655 seq:65535 rnd:65535 10889-08-02 05:31:50.655 +0000 UTC ID{ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff }
			"valid", "zzzzzzzzzzzzzzzz", ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, false,
		},
		{"invalid chars", "000000000000000u", nilID, true},
		{"invalid length too long", "12345678901", nilID, true},
		{"invalid length too short", "dfb7emm", nilID, true},
		{ // 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
			"valid id", "06bprg666xzm7hpg", ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// pre-fill so the error path's reset-to-nilID is actually exercised
			id := ID{0xde, 0xca, 0xfb, 0xad, 0xde, 0xca, 0xfb, 0xad, 0xde, 0xca}
			err := id.UnmarshalText([]byte(tt.encoded))
			if (err != nil) != tt.wantErr {
				t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				// on error, id must be reset to the nil ID
				if id != nilID {
					t.Errorf("ID.UnmarshalText(%s) got: %v, want nilID %v", tt.encoded, id, nilID)
				}
				return
			}
			// the decoded value must equal the expected ID, and roundtrip
			if id != tt.id {
				t.Errorf("ID.UnmarshalText(%s) decoded: %v, want: %v", tt.encoded, id, tt.id)
			}
			if id.String() != tt.encoded {
				t.Errorf("ID.UnmarshalText() roundtrip got: %v, want: %v", id.String(), tt.encoded)
			}
		})
	}
	id := ID{}
	if err := id.UnmarshalText([]byte("decafebad")); err != ErrInvalidID {
		t.Errorf("ID.UnmarshalText(\"foo\" got: %v, want err", err)
	}
	if err := id.UnmarshalText([]byte("decafebad")); err != nil && !id.IsNil() {
		t.Errorf("ID.UnmarshalText(\"foo\") got: %v, want %v", id, nilID)
	}
}

func TestFromBytes_Invariant(t *testing.T) {
	want := New()
	got, err := FromBytes(want.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got[:], want[:]) {
		t.Error("FromBytes(id.Bytes()) != id")
	}
	// invalid
	got, err = FromBytes([]byte{0x1, 0x2})
	if !bytes.Equal(got[:], nilID[:]) {
		t.Error("FromBytes([]byte{0x1, 0x2}) - invalid - != nilID")
	}
	if err == nil {
		t.Fatal(err)
	}
}

type jsonType struct {
	ID  *ID
	Str string
}

func TestIDMarshalJSON(t *testing.T) {
	id := ID{}
	got, err := id.MarshalJSON()
	if err != nil {
		t.Error("id.MarshalJSON()", err)
	}
	if id == nilID && !reflect.DeepEqual(string(got), "null") {
		t.Errorf("got: %v, want: \"null\"", string(got))
	}
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	id = ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	v := jsonType{ID: &id, Str: "valid"}
	data, err := json.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"ID":"06bprg666xzm7hpg","Str":"valid"}`; got != want {
		t.Errorf("json.Marshal() = %v, want %v", got, want)
	}
}

func TestIDUnmarshalJSON(t *testing.T) {
	id := ID{}
	if err := id.UnmarshalJSON([]byte("null")); err != nil || id != nilID {
		t.Errorf("id.UnmarshalJSON(\"null\") returns %v, %v, want nilID, nil", id, err)
	}
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	data := []byte(`{"ID":"06bprg666xzm7hpg","Str":"valid"}`)
	v := jsonType{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	if got := *v.ID; !bytes.Equal(got[:], want[:]) {
		t.Errorf("json.Unmarshal() = %v, want %v", got, want)
	}
}

func TestIDUnmarshalJSON_Error(t *testing.T) {
	v := jsonType{}
	// callers are responsible for forcing lower case input for Base32
	// otherwise valid id:
	err := json.Unmarshal([]byte(`{"ID":"06BPRG666XZM7HPG"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// too short
	err = json.Unmarshal([]byte(`{"ID":"06bprg666xzm"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// no 'a' in character set
	err = json.Unmarshal([]byte(`{"ID":"0000000000000a"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// invalid on multiple levels
	err = json.Unmarshal([]byte(`{"ID":1}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
}

func TestIDDriverValue(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	id := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	got, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	if want := "06bprg666xzm7hpg"; got != want {
		t.Errorf("Value() = %v, want %v", got, want)
	}
	got, err = nilID.Value()
	if got != nil && err != nil {
		t.Errorf("nilID.Value() should return nil, nil, got: %v, %v", got, err)
	}
}

func TestIDDriverScan(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	id := ID{}
	err := id.Scan("06bprg666xzm7hpg")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	if !bytes.Equal(id[:], want[:]) {
		t.Errorf("Scan() = %v, want %v", id, want)
	}
	id = ID{}
	err = id.Scan(nil)
	if err != nil || id != nilID {
		t.Errorf("nilID.Scan(\"\") should return nil err, nilID. got: %v %v", err, id)
	}
}

func TestIDDriverScanError(t *testing.T) {
	id := ID{}

	if got, want := id.Scan(0), errors.New("kid: scanning unsupported type: int"); got.Error() != want.Error() {
		t.Errorf("Scan() err=%v, want %v", got, want)
	}
	if got, want := id.Scan("0"), ErrInvalidID; got != want {
		t.Errorf("Scan() err=%v, want %v", got, want)
	}
	if id != nilID {
		t.Errorf("Scan() id=%v, want %v", id, nilID)
	}
}

func TestIDDriverScanByteFromDatabase(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	got := ID{}
	bs := []byte("06bprg666xzm7hpg")
	err := got.Scan(bs)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	if !bytes.Equal(got[:], want[:]) {
		t.Errorf("Scan() = %v, want %v", got, want)
	}
}

func TestIDDriverScanBinary(t *testing.T) {
	// Scan must also accept the 10-byte binary form, e.g. from a BLOB column
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	got := ID{}
	if err := got.Scan(want.Bytes()); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Scan(binary) = %v, want %v", got, want)
	}
	// a []byte of any other invalid length must fail
	if err := got.Scan([]byte{0x1, 0x2, 0x3}); err != ErrInvalidID {
		t.Errorf("Scan(3 bytes) err=%v, want %v", err, ErrInvalidID)
	}
}

func TestIDUnmarshalJSON_RejectsNonString(t *testing.T) {
	// A bare JSON number of length encodedLen+2 is composed entirely of
	// valid alphabet characters once the delimiters are stripped; without
	// the quote check in UnmarshalJSON it decoded silently.
	v := jsonType{}
	if err := json.Unmarshal([]byte(`{"ID":123456789012345678}`), &v); err != ErrInvalidID {
		t.Errorf("json.Unmarshal(18-digit number) err=%v, want %v", err, ErrInvalidID)
	}
	var id ID
	if err := id.UnmarshalJSON([]byte(`123456789012345678`)); err != ErrInvalidID {
		t.Errorf("UnmarshalJSON(number) err=%v, want %v", err, ErrInvalidID)
	}
	if id != nilID {
		t.Errorf("UnmarshalJSON(number) id=%v, want nilID", id)
	}
	// mismatched/absent quotes of the right total length must also fail
	for _, b := range []string{
		`'06bqer9xnm79tfnl'`,
		`06bqer9xnm79tfnl00`,
		`"06bqer9xnm79tfnl'`,
	} {
		if err := id.UnmarshalJSON([]byte(b)); err != ErrInvalidID {
			t.Errorf("UnmarshalJSON(%s) err=%v, want %v", b, err, ErrInvalidID)
		}
	}
}

func TestFromBytes_InvalidBytes(t *testing.T) {
	cases := []struct {
		length     int
		shouldFail bool
	}{
		{rawLen - 1, true},
		{rawLen, false},
		{rawLen + 1, true},
	}
	for _, c := range cases {
		b := make([]byte, c.length)
		_, err := FromBytes(b)
		if got, want := err != nil, c.shouldFail; got != want {
			t.Errorf("FromBytes() error got %v, want %v", got, want)
		}
	}
}

func TestCompare(t *testing.T) {
	pairs := []struct {
		left     ID
		right    ID
		expected int
	}{
		{tests[1].id, tests[0].id, 1},
		{ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, tests[2].id, 0},
		{tests[0].id, tests[0].id, 0},
		{tests[2].id, tests[1].id, -1},
		{tests[5].id, tests[4].id, -1},
		// identical timestamp+sequence, differing only in the random bytes;
		// Compare considers all 10 bytes and must not report equality
		{
			ID{0x0, 0xa2, 0x48, 0x34, 0xcd, 0x92, 0x0, 0x0, 0x68, 0x8e},
			ID{0x0, 0xa2, 0x48, 0x34, 0xcd, 0x92, 0x0, 0x0, 0x68, 0x8f},
			-1,
		},
	}
	for _, p := range pairs {
		if p.expected != p.left.Compare(p.right) {
			t.Errorf("%s Compare to %s should return %d", p.left, p.right, p.expected)
		}
		if -1*p.expected != p.right.Compare(p.left) {
			t.Errorf("%s Compare to %s should return %d", p.right, p.left, -1*p.expected)
		}
	}
}

var sortTests = []ID{tests[0].id, tests[1].id, tests[2].id, tests[3].id, tests[4].id, tests[5].id}

func TestSort(t *testing.T) {
	ids := make([]ID, 0)
	ids = append(ids, sortTests...)
	Sort(ids)
	// sorted (ascending) should be IDs 2, 3, 0, 5, 4, 1
	if got, want := ids, []ID{sortTests[2], sortTests[3], sortTests[0], sortTests[5], sortTests[4], sortTests[1]}; !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot %v\nwant %v\n", got, want)
	}
}

// Benchmarks
var (
	// avoid compiler over-optimization and silly results
	benchResultID     ID
	benchResultString string
)

// Create new ID
func BenchmarkNew(b *testing.B) {
	var r ID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = New()
		}
		benchResultID = r
	})
}

// common use case, generate an ID, encode as a string:
func BenchmarkNewString(b *testing.B) {
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = New().String()
		}
		benchResultString = r
	})
}

// encoding performance only
func BenchmarkString(b *testing.B) {
	id := New()
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = id.String()
		}
		benchResultString = r
	})
}

// decoding performance only
func BenchmarkFromString(b *testing.B) {
	var r ID
	str := "06bprlcm7q4z16vh"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ = FromString(str)
		}
		benchResultID = r
	})
}

// examples
func ExampleNew() {
	id := New()
	fmt.Printf(`ID:
    String()    %s
    Timestamp() %d
    Sequence()  %d
    Random()    %d
    Time()      %v
    Bytes()     %3v
`, id.String(), id.Timestamp(), id.Sequence(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := FromString("03f6nlxczw0018fz")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Timestamp(), id.Random())
	// Output: 946684799999 41439
}

// resetClock saves and restores the getTS globals so clock-manipulating
// tests leave the package in its original state. Tests using this must not
// run in parallel.
func resetClock(t *testing.T) {
	t.Helper()
	savedNow := timeNow
	savedLast := lastTime.Load()
	t.Cleanup(func() {
		timeNow = savedNow
		lastTime.Store(savedLast)
	})
}

// TestGetTSClockRegression verifies the monotonicity guarantee when the wall
// clock steps backwards (e.g. NTP correction): ts+seq must still increase.
func TestGetTSClockRegression(t *testing.T) {
	resetClock(t)

	base := time.Date(2026, 7, 6, 12, 0, 0, 500_000, time.UTC)
	timeNow = func() time.Time { return base }
	a := New()

	// step the clock back one hour
	timeNow = func() time.Time { return base.Add(-time.Hour) }
	b := New()
	if b.Compare(a) <= 0 {
		t.Errorf("ID generated after clock regression does not sort after predecessor: %v <= %v", b, a)
	}
	if b.Timestamp() < a.Timestamp() {
		t.Errorf("timestamp regressed: %d < %d", b.Timestamp(), a.Timestamp())
	}
}

// TestGetTSSequenceBorrow verifies that sequence overflow within a single
// millisecond carries into the timestamp rather than repeating or wrapping.
func TestGetTSSequenceBorrow(t *testing.T) {
	resetClock(t)

	fixed := time.Date(2026, 7, 6, 12, 0, 0, 250_000, time.UTC)
	timeNow = func() time.Time { return fixed }

	milli0, _ := getTS()
	// force the sequence to its 12-bit maximum for the current millisecond
	lastTime.Store(milli0<<12 | 0xfff)

	milli1, seq1 := getTS()
	if milli1 != milli0+1 || seq1 != 0 {
		t.Errorf("sequence overflow: got milli=%d seq=%d, want milli=%d seq=0", milli1, seq1, milli0+1)
	}
}

// TestGetTSBurstMonotonic verifies strictly increasing ts+seq under a frozen
// clock, where every call takes the catch-up path.
func TestGetTSBurstMonotonic(t *testing.T) {
	resetClock(t)

	fixed := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixed }

	prev := int64(-1)
	for i := range 10000 {
		m, s := getTS()
		if s < 0 || s > 0xfff {
			t.Fatalf("call %d: sequence %d out of 12-bit range", i, s)
		}
		now := m<<12 + s
		if now <= prev {
			t.Fatalf("call %d: ts+seq not strictly increasing (%d <= %d)", i, now, prev)
		}
		prev = now
	}
}

// TestEncodingPreservesOrder verifies the documented k-order property of the
// encoded form: lexicographic order of encoded strings must match byte order
// of the raw IDs (the alphabet is in ascending ASCII order).
func TestEncodingPreservesOrder(t *testing.T) {
	var prev ID
	prevStr := prev.String()
	for range 20000 {
		var id ID
		rand.Read(id[:])
		s := id.String()
		rawCmp := prev.Compare(id)
		strCmp := strings.Compare(prevStr, s)
		if (rawCmp < 0) != (strCmp < 0) || (rawCmp == 0) != (strCmp == 0) {
			t.Fatalf("order mismatch: raw=%d str=%d (%v %s / %v %s)", rawCmp, strCmp, prev, prevStr, id, s)
		}
		prev, prevStr = id, s
	}
}

// TestNewUniqueParallel exercises the lock-free getTS path under concurrent
// load: IDs generated across goroutines must never repeat a ts+seq pair, and
// each goroutine must observe strictly increasing IDs. Run with -race.
func TestNewUniqueParallel(t *testing.T) {
	const goroutines, per = 8, 50000
	results := make([][]ID, goroutines)
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ids := make([]ID, per)
			var prev ID
			for i := range per {
				ids[i] = New()
				if ids[i].Compare(prev) <= 0 {
					t.Errorf("goroutine %d: ID %d does not sort after predecessor", g, i)
					return
				}
				prev = ids[i]
			}
			results[g] = ids
		}()
	}
	wg.Wait()
	all := make([]ID, 0, goroutines*per)
	for _, r := range results {
		all = append(all, r...)
	}
	Sort(all)
	for i := 1; i < len(all); i++ {
		if bytes.Equal(all[i-1][:8], all[i][:8]) {
			t.Fatalf("duplicate ts+seq across goroutines: %v / %v", all[i-1], all[i])
		}
	}
}
