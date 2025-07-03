package token

import (
	"crypto/rand"
	"crypto/sha256"
	"time"
)

type Scope string

const (
	ScopeAuthentication Scope = "authentication"
)

const AuthenticationTokenDuration = 15 * time.Minute

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	AccountID string    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     Scope     `json:"-"`
}

func New(accountId string, ttl time.Duration, scope Scope) *Token {
	token := &Token{
		Plaintext: rand.Text(),
		AccountID: accountId,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token
}

type AuthenticationRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
