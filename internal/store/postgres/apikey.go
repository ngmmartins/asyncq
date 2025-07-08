package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ngmmartins/asyncq/internal/apikey"
	"github.com/ngmmartins/asyncq/internal/store"
)

type PostgresAPIKeyStore struct {
	*PostgresStore
}

func newPostgresAPIKeyStore(postgresStore *PostgresStore) store.APIKeyStore {
	s := &PostgresAPIKeyStore{
		PostgresStore: postgresStore,
	}

	return s
}

func (s *PostgresAPIKeyStore) Save(ctx context.Context, key *apikey.APIKey) error {
	query := `INSERT INTO api_keys (id, hash, account_id, name, expires_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

	args := []any{key.ID, key.Hash, key.AccountID, key.Name, key.ExpiresAt, key.CreatedAt}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return store.ErrNoRowsAffected
	}

	return nil
}

func (s *PostgresAPIKeyStore) GetByHash(ctx context.Context, hash []byte, now time.Time) (*apikey.APIKey, error) {
	query := `SELECT id, hash, account_id, name, expires_at, created_at
	FROM api_keys
	WHERE hash = $1
	AND (expires_at > $2 OR expires_at IS NULL)`

	args := []any{hash, now}

	var apiKey apikey.APIKey

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&apiKey.ID,
		&apiKey.Hash,
		&apiKey.AccountID,
		&apiKey.Name,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, store.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &apiKey, nil
}
