package store

import (
	"context"
	"errors"
	"time"

	"github.com/ngmmartins/asyncq/internal/account"
	"github.com/ngmmartins/asyncq/internal/apikey"
	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/pagination"
	"github.com/ngmmartins/asyncq/internal/token"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrNoRowsAffected = errors.New("no rows affected after query execution")
)

type Store interface {
	Job() JobStore
	Account() AccountStore
	Token() TokenStore
	APIKey() APIKeyStore
}

type JobStore interface {
	Save(ctx context.Context, job *job.Job) error
	Search(ctx context.Context, criteria *job.SearchCriteria) ([]*job.Job, *pagination.Metadata, error)
	Get(ctx context.Context, jobId string) (*job.Job, error)
	Update(ctx context.Context, job *job.Job) error
}

type AccountStore interface {
	Save(ctx context.Context, account *account.Account) error
	Get(ctx context.Context, id string) (*account.Account, error)
	GetByEmail(ctx context.Context, email string) (*account.Account, error)
	GetForToken(ctx context.Context, hash []byte, scope token.Scope, now time.Time) (*account.Account, error)
}

type TokenStore interface {
	Save(ctx context.Context, token *token.Token) error
	Delete(ctx context.Context, hash []byte) error
}

type APIKeyStore interface {
	Save(ctx context.Context, key *apikey.APIKey) error
	GetByHash(ctx context.Context, hash []byte, now time.Time) (*apikey.APIKey, error)
}
