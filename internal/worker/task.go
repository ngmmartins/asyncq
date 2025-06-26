package worker

import "github.com/ngmmartins/asyncq/internal/job"

type TaskExecutor interface {
	Execute(job.Job) error
}
