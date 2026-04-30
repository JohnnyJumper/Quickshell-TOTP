package importer

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Account struct {
	ID        string
	Issuer    string
	Label     string
	Secret    string
	Algorithm string
	Digits    int
	Period    int
}

func ParsePayload(payload string) ([]Account, error) {
	payload = strings.TrimSpace(payload)
	switch {
	case strings.HasPrefix(payload, "otpauth-migration://offline"):
		return ParseGoogleMigration(payload)
	case strings.HasPrefix(payload, "otpauth://totp/"):
		account, err := ParseOTPAuthURI(payload)
		if err != nil {
			return nil, err
		}
		return []Account{account}, nil
	default:
		return nil, fmt.Errorf("unsupported import payload")
	}
}

func ParseOTPAuthURI(uri string) (Account, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return Account{}, fmt.Errorf("parse otpauth URI: %w", err)
	}

	if parsed.Scheme != "otpauth" {
		return Account{}, fmt.Errorf("invalid otpauth scheme: %s", parsed.Scheme)
	}
	if parsed.Host != "totp" {
		return Account{}, fmt.Errorf("only totp type supported, got: %s", parsed.Host)
	}

	label := strings.TrimPrefix(parsed.Path, "/")
	issuer, accountName := splitLabel(label)

	params := parsed.Query()

	secret := strings.ToUpper(strings.TrimSpace(params.Get("secret")))
	if secret == "" {
		return Account{}, fmt.Errorf("otpauth URI missing secret parameter")
	}

	if v := params.Get("issuer"); v != "" {
		issuer = v
	}

	algorithm := "SHA1"
	if v := params.Get("algorithm"); v != "" {
		algorithm = strings.ToUpper(v)
	}

	digits := 6
	if v := params.Get("digits"); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			digits = d
		}
	}

	period := 30
	if v := params.Get("period"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			period = p
		}
	}

	return Account{
		ID:        buildAccountID(issuer, accountName, secret),
		Issuer:    issuer,
		Label:     accountName,
		Secret:    secret,
		Algorithm: algorithm,
		Digits:    digits,
		Period:    period,
	}, nil
}

func splitLabel(label string) (issuer, accountName string) {
	if idx := strings.IndexByte(label, ':'); idx >= 0 {
		return label[:idx], label[idx+1:]
	}
	return "", label
}
