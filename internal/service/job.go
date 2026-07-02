package service

import (
	"context"
	"errors"

	"github.com/stephensulimani/internlyapp/internal/db"
)

const (
	DefaultJobsLimit = 50
	MaxJobsLimit     = 100
)

var (
	ErrInvalidJobsLimit = errors.New("invalid jobs limit")
	ErrGetJobs          = errors.New("get jobs")
)

type JobService struct {
	store JobReader
}

func NewJobService(store JobReader) *JobService {
	return &JobService{store: store}
}

func (s *JobService) List(ctx context.Context, limit int) ([]db.Job, error) {
	if limit <= 0 {
		limit = DefaultJobsLimit
	}
	if limit > MaxJobsLimit {
		limit = MaxJobsLimit
	}

	jobs, err := s.store.GetJobsLimit(ctx, int32(limit))
	if err != nil {
		return nil, errors.Join(ErrGetJobs, err)
	}

	return jobs, nil
}

func NormalizeJobsLimit(limit int) (int, error) {
	if limit <= 0 {
		return DefaultJobsLimit, nil
	}
	if limit > MaxJobsLimit {
		return 0, ErrInvalidJobsLimit
	}
	return limit, nil
}
