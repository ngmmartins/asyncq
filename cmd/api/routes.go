package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthHandler)

	router.HandlerFunc(http.MethodPost, "/v1/jobs", app.createJobHandler)
	router.HandlerFunc(http.MethodGet, "/v1/jobs", app.searchJobsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/jobs/:id", app.getJobHandler)
	router.HandlerFunc(http.MethodPost, "/v1/jobs/:id/schedule", app.scheduleJobHandler)
	router.HandlerFunc(http.MethodGet, "/v1/jobs/:id/status", app.getJobStatusHandler)
	router.HandlerFunc(http.MethodPost, "/v1/jobs/:id/cancel", app.cancelJobHandler)

	return app.recoverPanic(app.enableCORS(router))
}
