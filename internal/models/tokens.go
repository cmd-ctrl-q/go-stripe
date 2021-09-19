package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

// Token is the type for authenticating tokens
type Token struct {
	// PlainText is the token
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// GenerateToken generates a token that lasts for ttl and returns it
func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	// create token
	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// secure token by making slice of bytes with size 16
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// create token plaintext
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// generate hash
	hash := sha256.Sum256([]byte(token.PlainText))
	// store token in hash
	token.Hash = hash[:]

	return token, nil
}
