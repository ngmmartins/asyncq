package service

import (
	"errors"

	"github.com/ngmmartins/asyncq/internal/store"
)

var (
	ErrRecordNotFound          = store.ErrRecordNotFound
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	ErrComparingPasswords = errors.New("error authenticating")
	ErrInvalidCredentials = errors.New("invalid credentials provided")

	ErrAccountInactive = errors.New("account is not activated")
)
