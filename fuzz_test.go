package kid

import (
	"bytes"
	"testing"
)

// FuzzFromString verifies that any input either fails cleanly (returning the
// nil ID) or roundtrips exactly through String().
func FuzzFromString(f *testing.F) {
	f.Add("06bqer9xnm79tfnl")
	f.Add("0000000000000000")
	f.Add("zzzzzzzzzzzzzzzz")
	f.Add("06BQER9XNR09HYQ5") // uppercase: invalid
	f.Add("o6bqer9xnr09hyq5") // 'o' not in alphabet
	f.Add("06bqer9")          // short
	f.Fuzz(func(t *testing.T, s string) {
		id, err := FromString(s)
		if err != nil {
			if id != nilID {
				t.Fatalf("FromString(%q) errored but returned non-nil ID %v", s, id)
			}
			return
		}
		if got := id.String(); got != s {
			t.Fatalf("roundtrip mismatch: FromString(%q).String() = %q", s, got)
		}
	})
}

// FuzzUnmarshalJSON verifies that only null or a correctly quoted 16-char kid
// encoding is ever accepted, and that accepted values roundtrip.
func FuzzUnmarshalJSON(f *testing.F) {
	f.Add([]byte(`"06bqer9xnm79tfnl"`))
	f.Add([]byte(`null`))
	f.Add([]byte(`123456789012345678`)) // regression: bare number, was accepted
	f.Add([]byte(`"0000000000000000"`))
	f.Add([]byte(`'06bqer9xnm79tfnl'`))
	f.Fuzz(func(t *testing.T, b []byte) {
		var id ID
		if err := id.UnmarshalJSON(b); err != nil {
			if id != nilID {
				t.Fatalf("UnmarshalJSON(%q) errored but left non-nil ID %v", b, id)
			}
			return
		}
		if string(b) == "null" {
			if id != nilID {
				t.Fatalf("UnmarshalJSON(null) = %v, want nilID", id)
			}
			return
		}
		if len(b) != encodedLen+2 || b[0] != '"' || b[len(b)-1] != '"' {
			t.Fatalf("UnmarshalJSON accepted non-string JSON: %q", b)
		}
		// the nil ID marshals to null (asymmetric by design); skip roundtrip
		if id == nilID {
			return
		}
		got, err := id.MarshalJSON()
		if err != nil || !bytes.Equal(got, b) {
			t.Fatalf("roundtrip mismatch: %q -> %v -> %q (%v)", b, id, got, err)
		}
	})
}

// FuzzFromBytes verifies binary roundtrips: FromBytes accepts exactly rawLen
// bytes, and any accepted ID roundtrips through both Bytes and String forms.
func FuzzFromBytes(f *testing.F) {
	f.Add([]byte{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf})
	f.Add([]byte{})
	f.Add(bytes.Repeat([]byte{0xff}, rawLen))
	f.Fuzz(func(t *testing.T, b []byte) {
		id, err := FromBytes(b)
		if err != nil {
			if len(b) == rawLen {
				t.Fatalf("FromBytes rejected valid %d-byte input", rawLen)
			}
			return
		}
		if len(b) != rawLen {
			t.Fatalf("FromBytes accepted %d-byte input", len(b))
		}
		if !bytes.Equal(id.Bytes(), b) {
			t.Fatalf("Bytes() = %v, want %v", id.Bytes(), b)
		}
		back, err := FromString(id.String())
		if err != nil || back != id {
			t.Fatalf("string roundtrip failed: %v -> %s -> %v (%v)", id, id.String(), back, err)
		}
	})
}
