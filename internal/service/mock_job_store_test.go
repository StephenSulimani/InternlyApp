package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
)

type mockJobStore struct {
	jobs         []db.Job
	limit        int32
	getJobsErr   error
	getJobsCalls int
	getJobsLimit func(ctx context.Context, limit int32) ([]db.Job, error)

	searchParams db.SearchJobsParams
	searchCalls  int
	searchErr    error
	searchTotal  *int64
	searchJobs   func(ctx context.Context, arg db.SearchJobsParams) ([]db.Job, error)
	countJobs    func(ctx context.Context, arg db.CountJobsParams) (int64, error)
	countErr     error
	countCalls   int
	locations    []string
	locationsErr error

	stats        db.GetJobsStatsRow
	statsErr     error
	statsCalls   int
	getJobsStats func(ctx context.Context) (db.GetJobsStatsRow, error)

	createErr   error
	createJob   func(ctx context.Context, arg db.CreateJobParams) (db.Job, error)
	createCalls []db.CreateJobParams

	atsCalls []db.UpsertCompanyATSParams
	atsErr   error

	getJobErr     error
	getJob        func(ctx context.Context, id pgtype.UUID) (db.Job, error)
	saveErr       error
	unsaveErr     error
	saveCalls     []db.SaveJobParams
	unsaveCalls   []db.UnsaveJobParams
	savedAmong    []pgtype.UUID
	savedAmongErr error
}

func (m *mockJobStore) CreateJob(ctx context.Context, arg db.CreateJobParams) (db.Job, error) {
	m.createCalls = append(m.createCalls, arg)
	if m.createJob != nil {
		return m.createJob(ctx, arg)
	}
	if m.createErr != nil {
		return db.Job{}, m.createErr
	}
	return db.Job{}, nil
}

func (m *mockJobStore) UpsertCompanyATS(ctx context.Context, arg db.UpsertCompanyATSParams) (db.CompanyAt, error) {
	m.atsCalls = append(m.atsCalls, arg)
	if m.atsErr != nil {
		return db.CompanyAt{}, m.atsErr
	}
	return db.CompanyAt{AtsUrl: arg.AtsUrl, AtsName: arg.AtsName, CompanyName: arg.CompanyName}, nil
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

func (m *mockJobStore) CountJobs(ctx context.Context, arg db.CountJobsParams) (int64, error) {
	m.countCalls++
	if m.countJobs != nil {
		return m.countJobs(ctx, arg)
	}
	if m.countErr != nil {
		return 0, m.countErr
	}
	if m.searchTotal != nil {
		return *m.searchTotal, nil
	}
	return int64(len(m.jobs)), nil
}

func (m *mockJobStore) SearchJobs(ctx context.Context, arg db.SearchJobsParams) ([]db.Job, error) {
	m.searchCalls++
	m.searchParams = arg
	m.limit = arg.RowLimit

	if m.searchJobs != nil {
		return m.searchJobs(ctx, arg)
	}
	if m.searchErr != nil {
		return nil, m.searchErr
	}

	return sliceJobsPage(m.jobs, arg.RowLimit, arg.RowOffset), nil
}

func (m *mockJobStore) ListJobLocations(ctx context.Context) ([]string, error) {
	if m.locationsErr != nil {
		return nil, m.locationsErr
	}
	if m.locations != nil {
		return m.locations, nil
	}
	return []string{}, nil
}

func (m *mockJobStore) GetJobsStats(ctx context.Context) (db.GetJobsStatsRow, error) {
	m.statsCalls++

	if m.getJobsStats != nil {
		return m.getJobsStats(ctx)
	}
	if m.statsErr != nil {
		return db.GetJobsStatsRow{}, m.statsErr
	}
	return m.stats, nil
}

func (m *mockJobStore) GetJob(ctx context.Context, id pgtype.UUID) (db.Job, error) {
	if m.getJob != nil {
		return m.getJob(ctx, id)
	}
	if m.getJobErr != nil {
		return db.Job{}, m.getJobErr
	}
	for _, job := range m.jobs {
		if job.ID == id {
			return job, nil
		}
	}
	return db.Job{ID: id}, nil
}

func (m *mockJobStore) ListSavedJobIDsAmong(ctx context.Context, arg db.ListSavedJobIDsAmongParams) ([]pgtype.UUID, error) {
	if m.savedAmongErr != nil {
		return nil, m.savedAmongErr
	}
	if m.savedAmong != nil {
		return m.savedAmong, nil
	}
	return []pgtype.UUID{}, nil
}

func (m *mockJobStore) SaveJob(ctx context.Context, arg db.SaveJobParams) error {
	m.saveCalls = append(m.saveCalls, arg)
	return m.saveErr
}

func (m *mockJobStore) UnsaveJob(ctx context.Context, arg db.UnsaveJobParams) error {
	m.unsaveCalls = append(m.unsaveCalls, arg)
	return m.unsaveErr
}

func sliceJobs(jobs []db.Job, limit int32) []db.Job {
	return sliceJobsPage(jobs, limit, 0)
}

func sliceJobsPage(jobs []db.Job, limit, offset int32) []db.Job {
	if len(jobs) == 0 {
		return []db.Job{}
	}
	start := int(offset)
	if start < 0 {
		start = 0
	}
	if start >= len(jobs) {
		return []db.Job{}
	}

	rest := jobs[start:]
	if limit <= 0 || int(limit) >= len(rest) {
		out := make([]db.Job, len(rest))
		copy(out, rest)
		return out
	}

	out := make([]db.Job, limit)
	copy(out, rest[:limit])
	return out
}
