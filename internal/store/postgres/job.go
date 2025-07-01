package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ngmmartins/asyncq/internal/job"
	"github.com/ngmmartins/asyncq/internal/pagination"
	"github.com/ngmmartins/asyncq/internal/store"
)

type PostgresJobStore struct {
	*PostgresStore
}

func newPostgresJobStore(postgresStore *PostgresStore) store.JobStore {
	s := &PostgresJobStore{
		PostgresStore: postgresStore,
	}

	return s
}

// Saves a new [job.Job] in the database.
//
// If the insert doesn't change any row, a [store.ErrNoRowsAffected] error is returned.
func (s *PostgresStore) Save(ctx context.Context, job *job.Job) error {
	query := `INSERT INTO jobs (id, task, payload, run_at, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

	args := []any{job.ID, job.Task, job.Payload, job.RunAt, job.Status, job.CreatedAt}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return store.ErrNoRowsAffected
	}

	return nil
}

func (s *PostgresStore) Search(ctx context.Context, criteria *job.SearchCriteria) ([]*job.Job, *pagination.Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id, task, payload, run_at, status, created_at
	FROM jobs
	WHERE (task = $1 OR $1::text IS NULL OR $1 = '')
	AND (run_at >= $2 OR $2::timestamptz IS NULL)
	AND (run_at <= $3 OR $3::timestamptz IS NULL)
	AND (status = $4 OR $4::text IS NULL OR $4 = '')
	ORDER BY %s %s, created_at DESC
	LIMIT $5 OFFSET $6`, criteria.SortColumn(), criteria.SortDirection())

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	args := []any{criteria.Task, criteria.RunAfter, criteria.RunBefore, criteria.Status, criteria.Limit(), criteria.Offset()}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	totalRecords := 0
	jobs := []*job.Job{}

	for rows.Next() {
		var j job.Job

		err := rows.Scan(
			&totalRecords,
			&j.ID,
			&j.Task,
			&j.Payload,
			&j.RunAt,
			&j.Status,
			&j.CreatedAt,
		)
		if err != nil {
			return nil, nil, err
		}

		jobs = append(jobs, &j)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	metadata := pagination.NewMetadata(totalRecords, criteria.Page, criteria.PageSize)

	return jobs, metadata, nil
}

// Gets the [job.Job] identified by the given jobId from the database.
//
// In case the record does not exist in the database a [store.ErrRecordNotFound] error is returned
func (s *PostgresStore) Get(ctx context.Context, jobId string) (*job.Job, error) {
	query := `SELECT id, task, payload, run_at, status, created_at
	FROM jobs
	WHERE id = $1`

	var job job.Job

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, jobId).Scan(
		&job.ID,
		&job.Task,
		&job.Payload,
		&job.RunAt,
		&job.Status,
		&job.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, store.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &job, nil
}

// Updates the given [job.Job] in the database.
// The fields that will be updated are: [job.Job].Task, [job.Job].Payload, [job.Job].RunAt and [job.Job].Status.
// All other changes provided in the struct will be ignored.
// The SQL Where clause will use the [job.Job].ID to update the record.
//
// If the update doesn't change any row, a [store.ErrNoRowsAffected] error is returned.
func (s *PostgresStore) Update(ctx context.Context, job *job.Job) error {
	query := `UPDATE jobs
	SET task = $1, payload = $2, run_at = $3, status = $4
	WHERE id = $5`

	args := []any{job.Task, job.Payload, job.RunAt, job.Status, job.ID}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return store.ErrNoRowsAffected
	}

	return nil
}
