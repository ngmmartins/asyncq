package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"log/slog"

	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/token"
	"github.com/ngmmartins/asyncq/internal/validator"
)

type TokenService struct {
	logger *slog.Logger
	store  store.Store
}

func NewTokenService(logger *slog.Logger, store store.Store) *TokenService {
	return &TokenService{logger: logger, store: store}
}

func (s *TokenService) CreateAuthenticationToken(ctx context.Context, request *token.AuthenticationRequest) (*token.Token, error) {
	v := validator.New()
	s.validateCreateAuthenticationToken(v, request)
	if !v.Valid() {
		return nil, &validator.ValidationError{Errors: v.Errors}
	}

	acc, err := s.store.Account().GetByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if !acc.Activated {
		return nil, ErrAccountInactive
	}

	match, err := acc.Password.Matches(request.Password)
	if err != nil {
		return nil, ErrComparingPasswords
	}
	if !match {
		return nil, ErrInvalidCredentials
	}

	t := token.New(acc.ID, token.AuthenticationTokenDuration, token.ScopeAuthentication)

	err = s.store.Token().Save(ctx, t)
	if err != nil {
		s.logger.Error("failed to store token", "err", err.Error())
		return nil, err
	}

	return t, nil
}

func (s *TokenService) DeleteToken(ctx context.Context, plaintext string) error {
	hash := sha256.Sum256([]byte(plaintext))

	return s.store.Token().Delete(ctx, hash[:])
}

func (s *TokenService) validateCreateAuthenticationToken(v *validator.Validator, request *token.AuthenticationRequest) {
	v.CheckRequired(request.Email != "", "email")
	v.CheckRequired(request.Password != "", "password")
}
