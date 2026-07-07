/*
Package kid (K-sortable ID) provides a goroutine-safe generator of short (10
byte binary, 16 bytes when base32 encoded), url-safe, k-sortable unique IDs.

The 10-byte binary representation of an ID is composed of:

  - 6-byte value representing Unix time in milliseconds
  - 2-byte sequence, and,
  - 2-byte random value (ChaCha8, seeded by the Go runtime from OS entropy).

IDs encode (base32) as 16-byte URL-friendly strings. The encoding alphabet is
in ascending ASCII order, so the encoded form preserves the sort order of the
binary form: IDs are k-orderable in either representation.

kid.ID features:

  - Size: 10 bytes as binary, 16 bytes if stored/transported as an encoded string.
  - Timestamp + sequence is guaranteed to be unique and monotonically
    increasing within a process, even if the wall clock steps backwards.
  - Lock-free generation: New scales with cores instead of serializing on a
    mutex.
  - 2 bytes of trailing randomness from math/rand/v2's ChaCha8 generator.
  - K-orderable in both binary and base32 encoded representations.
  - URL-friendly custom encoding without the vowels a, i, o, and u.
  - Automatic (un)/marshalling for SQL and JSON.
  - The cmd/kid tool for ID generation and introspection.

Capacity: uniqueness is carried entirely by the timestamp+sequence pair,
giving 4,096 IDs per millisecond (~4.1 million/second) sustained, per
process. Bursts beyond that rate borrow sequence slots from future
milliseconds: IDs remain unique and strictly k-sortable, but their embedded
timestamps lead the wall clock until generation slows. Treat the embedded
time as approximate metadata rather than an exact wall-clock instant.

Security note: an ID carries only 16 bits of randomness alongside values
derived from the clock; IDs are predictable by design. Do not use kid IDs
where unguessability matters, such as session tokens, API keys, or password
reset codes.

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

While the ID payload differs greatly, the API and much of this package borrows
heavily from https://github.com/rs/xid, a zero-configuration globally-unique
ID generator. The timestamp+sequence encoding is derived from the google/uuid
getV7Time() algorithm, with its mutex protection replaced by a lock-free
atomic claim. Third-party copyright notices and license texts are reproduced
in the NOTICES file.
*/
package kid

import (
	"bytes"
	"database/sql/driver"
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
	nilID ID        // nilID represents the zero-value of an ID
	dec   [256]byte // dec is the base32 decoding map

	// ErrInvalidID represents an error state, typically when decoding invalid input
	ErrInvalidID = errors.New("kid: invalid id")
)

func init() {
	// initialize the decoding map; used also for sanity checking input
	for i := range len(dec) {
		dec[i] = maxByte
	}
	for i := range len(encoding) {
		dec[encoding[i]] = byte(i)
	}
}

// New generates a new unique ID.
//
// This function is goroutine-safe. IDs are composed of:
//
//   - 6 bytes, timestamp, a Unix time in milliseconds
//   - 2 bytes, sequence, a derived value ensuring uniqueness and order
//   - 2 bytes, random value from math/rand/v2's ChaCha8 generator
//
// New is lock-free and free of retry loops: the timestamp+sequence is
// claimed with at most one atomic compare-and-swap followed, if needed, by a
// wait-free atomic increment, and the random bytes are drawn from
// math/rand/v2's per-goroutine ChaCha8 state (seeded by the Go runtime from
// OS entropy), so ID generation scales with cores rather than serializing
// on a mutex or spinning under contention.
//
// K-orderable: Each subsequent call to New() is guaranteed to produce an ID
// having a timestamp + sequence value greater than the previously generated ID.
func New() (id ID) {
	t, s := getTS() // milli << 12 + seq
	// timestamp, 6 bytes, big endian
	id[0] = byte(t >> 40)
	id[1] = byte(t >> 32)
	id[2] = byte(t >> 24)
	id[3] = byte(t >> 16)
	id[4] = byte(t >> 8)
	id[5] = byte(t)
	// sequence, 2 bytes, big endian
	id[6] = byte(s >> 8)
	id[7] = byte(s)
	// Two random bytes from the runtime-seeded ChaCha8 generator; see the
	// package documentation for the security properties of this choice.
	r := mrand.Uint32()
	id[8] = byte(r >> 8)
	id[9] = byte(r)
	return id
}

// IsNil returns true if ID == nilID.
func (id ID) IsNil() bool {
	return id == nilID
}

// IsZero is an alias of IsNil.
func (id ID) IsZero() bool {
	return id.IsNil()
}

// Encode the id using base32 encoding, writing 16 bytes to dst and returning
// it. Encode panics if len(dst) < 16.
func (id ID) Encode(dst []byte) []byte {
	encode(dst, id[:])
	return dst
}

// String implements `fmt.Stringer`, returning id as a base32 encoded string
// using the kid custom character set.
// https://pkg.go.dev/fmt#Stringer
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return string(text)
}

// MarshalText implements `encoding.TextMarshaler`.
//
// As any ID value will always encode, error is always nil.
// https://golang.org/pkg/encoding/#TextMarshaler
func (id ID) MarshalText() ([]byte, error) {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return text, nil
}

// encode encodes id bytes by unrolling the stdlib base32 algorithm and removing
// all safe checks for performance.
//
// dst will always contain 16 bytes. Base32 encoded 10-byte binary ids are never
// padded as base32 encoding returns 8 encoded bytes per 5 bytes of input.
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

// FromBytes copies []bytes into an ID value.
// Only a length-check is performed.
func FromBytes(b []byte) (ID, error) {
	var id ID
	if len(b) != rawLen {
		return nilID, ErrInvalidID
	}
	copy(id[:], b)
	return id, nil
}

// FromString decodes a base32-encoded string to return an ID.
func FromString(str string) (ID, error) {
	var id ID
	err := id.UnmarshalText([]byte(str))
	return id, err
}

// UnmarshalText implements `encoding.TextUnmarshaler`. text must be a 16-byte
// base32-encoded value over the kid alphabet; on error, id is set to the nil
// ID and ErrInvalidID is returned.
// https://pkg.go.dev/encoding#TextUnmarshaler
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) != encodedLen {
		*id = nilID
		return ErrInvalidID
	}
	for _, c := range text {
		if dec[c] == maxByte {
			*id = nilID
			return ErrInvalidID
		}
	}
	decode(id, text)
	return nil
}

// decode by unrolling the stdlib Base32 algorithm.
//
// decode cannot fail: 16 characters x 5 bits is exactly the 80 bits of a
// 10-byte ID, so every 16-character string over the kid alphabet is a valid
// encoding. (Contrast xid, where 20 characters carry 100 bits against a
// 96-bit ID and the final character must be range-checked.) Input length and
// alphabet membership are enforced by UnmarshalText, the only caller.
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

// Value implements package sql's driver.Valuer, returning the ID in its
// 16-byte encoded string form, or nil for the nil ID.
// https://pkg.go.dev/database/sql/driver#Valuer
func (id ID) Value() (driver.Value, error) {
	if id.IsNil() {
		return nil, nil
	}
	return id.String(), nil
}

// Scan implements the sql.Scanner interface, accepting the 16-byte encoded
// form as a string or []byte, the 10-byte binary form as a []byte, or nil,
// which yields the nil ID.
// https://pkg.go.dev/database/sql#Scanner
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
		*id = nilID
		return nil
	default:
		return fmt.Errorf("kid: scanning unsupported type: %T", value)
	}
}

// MarshalJSON implements the json.Marshaler interface.
//
// A json value will always be returned; as a nilID or any other binary ID will
// always encode, error will always be nil.
//
// https://golang.org/pkg/encoding/json/#Marshaler
func (id ID) MarshalJSON() ([]byte, error) {
	// endless loop if merely return json.Marshal(id)
	if id == nilID {
		return []byte("null"), nil
	}
	text := make([]byte, encodedLen+2) // +2 accounts for ""
	encode(text[1:encodedLen+1], id[:])
	text[0], text[encodedLen+1] = '"', '"'

	return text, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface, accepting only
// null or a quoted 16-character kid encoding.
// https://golang.org/pkg/encoding/json/#Unmarshaler
func (id *ID) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*id = nilID
		return nil
	}
	// Only a quoted string is acceptable. Without the quote check, a bare
	// JSON number of the right length would be accepted, as digits are valid
	// characters in the kid alphabet.
	if len(b) != encodedLen+2 || b[0] != '"' || b[len(b)-1] != '"' {
		*id = nilID
		return ErrInvalidID
	}
	return id.UnmarshalText(b[1 : len(b)-1])
}

// Bytes returns the binary representation of id, which is simply id[:].
func (id ID) Bytes() []byte {
	return id[:]
}

// Timestamp returns the timestamp component of id as milliseconds since the
// Unix epoch. Go timestamps are at location UTC.
func (id ID) Timestamp() int64 {
	b := id[0:6]
	// Big Endian, no overflow possible
	return int64(uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 | uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5])) //nolint:gosec
}

// Time returns the ID's timestamp as a Time value with millisecond resolution
// and location set to UTC
func (id ID) Time() time.Time {
	return time.UnixMilli(id.Timestamp()).UTC()
}

// Sequence returns the sequence component of id.
//
// For IDs produced by New, the sequence is a 12-bit value (0-4095); if a
// burst of calls would overflow the sequence within a single millisecond, the
// overflow carries into the timestamp, preserving order (see getTS). The
// field occupies two bytes, so IDs from other sources may carry larger
// values.
func (id ID) Sequence() int32 {
	b := id[6:8]
	// Big Endian
	return int32(uint32(b[0])<<8 | uint32(b[1])) //nolint:gosec
}

// Random returns the two-byte random component of the ID.
func (id ID) Random() int32 {
	b := id[8:]
	// Big Endian
	return int32(uint32(b[0])<<8 | uint32(b[1])) //nolint:gosec
}

// Compare returns an integer comparing two IDs with `bytes.Compare`
// semantics: 0 if the IDs are identical, -1 if id is less than other, and 1
// if id is greater than other. All 10 bytes participate, so Compare is
// consistent with ==; because the timestamp and sequence occupy the leading
// bytes, IDs order by creation time first.
func (id ID) Compare(other ID) int {
	return bytes.Compare(id[:], other[:])
}

// Sort sorts a slice of IDs in place, in ascending order.
func Sort(ids []ID) {
	slices.SortFunc(ids, ID.Compare)
}

// getTS provides the basis of ID timestamp uniqueness; the time encoding is
// borrowed from getV7Time, converted from mutex protection to a lock-free
// compare-and-swap:
// https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version7.go#L88

var (
	// lastTime is the last ts+seq we returned stored as:
	//
	//	52 bits of time in milliseconds since epoch
	//	12 bits of (fractional nanoseconds) >> 8
	lastTime atomic.Int64
	timeNow  = time.Now // for testing
)

const nanoPerMilli = 1000000

// getTS returns:
// - the number of milliseconds elapsed since January 1, 1970 UTC, and,
// - a sequence value
//
// The fast path claims a clock-derived value with a single compare-and-swap;
// if the clock is not ahead of the last issued value, or the swap loses a
// race, getTS instead claims the next sequence slot with a wait-free atomic
// increment. Both operations strictly increase lastTime and return exactly
// the value they installed, so every call — across all goroutines — returns
// a (milli << 12 + seq) strictly greater than that of any previous call,
// even if the wall clock steps backwards. There is no retry loop: under
// contention every caller completes in a bounded number of atomic
// operations, which avoids CAS retry storms on hardware where the shared
// cache line is expensive to bounce (notably multi-cluster arm64 CPUs such
// as Apple silicon). The clock path re-synchronizes the timestamp to real
// time whenever the wall clock is ahead.
//
// Note: At time of writing, the available timer resolution provided by the Go
// runtime, operating system and hardware can vary from < 1ms to several ms.
// https://pkg.go.dev/time#hdr-Timer_Resolution
func getTS() (milli, seq int64) {
	nano := timeNow().UnixNano()
	milli = nano / nanoPerMilli
	// Sequence number is between 0 and 3906 (nanoPerMilli>>8)
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
