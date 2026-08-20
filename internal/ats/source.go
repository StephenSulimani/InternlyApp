package ats

import (
	"context"

	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

// BoardSource adapts one company_ats row + Provider to service.JobSource.
type BoardSource struct {
	Ctx      context.Context
	Provider Provider
	Board    Board
}

func (s *BoardSource) Scrape(log *zap.SugaredLogger) ([]db.Job, error) {
	ctx := s.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	jobs, err := s.Provider.Scrape(ctx, s.Board)
	if err != nil {
		return nil, err
	}
	log.Infow("ats board scraped",
		"ats", s.Provider.Name(),
		"company", s.Board.Company,
		"board", s.Board.URL,
		"jobs", len(jobs),
	)
	return jobs, nil
}
