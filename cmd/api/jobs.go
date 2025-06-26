package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/validator"
)

func (app *application) createJobHandler(w http.ResponseWriter, r *http.Request) {
	var input job.CreateRequest

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	if job.ValidateCreateJob(v, &input); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	job, err := app.dispatcher.Enqueue(r.Context(), &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"job": job}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getJobHandler(w http.ResponseWriter, r *http.Request) {
	id := httprouter.ParamsFromContext(r.Context()).ByName("id")

	j, err := app.store.Job().Get(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"job": j}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := httprouter.ParamsFromContext(r.Context()).ByName("id")

	j, err := app.store.Job().Get(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"status": j.Status}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) cancelJobHandler(w http.ResponseWriter, r *http.Request) {
	id := httprouter.ParamsFromContext(r.Context()).ByName("id")

	j, err := app.store.Job().Get(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	v := validator.New()
	v.Check(j.Status == job.StatusQueued, "status", fmt.Sprintf("Unable to transition from %s to %s", j.Status, job.StatusCancelled))
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	j.Status = job.StatusCancelled

	err = app.dispatcher.Remove(r.Context(), id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.store.Job().Update(r.Context(), j)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"job": j}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
