package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID        string     `json:"id"`
	Hash      []byte     `json:"-"`
	AccountID string     `json:"account_id"`
	Name      string     `json:"name"`
	Key       string     `json:"key,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // Allow for keys that don't expire
	CreatedAt time.Time  `json:"created_at"`
}

func New(accountId string, name string, expiresAt *time.Time) *APIKey {
	plaintext := rand.Text()
	hash := sha256.Sum256([]byte(plaintext))

	now := time.Now()

	key := &APIKey{
		ID:        uuid.NewString(),
		Hash:      hash[:],
		AccountID: accountId,
		Name:      name,
		Key:       plaintext,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	return key
}

type CreateRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
