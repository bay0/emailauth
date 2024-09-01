package emailauth

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

const (
	codeLength     = 6
	codeExpiration = 10 * time.Minute
)

func generateSecureCode() (string, error) {
	b := make([]byte, codeLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:codeLength], nil
}
