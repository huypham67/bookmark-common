package base62

import "strconv"

// CorruptInputError reports that the input string contains a character
// outside the alphabet. Its value is the byte index of that character.
type CorruptInputError int64

func (e CorruptInputError) Error() string {
	return "base62: invalid character at index " + strconv.FormatInt(int64(e), 10)
}

// OutOfRangeError reports that the decoded value exceeds the range of a
// uint64. Its value is the byte index at which the overflow was detected.
type OutOfRangeError int64

func (e OutOfRangeError) Error() string {
	return "base62: value overflows uint64 at index " + strconv.FormatInt(int64(e), 10)
}
