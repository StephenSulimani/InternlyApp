package ats

import (
	"strings"
	"unicode"
)

// ApplicationMatchKeys are apply URLs that may already be stored for this posting
// (Simplify vs ATS host differences, trailing slashes).
func ApplicationMatchKeys(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	seen := make(map[string]struct{})
	var out []string
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	add(raw)
	trimmed := strings.TrimRight(raw, "/")
	add(trimmed)
	add(strings.Replace(trimmed, "://boards.greenhouse.io/", "://job-boards.greenhouse.io/", 1))
	add(strings.Replace(trimmed, "://job-boards.greenhouse.io/", "://boards.greenhouse.io/", 1))
	return out
}

// ApplicationMatchRegex matches Greenhouse job IDs when the stored apply URL
// uses a different path or embed token than the board API's absolute_url.
func ApplicationMatchRegex(raw string) string {
	id := greenhouseJobID(raw)
	if id == "" {
		return ""
	}
	return `/jobs/` + id + `([^0-9]|$)` + `|` + `[?&]token=` + id + `([^0-9]|$)`
}

func greenhouseJobID(raw string) string {
	link, ok := parseLink(raw)
	if !ok {
		return ""
	}
	for i, part := range link.parts {
		if part == "jobs" && i+1 < len(link.parts) && isDigits(link.parts[i+1]) {
			return link.parts[i+1]
		}
	}
	for _, key := range []string{"token", "gh_jid"} {
		value := link.query.Get(key)
		if isDigits(value) {
			return value
		}
	}
	return ""
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
