package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type ResetToken struct {
	Token        string
	HashedToken  string
	ExpiresAt    time.Time
}

func GenerateResetToken() (*ResetToken, error) {
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	token := hex.EncodeToString(b)

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	return &ResetToken{
		Token:       token,
		HashedToken: hashedToken,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}, nil
}