package main

import (
	"errors"
	"net/http"

	"github.com/ngmmartins/asyncq/internal/service"
	"github.com/ngmmartins/asyncq/internal/token"
	"github.com/ngmmartins/asyncq/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input token.AuthenticationRequest

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	t, err := app.tokenService.CreateAuthenticationToken(r.Context(), &input)
	if err != nil {
		var validationError *validator.ValidationError
		switch {
		case errors.As(err, &validationError):
			app.failedValidationResponse(w, r, validationError.Errors)
		case errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		case errors.Is(err, service.ErrAccountInactive):
			app.inactiveAccountResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"token": t}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
