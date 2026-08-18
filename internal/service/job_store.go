package service

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type JobWriter interface {
	CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error)
}

type ATSWriter interface {
	UpsertCompanyATS(ctx context.Context, arg db.UpsertCompanyATSParams) (db.CompanyAt, error)
}

type JobIngestStore interface {
	JobWriter
	ATSWriter
}

type JobReader interface {
	GetJobsLimit(ctx context.Context, limit int32) ([]db.Job, error)
	CountJobs(ctx context.Context, arg db.CountJobsParams) (int64, error)
	SearchJobs(ctx context.Context, arg db.SearchJobsParams) ([]db.Job, error)
	ListJobLocations(ctx context.Context) ([]string, error)
	GetJobsStats(ctx context.Context) (db.GetJobsStatsRow, error)
}

type JobStore interface {
	JobWriter
	JobReader
}
