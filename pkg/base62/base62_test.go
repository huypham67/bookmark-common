package base62

import (
	"errors"
	"math"
	"testing"
)

// TestAlphabetsValid checks that every built-in alphabet is exactly 62
// unique characters.
func TestAlphabetsValid(t *testing.T) {
	for name, s := range map[string]string{
		"ordered":  OrderedCharset,
		"shuffled": shuffledCharset,
	} {
		if len(s) != base {
			t.Fatalf("%s: length = %d, want %d", name, len(s), base)
		}
		seen := map[byte]bool{}
		for i := 0; i < len(s); i++ {
			if seen[s[i]] {
				t.Fatalf("%s: duplicate character %q", name, s[i])
			}
			seen[s[i]] = true
		}
	}
}

// TestKnownValues verifies encoding against OrderedEncoding, whose output is
// deterministic and easy to reason about.
func TestKnownValues(t *testing.T) {
	cases := []struct {
		num uint64
		out string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "a"},
		{61, "Z"},
		{62, "10"},
		{125, "21"},  // 2*62 + 1
		{3843, "ZZ"}, // 61*62 + 61
		{3844, "100"},
	}
	for _, c := range cases {
		if got := OrderedEncoding.FormatUint(c.num); got != c.out {
			t.Errorf("FormatUint(%d) = %q, want %q", c.num, got, c.out)
		}
		got, err := OrderedEncoding.ParseUint(c.out)
		if err != nil {
			t.Errorf("ParseUint(%q) unexpected error: %v", c.out, err)
		}
		if got != c.num {
			t.Errorf("ParseUint(%q) = %d, want %d", c.out, got, c.num)
		}
	}
}

// TestRoundTripUint checks that ParseUint(FormatUint(n)) == n across a range
// of values.
func TestRoundTripUint(t *testing.T) {
	vals := []uint64{
		0, 1, 61, 62, 63, 125, 1000, 1 << 16, 1 << 32, 1<<32 + 1,
		math.MaxUint32, math.MaxUint64 - 1, math.MaxUint64,
	}
	for _, enc := range []*Encoding{StdEncoding, OrderedEncoding} {
		for _, n := range vals {
			s := enc.FormatUint(n)
			got, err := enc.ParseUint(s)
			if err != nil {
				t.Fatalf("ParseUint(%q) error: %v", s, err)
			}
			if got != n {
				t.Errorf("round trip %d -> %q -> %d", n, s, got)
			}
		}
	}
}

// TestRoundTripInt covers signed values, including negatives, which are
// encoded via their two's complement representation.
func TestRoundTripInt(t *testing.T) {
	vals := []int64{0, 1, -1, 42, -42, math.MinInt64, math.MaxInt64}
	for _, n := range vals {
		s := FormatInt(n)
		got, err := ParseInt(s)
		if err != nil {
			t.Fatalf("ParseInt(%q) error: %v", s, err)
		}
		if got != n {
			t.Errorf("round trip int %d -> %q -> %d", n, s, got)
		}
	}
}

// TestShuffledIsNonSequential confirms that adjacent ids do not map to
// adjacent codes, which is the purpose of the shuffled alphabet.
func TestShuffledIsNonSequential(t *testing.T) {
	a := StdEncoding.FormatUint(1)
	b := StdEncoding.FormatUint(2)
	if a == "1" && b == "2" {
		t.Errorf("StdEncoding looks ordered: 1->%q, 2->%q", a, b)
	}
}

func TestParseInvalidChar(t *testing.T) {
	for _, s := range []string{"!", "ab*c", "with space", "日本"} {
		_, err := ParseUint(s)
		var cie CorruptInputError
		if !errors.As(err, &cie) {
			t.Errorf("ParseUint(%q) want CorruptInputError, got %v", s, err)
		}
	}
}

func TestParseEmpty(t *testing.T) {
	if _, err := ParseUint(""); err == nil {
		t.Error("ParseUint(\"\") want error, got nil")
	}
}

func TestParseIntInvalid(t *testing.T) {
	if _, err := ParseInt("!"); err == nil {
		t.Error("ParseInt(\"!\") want error, got nil")
	}
}

// TestParseOverflow checks that a string representing a value larger than
// maxUint64 returns an OutOfRangeError.
func TestParseOverflow(t *testing.T) {
	// maxUint64 needs 11 characters; appending one more digit overflows.
	big := OrderedEncoding.FormatUint(math.MaxUint64) + "1"
	_, err := OrderedEncoding.ParseUint(big)
	var oor OutOfRangeError
	if !errors.As(err, &oor) {
		t.Errorf("ParseUint(%q) want OutOfRangeError, got %v", big, err)
	}
}

func TestErrorMessages(t *testing.T) {
	if msg := CorruptInputError(3).Error(); msg == "" {
		t.Error("CorruptInputError.Error() is empty")
	}
	if msg := OutOfRangeError(5).Error(); msg == "" {
		t.Error("OutOfRangeError.Error() is empty")
	}
}

func TestNewEncodingPanics(t *testing.T) {
	cases := map[string]string{
		"too short": "abc",
		"duplicate": "00" + OrderedCharset[2:], // first character repeated
	}
	for name, alphabet := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("NewEncoding(%q...) want panic", name)
				}
			}()
			NewEncoding(alphabet)
		})
	}
}

// FuzzRoundTrip asserts that encoding then decoding any uint64 yields the
// original value.
func FuzzRoundTrip(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(123456789))
	f.Add(uint64(math.MaxUint64))
	f.Fuzz(func(t *testing.T, n uint64) {
		s := FormatUint(n)
		got, err := ParseUint(s)
		if err != nil {
			t.Fatalf("ParseUint(%q) error: %v", s, err)
		}
		if got != n {
			t.Fatalf("round trip %d -> %q -> %d", n, s, got)
		}
	})
}

func BenchmarkFormatUint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FormatUint(uint64(i) * 2654435761)
	}
}

func BenchmarkParseUint(b *testing.B) {
	s := FormatUint(math.MaxUint64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseUint(s)
	}
}
