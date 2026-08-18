package service

import (
	"context"
	"encoding/csv"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stephensulimani/internlyapp/internal/db"
)

const (
	DefaultJobsLimit = 50
	MaxJobsLimit     = 100
)

var (
	ErrInvalidJobsLimit   = errors.New("invalid jobs limit")
	ErrInvalidJobsOffset  = errors.New("invalid jobs offset")
	ErrInvalidJobsRecency = errors.New("invalid jobs recency")
	ErrInvalidJobsSort    = errors.New("invalid jobs sort")
	ErrInvalidJobsSaved   = errors.New("invalid jobs saved")
	ErrInvalidJobID       = errors.New("invalid job id")
	ErrJobNotFound        = errors.New("job not found")
	ErrGetJobs            = errors.New("get jobs")
	ErrGetJobsStats       = errors.New("get jobs stats")
	ErrSaveJob            = errors.New("save job")
	ErrUnsaveJob          = errors.New("unsave job")
)

type JobListQuery struct {
	Q            string
	Type         string
	Location     string
	Source       string
	RecencyHours int32
	SavedOnly    bool
	UserID       pgtype.UUID
	SortBy       string
	SortDir      string
	Limit        int
	Offset       int
}

type JobPage struct {
	Jobs     []db.Job
	SavedIDs map[string]struct{}
	Total    int64
	Limit    int
	Offset   int
}

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

func (s *JobService) Search(ctx context.Context, query JobListQuery) (JobPage, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = DefaultJobsLimit
	}
	if limit > MaxJobsLimit {
		limit = MaxJobsLimit
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	countParams := db.CountJobsParams{
		Q:               csvLikePatterns(query.Q),
		FilterType:      strings.TrimSpace(query.Type),
		FilterLocations: csvLikePatterns(query.Location),
		FilterSource:    strings.TrimSpace(query.Source),
		RecencyHours:    query.RecencyHours,
		FilterSaved:     query.SavedOnly,
		UserID:          query.UserID,
	}
	total, err := s.store.CountJobs(ctx, countParams)
	if err != nil {
		return JobPage{}, errors.Join(ErrGetJobs, err)
	}

	sortBy, sortDir := normalizeSort(query.SortBy, query.SortDir)

	jobs, err := s.store.SearchJobs(ctx, db.SearchJobsParams{
		Q:               countParams.Q,
		FilterType:      countParams.FilterType,
		FilterLocations: countParams.FilterLocations,
		FilterSource:    countParams.FilterSource,
		RecencyHours:    countParams.RecencyHours,
		FilterSaved:     countParams.FilterSaved,
		UserID:          countParams.UserID,
		SortBy:          sortBy,
		SortDir:         sortDir,
		RowLimit:        int32(limit),
		RowOffset:       int32(offset),
	})
	if err != nil {
		return JobPage{}, errors.Join(ErrGetJobs, err)
	}
	if jobs == nil {
		jobs = []db.Job{}
	}

	savedIDs, err := s.savedIDSet(ctx, query.UserID, jobs)
	if err != nil {
		return JobPage{}, err
	}

	return JobPage{
		Jobs:     jobs,
		SavedIDs: savedIDs,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
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

func NormalizeJobsOffset(offset int) (int, error) {
	if offset < 0 || offset > math.MaxInt32 {
		return 0, ErrInvalidJobsOffset
	}
	return offset, nil
}

func ParseRecency(raw string) (int32, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "any":
		return 0, nil
	case "24h":
		return 24, nil
	case "3d":
		return 72, nil
	case "7d":
		return 168, nil
	default:
		return 0, ErrInvalidJobsRecency
	}
}

func ParseSort(field, dir string) (string, string, error) {
	field = strings.ToLower(strings.TrimSpace(field))
	switch field {
	case "", "posted", "company", "role", "location", "type":
	default:
		return "", "", ErrInvalidJobsSort
	}

	dir = strings.ToLower(strings.TrimSpace(dir))
	switch dir {
	case "", "asc", "desc":
	default:
		return "", "", ErrInvalidJobsSort
	}

	field, dir = normalizeSort(field, dir)
	return field, dir, nil
}

func ParseSavedFilter(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "0", "false", "no":
		return false, nil
	case "1", "true", "yes":
		return true, nil
	default:
		return false, ErrInvalidJobsSaved
	}
}

func normalizeSort(field, dir string) (string, string) {
	if field == "" {
		field = "posted"
	}
	if dir == "" {
		if field == "posted" {
			dir = "desc"
		} else {
			dir = "asc"
		}
	}
	return field, dir
}

func (s *JobService) Locations(ctx context.Context) ([]string, error) {
	locations, err := s.store.ListJobLocations(ctx)
	if err != nil {
		return nil, errors.Join(ErrGetJobs, err)
	}
	if locations == nil {
		return []string{}, nil
	}
	return locations, nil
}

func csvLikePatterns(raw string) []string {
	parts := splitCSVQuery(raw)
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		pattern := likePattern(part)
		if pattern == "" {
			continue
		}
		key := strings.ToLower(pattern)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, pattern)
	}
	return out
}

// splitCSVQuery parses a comma-separated filter.
// Tokens that contain commas must be quoted, e.g. `"New York, NY", Remote`.
func splitCSVQuery(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	reader := csv.NewReader(strings.NewReader(raw))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	parts, err := reader.Read()
	if err != nil {
		return []string{raw}
	}
	return parts
}

func likePattern(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	return "%" + escapeLike(q) + "%"
}

func escapeLike(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)
	return q
}

func (s *JobService) Save(ctx context.Context, userID, jobID pgtype.UUID) error {
	if _, err := s.store.GetJob(ctx, jobID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrJobNotFound
		}
		return errors.Join(ErrGetJobs, err)
	}

	err := s.store.SaveJob(ctx, db.SaveJobParams{UserID: userID, JobID: jobID})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrJobNotFound
		}
		return errors.Join(ErrSaveJob, err)
	}
	return nil
}

func (s *JobService) Unsave(ctx context.Context, userID, jobID pgtype.UUID) error {
	err := s.store.UnsaveJob(ctx, db.UnsaveJobParams{UserID: userID, JobID: jobID})
	if err != nil {
		return errors.Join(ErrUnsaveJob, err)
	}
	return nil
}

func (s *JobService) savedIDSet(ctx context.Context, userID pgtype.UUID, jobs []db.Job) (map[string]struct{}, error) {
	saved := make(map[string]struct{})
	if !userID.Valid || len(jobs) == 0 {
		return saved, nil
	}

	ids := make([]pgtype.UUID, 0, len(jobs))
	for _, job := range jobs {
		ids = append(ids, job.ID)
	}

	savedIDs, err := s.store.ListSavedJobIDsAmong(ctx, db.ListSavedJobIDsAmongParams{
		UserID: userID,
		JobIds: ids,
	})
	if err != nil {
		return nil, errors.Join(ErrGetJobs, err)
	}
	for _, id := range savedIDs {
		if id.Valid {
			saved[id.String()] = struct{}{}
		}
	}
	return saved, nil
}

func (s *JobService) Stats(ctx context.Context) (db.GetJobsStatsRow, error) {
	stats, err := s.store.GetJobsStats(ctx)
	if err != nil {
		return db.GetJobsStatsRow{}, errors.Join(ErrGetJobsStats, err)
	}
	return stats, nil
}
