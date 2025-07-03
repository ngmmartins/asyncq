package store

import (
	"context"
	"errors"

	"github.com/ngmmartins/asyncq/internal/account"
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
}

type JobStore interface {
	Save(ctx context.Context, job *job.Job) error
	Search(ctx context.Context, criteria *job.SearchCriteria) ([]*job.Job, *pagination.Metadata, error)
	Get(ctx context.Context, jobId string) (*job.Job, error)
	Update(ctx context.Context, job *job.Job) error
}

type AccountStore interface {
	Save(ctx context.Context, account *account.Account) error
	GetByEmail(ctx context.Context, email string) (*account.Account, error)
}

type TokenStore interface {
	Save(ctx context.Context, token *token.Token) error
}
