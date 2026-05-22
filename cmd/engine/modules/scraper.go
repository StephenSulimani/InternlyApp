package modules

import (
	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

type Scraper interface {
	Scrape(log *zap.SugaredLogger) []db.Job
}
