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
	seq     uint16
	random  uint32
	time    string
	valid   bool
}

var tests = []test{
	// sorted (ascending) should be IDs 2, 3, 0, 5, 4, 1 and then the rest.
	{
		ID{0x0, 0xdc, 0x6a, 0xcf, 0xab, 0xff, 0x0, 0x0, 0x0, 0x0},
		"03f6nlxczw000000",
		946684799999,
		0,
		0,
		"1999-12-31 23:59:59.999 +0000 UTC",
		true,
	},
	{
		// zzzzzzzzzzzzzzzz ts:281474976710655 seq:4095 rnd:1048575 10889-08-02 05:31:50.655 +0000 UTC ID{ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff }
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		"zzzzzzzzzzzzzzzz",
		281474976710655,
		4095,
		1048575,
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
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0x9d, 0x3a, 0xb3}, "06bqer9xnm79tfnl", 1741456227757, 233, 866995, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xaa, 0x84, 0x0}, "06bqer9xnm7bn100", 1741456227757, 234, 689152, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xb7, 0xd9, 0x40}, "06bqer9xnm7cgpb0", 1741456227757, 235, 514368, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xc4, 0xdb, 0xb2}, "06bqer9xnm7d9pxk", 1741456227757, 236, 318386, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xd1, 0xd5, 0x4e}, "06bqer9xnm7e3nbf", 1741456227757, 237, 120142, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xe4, 0x19, 0xbb}, "06bqer9xnm7f86ev", 1741456227757, 238, 268731, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xf2, 0xad, 0x75}, "06bqer9xnm7g5ccn", 1741456227757, 239, 175477, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xe, 0xff, 0xc0, 0xb}, "06bqer9xnm7gzh0c", 1741456227757, 239, 1032203, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xf, 0xd, 0xca, 0x3b}, "06bqer9xnm7hvkjv", 1741456227757, 240, 903739, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xf, 0x21, 0x70, 0x79}, "06bqer9xnm7k2w3s", 1741456227757, 242, 94329, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xad, 0xf, 0x3b, 0xac, 0xdb}, "06bqer9xnm7lqc6v", 1741456227757, 243, 765147, "2025-03-08 17:50:27.757 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x12, 0x41, 0x49}, "06bqer9xnr014hb9", 1741456227758, 1, 147785, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x20, 0x75, 0x9b}, "06bqer9xnr020xdv", 1741456227758, 2, 30107, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x2d, 0x8d, 0x95}, "06bqer9xnr02v3dn", 1741456227758, 2, 888213, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x3b, 0xd3, 0xf7}, "06bqer9xnr03qmzq", 1741456227758, 3, 775159, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x48, 0xa4, 0xef}, "06bqer9xnr04j97g", 1741456227758, 4, 566511, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x55, 0x4f, 0x4f}, "06bqer9xnr05bltg", 1741456227758, 5, 347983, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x63, 0xc6, 0x81}, "06bqer9xnr067jm1", 1741456227758, 6, 247425, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x70, 0xd9, 0x2c}, "06bqer9xnr071p9d", 1741456227758, 7, 55596, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x7d, 0x5d, 0xac}, "06bqer9xnr07tqed", 1741456227758, 7, 875948, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x8b, 0x13, 0xb}, "06bqer9xnr08p4rc", 1741456227758, 8, 725771, "2025-03-08 17:50:27.758 +0000 UTC", true},
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x98, 0x7a, 0xe5}, "06bqer9xnr09hyq5", 1741456227758, 9, 555749, "2025-03-08 17:50:27.758 +0000 UTC", true},
	// invalid encoded values
	{ID{0x1, 0x95, 0x76, 0xe1, 0x3d, 0xae, 0x0, 0x98, 0x7a, 0xe5}, "06BQER9XNR09HYQ5", 1741456227758, 9, 555749, "2025-03-08 17:50:27.758 +0000 UTC", false}, // must be lowercase
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

func TestNewWithTime(t *testing.T) {
	tm := time.Date(2026, 9, 12, 23, 4, 5, 123_456_000, time.UTC)
	id, err := NewWithTime(tm)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := id.Timestamp(), tm.UnixMilli(); got != want {
		t.Errorf("Timestamp = %d, want %d", got, want)
	}
	if got, want := id.Sequence(), uint16((int64(tm.Nanosecond())%nanoPerMilli)>>8); got != want {
		t.Errorf("Sequence = %d, want %d", got, want)
	}
	if got := id.Time().UTC(); !got.Equal(tm.Truncate(time.Millisecond)) {
		t.Errorf("Time = %v, want %v", got, tm.Truncate(time.Millisecond))
	}
	// Round trip through the encoded form.
	if back, err := Parse(id.String()); err != nil || back != id {
		t.Errorf("round trip: %v, %v", back, err)
	}
	// IDs preserve time order.
	if earlier, err := NewWithTime(tm.Add(-time.Hour)); err != nil || earlier.Compare(id) >= 0 {
		t.Errorf("earlier ID does not sort first: %v, %v", earlier, err)
	}
	// A time with no sub-millisecond component gets sequence zero.
	if id, err := NewWithTime(time.UnixMilli(123456789)); err != nil || id.Sequence() != 0 {
		t.Errorf("UnixMilli ID: %v, seq=%d, err=%v", id, id.Sequence(), err)
	}
}

func TestNewWithTimeOutOfRange(t *testing.T) {
	for name, tm := range map[string]time.Time{
		"before epoch": time.Unix(0, 0).Add(-time.Millisecond),
		"past field":   time.UnixMilli(1 << 48),
	} {
		id, err := NewWithTime(tm)
		if !errors.Is(err, ErrTimestampOutOfRange) || id != ZeroID {
			t.Errorf("%s: got %v, %v; want ZeroID, ErrTimestampOutOfRange", name, id, err)
		}
	}
	// The largest representable millisecond is accepted.
	if _, err := NewWithTime(time.UnixMilli(1<<48 - 1)); err != nil {
		t.Errorf("max milli: %v", err)
	}
}

func TestNewUnique(t *testing.T) {
	// Generate N ids, see if all unique
	// Parallel generation test is in ./eval/uniqcheck/main.go
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
		// Each ID must sort strictly after its predecessor. This also
		// implies the embedded timestamp never decreases, since the
		// timestamp occupies the leading bytes. There is deliberately no
		// upper bound on the delta: a forward wall-clock step (e.g. an
		// NTP correction or a CI VM) legitimately jumps the timestamp
		// more than 1000 ms between two consecutive calls. The clock
		// behaviors are pinned deterministically by the stubbed-clock
		// tests TestGetTSClockRegression (backwards step),
		// TestGetTSSequenceBorrow (sequence overflow), and
		// TestGetTSBurstMonotonic (saturated burst).
		if ids[i].Compare(ids[i-1]) <= 0 {
			t.Errorf("ID %d does not sort after its predecessor", i)
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
		if v.valid {
			continue
		}
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			id, err := Parse(v.encoded)
			if err == nil {
				t.Errorf("invalid encoded %v, Parse() should be err", v.encoded)
			}
			if id != ZeroID {
				t.Errorf("invalid encoded %v returned %v, Parse() should return ZeroID", v.encoded, v.id[:])
			}
		})
	}
}

func TestIDComponents(t *testing.T) {
	for i, v := range tests {
		if v.valid {
			t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
				if got, want := v.id.Time().String(), v.time; got != want {
					t.Errorf("Time() = %v, want %v", got, want)
				}
				if got, want := v.id.Timestamp(), v.ts; got != want {
					t.Errorf("Timestamp() = %v, want %v", got, want)
				}
				if got, want := v.id.Sequence(), v.seq; got != want {
					t.Errorf("Sequence() = %v, want %v", got, want)
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
		lastSeq uint16
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
	ZeroIDTime := "1970-01-01 00:00:00 +0000 UTC"
	if ZeroID.Time().String() != ZeroIDTime {
		t.Errorf("got: %s, want:%s", ZeroID.Time(), ZeroIDTime)
	}
	// zero-valued ID (all bytes zero) must produce the same time
	zero := ID{}
	if zero.Time().String() != ZeroIDTime {
		t.Errorf("zero ID Time() = %s, want %s", zero.Time().String(), ZeroIDTime)
	}
}

func TestIDString(t *testing.T) {
	for _, v := range tests {
		if v.valid {
			if got, want := v.encoded, v.id.String(); got != want {
				t.Errorf("String() = %v, want %v", got, want)
			}
		}
	}
}

func TestParse(t *testing.T) {
	// 06bprdfln4x281hd ts:1741276959657 seq:930 rnd:263692 2025-03-06 16:02:39.657 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x31, 0xd3, 0xa9, 0x3a, 0x24,  0x6,  0xc }
	got, err := Parse("06bprdfln4x281hd")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x31, 0xd3, 0xa9, 0x3a, 0x24, 0x6, 0xc}
	if got != want {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
	// nil ID
	got, err = Parse("0000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	if got != want {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
	// max ID
	got, err = Parse("zzzzzzzzzzzzzzzz")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if got != want {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("012345")
	if err != ErrInvalidID {
		t.Errorf("Parse(invalid length) err=%v, want %v", err, ErrInvalidID)
	}
	id, err := Parse("062ez870acdtzd2y3qajilou") // i, l, o, u never in our IDs
	if err != ErrInvalidID {
		t.Errorf("Parse(062ez870acdtzd2y3qajilou - invalid chars) err=%v, want %v", err, ErrInvalidID)
	}
	if id != ZeroID {
		t.Errorf("Parse() = %v, want %v", id, ZeroID)
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
			"valid_zero", "0000000000000000", ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, false,
		},
		{ // zzzzzzzzzzzzzzzz ts:281474976710655 seq:4095 rnd:1048575 10889-08-02 05:31:50.655 +0000 UTC ID{ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff }
			"valid_max", "zzzzzzzzzzzzzzzz", ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, false,
		},
		{"invalid chars", "000000000000000u", ZeroID, true},
		{"invalid length too long", "12345678901", ZeroID, true},
		{"invalid length too short", "dfb7emm", ZeroID, true},
		{ // 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
			"valid id", "06bprg666xzm7hpg", ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// pre-fill so the error path's reset-to-ZeroID is actually exercised
			id := ID{0xde, 0xca, 0xfb, 0xad, 0xde, 0xca, 0xfb, 0xad, 0xde, 0xca}
			err := id.UnmarshalText([]byte(tt.encoded))
			if (err != nil) != tt.wantErr {
				t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				// on error, id must be reset to the zero ID
				if id != ZeroID {
					t.Errorf("ID.UnmarshalText(%s) got: %v, want ZeroID %v", tt.encoded, id, ZeroID)
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
		t.Errorf("ID.UnmarshalText(%q) err = %v, want %v", "decafebad", err, ErrInvalidID)
	}
	if id != ZeroID {
		t.Errorf("ID.UnmarshalText(%q) = %v, want %v", "decafebad", id, ZeroID)
	}
}

func TestIDMarshalText(t *testing.T) {
	id := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	b, err := id.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v, want nil", err)
	}
	if got, want := string(b), "06bprg666xzm7hpg"; got != want {
		t.Errorf("MarshalText() = %s, want %s", got, want)
	}
	// nil ID
	b, err = ZeroID.MarshalText()
	if err != nil {
		t.Fatalf("ZeroID.MarshalText() error = %v, want nil", err)
	}
	if got, want := string(b), "0000000000000000"; got != want {
		t.Errorf("ZeroID.MarshalText() = %s, want %s", got, want)
	}
}

func TestIDAppendText(t *testing.T) {
	id := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	// a non-empty, non-zero-capacity prefix proves this appends rather
	// than overwrites or ignores b.
	prefix := []byte("id=")
	b, err := id.AppendText(prefix)
	if err != nil {
		t.Fatalf("AppendText() error = %v, want nil", err)
	}
	if got, want := string(b), "id=06bprg666xzm7hpg"; got != want {
		t.Errorf("AppendText() = %s, want %s", got, want)
	}
	// the original prefix slice's contents must be untouched
	if got, want := string(prefix), "id="; got != want {
		t.Errorf("AppendText() mutated its argument: prefix = %s, want %s", got, want)
	}
	// nil b, nil ID
	b, err = ZeroID.AppendText(nil)
	if err != nil {
		t.Fatalf("ZeroID.AppendText(nil) error = %v, want nil", err)
	}
	if got, want := string(b), "0000000000000000"; got != want {
		t.Errorf("ZeroID.AppendText(nil) = %s, want %s", got, want)
	}
	// must agree with MarshalText for the same ID
	want, _ := id.MarshalText()
	got, _ := id.AppendText(nil)
	if !bytes.Equal(got, want) {
		t.Errorf("AppendText(nil) = %s, want %s (MarshalText)", got, want)
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
	if !bytes.Equal(got[:], ZeroID[:]) {
		t.Error("FromBytes([]byte{0x1, 0x2}) - invalid - != ZeroID")
	}
	if err == nil {
		t.Fatal(err)
	}
}

func TestIDMarshalBinary(t *testing.T) {
	id := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	b, err := id.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v, want nil", err)
	}
	if !bytes.Equal(b, id[:]) {
		t.Errorf("MarshalBinary() = %v, want %v", b, id[:])
	}
	// must be a copy: mutating the result must not alter id
	b[0] = 0xff
	if id[0] == 0xff {
		t.Error("MarshalBinary() did not return a copy")
	}
	// unlike Value/ValueBinary, MarshalBinary has no ZeroID special case:
	// it always returns the 10 raw bytes, even all-zero ones.
	b, err = ZeroID.MarshalBinary()
	if err != nil {
		t.Fatalf("ZeroID.MarshalBinary() error = %v, want nil", err)
	}
	if !bytes.Equal(b, ZeroID[:]) {
		t.Errorf("ZeroID.MarshalBinary() = %v, want %v", b, ZeroID[:])
	}
}

func TestIDAppendBinary(t *testing.T) {
	id := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	prefix := []byte{0xaa, 0xbb}
	b, err := id.AppendBinary(prefix)
	if err != nil {
		t.Fatalf("AppendBinary() error = %v, want nil", err)
	}
	want := append([]byte{0xaa, 0xbb}, id[:]...)
	if !bytes.Equal(b, want) {
		t.Errorf("AppendBinary() = %v, want %v", b, want)
	}
	// the original prefix slice's contents must be untouched
	if !bytes.Equal(prefix, []byte{0xaa, 0xbb}) {
		t.Errorf("AppendBinary() mutated its argument: prefix = %v", prefix)
	}
	// must agree with MarshalBinary for the same ID
	wantMB, _ := id.MarshalBinary()
	gotAB, _ := id.AppendBinary(nil)
	if !bytes.Equal(gotAB, wantMB) {
		t.Errorf("AppendBinary(nil) = %v, want %v (MarshalBinary)", gotAB, wantMB)
	}
}

func TestIDUnmarshalBinary(t *testing.T) {
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	data, err := want.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var got ID
	if err := got.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("UnmarshalBinary() = %v, want %v", got, want)
	}
	// invalid length resets to ZeroID and returns ErrInvalidID
	got = want // pre-fill with a non-zero value so the reset is exercised
	if err := got.UnmarshalBinary([]byte{0x1, 0x2}); err != ErrInvalidID {
		t.Errorf("UnmarshalBinary(short) error = %v, want %v", err, ErrInvalidID)
	}
	if got != ZeroID {
		t.Errorf("UnmarshalBinary(short) left id = %v, want ZeroID", got)
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
	if id == ZeroID && !reflect.DeepEqual(string(got), "null") {
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
	if err := id.UnmarshalJSON([]byte("null")); err != nil || id != ZeroID {
		t.Errorf("id.UnmarshalJSON(\"null\") returns %v, %v, want ZeroID, nil", id, err)
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
	// decoding is case-sensitive: an uppercase spelling of an otherwise
	// valid id is rejected:
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
	got, err = ZeroID.Value()
	if got != nil || err != nil {
		t.Errorf("ZeroID.Value() should return nil, nil, got: %v, %v", got, err)
	}
	got, err = id.ValueBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.([]byte), id[:]) {
		t.Errorf("ValueBinary() = %v, want %v", got, id[:])
	}
	got, err = ZeroID.ValueBinary()
	if got != nil || err != nil {
		t.Errorf("ZeroID.ValueBinary() should return nil, nil, got: %v, %v", got, err)
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
	if err != nil || id != ZeroID {
		t.Errorf("ZeroID.Scan(\"\") should return nil err, ZeroID. got: %v %v", err, id)
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
	if id != ZeroID {
		t.Errorf("Scan() id=%v, want %v", id, ZeroID)
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
	if id != ZeroID {
		t.Errorf("UnmarshalJSON(number) id=%v, want ZeroID", id)
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
	for b.Loop() {
		r = New()
	}
	benchResultID = r
}

// common use case, generate an ID, encode as a string:
func BenchmarkNewString(b *testing.B) {
	var r string
	for b.Loop() {
		r = New().String()
	}
	benchResultString = r
}

// encoding performance only
func BenchmarkString(b *testing.B) {
	id := New()
	var r string
	for b.Loop() {
		r = id.String()
	}
	benchResultString = r
}

// decoding performance only
func BenchmarkParse(b *testing.B) {
	var r ID
	str := "06bprlcm7q4z16vh"
	for b.Loop() {
		r, _ = Parse(str)
	}
	benchResultID = r
}

// JSON decoding performance, the path used when scanning IDs out of a
// JSON document
func BenchmarkUnmarshalJSON(b *testing.B) {
	var r ID
	payload := []byte(`"06bprlcm7q4z16vh"`)
	for b.Loop() {
		_ = r.UnmarshalJSON(payload)
	}
	benchResultID = r
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

	prev := uint64(0)
	for i := range 10000 {
		m, s := getTS()
		if s > 0xfff {
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
		if all[i-1].Timestamp() == all[i].Timestamp() && all[i-1].Sequence() == all[i].Sequence() {
			t.Fatalf("duplicate ts+seq across goroutines: %v / %v", all[i-1], all[i])
		}
	}
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

func ExampleParse() {
	id, err := Parse("03f6nlxczw0018fz")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Timestamp(), id.Random())
	// Output: 946684799999 41439
}
