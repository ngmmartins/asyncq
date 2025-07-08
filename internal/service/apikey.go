package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"log/slog"
	"time"

	"github.com/ngmmartins/asyncq/internal/apikey"
	"github.com/ngmmartins/asyncq/internal/store"
)

type APIKeyService struct {
	logger *slog.Logger
	store  store.Store
}

func NewAPIKeyService(logger *slog.Logger, store store.Store) *APIKeyService {
	return &APIKeyService{logger: logger, store: store}
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, accountId string, request *apikey.CreateRequest) (*apikey.APIKey, error) {
	acc, err := s.store.Account().Get(ctx, accountId)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if !acc.Activated {
		return nil, ErrAccountInactive
	}

	apiKey := apikey.New(accountId, request.Name, request.ExpiresAt)

	err = s.store.APIKey().Save(ctx, apiKey)
	if err != nil {
		s.logger.Error("failed to store API Key", "err", err.Error())
		return nil, err
	}

	return apiKey, nil
}

func (s *APIKeyService) GetValidAPIKey(ctx context.Context, plaintext string) (*apikey.APIKey, error) {
	hash := sha256.Sum256([]byte(plaintext))

	apiKey, err := s.store.APIKey().GetByHash(ctx, hash[:], time.Now())
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return apiKey, nil
}

func (s *APIKeyService) GetAPIKeys(ctx context.Context, accountId string) ([]*apikey.APIKey, error) {
	return s.store.APIKey().GetByAccountId(ctx, accountId)
}
