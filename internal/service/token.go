package service

import (
	"context"
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

func (ts *TokenService) CreateAuthenticationToken(ctx context.Context, request *token.AuthenticationRequest) (*token.Token, error) {
	v := validator.New()
	ts.validateCreateAuthenticationToken(v, request)
	if !v.Valid() {
		return nil, &validator.ValidationError{Errors: v.Errors}
	}

	acc, err := ts.store.Account().GetByEmail(ctx, request.Email)
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

	err = ts.store.Token().Save(ctx, t)
	if err != nil {
		ts.logger.Error("failed to store token", "err", err.Error())
		return nil, err
	}

	return t, nil
}

func (ts *TokenService) validateCreateAuthenticationToken(v *validator.Validator, request *token.AuthenticationRequest) {
	v.CheckRequired(request.Email != "", "email")
	v.CheckRequired(request.Password != "", "password")
}
