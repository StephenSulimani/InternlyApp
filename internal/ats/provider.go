package ats

import (
	"context"
	"errors"
	"net/http"

	"github.com/stephensulimani/internlyapp/internal/db"
)

const NameGreenhouse = "greenhouse"

// ErrBoardGone means the ATS board is gone or unpublished.
// The engine should set company_ats.working = false.
var ErrBoardGone = errors.New("ats board no longer available")

// Provider scrapes one company's board on a single ATS.
type Provider interface {
	Name() string
	Scrape(ctx context.Context, board Board) ([]db.Job, error)
}

// DefaultProviders is the engine registry, keyed by company_ats.ats_name.
func DefaultProviders(client *http.Client) map[string]Provider {
	gh := NewGreenhouse(client)
	return map[string]Provider{
		gh.Name(): gh,
	}
}
