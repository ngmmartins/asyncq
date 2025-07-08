package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"log/slog"
	"time"

	"github.com/ngmmartins/asyncq/internal/account"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/token"
)

type AccountService struct {
	logger *slog.Logger
	store  store.Store
}

func NewAccountService(logger *slog.Logger, store store.Store) *AccountService {
	return &AccountService{logger: logger, store: store}
}

func (s *AccountService) GetAccount(ctx context.Context, accountId string) (*account.Account, error) {
	acc, err := s.store.Account().Get(ctx, accountId)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return acc, nil
}

func (s *AccountService) GetForToken(ctx context.Context, plaintext string, scope token.Scope) (*account.Account, error) {
	hash := sha256.Sum256([]byte(plaintext))

	acc, err := s.store.Account().GetForToken(ctx, hash[:], scope, time.Now())
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return acc, nil
}
