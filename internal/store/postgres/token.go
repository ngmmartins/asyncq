package postgres

import (
	"context"
	"time"

	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/token"
)

type PostgresTokenStore struct {
	*PostgresStore
}

func newPostgresTokenStore(postgresStore *PostgresStore) store.TokenStore {
	s := &PostgresTokenStore{
		PostgresStore: postgresStore,
	}

	return s
}

func (s *PostgresTokenStore) Save(ctx context.Context, token *token.Token) error {
	query := `INSERT INTO tokens (hash, account_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.AccountID, token.Expiry, token.Scope}

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
