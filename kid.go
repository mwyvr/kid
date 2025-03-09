/*
Package kid (K-sortable ID) provides a goroutine-safe generator of short (10
byte binary, 16 bytes when base32 encoded), url-safe, k-sortable unique IDs.

The 10-byte binary representation of an ID is composed of:

  - 6-byte value representing Unix time in milliseconds
  - 2-byte sequence, and,
  - 2-byte random value.

IDs encode (base32) as 16-byte url-friendly strings.

kid.ID features:

  - Size: 10 bytes as binary, 16 bytes if stored/transported as an encoded string.
  - Timestamp + sequence is guaranteed to be unique.
  - 2 bytes of trailing randomness
  - K-orderable in both binary and base32 encoded representations.
  - URL-friendly custom encoding without the vowels a, i, o, and u.
  - Automatic (un)/marshalling for SQL and JSON.
  - The cmd/kid tool for ID generation and introspection.

Example usage:

	func main() {
	    id := kid.New()
		  fmt.Printf("%s %s %03v\n", id, id.String(), id[:])
		  // Example output: 06bq7xhnr03mlz6r 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]

		  id, err := kid.FromString("06bq7xhnr03mlz6r")
		  if err != nil {
		  	// do something
		  }
		  fmt.Printf("%s %s %03v\n", id, id.String(), id[:])
		  // Output: 06bq7xhnr03mlz6r 06bq7xhnr03mlz6r [001 149 115 246 021 192 007 073 252 216]
	}

Acknowledgments:

While the ID payload differs greatly, the API and much of this package borrows
heavily from https://github.com/rs/xid, a zero-configuration globally-unique
ID generator. ID unique timestamp+sequence pairs are generated from the
google/uuidV7 getV7Time() algorithm.
*/
package kid

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"errors"
	"fmt"
	"sort"
	"sync"
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
//   - 2 bytes, random value provided by crypto/rand
//
// K-orderable: Each subsequent call to New() is guaranteed to produce an ID
// having a timestamp + sequence value greater than the previously generated ID.
func New() (id ID) {
	_ = id[9] // bounds check

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
	// two random bytes
	rand.Read(id[8:])
	return id
}

// IsNil returns true if ID == nilID.
func (id ID) IsNil() bool {
	return id == nilID
}

// IsZero is an alias of is IsNil.
func (id ID) IsZero() bool {
	return id.IsNil()
}

// Encode the id using base32 encoding, writing 16 bytes to dst and return it.
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
	_ = dst[15]
	_ = id[9]

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
	id := &ID{}
	err := id.UnmarshalText([]byte(str))
	return *id, err
}

// UnmarshalText implements `encoding.TextUnmarshaler`, and performs a sanity
// check on text.
//
// Note: decode() is only called from here and should never fail.
// https://pkg.go.dev/encoding#TextUnmarshaler
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) != encodedLen {
		return ErrInvalidID
	}
	for _, c := range text {
		if dec[c] == maxByte {
			return ErrInvalidID
		}
	}
	// should never be reached due to checks
	if !decode(id, text) {
		*id = nilID
		return ErrInvalidID
	}
	return nil
}

// decode by unrolling the stdlib Base32 algorithm plus a custom safe check.
func decode(id *ID, src []byte) bool {
	_ = id[9]
	_ = src[15]

	id[9] = dec[src[14]]<<5 | dec[src[15]]
	// check the last byte
	if encoding[id[9]&0x1F] != src[15] {
		return false
	}
	id[8] = dec[src[12]]<<7 | dec[src[13]]<<2 | dec[src[14]]>>3
	id[7] = dec[src[11]]<<4 | dec[src[12]]>>1
	id[6] = dec[src[9]]<<6 | dec[src[10]]<<1 | dec[src[11]]>>4
	id[5] = dec[src[8]]<<3 | dec[src[9]]>>2
	id[4] = dec[src[6]]<<5 | dec[src[7]]
	id[3] = dec[src[4]]<<7 | dec[src[5]]<<2 | dec[src[6]]>>3
	id[2] = dec[src[3]]<<4 | dec[src[4]]>>1
	id[1] = dec[src[1]]<<6 | dec[src[2]]<<1 | dec[src[3]]>>4
	id[0] = dec[src[0]]<<3 | dec[src[1]]>>2
	return true
}

// Value implements package sql's driver.Valuer.
// https://pkg.go.dev/database/sql/driver#Valuer
func (id ID) Value() (driver.Value, error) {
	if id.IsNil() {
		return nil, nil
	}
	b, err := id.MarshalText()
	return string(b), err
}

// Scan implements the sql.Scanner interface.
// https://pkg.go.dev/database/sql#Scanner
func (id *ID) Scan(value any) error {
	switch val := value.(type) {
	case string:
		return id.UnmarshalText([]byte(val))
	case []byte:
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

// UnmarshalJSON implements the json.Unmarshaler interface.
// https://golang.org/pkg/encoding/json/#Unmarshaler
func (id *ID) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" {
		*id = nilID
		return nil
	}
	// Check the slice length to prevent runtime bounds check panic in UnmarshalText()
	if len(b) < 2 {
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
	// Big Endian
	return int64(uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 | uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5]))
}

// Time returns the ID's timestamp as a Time value with millisecond resolution
// and location set to UTC
func (id ID) Time() time.Time {
	return time.UnixMilli(id.Timestamp()).UTC()
}

// Sequence returns the ID sequence.
func (id ID) Sequence() int32 {
	b := id[6:8]
	// Big Endian
	return int32(uint32(b[0])<<8 | uint32(b[1]))
}

// Random returns the two-byte random component of the ID.
func (id ID) Random() int32 {
	b := id[8:]
	// Big Endian
	return int32(uint32(b[0])<<8 | uint32(b[1]))
}

// Compare makes IDs k-sortable, behaving like `bytes.Compare`, returning 0 if
// two IDs are identical, -1 if the current ID is less than the other, and 1 if
// current ID is greater than other.
//
// Note: only the first 8 bytes of the two IDs (timestamp+sequence) are compared.
func (id ID) Compare(other ID) int {
	return bytes.Compare(id[:8], other[:8])
}

type sorter []ID

func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Less(i, j int) bool {
	return s[i].Compare(s[j]) < 0
}

func (s sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Sort sorts an array of IDs in place.
func Sort(ids []ID) {
	sort.Sort(sorter(ids))
}

// getTS provides the basis of ID timestamp uniqueness; code borrowed directly from getV7Time:
// https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version7.go#L88

var (
	// lastTime is the last time we returned stored as:
	//
	//	52 bits of time in milliseconds since epoch
	//	12 bits of (fractional nanoseconds) >> 8
	lastTime int64
	timeMu   sync.Mutex
	timeNow  = time.Now // for testing
)

const nanoPerMilli = 1000000

// getTS returns:
// - the number of milliseconds elapsed since January 1, 1970 UTC, and,
// - a sequence value
//
// Note: At time of writing, the available timer resolution provided by the Go
// runtime, operating system and hardware can vary from < 1ms to several ms.
// https://pkg.go.dev/time#hdr-Timer_Resolution
func getTS() (milli, seq int64) {
	timeMu.Lock()
	defer timeMu.Unlock()

	nano := timeNow().UnixNano()
	milli = nano / nanoPerMilli
	// Sequence number is between 0 and 3906 (nanoPerMilli>>8)
	seq = (nano - milli*nanoPerMilli) >> 8
	now := milli<<12 + seq
	if now <= lastTime {
		now = lastTime + 1
		milli = now >> 12
		seq = now & 0xfff
	}
	lastTime = now
	// The returned (milli << 12 + seq) is guaranteed to be greater than
	// (milli << 12 + seq) returned by any previous call to getTS.
	return milli, seq
}
