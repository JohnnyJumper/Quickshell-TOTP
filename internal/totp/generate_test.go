package totp

import (
	"encoding/base32"
	"testing"
	"time"
)

func TestGenerateRFC6238(t *testing.T) {
	secret := base32.StdEncoding.EncodeToString([]byte("12345678901234567890"))

	cases := []struct {
		unix int64
		want string
	}{
		{59, "94287082"},
		{1111111109, "07081804"},
		{1111111111, "14050471"},
		{1234567890, "89005924"},
		{2000000000, "69279037"},
		{20000000000, "65353130"},
	}

	for _, c := range cases {
		code, err := Generate(secret, "SHA1", 8, 30, time.Unix(c.unix, 0))
		if err != nil {
			t.Errorf("T=%d: unexpected error: %v", c.unix, err)
			continue
		}
		if code.Value != c.want {
			t.Errorf("T=%d: got %s, want %s", c.unix, code.Value, c.want)
		}
	}
}
