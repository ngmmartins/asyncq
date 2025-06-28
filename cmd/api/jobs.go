package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/service"
	"github.com/ngmmartins/asyncq/internal/validator"
)

func (app *application) createJobHandler(w http.ResponseWriter, r *http.Request) {
	var input job.CreateRequest

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	job, err := app.jobService.CreateJob(r.Context(), &input)
	if err != nil {
		var validationError *validator.ValidationError
		if errors.As(err, &validationError) {
			app.failedValidationResponse(w, r, validationError.Errors)
			return
		}
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

	j, err := app.jobService.GetJob(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
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

	j, err := app.jobService.GetJob(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
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

func (app *application) scheduleJobHandler(w http.ResponseWriter, r *http.Request) {
	id := httprouter.ParamsFromContext(r.Context()).ByName("id")

	var input struct {
		RunAt time.Time `json:"run_at"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.jobService.ScheduleJob(r.Context(), id, input.RunAt)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, service.ErrInvalidStatusTransition):
			app.conflictResponse(w, r, map[string]string{"message": err.Error()})
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
}

func (app *application) cancelJobHandler(w http.ResponseWriter, r *http.Request) {
	id := httprouter.ParamsFromContext(r.Context()).ByName("id")

	err := app.jobService.UpdateJobStatus(r.Context(), id, job.StatusCancelled)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, service.ErrInvalidStatusTransition):
			app.conflictResponse(w, r, map[string]string{"message": err.Error()})
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.queue.Remove(r.Context(), id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, nil, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
