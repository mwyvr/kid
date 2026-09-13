/*
Package kid generates short, URL-safe, k-sortable unique IDs.

An ID is 10 bytes: 6 bytes of Unix time in milliseconds, 2 bytes of
sequence, and 2 random bytes (math/rand/v2, seeded by the runtime from OS
entropy). Base32-encoded with a lowercase alphabet that omits a, i, o, and u,
an ID is a 16-character URL-friendly string. The alphabet is in ascending
ASCII order, so encoded IDs sort the same as their binary form. Decoding is
case-sensitive: uppercase input is rejected.

New is goroutine-safe and lock-free. The timestamp+sequence is unique and
strictly increasing within a process, even if the wall clock steps backwards.
Sustained capacity is 4,096 IDs per millisecond per process; bursts beyond
that borrow sequence slots from future milliseconds, so IDs stay unique and
sortable but their embedded timestamps lead the wall clock until generation
slows.

Uniqueness is per process. Across processes there is no coordination; two
processes that derive the same timestamp+sequence are separated only by the
16 random bits (1 in 65,536). Use a coordinated or longer ID (xid, uuid)
where cross-machine uniqueness is required.

The zero value, ZeroID, is both the nil sentinel (IsNil) and a valid,
decodable ID: FromString("0000000000000000") decodes to it. A JSON null into
a *ID nils the pointer without calling UnmarshalJSON.

ID implements TextMarshaler/TextUnmarshaler, json.Marshaler/json.Unmarshaler,
and database/sql driver.Valuer and sql.Scanner.

Security note: an ID carries only 16 bits of randomness alongside values
derived from the clock; IDs are predictable by design. Do not use them where
unguessability matters, such as session tokens, API keys, or password reset
codes.

Example usage:

	func main() {
		id := kid.New()
		fmt.Printf("%s %03v\n", id, id[:])
		// Example output: 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]

		id, err := kid.FromString("06bq7xhnr03mlz6r")
		if err != nil {
			// handle the error
		}
		fmt.Printf("%s %03v\n", id, id[:])
		// Output: 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]
	}

Acknowledgments:

The API borrows from github.com/rs/xid, and the timestamp+sequence encoding
derives from google/uuid's getV7Time() algorithm, with its mutex replaced by
a lock-free atomic claim. Third-party license texts are in NOTICES.
*/
package kid

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	mrand "math/rand/v2"
	"slices"
	"sync/atomic"
	"time"
)

// ID represents a unique identifier
type ID [rawLen]byte

const (
	rawLen     = 10                                 // binary
	encodedLen = 16                                 // base32
	encoding   = "0123456789bcdefghjklmnpqrstvwxyz" // base32 encoding without: a,i,o,u
	maxByte    = 0xFF                               // used as a sentinel value in charmap
)

var (
	// ZeroID is the zero value of ID: the nil sentinel (see IsNil) and a
	// valid, decodable ID. It is compared against, never modified.
	ZeroID ID
	// dec maps an alphabet character to its 5-bit value; maxByte marks
	// characters not in the alphabet.
	dec [256]byte

	// ErrInvalidID is returned when decoding input that is not a valid
	// kid encoding.
	ErrInvalidID = errors.New("kid: invalid id")
	// ErrTimestampOutOfRange is returned by NewWithTime when t does not fit
	// the 6-byte millisecond field: before 1970 or after ~2262.
	ErrTimestampOutOfRange = errors.New("kid: timestamp out of range")
)

func init() {
	for i := range len(dec) {
		dec[i] = maxByte
	}
	for i := range len(encoding) {
		dec[encoding[i]] = byte(i)
	}
}

// New generates a new unique ID: 6 bytes of Unix time in milliseconds,
// 2 bytes of sequence, and 2 random bytes from math/rand/v2.
//
// New is goroutine-safe and lock-free: the timestamp+sequence is claimed
// with a single compare-and-swap, falling back to a wait-free atomic
// increment under contention. Every call returns an ID whose timestamp +
// sequence is strictly greater than the previously generated one, even if
// the wall clock steps backwards (see getTS).
func New() (id ID) {
	t, s := getTS()
	return buildID(t, s)
}

// NewWithTime generates a new ID with the given timestamp. The sequence is
// derived from t's sub-millisecond component, as in New.
//
// NewWithTime does not draw from New's monotonic sequence, so its IDs are
// not ordered with respect to New() output and ts+seq uniqueness against
// New is not guaranteed. Use it where you control generation for a key
// space: tests, backfills, replays.
func NewWithTime(t time.Time) (id ID, err error) {
	milli := t.UnixMilli()
	if milli < 0 || milli >= 1<<48 {
		return ZeroID, ErrTimestampOutOfRange
	}
	// Nanosecond stays well-defined where UnixNano would overflow.
	sub := int64(t.Nanosecond()) % nanoPerMilli
	return buildID(milli, sub>>8), nil
}

// buildID lays out milli (48 bits), seq (12 bits), and 2 random bytes.
// Callers must keep milli and seq in range.
func buildID(milli, seq int64) (id ID) {
	// timestamp, 6 bytes, big endian
	id[0] = byte(milli >> 40)
	id[1] = byte(milli >> 32)
	id[2] = byte(milli >> 24)
	id[3] = byte(milli >> 16)
	id[4] = byte(milli >> 8)
	id[5] = byte(milli)
	// sequence, 2 bytes, big endian
	id[6] = byte(seq >> 8)
	id[7] = byte(seq)
	// 2 random bytes; IDs are predictable by design, see the package doc
	r := mrand.Uint32()
	id[8] = byte(r >> 8)
	id[9] = byte(r)
	return id
}

// IsNil reports whether id is the zero value, ZeroID. Note that ZeroID is
// also a valid, decodable ID.
func (id ID) IsNil() bool {
	return id == ZeroID
}

// IsZero reports whether id is the zero value. It is an alias for IsNil.
func (id ID) IsZero() bool {
	return id.IsNil()
}

// Encode writes the 16-byte base32 encoding of id to dst and returns it.
// dst must have length at least 16; a shorter slice panics.
func (id ID) Encode(dst []byte) []byte {
	encode(dst, id[:])
	return dst
}

// String returns the 16-character base32 encoding of id.
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return string(text)
}

// MarshalText returns the 16-byte base32 encoding of id. Every ID encodes,
// so error is always nil.
func (id ID) MarshalText() ([]byte, error) {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return text, nil
}

// encode writes the 16-byte base32 encoding of id to dst, unrolling the
// stdlib algorithm without bounds checks. Base32 of 10 bytes needs no
// padding: 8 encoded bytes per 5 input bytes.
func encode(dst, id []byte) {
	_ = dst[15] // bounds check hint
	_ = id[9]   // bounds check hint

	dst[15] = encoding[id[9]&0x1F]
	dst[14] = encoding[(id[9]>>5)|(id[8]<<3)&0x1F]
	dst[13] = encoding[(id[8]>>2)&0x1F]
	dst[12] = encoding[id[8]>>7|(id[7]<<1)&0x1F]
	dst[11] = encoding[(id[7]>>4)&0x1F|(id[6]<<4)&0x1F]
	dst[10] = encoding[(id[6]>>1)&0x1F]
	dst[9] = encoding[(id[6]>>6)&0x1F|(id[5]<<2)&0x1F]
	dst[8] = encoding[id[5]>>3]
	dst[7] = encoding[id[4]&0x1F]
	dst[6] = encoding[id[4]>>5|(id[3]<<3)&0x1F]
	dst[5] = encoding[(id[3]>>2)&0x1F]
	dst[4] = encoding[id[3]>>7|(id[2]<<1)&0x1F]
	dst[3] = encoding[(id[2]>>4)&0x1F|(id[1]<<4)&0x1F]
	dst[2] = encoding[(id[1]>>1)&0x1F]
	dst[1] = encoding[(id[1]>>6)&0x1F|(id[0]<<2)&0x1F]
	dst[0] = encoding[id[0]>>3]
}

// FromBytes copies b into an ID. Only a length check is performed.
func FromBytes(b []byte) (ID, error) {
	var id ID
	if len(b) != rawLen {
		return ZeroID, ErrInvalidID
	}
	copy(id[:], b)
	return id, nil
}

// FromString decodes a 16-character base32-encoded string to return an ID.
// Decoding is case-sensitive: uppercase input is rejected with ErrInvalidID.
func FromString(str string) (ID, error) {
	var id ID
	err := id.UnmarshalText([]byte(str))
	return id, err
}

// UnmarshalText decodes a 16-byte lowercase base32 kid encoding. On error,
// id is set to ZeroID and ErrInvalidID is returned.
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) != encodedLen {
		*id = ZeroID
		return ErrInvalidID
	}
	for _, c := range text {
		if dec[c] == maxByte {
			*id = ZeroID
			return ErrInvalidID
		}
	}
	decode(id, text)
	return nil
}

// decode fills id from 16 characters of src, which the caller has
// validated. 16 characters x 5 bits is exactly the 80 bits of an ID, so
// every 16-character string over the alphabet is a valid encoding.
func decode(id *ID, src []byte) {
	_ = src[15] // bounds check hint

	id[9] = dec[src[14]]<<5 | dec[src[15]]
	id[8] = dec[src[12]]<<7 | dec[src[13]]<<2 | dec[src[14]]>>3
	id[7] = dec[src[11]]<<4 | dec[src[12]]>>1
	id[6] = dec[src[9]]<<6 | dec[src[10]]<<1 | dec[src[11]]>>4
	id[5] = dec[src[8]]<<3 | dec[src[9]]>>2
	id[4] = dec[src[6]]<<5 | dec[src[7]]
	id[3] = dec[src[4]]<<7 | dec[src[5]]<<2 | dec[src[6]]>>3
	id[2] = dec[src[3]]<<4 | dec[src[4]]>>1
	id[1] = dec[src[1]]<<6 | dec[src[2]]<<1 | dec[src[3]]>>4
	id[0] = dec[src[0]]<<3 | dec[src[1]]>>2
}

// Value implements driver.Valuer, returning the 16-character encoded
// string, or nil for ZeroID. Value only ever writes the string form; to
// store the 10-byte binary form (e.g. in a VARBINARY(10) column), use
// ValueBinary. Scan reads both back.
func (id ID) Value() (driver.Value, error) {
	if id.IsNil() {
		return nil, nil
	}
	return id.String(), nil
}

// ValueBinary implements driver.Valuer, returning a copy of the 10-byte
// binary form, or nil for ZeroID. Scan reads both forms back.
func (id ID) ValueBinary() (driver.Value, error) {
	if id.IsNil() {
		return nil, nil
	}
	return id.Bytes(), nil
}

// Scan implements sql.Scanner, accepting the encoded form as a string or
// []byte, the 10-byte binary form as a []byte, or nil, which yields ZeroID.
// The binary form can only be read through Scan; use ValueBinary to write
// it.
func (id *ID) Scan(value any) error {
	switch val := value.(type) {
	case string:
		return id.UnmarshalText([]byte(val))
	case []byte:
		if len(val) == rawLen {
			copy(id[:], val)
			return nil
		}
		return id.UnmarshalText(val)
	case nil:
		*id = ZeroID
		return nil
	default:
		*id = ZeroID
		return fmt.Errorf("kid: scanning unsupported type: %T", value)
	}
}

// MarshalJSON encodes id as a quoted string, or null for ZeroID. Every ID
// encodes, so error is always nil.
func (id ID) MarshalJSON() ([]byte, error) {
	if id == ZeroID {
		return []byte("null"), nil
	}
	text := make([]byte, encodedLen+2) // +2 accounts for ""
	encode(text[1:encodedLen+1], id[:])
	text[0], text[encodedLen+1] = '"', '"'

	return text, nil
}

// UnmarshalJSON accepts null (ZeroID) or a quoted 16-character kid encoding.
// A JSON null into a *ID nils the pointer before UnmarshalJSON is called;
// the null case applies when decoding into an ID value.
func (id *ID) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*id = ZeroID
		return nil
	}
	if len(b) != encodedLen+2 || b[0] != '"' || b[len(b)-1] != '"' {
		*id = ZeroID
		return ErrInvalidID
	}
	return id.UnmarshalText(b[1 : len(b)-1])
}

// Bytes returns a copy of the 10-byte binary form of id.
func (id ID) Bytes() []byte {
	b := make([]byte, rawLen)
	copy(b, id[:])
	return b
}

// Timestamp returns the timestamp component of id, milliseconds since the
// Unix epoch. The 6-byte field overflows around the year 2262.
func (id ID) Timestamp() int64 {
	// First 8 bytes as one big-endian uint64, shifted to drop the sequence.
	return int64(binary.BigEndian.Uint64(id[:]) >> 16)
}

// Time returns the timestamp component of id as time.Time, with millisecond
// resolution and location UTC.
func (id ID) Time() time.Time {
	return time.UnixMilli(id.Timestamp()).UTC()
}

// Sequence returns the sequence component of id. For IDs from New this is
// 12 bits (0-4095); overflow within a millisecond carries into the
// timestamp (see getTS). IDs from other sources may carry any 16-bit value.
func (id ID) Sequence() int32 {
	b := id[6:8]
	return int32(uint32(b[0])<<8 | uint32(b[1]))
}

// Random returns the two random bytes of id.
func (id ID) Random() int32 {
	b := id[8:]
	return int32(uint32(b[0])<<8 | uint32(b[1]))
}

// Compare reports whether id is less than, equal to, or greater than other
// with bytes.Compare semantics: -1, 0, or 1. All 10 bytes participate, so
// Compare is consistent with ==; the timestamp and sequence occupy the
// leading bytes, so IDs order by creation time first.
func (id ID) Compare(other ID) int {
	return bytes.Compare(id[:], other[:])
}

// Sort sorts ids in place, ascending.
func Sort(ids []ID) {
	slices.SortFunc(ids, ID.Compare)
}

var (
	// lastTime is the last issued ts+seq: 52 bits of milliseconds since
	// epoch (valid until ~2262) and 12 bits of (fractional nanoseconds >> 8).
	lastTime atomic.Int64
	timeNow  = time.Now // for testing
)

const nanoPerMilli = 1000000

// getTS returns the current Unix time in milliseconds and a sequence value.
// The fast path claims a clock-derived value with one compare-and-swap; if
// the clock is not ahead of the last issued value, or the swap loses a race,
// the next slot is claimed with a wait-free atomic increment. Both paths
// strictly increase lastTime and return exactly the value installed, so
// every (milli << 12 + seq) is strictly greater than any previous one, even
// if the wall clock steps backwards, with no retry loop.
func getTS() (milli, seq int64) {
	nano := timeNow().UnixNano()
	milli = nano / nanoPerMilli
	// seq is 0-3906 clock-derived; the increment path can return up to 4095
	seq = (nano - milli*nanoPerMilli) >> 8
	now := milli<<12 + seq
	if last := lastTime.Load(); now > last && lastTime.CompareAndSwap(last, now) {
		return milli, seq
	}
	// The wall clock is not ahead, or another goroutine won the race:
	// claim the next slot wait-free.
	now = lastTime.Add(1)
	return now >> 12, now & 0xfff
}
