package shortcode

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/huypham67/bookmark-common/pkg/base62"
)

// randReader is the source of randomness for prefix selection. It is a
// package variable so tests can substitute a failing reader.
var randReader = rand.Reader

// Prefix buckets. The two ranges must stay disjoint so Classify is
// unambiguous.
const (
	// RedisPrefixes routes a code to the Redis link store (a..h).
	RedisPrefixes = "abcdefgh"
	// SQLPrefixes routes a code to the SQL bookmark store (i..z).
	SQLPrefixes = "ijklmnopqrstuvwxyz"
)

// Store identifies which backing store a code belongs to.
type Store int

const (
	// StoreUnknown means the code does not carry a recognized prefix.
	StoreUnknown Store = iota
	// StoreRedis means the code routes to the Redis link store.
	StoreRedis
	// StoreSQL means the code routes to the SQL bookmark store.
	StoreSQL
)

// Classify reports which store a code routes to based on its first byte.
func Classify(code string) Store {
	if code == "" {
		return StoreUnknown
	}
	switch {
	case strings.IndexByte(RedisPrefixes, code[0]) >= 0:
		return StoreRedis
	case strings.IndexByte(SQLPrefixes, code[0]) >= 0:
		return StoreSQL
	default:
		return StoreUnknown
	}
}

// randomPrefix returns a uniformly random byte from bucket.
func randomPrefix(bucket string) (byte, error) {
	n, err := rand.Int(randReader, big.NewInt(int64(len(bucket))))
	if err != nil {
		return 0, err
	}
	return bucket[n.Int64()], nil
}

// EncodeSQLCode prepends a random SQL prefix to the base62 encoding of codeInt.
func EncodeSQLCode(codeInt uint64) (string, error) {
	p, err := randomPrefix(SQLPrefixes)
	if err != nil {
		return "", err
	}
	return string(p) + base62.FormatUint(codeInt), nil
}

// AddRedisPrefix prepends a random Redis prefix to a link payload.
func AddRedisPrefix(payload string) (string, error) {
	p, err := randomPrefix(RedisPrefixes)
	if err != nil {
		return "", err
	}
	return string(p) + payload, nil
}
