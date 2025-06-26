package main

import "net/http"

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status":      "available",
		"environment": app.config.env,
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
