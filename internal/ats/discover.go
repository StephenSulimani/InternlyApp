package ats

import (
	"net/url"
	"strings"
)

// Board is a company career site hosted on a known ATS.
type Board struct {
	Name    string // greenhouse, lever, ashby, workday, etc.
	URL     string
	Company string // from company_ats; empty when discovered from a posting URL
}

type parsedLink struct {
	scheme string
	host   string
	parts  []string
	query  url.Values
}

type provider struct {
	name  string
	match func(host string) bool
	// boardPath returns the path after origin ("stripe"), or "" for origin-only.
	boardPath func(parsedLink) (string, bool)
}

var providers = []provider{
	{name: NameGreenhouse, match: contains("greenhouse.io"), boardPath: greenhousePath},
	{name: "lever", match: exactOrSuffix("jobs.lever.co", "lever.co"), boardPath: firstSegment},
	{name: "ashby", match: exactOrSuffix("jobs.ashbyhq.com", "ashbyhq.com"), boardPath: ashbyPath},
	{name: "workday", match: contains("myworkdayjobs.com", "workdayjobs.com"), boardPath: workdayPath},
	{name: "smartrecruiters", match: exact("jobs.smartrecruiters.com"), boardPath: firstSegment},
	{name: "workable", match: exact("apply.workable.com"), boardPath: firstSegment},
	{name: "jobvite", match: contains("jobvite.com"), boardPath: firstSegment},
	{name: "icims", match: contains("icims.com"), boardPath: fixedPath("jobs")},
	{name: "rippling", match: exactOrSuffix("ats.rippling.com", "rippling.com"), boardPath: firstSegment},
}

// Discover extracts a scrapeable ATS board URL from a job application link.
// Returns false when the link is not a supported ATS posting.
func Discover(applicationURL string) (Board, bool) {
	link, ok := parseLink(applicationURL)
	if !ok {
		return Board{}, false
	}

	for _, p := range providers {
		if !p.match(link.host) {
			continue
		}
		path, ok := p.boardPath(link)
		if !ok {
			return Board{}, false
		}
		return Board{Name: p.name, URL: boardURL(link.scheme, link.host, path)}, true
	}

	return Board{}, false
}

func parseLink(raw string) (parsedLink, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return parsedLink{}, false
	}

	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}

	return parsedLink{
		scheme: scheme,
		host:   strings.ToLower(strings.TrimPrefix(parsed.Hostname(), "www.")),
		parts:  splitPath(parsed.Path),
		query:  parsed.Query(),
	}, true
}

func greenhousePath(link parsedLink) (string, bool) {
	if len(link.parts) == 0 {
		return "", false
	}
	slug := link.parts[0]
	if slug == "embed" {
		slug = link.query.Get("for")
	}
	if slug == "" || slug == "jobs" {
		return "", false
	}
	return slug, true
}

func ashbyPath(link parsedLink) (string, bool) {
	if link.host != "jobs.ashbyhq.com" {
		return "", true
	}
	return firstSegment(link)
}

func workdayPath(link parsedLink) (string, bool) {
	for i, part := range link.parts {
		if strings.EqualFold(part, "job") {
			return strings.Join(link.parts[:i], "/"), true
		}
	}
	return strings.Join(link.parts, "/"), true
}

func firstSegment(link parsedLink) (string, bool) {
	if len(link.parts) == 0 {
		return "", false
	}
	return link.parts[0], true
}

func fixedPath(path string) func(parsedLink) (string, bool) {
	return func(parsedLink) (string, bool) {
		return path, true
	}
}

func contains(needles ...string) func(string) bool {
	return func(host string) bool {
		for _, n := range needles {
			if strings.Contains(host, n) {
				return true
			}
		}
		return false
	}
}

func exact(want string) func(string) bool {
	return func(host string) bool { return host == want }
}

func exactOrSuffix(exactHost, suffix string) func(string) bool {
	return func(host string) bool {
		return host == exactHost || strings.HasSuffix(host, "."+suffix)
	}
}

func boardURL(scheme, host, path string) string {
	base := scheme + "://" + host
	if path == "" {
		return base
	}
	return base + "/" + strings.Trim(path, "/")
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
