package postgres

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/ngmmartins/asyncq/internal/store"
)

type PostgresConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

var _ store.Store = (*PostgresStore)(nil)

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) Job() store.JobStore {
	return newPostgresJobStore(s)
}

func (s *PostgresStore) Account() store.AccountStore {
	return newPostgresAccountStore(s)
}

func (s *PostgresStore) Token() store.TokenStore {
	return newPostgresTokenStore(s)
}

func New(cfg *PostgresConfig, logger *slog.Logger) *PostgresStore {
	store := &PostgresStore{}

	store.initConnection(cfg, logger)

	return store
}

func (s *PostgresStore) initConnection(cfg *PostgresConfig, logger *slog.Logger) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("Database connection successful")

	s.db = db
}
