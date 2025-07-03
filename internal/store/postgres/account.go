package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ngmmartins/asyncq/internal/account"
	"github.com/ngmmartins/asyncq/internal/store"
)

type PostgresAccountStore struct {
	*PostgresStore
}

func newPostgresAccountStore(postgresStore *PostgresStore) store.AccountStore {
	s := &PostgresAccountStore{
		PostgresStore: postgresStore,
	}

	return s
}

func (s *PostgresAccountStore) Save(ctx context.Context, account *account.Account) error {
	query := `INSERT INTO accounts (id, name, email, password_hash, activated, created_at)
	VALUES ($1, $2, $3, $4, $5)`

	args := []any{account.ID, account.Name, account.Email, account.Password.Hash, account.Activated, account.CreatedAt}

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

func (s *PostgresAccountStore) GetByEmail(ctx context.Context, email string) (*account.Account, error) {
	query := `SELECT id, name, email, password_hash, activated, created_at
	FROM accounts
	WHERE email = $1`

	var acc account.Account

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&acc.ID,
		&acc.Name,
		&acc.Email,
		&acc.Password.Hash,
		&acc.Activated,
		&acc.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, store.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &acc, nil
}
