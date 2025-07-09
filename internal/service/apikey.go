package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"log/slog"
	"time"

	"github.com/ngmmartins/asyncq/internal/apikey"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/validator"
)

type APIKeyService struct {
	logger *slog.Logger
	store  store.Store
}

func NewAPIKeyService(logger *slog.Logger, store store.Store) *APIKeyService {
	return &APIKeyService{logger: logger, store: store}
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, accountId string, request *apikey.CreateRequest) (*apikey.APIKey, error) {
	v := validator.New()
	s.validateCreateAPIKey(v, request)
	if !v.Valid() {
		return nil, &validator.ValidationError{Errors: v.Errors}
	}

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

func (s *APIKeyService) GetAPIKey(ctx context.Context, id, accountId string) (*apikey.APIKey, error) {
	apiKey, err := s.store.APIKey().Get(ctx, id, accountId)
	if err != nil {
		if errors.Is(err, store.ErrNoRowsAffected) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return apiKey, nil
}

func (s *APIKeyService) DeleteAPIKey(ctx context.Context, id string, accountId string) error {
	err := s.store.APIKey().Delete(ctx, id, accountId)
	if err != nil {
		if errors.Is(err, store.ErrNoRowsAffected) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (s *APIKeyService) validateCreateAPIKey(v *validator.Validator, request *apikey.CreateRequest) {
	v.CheckRequired(request.Name != "", "name")
	v.Check(request.ExpiresAt == nil || request.ExpiresAt.After(time.Now()), "expires_at", "must be in the future")
}
