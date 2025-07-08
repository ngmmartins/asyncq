package main

import (
	"errors"
	"net/http"

	"github.com/ngmmartins/asyncq/internal/apikey"
	"github.com/ngmmartins/asyncq/internal/service"
	"github.com/ngmmartins/asyncq/internal/util"
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
		switch {
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
