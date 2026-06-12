package shortcode

import (
	"errors"
	"strings"
	"testing"

	"github.com/huypham67/bookmark-common/pkg/base62"
)

// failingReader always errors, used to cover the randomness failure paths.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("rand failure") }

func TestEncodeErrorsWhenRandFails(t *testing.T) {
	orig := randReader
	randReader = failingReader{}
	defer func() { randReader = orig }()

	if _, err := EncodeSQLCode(1); err == nil {
		t.Error("EncodeSQLCode expected error when rand fails")
	}
	if _, err := AddRedisPrefix("abc"); err == nil {
		t.Error("AddRedisPrefix expected error when rand fails")
	}
}

func TestPrefixBucketsDisjoint(t *testing.T) {
	for i := 0; i < len(RedisPrefixes); i++ {
		if strings.IndexByte(SQLPrefixes, RedisPrefixes[i]) >= 0 {
			t.Fatalf("prefix %q is in both buckets", RedisPrefixes[i])
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		code string
		want Store
	}{
		{"", StoreUnknown},
		{"a", StoreRedis},
		{"habc", StoreRedis},   // 'h' is the last Redis prefix
		{"iZZ", StoreSQL},      // 'i' is the first SQL prefix
		{"zfoo", StoreSQL},     // 'z' is the last SQL prefix
		{"5xyz", StoreUnknown}, // digit prefix
		{"Axyz", StoreUnknown}, // uppercase prefix
	}
	for _, c := range cases {
		if got := Classify(c.code); got != c.want {
			t.Errorf("Classify(%q) = %v, want %v", c.code, got, c.want)
		}
	}
}

func TestEncodeSQLCode(t *testing.T) {
	for _, n := range []uint64{0, 1, 42, 62, 1_000_000} {
		code, err := EncodeSQLCode(n)
		if err != nil {
			t.Fatalf("EncodeSQLCode(%d) error: %v", n, err)
		}
		if Classify(code) != StoreSQL {
			t.Errorf("EncodeSQLCode(%d) = %q is not SQL-routed", n, code)
		}
		if want := base62.FormatUint(n); code[1:] != want {
			t.Errorf("EncodeSQLCode(%d) payload = %q, want %q", n, code[1:], want)
		}
	}
}

func TestAddRedisPrefix(t *testing.T) {
	payload := "abc1234"
	code, err := AddRedisPrefix(payload)
	if err != nil {
		t.Fatalf("AddRedisPrefix error: %v", err)
	}
	if Classify(code) != StoreRedis {
		t.Errorf("AddRedisPrefix(%q) = %q is not Redis-routed", payload, code)
	}
	if code[1:] != payload {
		t.Errorf("AddRedisPrefix payload = %q, want %q", code[1:], payload)
	}
}

// TestRandomPrefixSpread sanity-checks that the random prefix varies across
// the bucket rather than always returning the same character.
func TestRandomPrefixSpread(t *testing.T) {
	seen := map[byte]bool{}
	for i := 0; i < 200; i++ {
		code, err := EncodeSQLCode(1)
		if err != nil {
			t.Fatal(err)
		}
		seen[code[0]] = true
	}
	if len(seen) < 2 {
		t.Errorf("random SQL prefix not spread: only saw %d distinct values", len(seen))
	}
}
