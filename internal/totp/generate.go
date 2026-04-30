package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"strings"
	"time"
)

type Code struct {
	Value     string
	Remaining int
}

func Generate(secret, algorithm string, digits, period int, now time.Time) (Code, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return Code{}, fmt.Errorf("decode secret: %w", err)
	}

	counter := uint64(now.Unix()) / uint64(period)
	remaining := period - int(now.Unix())%period

	h := hmacFor(algorithm, key)
	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], counter)
	h.Write(counterBytes[:])
	sum := h.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	otp := truncated % uint32(math.Pow10(digits))

	return Code{
		Value:     fmt.Sprintf("%0*d", digits, otp),
		Remaining: remaining,
	}, nil
}

func decodeSecret(secret string) ([]byte, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	// add padding if needed
	if pad := len(secret) % 8; pad != 0 {
		secret += strings.Repeat("=", 8-pad)
	}
	return base32.StdEncoding.DecodeString(secret)
}

func hmacFor(algorithm string, key []byte) hash.Hash {
	switch strings.ToUpper(algorithm) {
	case "SHA256":
		return hmac.New(sha256.New, key)
	case "SHA512":
		return hmac.New(sha512.New, key)
	default:
		return hmac.New(sha1.New, key)
	}
}
