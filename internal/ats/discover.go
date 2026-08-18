package ats

import (
	"net/url"
	"strings"
)

// Board is a company career site hosted on a known ATS.
type Board struct {
	Name string // greenhouse, lever, ashby, workday, etc.
	URL  string
}

// Discover extracts a scrapeable ATS board URL from a job application link.
// Returns false when the link is not a supported ATS posting.
func Discover(applicationURL string) (Board, bool) {
	parsed, err := url.Parse(strings.TrimSpace(applicationURL))
	if err != nil || parsed.Host == "" {
		return Board{}, false
	}

	host := strings.ToLower(strings.TrimPrefix(parsed.Hostname(), "www."))
	path := strings.Trim(parsed.Path, "/")
	parts := splitPath(path)

	switch {
	case strings.Contains(host, "greenhouse.io"):
		return greenhouseBoard(parsed.Scheme, host, parts, parsed.Query())
	case host == "jobs.lever.co" || strings.HasSuffix(host, ".lever.co"):
		return firstSegmentBoard("lever", parsed.Scheme, host, parts)
	case host == "jobs.ashbyhq.com" || strings.HasSuffix(host, ".ashbyhq.com"):
		return ashbyBoard(parsed.Scheme, host, parts)
	case strings.Contains(host, "myworkdayjobs.com") || strings.Contains(host, "workdayjobs.com"):
		return workdayBoard(parsed.Scheme, host, parts)
	case host == "jobs.smartrecruiters.com":
		return firstSegmentBoard("smartrecruiters", parsed.Scheme, host, parts)
	case host == "apply.workable.com":
		return firstSegmentBoard("workable", parsed.Scheme, host, parts)
	case strings.Contains(host, "jobvite.com"):
		return firstSegmentBoard("jobvite", parsed.Scheme, host, parts)
	case strings.Contains(host, "icims.com"):
		return icimsBoard(parsed.Scheme, host)
	case host == "ats.rippling.com" || strings.Contains(host, "rippling.com"):
		return firstSegmentBoard("rippling", parsed.Scheme, host, parts)
	default:
		return Board{}, false
	}
}

func greenhouseBoard(scheme, host string, parts []string, query url.Values) (Board, bool) {
	if len(parts) == 0 {
		return Board{}, false
	}
	slug := parts[0]
	if slug == "embed" {
		slug = query.Get("for")
	}
	if slug == "" || slug == "jobs" {
		return Board{}, false
	}
	return Board{Name: "greenhouse", URL: origin(scheme, host) + "/" + slug}, true
}

func ashbyBoard(scheme, host string, parts []string) (Board, bool) {
	if host != "jobs.ashbyhq.com" {
		return Board{Name: "ashby", URL: origin(scheme, host)}, true
	}
	return firstSegmentBoard("ashby", scheme, host, parts)
}

func workdayBoard(scheme, host string, parts []string) (Board, bool) {
	cutoff := len(parts)
	for i, part := range parts {
		if strings.EqualFold(part, "job") {
			cutoff = i
			break
		}
	}
	if cutoff == 0 {
		return Board{Name: "workday", URL: origin(scheme, host)}, true
	}
	return Board{Name: "workday", URL: origin(scheme, host) + "/" + strings.Join(parts[:cutoff], "/")}, true
}

func icimsBoard(scheme, host string) (Board, bool) {
	return Board{Name: "icims", URL: origin(scheme, host) + "/jobs"}, true
}

func firstSegmentBoard(name, scheme, host string, parts []string) (Board, bool) {
	if len(parts) == 0 {
		return Board{}, false
	}
	return Board{Name: name, URL: origin(scheme, host) + "/" + parts[0]}, true
}

func origin(scheme, host string) string {
	if scheme == "" {
		scheme = "https"
	}
	return scheme + "://" + host
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	raw := strings.Split(path, "/")
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
