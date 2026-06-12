// Package base62 encodes and decodes integers as URL-safe base62 strings
// using only the characters 0-9, a-z and A-Z.
//
// It is intended for the URL-shortener pattern, mapping an auto-increment
// integer id to a short opaque code and back:
//
//	code := base62.FormatUint(uint64(b.CodeInt))
//	n, err := base62.ParseUint(code)
package base62

const base = 62

// OrderedCharset is the 62-character alphabet in natural order
// (0-9, a-z, A-Z). It produces codes that increase monotonically with the
// encoded value and is useful as a reference or when monotonic codes are
// desired.
const OrderedCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// shuffledCharset is a fixed permutation of OrderedCharset. Consecutive
// values (1, 2, 3, ...) map to codes that appear random, making them harder
// to guess or enumerate sequentially.
//
// This value must never change: altering it would cause every previously
// stored code to decode to a different number.
const shuffledCharset = "oTId3S4FiJPpHneVyXamgEML7OYCvtQ2KbA1zr5hwqsjWux68B0ZRlU9NGkfcD"

// invalidIndex marks a byte that is not part of the alphabet in decodeMap.
const invalidIndex = 0xFF

// Encoding holds the encode and decode tables for a 62-character alphabet.
// Create one with NewEncoding. An Encoding is read-only after creation and
// safe for concurrent use.
type Encoding struct {
	encode    [base]byte
	decodeMap [256]byte
}

// StdEncoding is the default Encoding. It uses a shuffled alphabet so that
// codes generated from sequential ids resist enumeration.
var StdEncoding = NewEncoding(shuffledCharset)

// OrderedEncoding uses the natural-order alphabet, producing codes that
// increase monotonically with the encoded value.
var OrderedEncoding = NewEncoding(OrderedCharset)

// NewEncoding returns a new Encoding defined by the given alphabet.
//
// The alphabet must be exactly 62 bytes with no duplicate characters;
// otherwise NewEncoding panics, since that indicates a programming error
// that should surface at initialization time.
func NewEncoding(alphabet string) *Encoding {
	if len(alphabet) != base {
		panic("base62: alphabet must be exactly 62 characters")
	}

	e := new(Encoding)
	for i := range e.decodeMap {
		e.decodeMap[i] = invalidIndex
	}
	for i := 0; i < base; i++ {
		c := alphabet[i]
		if e.decodeMap[c] != invalidIndex {
			panic("base62: alphabet contains duplicate characters")
		}
		e.encode[i] = c
		e.decodeMap[c] = byte(i)
	}
	return e
}

// FormatUint returns the base62 encoding of num. FormatUint(0) returns the
// first character of the alphabet.
func (e *Encoding) FormatUint(num uint64) string {
	if num == 0 {
		return string(e.encode[0])
	}

	// The largest uint64 needs at most 11 base62 characters.
	var buf [11]byte
	i := len(buf)
	for num > 0 {
		i--
		buf[i] = e.encode[num%base]
		num /= base
	}
	return string(buf[i:])
}

// FormatInt returns the base62 encoding of num. Negative values are
// interpreted as their two's complement uint64 representation, so FormatInt
// and ParseInt form a consistent round trip.
func (e *Encoding) FormatInt(num int64) string {
	return e.FormatUint(uint64(num))
}

// ParseUint returns the unsigned integer represented by the base62 string s.
//
// It returns a CorruptInputError if s contains a character outside the
// alphabet, or an OutOfRangeError if the value does not fit in a uint64.
func (e *Encoding) ParseUint(s string) (uint64, error) {
	if s == "" {
		return 0, CorruptInputError(0)
	}

	const cutoff = ^uint64(0) // maxUint64

	var num uint64
	for i := 0; i < len(s); i++ {
		x := e.decodeMap[s[i]]
		if x == invalidIndex {
			return 0, CorruptInputError(i)
		}
		// Detect overflow before computing num*base + x.
		if num > (cutoff-uint64(x))/base {
			return 0, OutOfRangeError(i)
		}
		num = num*base + uint64(x)
	}
	return num, nil
}

// ParseInt returns the signed integer represented by the base62 string s.
// It is the inverse of FormatInt.
func (e *Encoding) ParseInt(s string) (int64, error) {
	num, err := e.ParseUint(s)
	if err != nil {
		return 0, err
	}
	return int64(num), nil
}

// FormatUint returns the base62 encoding of num using StdEncoding.
func FormatUint(num uint64) string { return StdEncoding.FormatUint(num) }

// FormatInt returns the base62 encoding of num using StdEncoding.
func FormatInt(num int64) string { return StdEncoding.FormatInt(num) }

// ParseUint returns the unsigned integer represented by s using StdEncoding.
func ParseUint(s string) (uint64, error) { return StdEncoding.ParseUint(s) }

// ParseInt returns the signed integer represented by s using StdEncoding.
func ParseInt(s string) (int64, error) { return StdEncoding.ParseInt(s) }
