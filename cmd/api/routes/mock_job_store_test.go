package routes

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type mockJobStore struct {
	jobs         []db.Job
	limit        int32
	getJobsErr   error
	getJobsCalls int
	getJobsLimit func(ctx context.Context, limit int32) ([]db.Job, error)

	createErr error
}

func (m *mockJobStore) CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error) {
	if m.createErr != nil {
		return db.Job{}, m.createErr
	}
	return db.Job{}, nil
}

func (m *mockJobStore) GetJobsLimit(ctx context.Context, limit int32) ([]db.Job, error) {
	m.getJobsCalls++
	m.limit = limit

	if m.getJobsLimit != nil {
		return m.getJobsLimit(ctx, limit)
	}
	if m.getJobsErr != nil {
		return nil, m.getJobsErr
	}

	return sliceJobs(m.jobs, limit), nil
}

func sliceJobs(jobs []db.Job, limit int32) []db.Job {
	if len(jobs) == 0 {
		return []db.Job{}
	}
	if limit <= 0 || int(limit) >= len(jobs) {
		out := make([]db.Job, len(jobs))
		copy(out, jobs)
		return out
	}

	out := make([]db.Job, limit)
	copy(out, jobs[:limit])
	return out
}
