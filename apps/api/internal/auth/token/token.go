package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const byteLength = 32

func New() (string, []byte, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("token: generate: %w", err)
	}

	plain := base64.RawURLEncoding.EncodeToString(b)

	return plain, Hash(plain), nil
}

func Hash(plain string) []byte {
	sum := sha256.Sum256([]byte(plain))
	return sum[:]
}
