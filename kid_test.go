package kid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
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
	count := 10000
	ids := make([]ID, count)
	for i := range count {
		ids[i] = New()
	}
	for i := 1; i < count; i++ {
		prevID := ids[i-1]
		id := ids[i]
		// Test for uniqueness among all other generated ids
		for j, tid := range ids {
			if j != i {
				// can't use ID.Compare for this test as it compares only the time
				// component of IDs
				if bytes.Equal(id[:], tid[:]) {
					t.Errorf("generated ID is not unique (%d/%d)\n%v", i, j, ids)
				}
			}
		}
		// Check that timestamp was incremented and is within 1000 milliseconds of the previous one
		milli := id.Time().Sub(prevID.Time()).Milliseconds()
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
		tt := tt
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
		if !v.iskid {
			t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
				_, err := FromString(v.encoded)
				if err == nil {
					t.Errorf("invalid encoded %v, FromString() should be err", v.encoded)
				}
			})
		}
		if !v.iskid {
			t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
				id, _ := FromString(v.encoded)
				if id != nilID {
					t.Errorf("invalid encoded %v returned %v, FromString() should return nilID", v.encoded, v.id[:])
				}
			})
		}
	}
}

func TestIDComponents(t *testing.T) {
	for i, v := range tests {
		if v.iskid {
			t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
				if got, want := fmt.Sprintf("%s", v.id.Time()), v.time; got != want {
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
	if fmt.Sprintf("%s", nilID.Time()) != nilTime {
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
			var id ID
			if err := id.UnmarshalText([]byte(tt.encoded)); (err != nil) != tt.wantErr {
				t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if !tt.wantErr && tt.id.String() != tt.encoded {
					t.Errorf("ID.UnmarshalText() got: %v, want encoded: %v", tt.id.String(), tt.encoded)
				}
			}
			if err := id.UnmarshalText([]byte(tt.encoded)); err != nil {
				if id != nilID {
					t.Errorf("ID.UnmarshalText(%s) got: %v, want nilID %v", tt.encoded, id, nilID)
				}
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
		if id != nilID {
			t.Errorf("Scan() id=%v, want %v", got, nilID)
		}
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

func TestSorter_Len(t *testing.T) {
	if got, want := sorter([]ID{}).Len(), 0; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
	if got, want := sorter(sortTests).Len(), 6; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
}

func TestSorter_Less(t *testing.T) {
	// sorted (ascending) should be IDs 2, 3, 0, 1
	sorter := sorter(sortTests)
	if !sorter.Less(0, 1) {
		t.Errorf("Less(0, 1) not true")
	}
	if sorter.Less(3, 2) {
		t.Errorf("Less(2, 1) true")
	}
	if sorter.Less(0, 0) {
		t.Errorf("Less(0, 0) true")
	}
}

func TestSorter_Swap(t *testing.T) {
	ids := make([]ID, 0)
	ids = append(ids, sortTests...)
	sorter := sorter(ids)
	sorter.Swap(0, 1)
	if got, want := ids[0], sortTests[1]; !reflect.DeepEqual(got, want) {
		t.Error("ids[0] != IDList[1]")
	}
	if got, want := ids[1], sortTests[0]; !reflect.DeepEqual(got, want) {
		t.Error("ids[1] != IDList[0]")
	}
	sorter.Swap(2, 2)
	if got, want := ids[2], sortTests[2]; !reflect.DeepEqual(got, want) {
		t.Error("ids[2], IDList[2]")
	}
}

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
    String()  %s
    Timestamp() %d
    Sequence() %d
    Random()  %d 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Timestamp(), id.Sequence(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := FromString("03f6nlxczw0018fz")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Timestamp(), id.Random())
	// Output: 946684799999 41439
}
