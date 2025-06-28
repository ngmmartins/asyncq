package postgres

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/store"
	"github.com/ngmmartins/asyncq/internal/task"
)

const sendEmailPayload = `{"to":"receiver@example.com", "from": "sender@example.com", "subject":"Hi"}`

func TestSave(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		runAt := now.Add(1 * time.Hour)

		j := &job.Job{
			ID:        uuid.NewString(),
			Task:      task.SendEmailTask,
			Payload:   json.RawMessage(sendEmailPayload),
			RunAt:     &runAt,
			Status:    job.StatusQueued,
			CreatedAt: now,
		}

		err := s.Save(t.Context(), j)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}
	})

	t.Run("Error/PK duplicated", func(t *testing.T) {
		t.Parallel()

		id := uuid.NewString()
		now := time.Now()
		runAt := now.Add(1 * time.Hour)

		j1 := &job.Job{
			ID:        id,
			Task:      task.SendEmailTask,
			Payload:   json.RawMessage(sendEmailPayload),
			RunAt:     &runAt,
			Status:    job.StatusQueued,
			CreatedAt: now,
		}

		err := s.Save(t.Context(), j1)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		runAt2 := now.Add(2 * time.Hour)

		j2 := &job.Job{
			ID:        id,
			Task:      task.SendEmailTask,
			Payload:   json.RawMessage(sendEmailPayload),
			RunAt:     &runAt2,
			Status:    job.StatusQueued,
			CreatedAt: now,
		}

		err = s.Save(t.Context(), j2)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		id := uuid.NewString()
		now := time.Now().UTC()
		runAt := now.Add(1 * time.Hour)

		originalJob := &job.Job{
			ID:        id,
			Task:      task.SendEmailTask,
			Payload:   json.RawMessage(sendEmailPayload),
			RunAt:     &runAt,
			Status:    job.StatusQueued,
			CreatedAt: now,
		}

		err := s.Save(t.Context(), originalJob)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		foundJob, err := s.Get(t.Context(), id)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		if originalJob.ID != foundJob.ID {
			t.Fatalf("originalJob.ID=%s does not match foundJob.ID=%s", originalJob.ID, foundJob.ID)
		}
		if originalJob.Task != foundJob.Task {
			t.Fatalf("originalJob.Task=%s does not match foundJob.Task=%s", originalJob.Task, foundJob.Task)
		}
		if originalJob.Status != foundJob.Status {
			t.Fatalf("originalJob.Status=%s does not match foundJob.Status=%s", originalJob.Status, foundJob.Status)
		}
	})

	t.Run("Error/NotFound", func(t *testing.T) {
		t.Parallel()

		_, err := s.Get(t.Context(), uuid.NewString())
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if !errors.Is(err, store.ErrRecordNotFound) {
			t.Errorf("expected %s, got %s", store.ErrRecordNotFound, err.Error())
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		runAt := now.Add(1 * time.Hour)

		originalJob := &job.Job{
			ID:        uuid.NewString(),
			Task:      task.SendEmailTask,
			Payload:   json.RawMessage(sendEmailPayload),
			RunAt:     &runAt,
			Status:    job.StatusQueued,
			CreatedAt: now,
		}

		err := s.Save(t.Context(), originalJob)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		toUpdateJob := &job.Job{
			ID:        originalJob.ID,
			Task:      originalJob.Task,
			Payload:   originalJob.Payload,
			RunAt:     originalJob.RunAt,
			Status:    originalJob.Status,
			CreatedAt: originalJob.CreatedAt,
		}

		toUpdateJob.Task = task.WebhookTask
		toUpdateJob.Status = job.StatusCancelled

		err = s.Update(t.Context(), toUpdateJob)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		updatedJob, err := s.Get(t.Context(), originalJob.ID)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		if updatedJob.ID != originalJob.ID {
			t.Fatalf("originalJob.ID=%s does not match updatedJob.ID=%s", originalJob.ID, updatedJob.ID)
		}
		if updatedJob.Task != toUpdateJob.Task {
			t.Fatalf("updatedJob.Task=%s does not match toUpdateJob.Task=%s", updatedJob.Task, toUpdateJob.Task)
		}
		if updatedJob.Status != toUpdateJob.Status {
			t.Fatalf("updatedJob.Status=%s does not match toUpdateJob.Status=%s", updatedJob.Status, toUpdateJob.Status)
		}

	})

	t.Run("Error/IDNotUpdated", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		runAt := now.Add(1 * time.Hour)

		originalJob := &job.Job{
			ID:        uuid.NewString(),
			Task:      task.SendEmailTask,
			Payload:   json.RawMessage(sendEmailPayload),
			RunAt:     &runAt,
			Status:    job.StatusQueued,
			CreatedAt: now,
		}

		err := s.Save(t.Context(), originalJob)
		if err != nil {
			t.Fatalf("expected no error, got %s", err.Error())
		}

		toUpdateJob := &job.Job{
			ID:        originalJob.ID,
			Task:      originalJob.Task,
			Payload:   originalJob.Payload,
			RunAt:     originalJob.RunAt,
			Status:    originalJob.Status,
			CreatedAt: originalJob.CreatedAt,
		}

		toUpdateJob.ID = uuid.NewString()

		err = s.Update(t.Context(), toUpdateJob)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, store.ErrNoRowsAffected) {
			t.Errorf("expected %s, got %s", store.ErrNoRowsAffected, err.Error())
		}
	})
}
