package service

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/db"
)

type JobStore interface {
	CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error)
}
