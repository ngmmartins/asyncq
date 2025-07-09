package main

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ngmmartins/asyncq/internal/apikey"
	"github.com/ngmmartins/asyncq/internal/service"
	"github.com/ngmmartins/asyncq/internal/util"
	"github.com/ngmmartins/asyncq/internal/validator"
)

func (app *application) createAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	var input apikey.CreateRequest

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	acc := util.ContextGetAccount(r.Context())

	apiKey, err := app.apiKeyService.CreateAPIKey(r.Context(), acc.ID, &input)
	if err != nil {
		var validationError *validator.ValidationError
		switch {
		case errors.As(err, &validationError):
			app.failedValidationResponse(w, r, validationError.Errors)
		case errors.Is(err, service.ErrAccountInactive):
			app.inactiveAccountResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"apiKey": apiKey}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	acc := util.ContextGetAccount(r.Context())

	apiKeys, err := app.apiKeyService.GetAPIKeys(r.Context(), acc.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"apiKeys": apiKeys}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	id := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// we use the accountId to ensure that the user doesn't get a key from other account
	acc := util.ContextGetAccount(r.Context())

	apiKey, err := app.apiKeyService.GetAPIKey(r.Context(), id, acc.ID)
	if err != nil {
		if errors.Is(err, service.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"apiKey": apiKey}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	id := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// we use the accountId to ensure that the user doesn't delete a key from other account
	acc := util.ContextGetAccount(r.Context())

	err := app.apiKeyService.DeleteAPIKey(r.Context(), id, acc.ID)
	if err != nil {
		if errors.Is(err, service.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusNoContent, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
