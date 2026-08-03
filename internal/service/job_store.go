package service

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type JobWriter interface {
	CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error)
}

type JobReader interface {
	GetJobsLimit(ctx context.Context, limit int32) ([]db.Job, error)
	GetJobsStats(ctx context.Context) (db.GetJobsStatsRow, error)
}

type JobStore interface {
	JobWriter
	JobReader
}
