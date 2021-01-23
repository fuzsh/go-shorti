package util

import (
	"crypto/rand"
	"encoding/base64"
)

const otpChars = "1234567890"

func GenerateVerificationCode(max int) (string, error) {
	b := make([]byte, max)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	otpCharsLength := len(otpChars)
	for i := 0; i < max; i++ {
		b[i] = otpChars[int(b[i])%otpCharsLength]
	}

	return string(b), nil
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
