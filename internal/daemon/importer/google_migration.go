package importer

import (
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/protobuf/proto"
)

func ParseGoogleMigration(payload string) ([]Account, error) {
	parsedURL, err := url.Parse(payload)
	if err != nil {
		return nil, fmt.Errorf("parse google migration URL: %w", err)
	}

	if parsedURL.Scheme != "otpauth-migration" {
		return nil, fmt.Errorf("invalid google migration scheme: %s", parsedURL.Scheme)
	}

	data := parsedURL.Query().Get("data")
	if data == "" {
		return nil, fmt.Errorf("google migration payload is missing data")
	}

	rawPayload, err := decodeMigrationData(data)
	if err != nil {
		return nil, err
	}

	var migrationPayload MigrationPayload
	if err := proto.Unmarshal(rawPayload, &migrationPayload); err != nil {
		return nil, fmt.Errorf("decode google migration protobuf: %w", err)
	}

	accounts := make([]Account, 0, len(migrationPayload.OtpParameters))
	for _, otp := range migrationPayload.OtpParameters {
		if otp.Type != OtpType_OTP_TYPE_TOTP {
			continue
		}

		account, err := accountFromGoogleOtp(otp)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("google migration payload contains no TOTP accounts")
	}

	return accounts, nil
}

func decodeMigrationData(data string) ([]byte, error) {
	rawPayload, err := base64.StdEncoding.DecodeString(data)
	if err == nil {
		return rawPayload, nil
	}

	rawPayload, err = base64.URLEncoding.DecodeString(data)
	if err == nil {
		return rawPayload, nil
	}

	rawPayload, err = base64.RawURLEncoding.DecodeString(data)
	if err == nil {
		return rawPayload, nil
	}

	return nil, fmt.Errorf("decode google migration data: %w", err)
}

func accountFromGoogleOtp(otp *OtpParameters) (Account, error) {
	if len(otp.Secret) == 0 {
		return Account{}, fmt.Errorf("google migration account is missing secret")
	}

	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(otp.Secret)

	account := Account{
		ID:        buildAccountID(otp.Issuer, otp.Name, secret),
		Issuer:    otp.Issuer,
		Label:     otp.Name,
		Secret:    secret,
		Algorithm: googleAlgorithm(otp.Algorithm),
		Digits:    googleDigits(otp.Digits),
		Period:    30,
	}

	return account, nil
}

func googleAlgorithm(algorithm Algorithm) string {
	switch algorithm {
	case Algorithm_ALGORITHM_SHA256:
		return "SHA256"
	case Algorithm_ALGORITHM_SHA512:
		return "SHA512"
	default:
		return "SHA1"
	}
}

func googleDigits(digits DigitCount) int {
	switch digits {
	case DigitCount_DIGIT_COUNT_EIGHT:
		return 8
	default:
		return 6
	}
}

func buildAccountID(issuer string, label string, secret string) string {
	value := strings.ToLower(strings.TrimSpace(issuer + ":" + label + ":" + secret))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, "\\", "-")

	return value
}
