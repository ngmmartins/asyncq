package util

import (
	"context"

	"github.com/ngmmartins/asyncq/internal/account"
)

type contextKey string

const accountContextKey = contextKey("account")

func ContextSetAccount(ctx context.Context, acc *account.Account) context.Context {
	return context.WithValue(ctx, accountContextKey, acc)
}

func ContextGetAccount(ctx context.Context) *account.Account {
	acc, ok := ctx.Value(accountContextKey).(*account.Account)
	if !ok {
		panic("missing account value in request context")
	}

	return acc
}
