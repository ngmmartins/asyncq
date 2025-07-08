package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Public routes
	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Protected routes - authentication required
	router.Handler(http.MethodPost, "/v1/api-keys", app.requireAuthenticatedAccount(
		app.requireActivatedAccount(http.HandlerFunc(app.createAPIKeyHandler))))
	router.Handler(http.MethodGet, "/v1/api-keys", app.requireAuthenticatedAccount(
		app.requireActivatedAccount(http.HandlerFunc(app.getAPIKeysHandler))))

	// Protected routes - API-Key required
	router.Handler(http.MethodPost, "/v1/jobs", app.requireAPIKey(
		app.requireActivatedAccount(http.HandlerFunc(app.createJobHandler))))
	router.Handler(http.MethodGet, "/v1/jobs", app.requireAPIKey(
		app.requireActivatedAccount(http.HandlerFunc(app.searchJobsHandler))))
	router.Handler(http.MethodGet, "/v1/jobs/:id", app.requireAPIKey(
		app.requireActivatedAccount(http.HandlerFunc(app.getJobHandler))))
	router.Handler(http.MethodPost, "/v1/jobs/:id/schedule", app.requireAPIKey(
		app.requireActivatedAccount(http.HandlerFunc(app.scheduleJobHandler))))
	router.Handler(http.MethodGet, "/v1/jobs/:id/status", app.requireAPIKey(
		app.requireActivatedAccount(http.HandlerFunc(app.getJobStatusHandler))))
	router.Handler(http.MethodPost, "/v1/jobs/:id/cancel", app.requireAPIKey(
		app.requireActivatedAccount(http.HandlerFunc(app.cancelJobHandler))))

	return app.recoverPanic(app.enableCORS(router))
}
